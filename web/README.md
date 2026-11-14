# Matrix Synapse Dashboard

Laravel 12 + Inertia + Vue 3/TypeScript + Vuetify 3 dashboard with a Go mx-core service for Synapse Admin API operations.

## First run

1. Copy the root .env.example to .env (or run `./start.sh`).
2. The setup script asks for the internal Synapse URL, Matrix server name, dashboard super-admin username/password, and a short Synapse admin localpart/password.
3. The script builds the Matrix ID as `@localpart:server-name`. The Matrix server name is not inferred from the Docker hostname.
4. The Go service logs in to Synapse's `/_matrix/client/v3/login` endpoint, keeps the admin access token only in process memory, and refreshes it after an Admin API 401. The token is never passed to Laravel or the browser.
5. Run `docker compose up --build -d` from the repository root if you are not using the setup script.
6. Open `http://localhost:8080/login` and sign in with the dashboard username/password.

The legacy `SYNAPSE_ADMIN_TOKEN` remains optional for existing deployments. Prefer a dedicated Synapse administrator service account with `SYNAPSE_ADMIN_USER` and `SYNAPSE_ADMIN_PASSWORD`. Keep .env private and never commit it.

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
- Session login with dashboard username, role/policy gates, audit log persistence, shared RTL layout and persistent light/dark theme.
