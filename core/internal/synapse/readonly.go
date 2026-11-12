package synapse
import("bytes";"context";"encoding/json";"fmt";"io";"net/http";"net/url")
func(c *Client)SetRoomReadOnly(ctx context.Context,roomID string)error{
 room:=url.PathEscape(roomID)
 _,raw,err:=c.AdminRequest(ctx,http.MethodGet,"/_synapse/admin/v1/rooms/"+room+"/state",nil,nil);if err!=nil{return err}
 var state map[string]any
 if err=json.Unmarshal(raw,&state);err!=nil{return err}
 events,ok:=state["state"].([]any);if !ok{return fmt.Errorf("invalid room state response")}
 var levels map[string]any
 for _,item:=range events{event,ok:=item.(map[string]any);if !ok||event["type"]!="m.room.power_levels"{continue};levels,_=event["content"].(map[string]any);break}
 if levels==nil{return fmt.Errorf("room has no m.room.power_levels state event")}
 levels["events_default"]=float64(100);levels["state_default"]=float64(100)
 if explicit,ok:=levels["events"].(map[string]any);ok{for key:=range explicit{explicit[key]=float64(100)};levels["events"]=explicit}
 // make_room_admin invites the local admin account to a non-public room. Join
 // explicitly as well: public rooms do not require an invite, and the caller
 // must be a member before it can send the power-level state event.
 makeAdmin:=func()error{_,_,e:=c.AdminRequest(ctx,http.MethodPost,"/_synapse/admin/v1/rooms/"+room+"/make_room_admin",nil,[]byte("{}"));return e}
 if err=makeAdmin();err!=nil{return fmt.Errorf("grant room admin power: %w",err)}
 joinReq,err:=http.NewRequestWithContext(ctx,http.MethodPost,c.baseURL+"/_matrix/client/v3/join/"+room,bytes.NewReader([]byte("{}")));if err!=nil{return err}
 joinReq.Header.Set("Authorization","Bearer "+c.token);joinReq.Header.Set("Content-Type","application/json")
 joinResp,err:=c.httpClient.Do(joinReq);if err!=nil{return fmt.Errorf("join room as server admin: %w",err)}
 joinBody,_:=io.ReadAll(io.LimitReader(joinResp.Body,1<<20));joinResp.Body.Close()
 if joinResp.StatusCode<200||joinResp.StatusCode>=300{return fmt.Errorf("joining room as server admin returned HTTP %d: %s",joinResp.StatusCode,string(joinBody))}
 if err=makeAdmin();err!=nil{return fmt.Errorf("grant room admin power after joining: %w",err)}
 body,err:=json.Marshal(levels);if err!=nil{return err}
 req,err:=http.NewRequestWithContext(ctx,http.MethodPut,c.baseURL+"/_matrix/client/v3/rooms/"+room+"/state/m.room.power_levels",bytes.NewReader(body));if err!=nil{return err}
 req.Header.Set("Authorization","Bearer "+c.token);req.Header.Set("Content-Type","application/json")
 resp,err:=c.httpClient.Do(req);if err!=nil{return err};defer resp.Body.Close()
 if resp.StatusCode<200||resp.StatusCode>=300{return fmt.Errorf("setting room power levels returned HTTP %d",resp.StatusCode)}
 return nil
}
