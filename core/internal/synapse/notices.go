package synapse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// SendServerNotice sends an idempotent Synapse server notice. A stable
// transaction ID is required so a retry after a network timeout cannot
// create a duplicate Matrix event.
func (c *Client) SendServerNotice(ctx context.Context, txnID, userID, message string) (string, error) {
	if strings.TrimSpace(txnID) == "" || len(txnID) > 128 {
		return "", fmt.Errorf("invalid server notice transaction ID")
	}
	if !strings.HasPrefix(userID, "@") || !strings.Contains(userID, ":") {
		return "", fmt.Errorf("invalid Matrix user ID")
	}
	if strings.TrimSpace(message) == "" {
		return "", fmt.Errorf("server notice message cannot be empty")
	}

	payload := map[string]any{
		"user_id": userID,
		"content": map[string]string{
			"msgtype": "m.text",
			"body":    message,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode server notice: %w", err)
	}

	path := "/_synapse/admin/v1/send_server_notice/" + url.PathEscape(txnID)
	status, response, err := c.AdminRequest(ctx, http.MethodPut, path, nil, body)
	if err != nil {
		return "", fmt.Errorf("Synapse send_server_notice failed (HTTP %d): %w", status, err)
	}

	var result struct {
		EventID string `json:"event_id"`
	}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("decode Synapse server notice response: %w", err)
	}
	if result.EventID == "" {
		return "", fmt.Errorf("Synapse server notice response did not include event_id")
	}
	return result.EventID, nil
}
