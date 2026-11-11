package httpapi
import("context";"net/http";"time";"github.com/go-chi/chi/v5")
func(h *Handler)mediaUsage(w http.ResponseWriter,r *http.Request){ctx,cancel:=context.WithTimeout(r.Context(),30*time.Second);defer cancel();userID:=chi.URLParam(r,"userID");bytes,err:=h.synapse.MediaBytesSince(ctx,userID,time.Unix(0,0));if err!=nil{write(w,http.StatusBadGateway,envelope{OK:false,Error:&apiError{Code:"media_usage_unavailable",Message:"Unable to calculate user media usage"}});return};write(w,http.StatusOK,envelope{OK:true,Data:map[string]any{"user_id":userID,"total_bytes":bytes}})}
