# Matrix Synapse Dashboard

Laravel 12 + Inertia + Vue 3/TypeScript + Vuetify 3 dashboard with a Go mx-core service for Synapse Admin API operations.

## First run

1. Copy the root .env.example to .env.
2. Generate a persistent Laravel key with: APP_KEY="base64:$(openssl rand -base64 32)".
3. Replace every change-this-* value with unique secrets. Set SYNAPSE_BASE_URL to the actual Synapse Admin API base URL and SYNAPSE_ADMIN_TOKEN to a Synapse administrator access token.
4. Set INITIAL_ADMIN_EMAIL and a strong INITIAL_ADMIN_PASSWORD for the first dashboard administrator.
5. Run docker compose up --build from the repository root.
6. Open http://localhost:8080/login. The detected Synapse version is shown at /server.

The Synapse administrator token is injected into mx-core only; it is deliberately not passed to Laravel/PHP or the browser. Public registration is disabled. Never commit .env.

## Optional data sources

- SYNAPSE_DB_URL: use a dedicated read-only PostgreSQL account. mx-core additionally sets default_transaction_read_only=on. Enables message counts, recent-online counts, last-seen aggregation, and Synapse database size in the activity/metrics collectors.
- SYNAPSE_PROMETHEUS_URL: Synapse Prometheus exposition endpoint. Scraped every 15 seconds.
- Without SYNAPSE_DB_URL, message counts, online counts and Synapse database size are shown as unavailable; last-seen falls back to the devices Admin API. Activity media totals use Synapse's paginated global user-media statistics endpoint.

## Media Gate

mx-core listens on port 8082 for media uploads. It checks the Matrix access token, the user's configured quota and maximum file size, reserves usage in Redis, and returns Matrix M_TOO_LARGE on quota/size rejection. Other media routes are passed upstream unchanged.

To activate the gate, include docs/nginx-media-gate.conf in the existing Synapse Nginx server block. The example assumes Nginx runs on the Docker host and reaches the gate at 127.0.0.1:8082. If Nginx runs in the same Docker network, change the upstream to http://mx-core:8082.

## Implemented areas

- Synapse users: search/filter/pagination, create/update, admin/deactivated/locked flags, password reset, deactivate, shadow-ban/unban, rate-limit overrides, devices, joined rooms, whois and media usage/deletion.
- Synapse rooms: search/sort/filter/pagination, details, members, state, block/unblock, asynchronous delete/purge with status polling, make-room-admin and read-only power-level operation.
- Media & Quotas: per-user total/file quotas, usage display, quota donut/progress, upload reverse proxy, Redis usage cache and cache invalidation after administrative media deletion.
- Activity: hourly collection of last-seen, 30-day message counts and media usage; configurable heavy-active/active/low-active/dormant/suspicious categories.
- Metrics: 15-second CPU/RAM/disk/database/user/room/online/Prometheus samples, seven-day metric retention, SSE stream and configurable alerts.
- Broadcast: queued server notices to one user, all active users, or an activity category, with throttle, retry, recipient status, progress and confirmation.
- Session login, role/policy gates, audit log persistence, shared RTL layout and persistent light/dark theme.

## Verification and known boundaries

- Go unit and integration tests run in CI. The Media Gate integration test uses a real Redis service and a test Synapse upstream.
- CI checks PHP syntax, builds the Vue application, validates Docker Compose and builds the containers.
- docs/SYNAPSE_ADMIN_API.md records the Synapse routes checked against the current v1.163 documentation.
- The repository does not reveal which Synapse version is installed in your environment. Connect the actual server, then compare the version shown at /server with the API compatibility notes before enabling version-sensitive operations.
- Role-management screens and a complete interactive component/style guide remain outside this implementation stage.

## Media Gate reverse-proxy wiring

The Media Gate listens on host loopback port 8082. It is **not** automatically inserted into your existing Synapse reverse proxy. Route the Matrix media namespace to it at the public reverse proxy, for example:

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

Keep the rest of `/_matrix/` routed to Synapse as before. Set `SYNAPSE_BASE_URL` to Synapse's **internal direct listener** (for example `http://synapse:8008`), not the public hostname whose media paths now route back to the gate; otherwise the proxy can loop. Do not expose port 8082 publicly. Use a dedicated read-only PostgreSQL account for `SYNAPSE_DB_URL`.
