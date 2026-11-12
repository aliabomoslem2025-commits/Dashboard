package synapse
import("context";"net/http";"net/http/httptest";"testing";"time")
func TestUserMediaStatsUsesSynapseStatisticsAdminAPI(t *testing.T){
 srv:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path!="/_synapse/admin/v1/statistics/users/media"{t.Errorf("unexpected endpoint: %s",r.URL.Path)}
  if r.Header.Get("Authorization")!="Bearer admin-token"{t.Error("admin token missing")}
  q:=r.URL.Query();if q.Get("from")!="100"||q.Get("limit")!="250"||q.Get("search_term")!="alice"||q.Get("order_by")!="media_length"||q.Get("dir")!="b"{t.Errorf("query mismatch: %v",q)}
  w.Header().Set("Content-Type","application/json");_,_=w.Write([]byte(`{"users":[{"user_id":"@alice:example.org","displayname":"Alice","media_count":3,"media_length":4096}],"next_token":350,"total":900}`))
 }))
 defer srv.Close();client:=New(srv.URL,"admin-token",time.Second)
 page,err:=client.UserMediaStats(context.Background(),"100",250,"alice","media_length","b");if err!=nil{t.Fatal(err)}
 if len(page.Users)!=1||page.Users[0].UserID!="@alice:example.org"||page.Users[0].MediaLength!=4096||page.Users[0].MediaCount!=3{t.Fatalf("unexpected media stats: %#v",page)}
 if page.NextToken==nil||*page.NextToken!=350||page.Total!=900{t.Fatalf("pagination fields missing: %#v",page)}
}
