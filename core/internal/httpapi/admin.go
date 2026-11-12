package httpapi
import("bytes";"context";"encoding/json";"fmt";"io";"net/http";"net/url";"strconv";"strings";"time";"github.com/go-chi/chi/v5")
func segment(v string)string{return url.PathEscape(v)}
func(h *Handler)proxy(w http.ResponseWriter,r *http.Request,method,path string){ctx,cancel:=context.WithTimeout(r.Context(),20*time.Second);defer cancel();var body []byte;if r.Body!=nil{b,err:=io.ReadAll(http.MaxBytesReader(w,r.Body,2<<20));if err!=nil{write(w,413,envelope{OK:false,Error:&apiError{Code:"request_too_large",Message:"Request body exceeds 2 MiB"}});return};body=b};status,data,err:=h.synapse.AdminRequest(ctx,method,path,r.URL.Query(),body);if err!=nil{code:="synapse_error";msg:="Synapse Admin API request failed";var remote map[string]any;if json.Unmarshal(data,&remote)==nil{if s,ok:=remote["error"].(string);ok&&s!=""{msg=s};if s,ok:=remote["errcode"].(string);ok&&s!=""{code=s}};h.logger.Warn().Err(err).Str("path",path).Int("status",status).Msg("Synapse Admin API request failed");if status==0||status<400||status>599{status=http.StatusBadGateway};write(w,status,envelope{OK:false,Error:&apiError{Code:code,Message:msg}});return};var payload any;if len(data)>0&&json.Unmarshal(data,&payload)!=nil{payload=string(data)};write(w,status,envelope{OK:true,Data:payload})}
func validateUserPayload(kind string, body []byte) error {
 obj:=map[string]json.RawMessage{}
 if len(body)==0 { if kind=="deactivate"||kind=="ratelimit" { return nil }; return fmt.Errorf("JSON object body is required") }
 if err:=json.Unmarshal(body,&obj);err!=nil||obj==nil { return fmt.Errorf("request body must be a JSON object") }
 stringField:=func(key string,min,max int,required bool)error{
  raw,ok:=obj[key];if !ok {if required{return fmt.Errorf("%s is required",key)};return nil}
  var value string;if err:=json.Unmarshal(raw,&value);err!=nil{return fmt.Errorf("%s must be a string",key)}
  if len(value)<min||len(value)>max{return fmt.Errorf("%s length must be between %d and %d",key,min,max)}
  return nil
 }
 boolField:=func(key string,required bool)error{
  raw,ok:=obj[key];if !ok {if required{return fmt.Errorf("%s is required",key)};return nil}
  var value bool;if err:=json.Unmarshal(raw,&value);err!=nil{return fmt.Errorf("%s must be a boolean",key)}
  return nil
 }
 intField:=func(key string,required bool)error{
  raw,ok:=obj[key];if !ok {if required{return fmt.Errorf("%s is required",key)};return nil}
  var value int64;if err:=json.Unmarshal(raw,&value);err!=nil||value<0{return fmt.Errorf("%s must be a non-negative integer",key)}
  return nil
 }
 switch kind {
 case "user":
  allowed:=map[string]bool{"password":true,"logout_devices":true,"displayname":true,"admin":true,"deactivated":true,"locked":true}
  if len(obj)==0{return fmt.Errorf("at least one user field is required")}
  for key:=range obj{if !allowed[key]{return fmt.Errorf("unsupported user field: %s",key)}}
  if raw,ok:=obj["displayname"];ok&&string(raw)!="null"{if err:=stringField("displayname",0,255,false);err!=nil{return err}}
  if err:=stringField("password",12,255,false);err!=nil{return err}
  for _,key:=range []string{"admin","deactivated","locked","logout_devices"}{if err:=boolField(key,false);err!=nil{return err}}
 case "reset_password":
  if err:=stringField("new_password",12,255,true);err!=nil{return err}
  if err:=boolField("logout_devices",false);err!=nil{return err}
 case "deactivate":
  if err:=boolField("erase",false);err!=nil{return err}
 case "suspend":
  if err:=boolField("suspend",true);err!=nil{return err}
 case "ratelimit":
  if err:=intField("messages_per_second",false);err!=nil{return err}
  if err:=intField("burst_count",false);err!=nil{return err}
 default:
  return fmt.Errorf("unknown validation profile")
 }
 return nil
}
func(h *Handler)proxyUserWrite(w http.ResponseWriter,r *http.Request,method,path,kind string){
 body,err:=io.ReadAll(http.MaxBytesReader(w,r.Body,2<<20))
 if err!=nil{write(w,413,envelope{OK:false,Error:&apiError{Code:"request_too_large",Message:"Request body exceeds 2 MiB"}});return}
 if err=validateUserPayload(kind,body);err!=nil{write(w,400,envelope{OK:false,Error:&apiError{Code:"invalid_request",Message:err.Error()}});return}
 r.Body=io.NopCloser(bytes.NewReader(body))
 h.proxy(w,r,method,path)
}
func(h *Handler)RegisterAdminRoutes(r chi.Router){
 r.Get("/api/v1/users",h.listUsers)
 r.Get("/api/v1/media/stats",h.mediaStats)
 r.Get("/api/v1/users/{userID}",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodGet,"/_synapse/admin/v2/users/"+segment(chi.URLParam(q,"userID")))})
 r.Put("/api/v1/users/{userID}",func(w http.ResponseWriter,q *http.Request){h.proxyUserWrite(w,q,http.MethodPut,"/_synapse/admin/v2/users/"+segment(chi.URLParam(q,"userID")),"user")})
 r.Post("/api/v1/users/{userID}/reset-password",func(w http.ResponseWriter,q *http.Request){h.proxyUserWrite(w,q,http.MethodPost,"/_synapse/admin/v1/reset_password/"+segment(chi.URLParam(q,"userID")),"reset_password")})
 r.Post("/api/v1/users/{userID}/deactivate",func(w http.ResponseWriter,q *http.Request){h.proxyUserWrite(w,q,http.MethodPost,"/_synapse/admin/v1/deactivate/"+segment(chi.URLParam(q,"userID")),"deactivate")})
 r.Put("/api/v1/users/{userID}/suspend",func(w http.ResponseWriter,q *http.Request){h.proxyUserWrite(w,q,http.MethodPut,"/_synapse/admin/v1/suspend/"+segment(chi.URLParam(q,"userID")),"suspend")})
 r.Post("/api/v1/users/{userID}/shadow-ban",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodPost,"/_synapse/admin/v1/users/"+segment(chi.URLParam(q,"userID"))+"/shadow_ban")})
 r.Delete("/api/v1/users/{userID}/shadow-ban",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodDelete,"/_synapse/admin/v1/users/"+segment(chi.URLParam(q,"userID"))+"/shadow_ban")})
 r.Get("/api/v1/users/{userID}/ratelimit",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodGet,"/_synapse/admin/v1/users/"+segment(chi.URLParam(q,"userID"))+"/override_ratelimit")})
 r.Post("/api/v1/users/{userID}/ratelimit",func(w http.ResponseWriter,q *http.Request){h.proxyUserWrite(w,q,http.MethodPost,"/_synapse/admin/v1/users/"+segment(chi.URLParam(q,"userID"))+"/override_ratelimit","ratelimit")})
 r.Delete("/api/v1/users/{userID}/ratelimit",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodDelete,"/_synapse/admin/v1/users/"+segment(chi.URLParam(q,"userID"))+"/override_ratelimit")})
 r.Get("/api/v1/users/{userID}/devices",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodGet,"/_synapse/admin/v2/users/"+segment(chi.URLParam(q,"userID"))+"/devices")})
 r.Delete("/api/v1/users/{userID}/devices/{deviceID}",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodDelete,"/_synapse/admin/v2/users/"+segment(chi.URLParam(q,"userID"))+"/devices/"+segment(chi.URLParam(q,"deviceID")))})
 r.Get("/api/v1/users/{userID}/joined-rooms",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodGet,"/_synapse/admin/v1/users/"+segment(chi.URLParam(q,"userID"))+"/joined_rooms")})
 r.Get("/api/v1/users/{userID}/whois",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodGet,"/_synapse/admin/v1/whois/"+segment(chi.URLParam(q,"userID")))})
 r.Get("/api/v1/users/{userID}/media",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodGet,"/_synapse/admin/v1/users/"+segment(chi.URLParam(q,"userID"))+"/media")})
 r.Get("/api/v1/users/{userID}/media-usage",h.mediaUsage)
 r.Delete("/api/v1/users/{userID}/media",h.deleteUserMedia)
 r.Get("/api/v1/rooms",h.listRooms)
 r.Get("/api/v1/rooms/{roomID}",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodGet,"/_synapse/admin/v1/rooms/"+segment(chi.URLParam(q,"roomID")))})
 r.Get("/api/v1/rooms/{roomID}/members",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodGet,"/_synapse/admin/v1/rooms/"+segment(chi.URLParam(q,"roomID"))+"/members")})
 r.Get("/api/v1/rooms/{roomID}/state",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodGet,"/_synapse/admin/v1/rooms/"+segment(chi.URLParam(q,"roomID"))+"/state")})
 r.Get("/api/v1/rooms/{roomID}/block",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodGet,"/_synapse/admin/v1/rooms/"+segment(chi.URLParam(q,"roomID"))+"/block")})
 r.Put("/api/v1/rooms/{roomID}/block",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodPut,"/_synapse/admin/v1/rooms/"+segment(chi.URLParam(q,"roomID"))+"/block")})
 r.Get("/api/v1/rooms/{roomID}/delete-status",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodGet,"/_synapse/admin/v2/rooms/"+segment(chi.URLParam(q,"roomID"))+"/delete_status")})
 r.Delete("/api/v1/rooms/{roomID}",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodDelete,"/_synapse/admin/v2/rooms/"+segment(chi.URLParam(q,"roomID")))})
 r.Post("/api/v1/rooms/{roomID}/make-admin",func(w http.ResponseWriter,q *http.Request){h.proxy(w,q,http.MethodPost,"/_synapse/admin/v1/rooms/"+segment(chi.URLParam(q,"roomID"))+"/make_room_admin")})
 r.Post("/api/v1/rooms/{roomID}/read-only",h.roomReadOnly)
}

