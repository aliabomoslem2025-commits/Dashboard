package synapse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSendServerNoticeUsesIdempotentAdminEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/_synapse/admin/v1/send_server_notice/broadcast-7-recipient-9" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-admin-token" {
			t.Fatalf("unexpected authorization header")
		}
		var payload struct {
			UserID string `json:"user_id"`
			Content struct {
				MsgType string `json:"msgtype"`
				Body    string `json:"body"`
			} `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.UserID != "@alice:example.org" || payload.Content.MsgType != "m.text" || payload.Content.Body != "Maintenance notice" {
			t.Fatalf("unexpected notice payload: %#v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"event_id":"$notice-event"}`))
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, token: "test-admin-token", httpClient: &http.Client{Timeout: time.Second}}
	eventID, err := client.SendServerNotice(context.Background(), "broadcast-7-recipient-9", "@alice:example.org", "Maintenance notice")
	if err != nil {
		t.Fatalf("SendServerNotice returned error: %v", err)
	}
	if eventID != "$notice-event" {
		t.Fatalf("event ID = %q, want %q", eventID, "$notice-event")
	}
}

func TestSendServerNoticeRejectsInvalidInput(t *testing.T) {
	client := &Client{}
	if _, err := client.SendServerNotice(context.Background(), "", "@alice:example.org", "hello"); err == nil {
		t.Fatal("expected transaction ID validation error")
	}
	if _, err := client.SendServerNotice(context.Background(), "txn", "alice", "hello"); err == nil {
		t.Fatal("expected Matrix user ID validation error")
	}
	if _, err := client.SendServerNotice(context.Background(), "txn", "@alice:example.org", "  "); err == nil {
		t.Fatal("expected empty message validation error")
	}
}
