package synapse
import("context";"encoding/json";"net/http";"net/http/httptest";"testing";"time")
func TestSetRoomReadOnlyJoinsAdminBeforeApplyingPowerLevels(t *testing.T){
 calls:=[]string{}
 srv:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  calls=append(calls,r.Method+" "+r.URL.Path)
  switch r.URL.Path{
  case "/_synapse/admin/v1/rooms/!r:hs/state":
   w.Header().Set("Content-Type","application/json");_,_=w.Write([]byte(\`{"state":[{"type":"m.room.power_levels","content":{"users_default":0,"events_default":0,"state_default":50,"events":{"m.room.message":0}}}]}\`))
  case "/_synapse/admin/v1/rooms/!r:hs/make_room_admin":
   w.WriteHeader(200)
  case "/_matrix/client/v3/join/!r:hs":
   if r.Header.Get("Authorization")!="Bearer admin-token"{t.Error("join did not use admin token")};w.Header().Set("Content-Type","application/json");_,_=w.Write([]byte(\`{"room_id":"!r:hs"}\`))
  case "/_matrix/client/v3/rooms/!r:hs/state/m.room.power_levels":
   var levels map[string]any;if err:=json.NewDecoder(r.Body).Decode(&levels);err!=nil{t.Fatal(err)}
   if levels["events_default"]!=float64(100)||levels["state_default"]!=float64(100){t.Errorf("power levels not read-only: %#v",levels)}
   if levels["events"].(map[string]any)["m.room.message"]!=float64(100){t.Errorf("message threshold not raised: %#v",levels)}
   w.WriteHeader(200)
  default:http.NotFound(w,r)
 }
 }))
 defer srv.Close();client:=New(srv.URL,"admin-token",time.Second)
 if err:=client.SetRoomReadOnly(context.Background(),"!r:hs");err!=nil{t.Fatal(err)}
 if len(calls)!=4{t.Fatalf("expected state, make-admin, join, make-admin, power-level sequence; got %#v",calls)}
}
