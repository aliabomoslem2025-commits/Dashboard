package synapse

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (c *Client) AdminRequest(ctx context.Context, method, path string, query url.Values, body []byte) (int, []byte, error) {
	endpoint := c.baseURL + "/" + strings.TrimLeft(path, "/")
	if len(query)>0 { endpoint += "?" + query.Encode() }
	retryable := method==http.MethodGet || method==http.MethodHead || method==http.MethodPut
	attempts := 1; if retryable { attempts = 3 }
	var lastStatus int; var lastBody []byte; var lastErr error
	for attempt:=0; attempt<attempts; attempt++ {
		resp, err := c.doAuthenticated(ctx, method, endpoint, body, func() string { if len(body)>0 { return "application/json" }; return "" }())
		if err != nil { lastErr = err
		} else {
			lastStatus = resp.StatusCode
			data, readErr := io.ReadAll(io.LimitReader(resp.Body, 8<<20)); _ = resp.Body.Close()
			if readErr != nil { lastErr = readErr
			} else {
				lastBody = data
				if resp.StatusCode>=200 && resp.StatusCode<300 { return resp.StatusCode, data, nil }
				lastErr = fmt.Errorf("Synapse Admin API returned HTTP %d", resp.StatusCode)
				if resp.StatusCode<500 || !retryable { return resp.StatusCode, data, lastErr }
			}
		}
		if attempt+1<attempts {
			delay:=time.Duration(150*(1<<attempt))*time.Millisecond
			timer:=time.NewTimer(delay)
			select { case <-ctx.Done(): timer.Stop(); return lastStatus,lastBody,ctx.Err(); case <-timer.C: }
		}
	}
	if lastErr==nil { lastErr=fmt.Errorf("Synapse Admin API request failed") }
	return lastStatus,lastBody,lastErr
}
