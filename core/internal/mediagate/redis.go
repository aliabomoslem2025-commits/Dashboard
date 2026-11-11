package mediagate
import("context";"fmt";"time";"github.com/redis/go-redis/v9")
type RedisCache struct{Client *redis.Client}
func(r RedisCache)GetUser(ctx context.Context,k string)(string,error){v,e:=r.Client.Get(ctx,"matrix:media:token:"+k).Result();if e==redis.Nil{return "",nil};return v,e}
func(r RedisCache)SetUser(ctx context.Context,k,u string,t time.Duration)error{return r.Client.Set(ctx,"matrix:media:token:"+k,u,t).Err()}
func(r RedisCache)GetUsage(ctx context.Context,k string)(int64,bool,error){v,e:=r.Client.Get(ctx,"matrix:media:usage:"+k).Int64();if e==redis.Nil{return 0,false,nil};return v,e==nil,e}
func(r RedisCache)SetUsage(ctx context.Context,k string,n int64,t time.Duration)error{return r.Client.Set(ctx,"matrix:media:usage:"+k,n,t).Err()}
func(r RedisCache)AddUsage(ctx context.Context,k string,n int64)error{if n<0{return fmt.Errorf("negative upload size")};return r.Client.IncrBy(ctx,"matrix:media:usage:"+k,n).Err()}
func(r RedisCache)ReserveUsage(ctx context.Context,k string,n,limit int64)(bool,error){if n<0||limit<0{return false,fmt.Errorf("invalid quota reservation")};script:=`
local current=tonumber(redis.call('GET',KEYS[1]) or '0')
local amount=tonumber(ARGV[1])
local limit=tonumber(ARGV[2])
if current+amount>limit then return 0 end
redis.call('INCRBY',KEYS[1],amount)
redis.call('EXPIRE',KEYS[1],ARGV[3])
return 1
`;v,e:=r.Client.Eval(ctx,script,[]string{"matrix:media:usage:"+k},n,limit,600).Int();return v==1,e}
func(r RedisCache)ReleaseUsage(ctx context.Context,k string,n int64)error{script:=`
local current=tonumber(redis.call('GET',KEYS[1]) or '0')
local amount=tonumber(ARGV[1])
local next=current-amount
if next<0 then next=0 end
redis.call('SET',KEYS[1],next,'EX',600)
return next
`;return r.Client.Eval(ctx,script,[]string{"matrix:media:usage:"+k},n).Err()}
func(r RedisCache)GetMediaOwner(ctx context.Context,id string)(string,error){v,e:=r.Client.Get(ctx,"matrix:media:owner:"+id).Result();if e==redis.Nil{return "",nil};return v,e}
func(r RedisCache)SetMediaOwner(ctx context.Context,id,user string,t time.Duration)error{return r.Client.Set(ctx,"matrix:media:owner:"+id,user,t).Err()}
func(r RedisCache)DeleteMediaOwner(ctx context.Context,id string)error{return r.Client.Del(ctx,"matrix:media:owner:"+id).Err()}
