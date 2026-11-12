package synapse

import (
 "context"
 "net/http"
 "net/http/httptest"
 "sync/atomic"
 "testing"
 "time"
)

func TestAdminRequestDoesNotRetryDelete(t *testing.T) {
 var calls atomic.Int32
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){calls.Add(1);w.WriteHeader(http.StatusInternalServerError)}))
 defer server.Close()
 client:=New(server.URL,"admin-token",time.Second)
 _,_,err:=client.AdminRequest(context.Background(),http.MethodDelete,"/_synapse/admin/v1/resource",nil,nil)
 if err==nil { t.Fatal("expected upstream error") }
 if got:=calls.Load();got!=1 { t.Fatalf("DELETE retried %d times; want exactly one attempt",got) }
}
