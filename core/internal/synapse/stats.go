package synapse
import("context";"encoding/json";"net/http";"net/url";"strconv";"time")
type UserSummary struct{Name string `json:"name"`;DisplayName string `json:"displayname"`;LastSeenTS int64 `json:"last_seen_ts"`;CreationTS int64 `json:"creation_ts"`}
type userList struct{Users []UserSummary `json:"users"`;NextToken string `json:"next_token"`;Total int64 `json:"total"`}
type roomList struct{TotalRooms int64 `json:"total_rooms"`}
func(c *Client)ListUsers(ctx context.Context,from string,limit int)([]UserSummary,string,int64,error){q:=url.Values{"limit":[]string{strconv.Itoa(limit)}};if from!=""{q.Set("from",from)};_,raw,err:=c.AdminRequest(ctx,http.MethodGet,"/_synapse/admin/v2/users",q,nil);if err!=nil{return nil,"",0,err};var data userList;if err=json.Unmarshal(raw,&data);err!=nil{return nil,"",0,err};return data.Users,data.NextToken,data.Total,nil}
func(c *Client)ServerCounts(ctx context.Context)(users,rooms,online int64,err error){_,_,users,err=c.ListUsers(ctx,"",1);if err!=nil{return};q:=url.Values{"limit":[]string{"1"}};_,raw,e:=c.AdminRequest(ctx,http.MethodGet,"/_synapse/admin/v1/rooms",q,nil);if e!=nil{err=e;return};var rr roomList;if e=json.Unmarshal(raw,&rr);e!=nil{err=e;return};rooms=rr.TotalRooms
 from:="";for page:=0;page<10000;page++{items,next,_,e:=c.ListUsers(ctx,from,100);if e!=nil{err=e;return};cutoff:=time.Now().Add(-5*time.Minute).UnixMilli();for _,item:=range items{if item.LastSeenTS>=cutoff{online++}};if next==""{break};from=next};return
}
func(c *Client)SendServerNotice(ctx context.Context,userID,message string)(string,error){body:=map[string]any{"user_id":userID,"content":map[string]string{"msgtype":"m.text","body":message}};raw,_:=json.Marshal(body);_,data,err:=c.AdminRequest(ctx,http.MethodPost,"/_synapse/admin/v1/send_server_notice",nil,raw);if err!=nil{return "",err};var response struct{EventID string `json:"event_id"`};if err=json.Unmarshal(data,&response);err!=nil{return "",err};return response.EventID,nil}
