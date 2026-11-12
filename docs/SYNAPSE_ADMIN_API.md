# Synapse API compatibility and implementation map

This document records the official Synapse and Matrix API paths used by mx-core. The repository does not contain a reachable Synapse URL or a running homeserver version, so these paths were checked against the current Element-HQ Synapse Admin API documentation (Synapse v1.163 documentation available at review time) and Matrix Client-Server specification, **not yet against the user's installed server**. Once SYNAPSE_BASE_URL is configured, the dashboard's Server page reports the live Synapse version. Do not treat version-sensitive endpoints below as runtime-verified until that connection succeeds.

## User administration

| Dashboard operation | Synapse endpoint | Notes |
|---|---|---|
| List users | GET /_synapse/admin/v2/users | from, limit, name, order_by, dir, admins, deactivated, locked filters |
| User details | GET /_synapse/admin/v2/users/{user_id} | Includes admin, deactivated, shadow_banned, locked, suspended |
| Create/update | PUT /_synapse/admin/v2/users/{user_id} | Accepts displayname, admin, deactivated, locked, password |
| Reset password | POST /_synapse/admin/v1/reset_password/{user_id} | Disabled when Matrix Authentication Service (MAS) integration manages authentication |
| Deactivate | POST /_synapse/admin/v1/deactivate/{user_id} | Synapse removes access tokens/devices and room memberships; not reversible by simply toggling a flag |
| Suspend/unsuspend | PUT /_synapse/admin/v1/suspend/{user_id} | JSON body {"suspend":true|false} |
| Shadow-ban/unban | POST/DELETE /_synapse/admin/v1/users/{user_id}/shadow_ban | Shadow ban is a last-resort moderation action |
| Rate-limit override | GET/POST/DELETE /_synapse/admin/v1/users/{user_id}/override_ratelimit | Set messages_per_second and burst_count |
| List devices | GET /_synapse/admin/v2/users/{user_id}/devices | Includes device ID, IP, user-agent and last-seen timestamp |
| Delete one device | DELETE /_synapse/admin/v2/users/{user_id}/devices/{device_id} | Invalidates that device's access token |
| Joined rooms | GET /_synapse/admin/v1/users/{user_id}/joined_rooms | Returns joined room IDs and total |
| Whois | GET /_synapse/admin/v1/whois/{user_id} | Returns sessions and connection IP/user-agent data |
| List user media | GET /_synapse/admin/v1/users/{user_id}/media | Uses numeric next_token pagination |
| Global per-user media statistics | GET /_synapse/admin/v1/statistics/users/media | Paginated per-user media_count and media_length; used by the hourly activity collector |
| Delete user media | DELETE /_synapse/admin/v1/users/{user_id}/media | Deletes at most 100 files per request by default; mx-core repeats bounded batches |

### Version caveats

- locked in the v2 users API was introduced in Synapse 1.93. On older installations, omit that filter/field and do not assume the server accepts it.
- Reset-password and server-admin role changes are unavailable through these Synapse endpoints when MAS integration is enabled.
- The v2 users list returns creation_ts in milliseconds; user detail examples have historically shown seconds. The activity collector uses the v2 list's millisecond value.
- The list-users v2 deactivated=true parameter means “include deactivated accounts”, not “return only deactivated accounts”. The separate v3 users API has different filtering semantics.

## Room administration

| Dashboard operation | Synapse endpoint |
|---|---|
| List/search/sort/paginate | GET /_synapse/admin/v1/rooms |
| Room details | GET /_synapse/admin/v1/rooms/{room_id} |
| Members | GET /_synapse/admin/v1/rooms/{room_id}/members |
| State events | GET /_synapse/admin/v1/rooms/{room_id}/state |
| Block/unblock | PUT /_synapse/admin/v1/rooms/{room_id}/block with {"block":true|false} |
| Block status | GET /_synapse/admin/v1/rooms/{room_id}/block |
| Make room admin | POST /_synapse/admin/v1/rooms/{room_id}/make_room_admin |
| Delete/purge | DELETE /_synapse/admin/v2/rooms/{room_id} with a non-empty JSON body |

The v2 room-deletion endpoint is documented as experimental/asynchronous. It returns a delete_id; the dashboard polls GET /_synapse/admin/v2/rooms/{room_id}/delete_status every three seconds until the task completes or fails. The room read-only action first grants the configured Synapse admin account room power, then writes m.room.power_levels through the Matrix Client-Server state endpoint. It raises default and explicit event/state thresholds to 100. This action is not automatically reversible; restore the original power-level state manually if needed.

## Media API and Media Gate

The gate recognises:

- Legacy POST /_matrix/media/v1/upload.
- Current POST /_matrix/media/v3/upload.
- The authenticated POST /_matrix/media/v1/create flow, which returns an MXC URI.
- The follow-up PUT /_matrix/media/v3/upload/{serverName}/{mediaId} upload created by the v1/create flow.

The v1/create response's MXC media ID is associated with the authenticated Matrix user in Redis so the subsequent PUT can be quota-checked even when the PUT itself has no access token. Other media routes are reverse-proxied without quota transformation.

Quota errors use Matrix JSON error format with HTTP 413 and errcode M_TOO_LARGE. The gate currently requires a known Content-Length for a quota-enforced file upload; unknown/chunked upload lengths are rejected with HTTP 411 rather than bypassing the quota. User-media usage is based on Synapse's local media Admin API; it does not account for media stored in a separate external media repository.

## Activity, metrics and broadcasts

- SYNAPSE_DB_URL is optional and should use a dedicated read-only PostgreSQL account. mx-core also sets default_transaction_read_only=on on every connection. The collector reads Synapse events and user_ips tables for message counts, device last-seen and recent-online counts. Synapse database schema changes may require query updates. Metrics samples are retained for seven days; activity-history samples are retained for 90 days.
- Without SYNAPSE_DB_URL, message counts, online counts, and Synapse database size are unavailable; last-seen falls back to the per-user devices Admin API, which can be slow on large servers. The dashboard database size is deliberately not presented as Synapse's database size. The activity collector uses GET /_synapse/admin/v1/statistics/users/media for one paginated media-usage snapshot instead of issuing one request per user; the per-user media endpoint remains available for the quota detail page and Media Gate cache refresh.
- SYNAPSE_PROMETHEUS_URL is optional. If set, mx-core scrapes its exposition endpoint every 15 seconds and stores parsed samples.
- Broadcasts use POST /_synapse/admin/v1/send_server_notice. Synapse server notices must be enabled in homeserver.yaml.

## Live verification checklist

1. Configure SYNAPSE_BASE_URL and SYNAPSE_ADMIN_TOKEN in the root .env.
2. Start the stack and open the Server page; record the exact Synapse version shown.
3. Compare that version with the compatibility caveats above, especially the 1.93 locked requirement and the experimental v2 room-delete endpoint.
4. Exercise list/detail requests with a test account and test room before enabling destructive operations.
5. Verify media upload routing at the front reverse proxy: all supported upload/create paths must reach mx-core:8082, while other media requests may continue to reach Synapse directly or pass through the gate.

See `docs/nginx-media-gate.conf` for a ready-to-include Nginx snippet. It routes only media upload/create requests through the gate; downloads and all other paths remain on the existing Synapse upstream.
