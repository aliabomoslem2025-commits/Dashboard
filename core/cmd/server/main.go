package main

import (
 "context"
 "errors"
 "net/http"
 "os"
 "os/signal"
 "syscall"
 "time"
 "matrix-dashboard/core/internal/config"
 "matrix-dashboard/core/internal/httpapi"
 "matrix-dashboard/core/internal/synapse"
 "github.com/go-chi/chi/v5"
 "github.com/rs/zerolog"
 "github.com/rs/zerolog/log"
)
func main() {
 if len(os.Args)>1 && os.Args[1]=="healthcheck" { os.Exit(0) }
 cfg,err:=config.Load(); if err!=nil { log.Fatal().Err(err).Msg("invalid configuration") }
 logger:=zerolog.New(os.Stdout).With().Timestamp().Logger()
 client:=synapse.New(cfg.SynapseBaseURL,cfg.SynapseAdminToken,cfg.RequestTimeout)
 h:=httpapi.New(client,cfg.CoreServiceToken,logger)
 router:=chi.NewRouter(); router.Get("/health",h.Health)
 router.Group(func(r chi.Router){ r.Use(h.Authenticate); r.Get("/api/v1/server/version",h.ServerVersion); h.registerAdminRoutes(r) })
 server:=&http.Server{Addr:cfg.HTTPAddr,Handler:router,ReadHeaderTimeout:5*time.Second}
 go func(){ logger.Info().Str("addr",cfg.HTTPAddr).Msg("mx-core listening"); if err:=server.ListenAndServe(); err!=nil && !errors.Is(err,http.ErrServerClosed){ log.Fatal().Err(err).Msg("http server failed") } }()
 stop:=make(chan os.Signal,1); signal.Notify(stop,syscall.SIGINT,syscall.SIGTERM); <-stop
 ctx,cancel:=context.WithTimeout(context.Background(),10*time.Second); defer cancel(); _=server.Shutdown(ctx)
}
