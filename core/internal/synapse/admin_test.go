package synapse

import (
 "context"
 "net/http"
 "net/http/httptest"
 "sync/atomic"
 "testing"
 "time"
)

func TestAdminRequestRetriesTransientServerErrors(t *testing.T) {
 var calls atomic.Int32
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if calls.Add(1)<3 {http.Error(w,"temporary",http.StatusBadGateway);return}
  w.Header().Set("Content-Type","application/json")
  _,_=w.Write([]byte(`{"ok":true}`))
 }))
 defer server.Close()
 c:=New(server.URL,"admin-token",time.Second)
 status,body,err:=c.AdminRequest(context.Background(),http.MethodGet,"/_synapse/admin/v1/test",nil,nil)
 if err!=nil||status!=http.StatusOK||string(body)!=`{"ok":true}` {t.Fatalf("request did not recover: status=%d body=%s err=%v",status,body,err)}
 if calls.Load()!=3 {t.Fatalf("expected 3 attempts, got %d",calls.Load())}
}

func TestAdminRequestDoesNotRetryNonIdempotentPost(t *testing.T) {
 var calls atomic.Int32
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){calls.Add(1);http.Error(w,"temporary",http.StatusBadGateway)}))
 defer server.Close()
 c:=New(server.URL,"admin-token",time.Second)
 status,_,err:=c.AdminRequest(context.Background(),http.MethodPost,"/_synapse/admin/v1/test",nil,[]byte("{}"))
 if err==nil||status!=http.StatusBadGateway {t.Fatalf("expected upstream error, got status=%d err=%v",status,err)}
 if calls.Load()!=1 {t.Fatalf("POST must not be retried automatically; got %d attempts",calls.Load())}
}
