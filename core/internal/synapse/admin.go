package synapse

import (
 "bytes"
 "context"
 "fmt"
 "io"
 "net/http"
 "net/url"
 "strings"
)

func (c *Client) AdminRequest(ctx context.Context, method, path string, query url.Values, body []byte) (int, []byte, error) {
 endpoint := c.baseURL + "/" + strings.TrimLeft(path, "/")
 if len(query)>0 { endpoint += "?" + query.Encode() }
 req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
 if err != nil { return 0,nil,err }
 req.Header.Set("Authorization","Bearer "+c.token)
 if len(body)>0 { req.Header.Set("Content-Type","application/json") }
 resp,err:=c.httpClient.Do(req); if err!=nil{return 0,nil,err}; defer resp.Body.Close()
 data,err:=io.ReadAll(io.LimitReader(resp.Body,8<<20)); if err!=nil{return resp.StatusCode,nil,err}
 if resp.StatusCode<200||resp.StatusCode>=300 { return resp.StatusCode,data,fmt.Errorf("Synapse Admin API returned HTTP %d",resp.StatusCode) }
 return resp.StatusCode,data,nil
}
