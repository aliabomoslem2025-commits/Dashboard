package httpapi

import("context";"crypto/subtle";"encoding/json";"net/http";"strings";"time";"matrix-dashboard/core/internal/synapse";"github.com/rs/zerolog")
type Handler struct{synapse *synapse.Client;token string;logger zerolog.Logger}
type envelope struct{OK bool `json:"ok"`;Data any `json:"data"`;Error *apiError `json:"error"`}
type apiError struct{Code string `json:"code"`;Message string `json:"message"`}
func New(c *synapse.Client,t string,l zerolog.Logger)*Handler{return &Handler{synapse:c,token:t,logger:l}}
func write(w http.ResponseWriter,status int,p envelope){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_=json.NewEncoder(w).Encode(p)}
func(h *Handler)Health(w http.ResponseWriter,r *http.Request){write(w,200,envelope{OK:true,Data:map[string]string{"status":"ok","service":"mx-core"}})}
func(h *Handler)Authenticate(next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){v:=strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"),"Bearer "));if v==""||len(v)!=len(h.token)||subtle.ConstantTimeCompare([]byte(v),[]byte(h.token))!=1{write(w,401,envelope{OK:false,Error:&apiError{Code:"unauthorized",Message:"Valid service token required"}});return};next.ServeHTTP(w,r)})}
func(h *Handler)ServerVersion(w http.ResponseWriter,r *http.Request){ctx,cancel:=context.WithTimeout(r.Context(),10*time.Second);defer cancel();v,err:=h.synapse.ServerVersion(ctx);if err!=nil{h.logger.Error().Err(err).Msg("Synapse version request failed");write(w,502,envelope{OK:false,Error:&apiError{Code:"synapse_unavailable",Message:"Unable to retrieve Synapse server version"}});return};write(w,200,envelope{OK:true,Data:v})}
