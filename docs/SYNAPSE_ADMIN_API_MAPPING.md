# Synapse Admin API mapping and verification notes

> Installed-version limitation: this repository has no reachable Synapse instance. Routes below were checked against current official Synapse Admin API documentation, but are not yet runtime-certified against the user's installed release. Read `/api/v1/server/version` in the dashboard before deploying. `locked` was introduced in Synapse v1.93.

## User administration

| Dashboard route | Synapse route |
|---|---|
| `GET /api/v1/users` | `GET /_synapse/admin/v2/users` |
| `GET/PUT /api/v1/users/{user_id}` | `GET/PUT /_synapse/admin/v2/users/{user_id}` |
| `POST /api/v1/users/{user_id}/reset-password` | `POST /_synapse/admin/v1/reset_password/{user_id}` |
| `POST /api/v1/users/{user_id}/deactivate` | `POST /_synapse/admin/v1/deactivate/{user_id}` |
| `PUT /api/v1/users/{user_id}/suspend` | `PUT /_synapse/admin/v1/suspend/{user_id}` |
| `POST/DELETE /api/v1/users/{user_id}/shadow-ban` | `POST/DELETE /_synapse/admin/v1/users/{user_id}/shadow_ban` |
| `GET/POST/DELETE /api/v1/users/{user_id}/ratelimit` | `GET/POST/DELETE /_synapse/admin/v1/users/{user_id}/override_ratelimit` |
| `GET /api/v1/users/{user_id}/devices` | `GET /_synapse/admin/v2/users/{user_id}/devices` |
| `DELETE /api/v1/users/{user_id}/devices/{device_id}` | `DELETE /_synapse/admin/v2/users/{user_id}/devices/{device_id}` |
| `GET /api/v1/users/{user_id}/joined-rooms` | `GET /_synapse/admin/v1/users/{user_id}/joined_rooms` |
| `GET /api/v1/users/{user_id}/whois` | `GET /_synapse/admin/v1/whois/{user_id}` |
| `GET/DELETE /api/v1/users/{user_id}/media` | `GET/DELETE /_synapse/admin/v1/users/{user_id}/media` |
| `GET /api/v1/media/stats` | `GET /_synapse/admin/v1/statistics/users/media` |

Password reset and administrator-management endpoints may be unavailable when Matrix Authentication Service (MAS) is enabled; verify the deployment authentication architecture.

## Room administration

| Dashboard route | Synapse route |
|---|---|
| `GET /api/v1/rooms` | `GET /_synapse/admin/v1/rooms` |
| `GET /api/v1/rooms/{room_id}` | `GET /_synapse/admin/v1/rooms/{room_id}` |
| `GET /api/v1/rooms/{room_id}/members` | `GET /_synapse/admin/v1/rooms/{room_id}/members` |
| `GET /api/v1/rooms/{room_id}/state` | `GET /_synapse/admin/v1/rooms/{room_id}/state` |
| `GET/PUT /api/v1/rooms/{room_id}/block` | `GET/PUT /_synapse/admin/v1/rooms/{room_id}/block` |
| `DELETE /api/v1/rooms/{room_id}` | `DELETE /_synapse/admin/v2/rooms/{room_id}` |
| `POST /api/v1/rooms/{room_id}/make-admin` | `POST /_synapse/admin/v1/rooms/{room_id}/make_room_admin` |
| `POST /api/v1/rooms/{room_id}/read-only` | Client state event `m.room.power_levels` |

`order_by=size` uses `GET /_synapse/admin/v1/statistics/database/rooms`, which returns only the ten largest estimated PostgreSQL rooms and is not a paginated full-room listing. The estimate excludes media storage.

## Media Gate

- User media usage uses `GET /_synapse/admin/v1/statistics/users/media`; per-user listing/deletion uses `GET/DELETE /_synapse/admin/v1/users/{user_id}/media`.
- MSC3916 moved authenticated downloads, thumbnails, config and preview_url to `/_matrix/client/v1/media/*`. It did not move upload: standard uploads remain `POST /_matrix/media/v3/upload`, and resumable uploads use `POST /_matrix/media/v1/create` followed by `PUT /_matrix/media/v3/upload/{serverName}/{mediaId}`.
- Media Gate intercepts `/_matrix/media/{r0,v1,v3}/upload` and `/_matrix/media/v1/create`. Other media routes, including `/_matrix/client/v1/media/*`, pass through unchanged. It does not assume a non-standard `/_matrix/client/v1/media/upload` endpoint.
- The external reverse proxy must route upload/create requests to `mx-core` port 8082. Publishing that port alone does not insert the gate into Matrix traffic.

## Official documentation checked

- https://element-hq.github.io/synapse/latest/admin_api/user_admin_api.html
- https://element-hq.github.io/synapse/latest/admin_api/rooms.html
- https://element-hq.github.io/synapse/latest/admin_api/media_admin_api.html
- https://element-hq.github.io/synapse/v1.163/admin_api/statistics.html
- https://github.com/matrix-org/matrix-spec-proposals/blob/main/proposals/3916-authentication-for-media.md
