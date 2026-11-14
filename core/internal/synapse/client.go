package synapse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Client struct {
	baseURL string
	token string
	adminUser string
	adminPassword string
	httpClient *http.Client
	tokenMu sync.RWMutex
	loginMu sync.Mutex
}

type ServerVersion struct {
	ServerVersion string `json:"server_version"`
	PythonVersion string `json:"python_version"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	UserID string `json:"user_id"`
}

func New(baseURL, token string, timeout time.Duration) *Client {
	return newClient(baseURL, token, "", "", timeout)
}

// NewWithCredentials obtains a Synapse admin access token using a dedicated
// Matrix account. The token remains in memory and is never returned to callers.
func NewWithCredentials(baseURL, username, password, fallbackToken string, timeout time.Duration) *Client {
	return newClient(baseURL, fallbackToken, username, password, timeout)
}

func newClient(baseURL, token, username, password string, timeout time.Duration) *Client {
	if timeout <= 0 { timeout = 8*time.Second }
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), token: strings.TrimSpace(token),
		adminUser: strings.TrimSpace(username), adminPassword: password,
		httpClient: &http.Client{Timeout: timeout}}
}

func (c *Client) currentToken() string {
	c.tokenMu.RLock(); defer c.tokenMu.RUnlock()
	return c.token
}

func (c *Client) login(ctx context.Context) error {
	if c.adminUser == "" || c.adminPassword == "" {
		return fmt.Errorf("Synapse admin credentials are not configured")
	}
	payload := map[string]any{
		"type": "m.login.password",
		"identifier": map[string]string{"type":"m.id.user", "user":c.adminUser},
		"password": c.adminPassword,
	}
	body, err := json.Marshal(payload); if err != nil { return err }
	endpoint := c.baseURL + "/_matrix/client/v3/login"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body)); if err != nil { return err }
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req); if err != nil { return fmt.Errorf("Synapse admin login request failed: %w", err) }
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)); if err != nil { return err }
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { return fmt.Errorf("Synapse admin login failed with HTTP %d", resp.StatusCode) }
	var result loginResponse
	if err := json.Unmarshal(data, &result); err != nil { return fmt.Errorf("invalid Synapse login response: %w", err) }
	if strings.TrimSpace(result.AccessToken) == "" { return fmt.Errorf("Synapse login response did not contain an access token") }
	c.tokenMu.Lock(); c.token = result.AccessToken; c.tokenMu.Unlock()
	return nil
}

func (c *Client) tokenForRequest(ctx context.Context) (string, error) {
	if token := c.currentToken(); token != "" { return token, nil }
	c.loginMu.Lock(); defer c.loginMu.Unlock()
	if token := c.currentToken(); token != "" { return token, nil }
	if err := c.login(ctx); err != nil { return "", err }
	return c.currentToken(), nil
}

func (c *Client) refreshAfterUnauthorized(ctx context.Context, rejected string) error {
	if c.adminUser == "" || c.adminPassword == "" { return fmt.Errorf("Synapse returned HTTP 401 and automatic login is not configured") }
	c.loginMu.Lock(); defer c.loginMu.Unlock()
	// Another request may already have refreshed the rejected token.
	if current := c.currentToken(); current != "" && current != rejected { return nil }
	c.tokenMu.Lock(); c.token = ""; c.tokenMu.Unlock()
	return c.login(ctx)
}

func (c *Client) doAuthenticated(ctx context.Context, method, endpoint string, body []byte, contentType string) (*http.Response, error) {
	token, err := c.tokenForRequest(ctx); if err != nil { return nil, err }
	send := func(auth string) (*http.Response, error) {
		req, e := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body)); if e != nil { return nil, e }
		req.Header.Set("Authorization", "Bearer "+auth)
		if contentType != "" { req.Header.Set("Content-Type", contentType) }
		return c.httpClient.Do(req)
	}
	resp, err := send(token); if err != nil { return nil, err }
	if resp.StatusCode != http.StatusUnauthorized { return resp, nil }
	_ = resp.Body.Close()
	if err := c.refreshAfterUnauthorized(ctx, token); err != nil { return nil, err }
	return send(c.currentToken())
}

func (c *Client) ServerVersion(ctx context.Context) (ServerVersion, error) {
	var out ServerVersion
	endpoint, err := url.JoinPath(c.baseURL, "/_synapse/admin/v1/server_version"); if err != nil { return out, err }
	var last error
	for attempt:=0; attempt<2; attempt++ {
		resp, e := c.doAuthenticated(ctx, http.MethodGet, endpoint, nil, "")
		if e == nil {
			body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20)); resp.Body.Close()
			if readErr != nil { last = readErr
			} else if resp.StatusCode >= 500 && attempt == 0 { last = fmt.Errorf("Synapse HTTP %d", resp.StatusCode)
			} else if resp.StatusCode < 200 || resp.StatusCode >= 300 { return out, fmt.Errorf("Synapse admin API HTTP %d", resp.StatusCode)
			} else if e = json.Unmarshal(body, &out); e != nil { return out, fmt.Errorf("invalid Synapse response: %w", e)
			} else { return out, nil }
		} else { last = e }
		if attempt == 0 { select { case <-ctx.Done(): return out, ctx.Err(); case <-time.After(150*time.Millisecond): } }
	}
	return out, fmt.Errorf("Synapse request failed: %w", last)
}
