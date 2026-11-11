package synapse

import("context";"net/http";"net/http/httptest";"testing";"time")
func TestServerVersionUsesAdminBearerToken(t *testing.T){s:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/_synapse/admin/v1/server_version"{t.Errorf("unexpected path %s",r.URL.Path)};if r.Header.Get("Authorization")!="Bearer secret"{t.Error("missing admin bearer token")};w.Header().Set("Content-Type","application/json");_,_=w.Write([]byte("{\"server_version\":\"1.120.0\",\"python_version\":\"3.12.0\"}"))}));defer s.Close();c:=New(s.URL,"secret",time.Second);v,e:=c.ServerVersion(context.Background());if e!=nil{t.Fatal(e)};if v.ServerVersion!="1.120.0"{t.Fatalf("unexpected version %s",v.ServerVersion)}}
