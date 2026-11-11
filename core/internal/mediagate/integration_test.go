package mediagate
import("context";"net/http";"net/http/httptest";"net/url";"os";"strings";"sync/atomic";"testing";"github.com/redis/go-redis/v9")
func TestMediaGateWithRedisAndSynapseUpstream(t *testing.T){
 addr:=os.Getenv("REDIS_TEST_ADDR");if addr==""{t.Skip("REDIS_TEST_ADDR is not configured")}
 ctx:=context.Background();client:=redis.NewClient(&redis.Options{Addr:addr});defer client.Close();if err:=client.FlushDB(ctx).Err();err!=nil{t.Fatal(err)}
 cache:=RedisCache{Client:client};var forwarded atomic.Int32
 upstream:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){forwarded.Add(1);_,_=w.Write([]byte(`{"content_uri":"mxc://example.org/uploaded"}`))}));defer upstream.Close()
 target,_:=url.Parse(upstream.URL);gate:=New(target,fakeQuota{&Quota{MaxTotalBytes:100,MaxFileBytes:50,Period:"lifetime"}},cache,fakeResolver{})
 upload:=func(body string) *httptest.ResponseRecorder {req:=httptest.NewRequest(http.MethodPost,"/_matrix/media/v3/upload",strings.NewReader(body));req.Header.Set("Authorization","Bearer integration-token");req.ContentLength=int64(len(body));rec:=httptest.NewRecorder();gate.ServeHTTP(rec,req);return rec}
 first:=upload("12345");if first.Code!=200{t.Fatalf("first upload should pass, got %d: %s",first.Code,first.Body.String())}
 used,ok,err:=cache.GetUsage(ctx,"@alice:example.org:0");if err!=nil||!ok||used!=95{t.Fatalf("expected Redis usage 95, got %d ok=%v err=%v",used,ok,err)}
 second:=upload("1234567890");if second.Code!=413||!strings.Contains(second.Body.String(),"M_TOO_LARGE"){t.Fatalf("second upload should be rejected by total quota, got %d: %s",second.Code,second.Body.String())}
 if forwarded.Load()!=1{t.Fatalf("rejected upload reached Synapse; forwarded=%d",forwarded.Load())}
 if err:=cache.InvalidateUserUsage(ctx,"@alice:example.org");err!=nil{t.Fatal(err)}
 if _,ok,err=cache.GetUsage(ctx,"@alice:example.org:0");err!=nil||ok{t.Fatalf("usage key should be invalidated, ok=%v err=%v",ok,err)}
 third:=upload("12345");if third.Code!=200{t.Fatalf("upload after invalidation should re-check usage and pass, got %d: %s",third.Code,third.Body.String())}
 if forwarded.Load()!=2{t.Fatalf("expected two uploads forwarded to Synapse, got %d",forwarded.Load())}
}
