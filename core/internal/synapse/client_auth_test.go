package synapse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestAutomaticAdminLoginAndRefreshAfterUnauthorized(t *testing.T) {
	var logins atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_matrix/client/v3/login":
			if r.Method != http.MethodPost { t.Fatalf("login method = %s", r.Method) }
			var payload struct {
				Identifier struct { Type string `json:"type"`; User string `json:"user"` } `json:"identifier"`
				Password string `json:"password"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil { t.Fatal(err) }
			if payload.Identifier.Type != "m.id.user" || payload.Identifier.User != "@ali:example.org" || payload.Password != "secret-password" {
				t.Fatalf("unexpected login payload: %+v", payload)
			}
			n := logins.Add(1)
			w.Header().Set("Content-Type", "application/json")
			if n == 1 { _, _ = w.Write([]byte(`{"access_token":"token-one","user_id":"@ali:example.org"}`))
			} else { _, _ = w.Write([]byte(`{"access_token":"token-two","user_id":"@ali:example.org"}`)) }
		case "/_synapse/admin/v1/server_version":
			if r.Header.Get("Authorization") == "Bearer token-one" { http.Error(w, "expired", http.StatusUnauthorized); return }
			if r.Header.Get("Authorization") != "Bearer token-two" { t.Errorf("unexpected authorization header: %q", r.Header.Get("Authorization")) }
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"server_version":"1.120.0","python_version":"3.12.0"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewWithCredentials(server.URL, "@ali:example.org", "secret-password", "", time.Second)
	if err := client.Authenticate(context.Background()); err != nil { t.Fatal(err) }
	version, err := client.ServerVersion(context.Background())
	if err != nil { t.Fatal(err) }
	if version.ServerVersion != "1.120.0" { t.Fatalf("unexpected version: %s", version.ServerVersion) }
	if got := logins.Load(); got != 2 { t.Fatalf("expected initial login and one refresh, got %d logins", got) }
}
