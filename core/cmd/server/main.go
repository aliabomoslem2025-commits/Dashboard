package main
import(
 "context";"errors";"fmt";"net/http";"net/url";"os";"os/signal";"syscall";"time"
 "matrix-dashboard/core/internal/config";"matrix-dashboard/core/internal/httpapi";"matrix-dashboard/core/internal/jobs";"matrix-dashboard/core/internal/mediagate";"matrix-dashboard/core/internal/synapse"
 "github.com/go-chi/chi/v5";"github.com/jackc/pgx/v5/pgxpool";"github.com/redis/go-redis/v9";"github.com/rs/zerolog";"github.com/rs/zerolog/log"
)
func main(){
 if len(os.Args)>1&&os.Args[1]=="healthcheck"{resp,err:=http.Get("http://127.0.0.1:8081/health");if err!=nil{os.Exit(1)};defer resp.Body.Close();if resp.StatusCode!=http.StatusOK{os.Exit(1)};os.Exit(0)}
 cfg,err:=config.Load();if err!=nil{log.Fatal().Err(err).Msg("invalid configuration")}
 logger:=zerolog.New(os.Stdout).With().Timestamp().Logger()
 client:=synapse.New(cfg.SynapseBaseURL,cfg.SynapseAdminToken,cfg.RequestTimeout)
 h:=httpapi.New(client,cfg.CoreServiceToken,logger)
 ctx:=context.Background()
 dsn:=fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",url.QueryEscape(cfg.DBUser),url.QueryEscape(cfg.DBPassword),cfg.DBHost,cfg.DBPort,cfg.DBName)
 pool,err:=pgxpool.New(ctx,dsn);if err!=nil{log.Fatal().Err(err).Msg("database pool initialization failed")};defer pool.Close()
 if err=pool.Ping(ctx);err!=nil{log.Fatal().Err(err).Msg("dashboard database unavailable")};h.SetDB(pool)
 redisClient:=redis.NewClient(&redis.Options{Addr:cfg.RedisAddr,Password:os.Getenv("REDIS_PASSWORD"),DB:0});if err=redisClient.Ping(ctx).Err();err!=nil{log.Fatal().Err(err).Msg("Redis unavailable")};defer redisClient.Close()
 target,err:=url.Parse(cfg.SynapseBaseURL);if err!=nil{log.Fatal().Err(err).Msg("invalid Synapse URL")}
 gate:=mediagate.New(target,mediagate.PostgresQuotas{Pool:pool},mediagate.RedisCache{Client:redisClient},client)
 router:=chi.NewRouter();router.Get("/health",h.Health);router.Group(func(r chi.Router){r.Use(h.Authenticate);r.Get("/api/v1/server/version",h.ServerVersion);r.Get("/api/v1/metrics/current",h.MetricsCurrent);r.Get("/api/v1/metrics/stream",h.MetricsStream);h.registerAdminRoutes(r)})
 jobCtx,cancelJobs:=context.WithCancel(context.Background());defer cancelJobs()
 var synapseDB *pgxpool.Pool
 if synapseDSN:=os.Getenv("SYNAPSE_DB_URL");synapseDSN!=""{synapseDB,err=pgxpool.New(jobCtx,synapseDSN);if err!=nil{logger.Warn().Err(err).Msg("Synapse read-only DB configuration invalid; message counts disabled")}else if err=synapseDB.Ping(jobCtx);err!=nil{logger.Warn().Err(err).Msg("Synapse DB unavailable; message counts disabled");synapseDB.Close();synapseDB=nil}else{defer synapseDB.Close()}}
 collector:=&jobs.Collector{DB:pool,SynapseDB:synapseDB,Synapse:client,PrometheusURL:os.Getenv("SYNAPSE_PROMETHEUS_URL"),Logger:logger}
 broadcaster:=&jobs.Broadcaster{DB:pool,Synapse:client,Logger:logger}
 go collector.Run(jobCtx);go broadcaster.Run(jobCtx)
 servers:=[]*http.Server{{Addr:cfg.HTTPAddr,Handler:router,ReadHeaderTimeout:5*time.Second},{Addr:cfg.MediaGateAddr,Handler:gate,ReadHeaderTimeout:5*time.Second,MaxHeaderBytes:1<<20}}
 for _,s:=range servers{s:=s;go func(){logger.Info().Str("addr",s.Addr).Msg("listener starting");if e:=s.ListenAndServe();e!=nil&&!errors.Is(e,http.ErrServerClosed){log.Fatal().Err(e).Str("addr",s.Addr).Msg("listener failed")}}()}
 stop:=make(chan os.Signal,1);signal.Notify(stop,syscall.SIGINT,syscall.SIGTERM);<-stop;cancelJobs()
 shutdown,cancel:=context.WithTimeout(context.Background(),15*time.Second);defer cancel();for _,s:=range servers{_=s.Shutdown(shutdown)}
}
