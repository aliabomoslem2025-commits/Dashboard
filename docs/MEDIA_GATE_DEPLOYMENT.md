# Media Gate deployment notes

1. In root `.env`, set `SYNAPSE_BASE_URL` to the internal Synapse HTTP listener, not the public reverse-proxy hostname.
2. Configure the public reverse proxy to send `/_matrix/media/*` to `127.0.0.1:8082`. The Gate proxies unhandled media paths back to the internal Synapse URL.
3. Leave `/_matrix/client/v1/media/*` and all other Matrix API routes on the normal Synapse route. The authenticated media download namespace is passed through by the Gate unchanged.
4. Do not expose port 8082 directly to the internet. Terminate TLS and apply the normal proxy security controls at the existing public reverse proxy.
5. Disable request buffering for upload requests and allow the application to enforce the per-user byte limits. Synapse's own maximum upload size remains a second independent upper bound.
6. If activity message counts are needed, configure `SYNAPSE_DB_URL` with a PostgreSQL role granted SELECT only on the required Synapse tables. The Go pool also sets default transactions read-only.
7. Confirm that the Media Gate integration test is run with a disposable Redis instance via `REDIS_TEST_ADDR`; the test flushes that Redis database.

Example Nginx location:

```nginx
location ^~ /_matrix/media/ {
    proxy_pass http://127.0.0.1:8082;
    proxy_http_version 1.1;
    proxy_request_buffering off;
    proxy_buffering off;
    client_max_body_size 0;
    proxy_read_timeout 300s;
}
```

Keep the remaining `/_matrix/` routes pointed at Synapse. Do not point `SYNAPSE_BASE_URL` at the public hostname after enabling this route, or a proxy loop may occur.