func pickQuery(q url.Values,keys []string)url.Values{out:=url.Values{};for _,k:=range keys{for _,v:=range q[k]{out.Add(k,v)}};return out}
func validateListQuery(w http.ResponseWriter,r *http.Request,orders,boolKeys,textKeys []string)(url.Values,bool){q:=r.URL.Query();if raw:=q.Get("from");raw!=""{if _,e:=strconv.ParseUint(raw,10,64);e!=nil{write(w,400,envelope{OK:false,Error:&apiError{Code:"invalid_from",Message:"from must be a non-negative numeric pagination token"}});return nil,false}};limit:=50;if raw:=q.Get("limit");raw!=""{n,e:=strconv.Atoi(raw);if e!=nil||n<1||n>500{write(w,400,envelope{OK:false,Error:&apiError{Code:"invalid_limit",Message:"limit must be between 1 and 500"}});return nil,false};limit=n};q.Set("limit",strconv.Itoa(limit));if d:=q.Get("dir");d!=""&&d!="f"&&d!="b"{write(w,400,envelope{OK:false,Error:&apiError{Code:"invalid_direction",Message:"dir must be f or b"}});return nil,false};if o:=q.Get("order_by");o!=""{found:=false;for _,v:=range orders{if o==v{found=true;break}};if !found{write(w,400,envelope{OK:false,Error:&apiError{Code:"invalid_order_by",Message:"Unsupported order_by value"}});return nil,false}};for _,k:=range boolKeys{if raw:=q.Get(k);raw!=""{if _,e:=strconv.ParseBool(raw);e!=nil{write(w,400,envelope{OK:false,Error:&apiError{Code:"invalid_filter",Message:k+" must be true or false"}});return nil,false}}};for _,k:=range textKeys{if len(q.Get(k))>255{write(w,400,envelope{OK:false,Error:&apiError{Code:"invalid_search",Message:k+" must be at most 255 characters"}});return nil,false}};return q,true}
func(h *Handler)listUsers(w http.ResponseWriter,r *http.Request){q,ok:=validateListQuery(w,r,[]string{"name","is_guest","admin","user_type","deactivated","shadow_banned","displayname","avatar_url","creation_ts","last_seen_ts"},[]string{"guests","admins","deactivated","locked"},[]string{"name","user_id"});if !ok{return};r.URL.RawQuery=pickQuery(q,[]string{"from","limit","name","user_id","guests","admins","deactivated","locked","order_by","dir"}).Encode();h.proxy(w,r,http.MethodGet,"/_synapse/admin/v2/users")}
func(h *Handler)listRooms(w http.ResponseWriter,r *http.Request){
 q,ok:=validateListQuery(w,r,[]string{"name","canonical_alias","joined_members","joined_local_members","version","creator","encryption","federatable","public","join_rules","guest_access","history_visibility","state_events","size"},[]string{"public_rooms","empty_rooms"},[]string{"search_term"});if !ok{return}
 if q.Get("order_by")=="size"{h.listLargestRooms(w,r,q.Get("search_term"));return}
 r.URL.RawQuery=pickQuery(q,[]string{"from","limit","search_term","order_by","dir","public_rooms","empty_rooms"}).Encode()
 h.proxy(w,r,http.MethodGet,"/_synapse/admin/v1/rooms")
}
func(h *Handler)listLargestRooms(w http.ResponseWriter,r *http.Request,searchTerm string){
 ctx,cancel:=context.WithTimeout(r.Context(),30*time.Second);defer cancel()
 _,raw,err:=h.synapse.AdminRequest(ctx,http.MethodGet,"/_synapse/admin/v1/statistics/database/rooms",nil,nil)
 if err!=nil{write(w,http.StatusBadGateway,envelope{OK:false,Error:&apiError{Code:"room_size_unavailable",Message:"Unable to retrieve largest room size estimates"}});return}
 var result struct{Rooms []struct{RoomID string `json:"room_id"`;EstimatedSize int64 `json:"estimated_size"`} `json:"rooms"`}
 if err=json.Unmarshal(raw,&result);err!=nil{write(w,http.StatusBadGateway,envelope{OK:false,Error:&apiError{Code:"invalid_synapse_response",Message:"Invalid room size statistics response"}});return}
 rooms:=make([]map[string]any,0,len(result.Rooms))
 for _,item:=range result.Rooms{
  if item.RoomID==""{continue}
  room:=map[string]any{"room_id":item.RoomID,"estimated_size":item.EstimatedSize}
  _,detail,detailErr:=h.synapse.AdminRequest(ctx,http.MethodGet,"/_synapse/admin/v1/rooms/"+segment(item.RoomID),nil,nil)
  if detailErr==nil{var fields map[string]any;if json.Unmarshal(detail,&fields)==nil{for k,v:=range fields{room[k]=v}}}
  if searchTerm!=""&&!strings.Contains(strings.ToLower(item.RoomID+" "+fmt.Sprint(room["name"])),strings.ToLower(searchTerm)){continue}
  rooms=append(rooms,room)
 }
 write(w,http.StatusOK,envelope{OK:true,Data:map[string]any{"rooms":rooms,"total_rooms":len(rooms),"next_batch":nil,"size_sort_limited":true}})
}
func(h *Handler)mediaStats(w http.ResponseWriter,r *http.Request){q,ok:=validateListQuery(w,r,[]string{"user_id","displayname","media_length","media_count"},nil,[]string{"search_term"});if !ok{return};r.URL.RawQuery=pickQuery(q,[]string{"from","limit","search_term","order_by","dir"}).Encode();h.proxy(w,r,http.MethodGet,"/_synapse/admin/v1/statistics/users/media")}
