package mediagate
import("context";"crypto/sha256";"encoding/hex";"encoding/json";"net/http";"net/http/httputil";"net/url";"strconv";"strings";"time")
type Quota struct{MaxTotalBytes int64;MaxFileBytes int64;Period string}
type QuotaRepository interface{GetQuota(context.Context,string)(*Quota,error)}
type Cache interface{GetUser(context.Context,string)(string,error);SetUser(context.Context,string,string,time.Duration)error;GetUsage(context.Context,string)(int64,bool,error);SetUsage(context.Context,string,int64,time.Duration)error;AddUsage(context.Context,string,int64)error}
type Resolver interface{WhoAmI(context.Context,string)(string,error);MediaBytesSince(context.Context,string,time.Time)(int64,error)}
type Gate struct{proxy http.Handler;quotas QuotaRepository;cache Cache;resolver Resolver}
func New(target *url.URL,q QuotaRepository,c Cache,r Resolver)*Gate{p:=httputil.NewSingleHostReverseProxy(target);p.ErrorHandler=func(w http.ResponseWriter,_ *http.Request,_ error){http.Error(w,"Upstream media service unavailable",http.StatusBadGateway)};return &Gate{p,q,c,r}}
func isUpload(p string)bool{return p=="/_matrix/media/v1/upload"||p=="/_matrix/media/v3/upload"||p=="/_matrix/client/v1/media/upload"}
func periodStart(p string,n time.Time)time.Time{switch strings.ToLower(p){case"day","daily":y,m,d:=n.Date();return time.Date(y,m,d,0,0,0,0,n.Location());case"week","weekly":y,m,d:=n.Date();return time.Date(y,m,d-((int(n.Weekday())+6)%7),0,0,0,0,n.Location());case"month","monthly":y,m,_:=n.Date();return time.Date(y,m,1,0,0,0,0,n.Location());default:return time.Unix(0,0)}}
func matrixError(w http.ResponseWriter,status int,code,msg string){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_=json.NewEncoder(w).Encode(map[string]string{"errcode":code,"error":msg})}
func(g *Gate)ServeHTTP(w http.ResponseWriter,r *http.Request){
 if !isUpload(r.URL.Path){g.proxy.ServeHTTP(w,r);return}
 token:=strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"),"Bearer "));if token==""{token=r.URL.Query().Get("access_token")};if token==""{matrixError(w,401,"M_MISSING_TOKEN","Missing access token");return}
 sum:=sha256.Sum256([]byte(token));tokenKey:=hex.EncodeToString(sum[:]);ctx,cancel:=context.WithTimeout(r.Context(),8*time.Second);defer cancel()
 userID,err:=g.cache.GetUser(ctx,tokenKey);if err!=nil||userID==""{userID,err=g.resolver.WhoAmI(ctx,token);if err!=nil{matrixError(w,401,"M_UNKNOWN_TOKEN","Invalid or expired access token");return};_=g.cache.SetUser(ctx,tokenKey,userID,60*time.Second)}
 quota,err:=g.quotas.GetQuota(ctx,userID);if err!=nil{matrixError(w,503,"M_UNKNOWN","Unable to read media quota");return};if quota==nil{g.proxy.ServeHTTP(w,r);return}
 if r.ContentLength<0{matrixError(w,411,"M_UNKNOWN","Content-Length is required for quota enforcement");return}
 if quota.MaxFileBytes>0&&r.ContentLength>quota.MaxFileBytes{matrixError(w,413,"M_TOO_LARGE","Upload exceeds the per-file media quota");return}
 usageKey:=""
 if quota.MaxTotalBytes>0{start:=periodStart(quota.Period,time.Now());usageKey=userID+":"+strconv.FormatInt(start.Unix(),10);used,ok,e:=g.cache.GetUsage(ctx,usageKey);if e!=nil{matrixError(w,503,"M_UNKNOWN","Unable to read media usage");return};if !ok{used,e=g.resolver.MediaBytesSince(ctx,userID,start);if e!=nil{matrixError(w,502,"M_UNKNOWN","Unable to calculate current media usage");return};_=g.cache.SetUsage(ctx,usageKey,used,2*time.Minute)};if used+r.ContentLength>quota.MaxTotalBytes{matrixError(w,413,"M_TOO_LARGE","Upload would exceed the total media quota");return}}
 max:=quota.MaxFileBytes;if max<=0{max=r.ContentLength};r.Body=http.MaxBytesReader(w,r.Body,max)
 recorder:=&statusRecorder{ResponseWriter:w,status:200};g.proxy.ServeHTTP(recorder,r)
 if usageKey!=""&&recorder.status>=200&&recorder.status<300{_=g.cache.AddUsage(context.Background(),usageKey,r.ContentLength)}
}
type statusRecorder struct{http.ResponseWriter;status int}
func(s *statusRecorder)WriteHeader(code int){s.status=code;s.ResponseWriter.WriteHeader(code)}
func(s *statusRecorder)Flush(){if f,ok:=s.ResponseWriter.(http.Flusher);ok{f.Flush()}}
func(s *statusRecorder)Unwrap()http.ResponseWriter{return s.ResponseWriter}
