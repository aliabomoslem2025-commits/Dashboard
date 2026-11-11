package mediagate
import("context";"fmt";"time";"github.com/redis/go-redis/v9")
type RedisCache struct{Client *redis.Client}
func(r RedisCache)GetUser(ctx context.Context,k string)(string,error){v,e:=r.Client.Get(ctx,"matrix:media:token:"+k).Result();if e==redis.Nil{return "",nil};return v,e}
func(r RedisCache)SetUser(ctx context.Context,k,u string,t time.Duration)error{return r.Client.Set(ctx,"matrix:media:token:"+k,u,t).Err()}
func(r RedisCache)GetUsage(ctx context.Context,k string)(int64,bool,error){v,e:=r.Client.Get(ctx,"matrix:media:usage:"+k).Int64();if e==redis.Nil{return 0,false,nil};return v,e==nil,e}
func(r RedisCache)SetUsage(ctx context.Context,k string,n int64,t time.Duration)error{return r.Client.Set(ctx,"matrix:media:usage:"+k,n,t).Err()}
func(r RedisCache)AddUsage(ctx context.Context,k string,n int64)error{if n<0{return fmt.Errorf("negative upload size")};return r.Client.IncrBy(ctx,"matrix:media:usage:"+k,n).Err()}
func(r RedisCache)GetMediaOwner(ctx context.Context,id string)(string,error){v,e:=r.Client.Get(ctx,"matrix:media:owner:"+id).Result();if e==redis.Nil{return "",nil};return v,e}
func(r RedisCache)SetMediaOwner(ctx context.Context,id,user string,t time.Duration)error{return r.Client.Set(ctx,"matrix:media:owner:"+id,user,t).Err()}
func(r RedisCache)DeleteMediaOwner(ctx context.Context,id string)error{return r.Client.Del(ctx,"matrix:media:owner:"+id).Err()}
