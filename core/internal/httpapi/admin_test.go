package httpapi
import(
 "net/http";"net/http/httptest";"testing";"time"
 "matrix-dashboard/core/internal/synapse"
 "github.com/go-chi/chi/v5"
 "github.com/rs/zerolog"
)
func TestUserListUsesSynapseV2AdminPath(t *testing.T){
 upstream:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path!="/_synapse/admin/v2/users"{t.Errorf("unexpected Synapse path: %s",r.URL.Path)}
  if r.Header.Get("Authorization")!="Bearer synapse-admin"{t.Errorf("admin token not forwarded")}
  if r.URL.Query().Get("limit")!="25"{t.Errorf("limit query not forwarded")}
  w.Header().Set("Content-Type","application/json");w.WriteHeader(http.StatusOK);_,_=w.Write([]byte(`{"users":[],"total":0}`))
 }))
 defer upstream.Close()
 client:=synapse.New(upstream.URL,"synapse-admin",time.Second)
 h:=New(client,"core-token",zerolog.Nop())
 router:=chi.NewRouter();router.Group(func(r chi.Router){r.Use(h.Authenticate);h.registerAdminRoutes(r)})
 req:=httptest.NewRequest(http.MethodGet,"/api/v1/users?limit=25",nil);req.Header.Set("Authorization","Bearer core-token")
 rec:=httptest.NewRecorder();router.ServeHTTP(rec,req)
 if rec.Code!=200{t.Fatalf("expected 200, got %d: %s",rec.Code,rec.Body.String())}
}
func TestAdminProxyRequiresCoreServiceToken(t *testing.T){
 client:=synapse.New("http://127.0.0.1:9","admin",time.Second);h:=New(client,"core-token",zerolog.Nop())
 router:=chi.NewRouter();router.Group(func(r chi.Router){r.Use(h.Authenticate);h.registerAdminRoutes(r)})
 req:=httptest.NewRequest(http.MethodGet,"/api/v1/users",nil);req.Header.Set("Authorization","Bearer wrong")
 rec:=httptest.NewRecorder();router.ServeHTTP(rec,req)
 if rec.Code!=401{t.Fatalf("expected 401, got %d",rec.Code)}
}
