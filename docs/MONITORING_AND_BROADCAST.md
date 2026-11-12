# Monitoring and Server Notices

## Monitoring pipeline

The Go collector starts with `mx-core` and samples every **15 seconds**. It writes one row per sample to the dashboard PostgreSQL table `system_metrics`. Samples older than **7 days** are removed by the collector's hourly retention task.

Collected fields:

- CPU utilization from Linux `/proc/stat` deltas.
- RAM used/total from `/proc/meminfo`.
- Disk used/total from `statfs("/")`.
- Synapse PostgreSQL database size, when `SYNAPSE_DB_URL` is configured.
- Synapse user and room counts through the Synapse Admin API.
- Online users, based on distinct `user_ips` entries seen in the last five minutes, when `SYNAPSE_DB_URL` is configured.
- Prometheus metrics, when `SYNAPSE_PROMETHEUS_URL` is configured.

The Laravel dashboard receives live samples through the authenticated Go SSE endpoint `GET /api/v1/metrics/stream`, proxied through `GET /metrics/stream`. A snapshot is also available at `GET /api/v1/metrics/current`.

### Important deployment notes

1. `/proc` and `statfs("/")` are read from inside the `mx-core` container. Disk figures therefore describe the filesystem visible at the container root, not necessarily the host's complete physical disks. For host-wide figures, deploy a host metrics exporter and use its Prometheus metrics.
2. `SYNAPSE_DB_URL` is optional for graceful startup, but database size, online-user count, and SQL-based activity classification require it. Use a dedicated PostgreSQL role with read-only permissions to Synapse's database. The Go collector sets `default_transaction_read_only=on` on that connection.
3. `SYNAPSE_PROMETHEUS_URL` should point to the Synapse Prometheus endpoint reachable from the `mx-core` container, typically the configured metrics listener's `/metrics` path. Enable Synapse metrics in `homeserver.yaml` and restrict that listener to trusted internal networks.
4. Resource alert thresholds are stored in `metric_thresholds` and can be changed from the Monitoring page. Only server managers can update them; changes are written to `audit_logs`.

## Server notices and broadcasts

### Single-recipient send

Laravel `POST /broadcasts/send-one` calls the authenticated Go endpoint `PUT /api/v1/notices/{transaction_id}`. Go forwards it to Synapse:

`PUT /_synapse/admin/v1/send_server_notice/{txnId}`

The transaction ID makes a repeated request idempotent. Successful sends are recorded in `audit_logs` with the target user and returned event ID.

### Bulk delivery

A bulk broadcast creates a row in `broadcasts` and one `broadcast_recipients` row per target. The Go worker claims queued broadcasts, sends one notice at a time, applies the configured inter-recipient throttle, retries failed attempts with exponential backoff, and stores the result/event ID/error/attempt count for each recipient. Recipient-level success/failure events are written to `audit_logs`. The Laravel detail page refreshes delivery progress while the broadcast is running.

The recipient selector supports all non-deactivated users, activity categories (`heavy_active`, `active`, `low_active`, `dormant`, `suspicious`), or an explicitly supplied list of Matrix IDs. The dedicated single-user flow sends directly and is audited.

### Required Synapse configuration

Server Notices must be enabled in Synapse `homeserver.yaml`, for example:

```yaml
server_notices:
  system_mxid_localpart: server
  system_mxid_display_name: "Server Notices"
  room_name: "Server Notices"
```

The configured server-notices user must be reserved for this purpose. Synapse only allows notices to be sent to local users. The homeserver must have Server Notices enabled; otherwise the API returns an error and the recipient is marked failed after retry attempts.

### Upstream API references

- Synapse Server Notices Admin API: https://matrix-org.github.io/synapse/develop/admin_api/server_notices.html
- Synapse Server Notices configuration: https://github.com/matrix-org/synapse/blob/develop/docs/server_notices.md
- Synapse Prometheus metrics: https://element-hq.github.io/synapse/latest/usage/administration/monitoring.html
