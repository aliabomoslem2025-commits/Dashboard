package synapse
import("context";"encoding/json";"fmt";"net/http";"net/url";"strconv";"strings")
type UserSummary struct{Name string `json:"name"`;DisplayName string `json:"displayname"`;CreationTS int64 `json:"creation_ts"`}
type userList struct{Users []UserSummary `json:"users"`;NextToken string `json:"next_token"`;Total int64 `json:"total"`}
type roomList struct{TotalRooms int64 `json:"total_rooms"`}
func(c *Client)ListUsers(ctx context.Context,from string,limit int)([]UserSummary,string,int64,error){q:=url.Values{"limit":[]string{strconv.Itoa(limit)}};if from!=""{q.Set("from",from)};_,raw,err:=c.AdminRequest(ctx,http.MethodGet,"/_synapse/admin/v2/users",q,nil);if err!=nil{return nil,"",0,err};var data userList;if err=json.Unmarshal(raw,&data);err!=nil{return nil,"",0,err};return data.Users,data.NextToken,data.Total,nil}
func(c *Client)ServerCounts(ctx context.Context)(users,rooms,online int64,err error){_,_,users,err=c.ListUsers(ctx,"",1);if err!=nil{return};q:=url.Values{"limit":[]string{"1"}};_,raw,e:=c.AdminRequest(ctx,http.MethodGet,"/_synapse/admin/v1/rooms",q,nil);if e!=nil{err=e;return};var rr roomList;if e=json.Unmarshal(raw,&rr);e!=nil{err=e;return};rooms=rr.TotalRooms;return}
type deviceList struct{Devices []struct{LastSeenTS int64 `json:"last_seen_ts"`} `json:"devices"`}
func(c *Client)UserLastSeenTS(ctx context.Context,userID string)(int64,error){_,raw,err:=c.AdminRequest(ctx,http.MethodGet,"/_synapse/admin/v2/users/"+url.PathEscape(userID)+"/devices",nil,nil);if err!=nil{return 0,err};var data deviceList;if err=json.Unmarshal(raw,&data);err!=nil{return 0,err};var latest int64;for _,d:=range data.Devices{if d.LastSeenTS>latest{latest=d.LastSeenTS}};return latest,nil}
func(c *Client)SendServerNotice(ctx context.Context,txnID,userID,message string)(string,error){
 if strings.TrimSpace(txnID)==""||len(txnID)>128{return "",fmt.Errorf("invalid server notice transaction ID")}
 if !strings.HasPrefix(userID,"@")||!strings.Contains(userID,":"){return "",fmt.Errorf("invalid Matrix user ID")}
 if strings.TrimSpace(message)==""{return "",fmt.Errorf("server notice message cannot be empty")}
 body:=map[string]any{"user_id":userID,"content":map[string]string{"msgtype":"m.text","body":message}}
 raw,err:=json.Marshal(body);if err!=nil{return "",fmt.Errorf("encode server notice: %w",err)}
 status,data,err:=c.AdminRequest(ctx,http.MethodPut,"/_synapse/admin/v1/send_server_notice/"+url.PathEscape(txnID),nil,raw)
 if err!=nil{return "",fmt.Errorf("Synapse send_server_notice failed (HTTP %d): %w",status,err)}
 var response struct{EventID string `json:"event_id"`}
 if err=json.Unmarshal(data,&response);err!=nil{return "",fmt.Errorf("decode server notice response: %w",err)}
 if response.EventID==""{return "",fmt.Errorf("Synapse server notice response did not include event_id")}
 return response.EventID,nil
}
