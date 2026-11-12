# Synapse User Admin API mapping

This dashboard connects to an externally configured homeserver through `SYNAPSE_BASE_URL`; the repository does not pin or run a Synapse image in Docker Compose. Therefore the exact installed Synapse release is deployment-specific and must be read at runtime from:

- Dashboard core: `GET /api/v1/server/version`
- Synapse: `GET /_synapse/admin/v1/server_version`

The Laravel Users controller uses the returned `server_version` to gate the `locked` field/filter. Synapse introduced the `locked` user field and list filter in **v1.93.0**. Do not claim a specific installed version from the repository alone.

All upstream requests are made by `core/internal/synapse.Client.AdminRequest` with the server-side `SYNAPSE_ADMIN_TOKEN`. The token is not returned to the browser.

## User endpoint mapping

| Dashboard core route | Method | Synapse Admin API route | Purpose / notes |
|---|---:|---|---|
| `/api/v1/users` | GET | `/_synapse/admin/v2/users` | User list, Synapse `from`/ `next_token` pagination, `limit`, `name`/ `user_id` search, `admins`, `deactivated`, `guests`, `locked`, `order_by`, `dir`. |
| `/api/v1/users/{userID}` | GET | `/_synapse/admin/v2/users/{user_id}` | User details. |
| `/api/v1/users/{userID}` | PUT | `/_synapse/admin/v2/users/{user_id}` | Create/update: `password`, `displayname`, `admin`, `deactivated`, `locked`. |
| `/api/v1/users/{userID}/reset-password` | POST | `/_synapse/admin/v1/reset_password/{user_id}` | Reset password; `new_password`, optional `logout_devices`. Synapse documents this endpoint as unavailable when MSC3861 is enabled. |
| `/api/v1/users/{userID}/deactivate` | POST | `/_synapse/admin/v1/deactivate/{user_id}` | Deactivate account; optional `erase` (defaults to false). |
| `/api/v1/users/{userID}/shadow-ban` | POST | `/_synapse/admin/v1/users/{user_id}/shadow_ban` | Enable shadow ban. |
| `/api/v1/users/{userID}/shadow-ban` | DELETE | `/_synapse/admin/v1/users/{user_id}/shadow_ban` | Remove shadow ban. |
| `/api/v1/users/{userID}/ratelimit` | GET | `/_synapse/admin/v1/users/{user_id}/override_ratelimit` | Read override. |
| `/api/v1/users/{userID}/ratelimit` | POST | `/_synapse/admin/v1/users/{user_id}/override_ratelimit` | Set `messages_per_second` and `burst_count`. |
| `/api/v1/users/{userID}/ratelimit` | DELETE | `/_synapse/admin/v1/users/{user_id}/override_ratelimit` | Remove override. |
| `/api/v1/users/{userID}/devices` | GET | `/_synapse/admin/v2/users/{user_id}/devices` | List devices. |
| `/api/v1/users/{userID}/devices/{deviceID}` | DELETE | `/_synapse/admin/v2/users/{user_id}/devices/{device_id}` | Delete device and invalidate its access token. |
| `/api/v1/users/{userID}/joined-rooms` | GET | `/_synapse/admin/v1/users/{user_id}/joined_rooms` | List joined room IDs. |
| `/api/v1/users/{userID}/whois` | GET | `/_synapse/admin/v1/whois/{user_id}` | Session and IP connection details. Restrict access through the Whois Policy/Gate. |

## Version and compatibility notes

- `locked` is sent only when the reported Synapse version is at least `1.93.0`; older servers must not receive the field or list filter.
- The list API returns `next_token`; clients must pass that value as `from` for the next page instead of calculating offsets themselves.
- User list `name` searches user IDs/localparts and display names. Synapse documents `user_id` as a separate substring filter; when both are supplied, `name` takes precedence.
- Synapse's Admin API requires an administrator access token. Only the Go core holds that token; Laravel calls the authenticated core service.
- Reset-password availability depends on the homeserver authentication configuration (MSC3861 can disable the endpoint).
- Deactivation is destructive: it removes devices, tokens, and room memberships. The dashboard's normal action uses `erase=false`; GDPR erasure must be an explicit separate decision.

## Deployment verification

After deployment, record the exact response of `GET /api/v1/server/version` in the release/deployment report. This repository cannot know the installed Synapse version because `SYNAPSE_BASE_URL` points to an external service and there is no pinned Synapse image in `docker-compose.yml`.

## Official references

- Synapse User Admin API: https://github.com/matrix-org/synapse/blob/develop/docs/admin_api/user_admin_api.md
- Synapse Admin API overview: https://github.com/matrix-org/synapse/blob/develop/docs/admin_api/index.md
