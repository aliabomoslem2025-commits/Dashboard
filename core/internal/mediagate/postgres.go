package mediagate
import("context";"errors";"fmt";"github.com/jackc/pgx/v5";"github.com/jackc/pgx/v5/pgxpool")
type PostgresQuotas struct{Pool *pgxpool.Pool}
func(p PostgresQuotas)GetQuota(ctx context.Context,userID string)(*Quota,error){var q Quota;err:=p.Pool.QueryRow(ctx,"SELECT max_total_bytes,max_file_bytes,period FROM user_quotas WHERE user_id=$1",userID).Scan(&q.MaxTotalBytes,&q.MaxFileBytes,&q.Period);if err!=nil{if errors.Is(err,pgx.ErrNoRows){return nil,nil};return nil,fmt.Errorf("query media quota: %w",err)};return &q,nil}
