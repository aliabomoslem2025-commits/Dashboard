package mediagate
import(
 "bytes";"context";"crypto/sha256";"encoding/hex";"encoding/json";"fmt";"io";"net/http";"net/http/httputil";"net/url";"path";"strconv";"strings";"time"
)
type Quota struct{MaxTotalBytes int64;MaxFileBytes int64;Period string}
type QuotaRepository interface{GetQuota(context.Context,string)(*Quota,error)}
type Cache interface{GetUser(context.Context,string)(string,error);SetUser(context.Context,string,string,time.Duration)error;GetUsage(context.Context,string)(int64,bool,error);SetUsage(context.Context,string,int64,time.Duration)error;AddUsage(context.Context,string,int64)error;ReserveUsage(context.Context,string,int64,int64)(bool,error);ReleaseUsage(context.Context,string,int64)error;GetMediaOwner(context.Context,string)(string,error);SetMediaOwner(context.Context,string,string,time.Duration)error;DeleteMediaOwner(context.Context,string)error}
type Resolver interface{WhoAmI(context.Context,string)(string,error);MediaBytesSince(context.Context,string,time.Time)(int64,error)}
type ownerContextKey struct{}
type Gate struct{proxy *httputil.ReverseProxy;quotas QuotaRepository;cache Cache;resolver Resolver}
func New(target *url.URL,q QuotaRepository,c Cache,r Resolver)*Gate{
 p:=httputil.NewSingleHostReverseProxy(target)
 p.ErrorHandler=func(w http.ResponseWriter,_ *http.Request,_ error){http.Error(w,"Upstream media service unavailable",http.StatusBadGateway)}
 p.ModifyResponse=func(resp *http.Response)error{owner,_:=resp.Request.Context().Value(ownerContextKey{}).(string);if owner==""||resp.StatusCode<200||resp.StatusCode>=300{return nil};body,err:=io.ReadAll(io.LimitReader(resp.Body,1<<20));if err!=nil{return err};_=resp.Body.Close();var result struct{ContentURI string `json:"content_uri"`};if json.Unmarshal(body,&result)==nil&&result.ContentURI!=""{mxc,err:=url.Parse(result.ContentURI);if err==nil{mediaID:=path.Base(mxc.Path);if mediaID!=""&&mediaID!="."{if err=c.SetMediaOwner(context.Background(),mediaID,owner,24*time.Hour);err!=nil{return err}}}};resp.Body=io.NopCloser(bytes.NewReader(body));resp.ContentLength=int64(len(body));resp.Header.Set("Content-Length",strconv.Itoa(len(body)));return nil}
 return &Gate{proxy:p,quotas:q,cache:c,resolver:r}
}
func isCreate(p string)bool{return p=="/_matrix/media/v1/create"}
func isUpload(r *http.Request)bool{p:=r.URL.Path;if r.Method==http.MethodPost&&(p=="/_matrix/media/v1/upload"||p=="/_matrix/media/v3/upload"){return true};return r.Method==http.MethodPut&&(strings.HasPrefix(p,"/_matrix/media/v3/upload/")||strings.HasPrefix(p,"/_matrix/media/v1/upload/"))}
func periodStart(p string,n time.Time)time.Time{switch strings.ToLower(p){case"day","daily":y,m,d:=n.Date();return time.Date(y,m,d,0,0,0,0,n.Location());case"week","weekly":y,m,d:=n.Date();return time.Date(y,m,d-((int(n.Weekday())+6)%7),0,0,0,0,n.Location());case"month","monthly":y,m,_:=n.Date();return time.Date(y,m,1,0,0,0,0,n.Location());default:return time.Unix(0,0)}}
func matrixError(w http.ResponseWriter,status int,code,msg string){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_=json.NewEncoder(w).Encode(map[string]string{"errcode":code,"error":msg})}
func(g *Gate)resolveUser(ctx context.Context,r *http.Request)(string,error){token:=strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"),"Bearer "));if token==""{token=r.URL.Query().Get("access_token")};if token==""{return "",fmt.Errorf("missing access token")};sum:=sha256.Sum256([]byte(token));key:=hex.EncodeToString(sum[:]);user,err:=g.cache.GetUser(ctx,key);if err==nil&&user!=""{return user,nil};user,err=g.resolver.WhoAmI(ctx,token);if err!=nil{return "",err};_=g.cache.SetUser(ctx,key,user,60*time.Second);return user,nil}
func(g *Gate)usageKey(ctx context.Context,q *Quota,user string)(string,int64,error){if q.MaxTotalBytes<=0{return "",0,nil};start:=periodStart(q.Period,time.Now());key:=user+":"+strconv.FormatInt(start.Unix(),10);used,ok,err:=g.cache.GetUsage(ctx,key);if err!=nil{return "",0,err};if !ok{observed,e:=g.resolver.MediaBytesSince(ctx,user,start);if e!=nil{return "",0,e};if e=g.cache.SetUsage(ctx,key,observed,2*time.Minute);e!=nil{return "",0,e};used,ok,err=g.cache.GetUsage(ctx,key);if err!=nil{return "",0,err};if !ok{used=observed}};return key,used,nil}
func(g *Gate)enforce(w http.ResponseWriter,r *http.Request,user string,create bool,mediaID string){
 ctx,cancel:=context.WithTimeout(r.Context(),10*time.Second);defer cancel()
 quota,err:=g.quotas.GetQuota(ctx,user);if err!=nil{matrixError(w,503,"M_UNKNOWN","Unable to read media quota");return}
 if quota==nil{if create{r=r.WithContext(context.WithValue(r.Context(),ownerContextKey{},user))};recorder:=&statusRecorder{ResponseWriter:w,status:200};g.proxy.ServeHTTP(recorder,r);if !create&&mediaID!=""&&recorder.status>=200&&recorder.status<300{_=g.cache.DeleteMediaOwner(context.Background(),mediaID)};return}
 if !create&&r.ContentLength<0{matrixError(w,411,"M_UNKNOWN","Content-Length is required for quota enforcement");return}
 if !create&&quota.MaxFileBytes>0&&r.ContentLength>quota.MaxFileBytes{matrixError(w,413,"M_TOO_LARGE","Upload exceeds the per-file media quota");return}
 usageKey,used,err:=g.usageKey(ctx,quota,user);if err!=nil{matrixError(w,502,"M_UNKNOWN","Unable to calculate current media usage");return}
 reserved:=false
 if quota.MaxTotalBytes>0{
  if create&&used>=quota.MaxTotalBytes{matrixError(w,413,"M_TOO_LARGE","User media quota is already exhausted");return}
  if !create{allowed,e:=g.cache.ReserveUsage(ctx,usageKey,r.ContentLength,quota.MaxTotalBytes);if e!=nil{matrixError(w,503,"M_UNKNOWN","Unable to reserve media quota");return};if !allowed{matrixError(w,413,"M_TOO_LARGE","Upload would exceed the total media quota");return};reserved=true}
 }
 if create{r=r.WithContext(context.WithValue(r.Context(),ownerContextKey{},user));g.proxy.ServeHTTP(w,r);return}
 max:=quota.MaxFileBytes;if max<=0{max=r.ContentLength};r.Body=http.MaxBytesReader(w,r.Body,max)
 recorder:=&statusRecorder{ResponseWriter:w,status:200};g.proxy.ServeHTTP(recorder,r)
 if reserved&&!(recorder.status>=200&&recorder.status<300){_=g.cache.ReleaseUsage(context.Background(),usageKey,r.ContentLength)}
 if recorder.status>=200&&recorder.status<300&&mediaID!=""{_=g.cache.DeleteMediaOwner(context.Background(),mediaID)}
}
func(g *Gate)ServeHTTP(w http.ResponseWriter,r *http.Request){
 ctx,cancel:=context.WithTimeout(r.Context(),8*time.Second);defer cancel()
 if isCreate(r.URL.Path){user,err:=g.resolveUser(ctx,r);if err!=nil{matrixError(w,401,"M_UNKNOWN_TOKEN","A valid access token is required to create media");return};g.enforce(w,r,user,true,"");return}
 if !isUpload(r){g.proxy.ServeHTTP(w,r);return}
 user:="";mediaID:=""
 if r.Method==http.MethodPut{mediaID=path.Base(r.URL.Path);user,_=g.cache.GetMediaOwner(ctx,mediaID)}
 if user==""{var err error;user,err=g.resolveUser(ctx,r);if err!=nil{matrixError(w,401,"M_MISSING_TOKEN","Upload requires an access token or a media ID created through this gate");return}}
 g.enforce(w,r,user,false,mediaID)
}
type statusRecorder struct{http.ResponseWriter;status int}
func(s *statusRecorder)WriteHeader(code int){s.status=code;s.ResponseWriter.WriteHeader(code)}
func(s *statusRecorder)Flush(){if f,ok:=s.ResponseWriter.(http.Flusher);ok{f.Flush()}}
func(s *statusRecorder)Unwrap()http.ResponseWriter{return s.ResponseWriter}
