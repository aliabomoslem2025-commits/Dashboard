package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
)

// RegisterNoticeRoutes exposes an idempotent server-notice operation. The
// transaction ID is forwarded to Synapse's PUT endpoint so retries cannot
// create duplicate notices.
func (h *Handler) RegisterNoticeRoutes(r chi.Router) {
	r.Put("/api/v1/notices/{txnID}", h.sendServerNotice)
}

func (h *Handler) sendServerNotice(w http.ResponseWriter, r *http.Request) {
	txnID := chi.URLParam(r, "txnID")
	if txnID == "" || len(txnID) > 128 {
		write(w, http.StatusBadRequest, envelope{OK: false, Error: &apiError{Code: "invalid_transaction_id", Message: "A transaction ID of 1-128 characters is required"}})
		return
	}
	for _, ch := range txnID {
		if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || strings.ContainsRune("._~-", ch)) {
			write(w, http.StatusBadRequest, envelope{OK: false, Error: &apiError{Code: "invalid_transaction_id", Message: "Transaction ID contains unsupported characters"}})
			return
		}
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 256<<10))
	if err != nil {
		write(w, http.StatusRequestEntityTooLarge, envelope{OK: false, Error: &apiError{Code: "request_too_large", Message: "Notice payload exceeds 256 KiB"}})
		return
	}
	var payload struct {
		UserID  string         `json:"user_id"`
		Content map[string]any `json:"content"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || !strings.HasPrefix(payload.UserID, "@") || !strings.Contains(payload.UserID, ":") || strings.TrimSpace(noticeString(payload.Content["body"])) == "" {
		write(w, http.StatusBadRequest, envelope{OK: false, Error: &apiError{Code: "invalid_notice", Message: "user_id and content.body are required"}})
		return
	}
	// Synapse's PUT form is idempotent for a given transaction ID.
	path := "/_synapse/admin/v1/send_server_notice/" + url.PathEscape(txnID)
	status, data, err := h.synapse.AdminRequest(r.Context(), http.MethodPut, path, nil, body)
	if err != nil {
		code, message := "synapse_error", "Synapse could not send the server notice"
		var remote map[string]any
		if json.Unmarshal(data, &remote) == nil {
			if v, ok := remote["errcode"].(string); ok && v != "" { code = v }
			if v, ok := remote["error"].(string); ok && v != "" { message = v }
		}
		if status < 400 || status > 599 { status = http.StatusBadGateway }
		h.logger.Warn().Err(err).Int("status", status).Msg("server notice delivery failed")
		write(w, status, envelope{OK: false, Error: &apiError{Code: code, Message: message}})
		return
	}
	var result any
	if len(data) > 0 { _ = json.Unmarshal(data, &result) }
	write(w, status, envelope{OK: true, Data: result})
}

func noticeString(value any) string {
	if value == nil { return "" }
	if text, ok := value.(string); ok { return text }
	return ""
}
