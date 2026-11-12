package httpapi

import (
 "net/http"
 "net/http/httptest"
 "net/url"
 "strings"
 "testing"
 "time"
 "matrix-dashboard/core/internal/synapse"
 "github.com/go-chi/chi/v5"
 "github.com/rs/zerolog"
)
func userRouter(upstream *httptest.Server) http.Handler {
 client:=synapse.New(upstream.URL,"synapse-admin",time.Second)
 h:=New(client,"core-token",zerolog.Nop())
 router:=chi.NewRouter()
 router.Group(func(r chi.Router){r.Use(h.Authenticate);h.RegisterAdminRoutes(r)})
 return router
}
func TestUserAdminRoutesMapToDocumentedSynapsePaths(t *testing.T) {
 userID:="@alice:example.org"
 cases:=[]struct{name,method,path,upstreamPath,body string}{
  {"details",http.MethodGet,"/api/v1/users/"+url.PathEscape(userID),"/_synapse/admin/v2/users/"+userID,""},
  {"create or update",http.MethodPut,"/api/v1/users/"+url.PathEscape(userID),"/_synapse/admin/v2/users/"+userID,"{\"displayname\":\"Alice\",\"admin\":false,\"deactivated\":false,\"locked\":false}"},
  {"reset password",http.MethodPost,"/api/v1/users/"+url.PathEscape(userID)+"/reset-password","/_synapse/admin/v1/reset_password/"+userID,"{\"new_password\":\"a-long-password\",\"logout_devices\":true}"},
  {"deactivate",http.MethodPost,"/api/v1/users/"+url.PathEscape(userID)+"/deactivate","/_synapse/admin/v1/deactivate/"+userID,"{\"erase\":false}"},
  {"suspend",http.MethodPut,"/api/v1/users/"+url.PathEscape(userID)+"/suspend","/_synapse/admin/v1/suspend/"+userID,"{\"suspend\":true}"},
  {"shadow ban",http.MethodPost,"/api/v1/users/"+url.PathEscape(userID)+"/shadow-ban","/_synapse/admin/v1/users/"+userID+"/shadow_ban",""},
  {"shadow unban",http.MethodDelete,"/api/v1/users/"+url.PathEscape(userID)+"/shadow-ban","/_synapse/admin/v1/users/"+userID+"/shadow_ban",""},
  {"rate limit get",http.MethodGet,"/api/v1/users/"+url.PathEscape(userID)+"/ratelimit","/_synapse/admin/v1/users/"+userID+"/override_ratelimit",""},
  {"rate limit set",http.MethodPost,"/api/v1/users/"+url.PathEscape(userID)+"/ratelimit","/_synapse/admin/v1/users/"+userID+"/override_ratelimit","{\"messages_per_second\":2,\"burst_count\":5}"},
  {"rate limit delete",http.MethodDelete,"/api/v1/users/"+url.PathEscape(userID)+"/ratelimit","/_synapse/admin/v1/users/"+userID+"/override_ratelimit",""},
  {"list devices",http.MethodGet,"/api/v1/users/"+url.PathEscape(userID)+"/devices","/_synapse/admin/v2/users/"+userID+"/devices",""},
  {"delete device",http.MethodDelete,"/api/v1/users/"+url.PathEscape(userID)+"/devices/PHONE","/_synapse/admin/v2/users/"+userID+"/devices/PHONE",""},
  {"joined rooms",http.MethodGet,"/api/v1/users/"+url.PathEscape(userID)+"/joined-rooms","/_synapse/admin/v1/users/"+userID+"/joined_rooms",""},
  {"whois",http.MethodGet,"/api/v1/users/"+url.PathEscape(userID)+"/whois","/_synapse/admin/v1/whois/"+userID,""},
 }
 for _,tc:=range cases { t.Run(tc.name,func(t *testing.T){
  upstream:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
   if r.URL.Path!=tc.upstreamPath { t.Errorf("path = %q, want %q",r.URL.Path,tc.upstreamPath) }
   if r.Method!=tc.method { t.Errorf("method = %s, want %s",r.Method,tc.method) }
   if r.Header.Get("Authorization")!="Bearer synapse-admin" { t.Error("Synapse admin bearer token was not forwarded") }
   w.Header().Set("Content-Type","application/json");w.WriteHeader(http.StatusOK);_,_=w.Write([]byte("{}"))
  }))
  defer upstream.Close()
  router:=userRouter(upstream)
  req:=httptest.NewRequest(tc.method,tc.path,strings.NewReader(tc.body));req.Header.Set("Authorization","Bearer core-token")
  rec:=httptest.NewRecorder();router.ServeHTTP(rec,req)
  if rec.Code!=http.StatusOK { t.Fatalf("status=%d response=%s",rec.Code,rec.Body.String()) }
 })}
}
func TestUserListForwardsPaginationSearchAndFilters(t *testing.T) {
 upstream:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path!="/_synapse/admin/v2/users" { t.Errorf("unexpected path %s",r.URL.Path) }
  for key,want:=range map[string]string{"from":"25","limit":"25","name":"alice","admins":"true","deactivated":"false","locked":"true","order_by":"last_seen_ts","dir":"b"} {
   if got:=r.URL.Query().Get(key);got!=want { t.Errorf("%s=%q, want %q",key,got,want) }
  }
  w.Header().Set("Content-Type","application/json");_,_=w.Write([]byte("{\"users\":[],\"total\":0}"))
 }))
 defer upstream.Close()
 router:=userRouter(upstream)
 req:=httptest.NewRequest(http.MethodGet,"/api/v1/users?from=25&limit=25&name=alice&admins=true&deactivated=false&locked=true&order_by=last_seen_ts&dir=b",nil)
 req.Header.Set("Authorization","Bearer core-token");rec:=httptest.NewRecorder();router.ServeHTTP(rec,req)
 if rec.Code!=200 { t.Fatalf("status=%d body=%s",rec.Code,rec.Body.String()) }
}
func TestUserListRejectsInvalidPaginationAndFilters(t *testing.T) {
 upstream:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){t.Error("upstream must not be called")}))
 defer upstream.Close()
 router:=userRouter(upstream)
 req:=httptest.NewRequest(http.MethodGet,"/api/v1/users?limit=9999&locked=maybe",nil);req.Header.Set("Authorization","Bearer core-token")
 rec:=httptest.NewRecorder();router.ServeHTTP(rec,req)
 if rec.Code!=400 { t.Fatalf("expected 400, got %d: %s",rec.Code,rec.Body.String()) }
}
