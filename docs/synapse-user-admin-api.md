# Synapse User Administration API mapping

Reference: https://element-hq.github.io/synapse/latest/admin_api/user_admin_api.html

**Deployment caveat:** the actual Synapse instance is not connected to this GitHub workspace, so its installed version and enabled authentication mode have not been independently queried. At runtime, mx-core exposes GET /_synapse/admin/v1/server_version; Laravel enables the locked field only for Synapse >= 1.93. The paths below are verified against the current official documentation, not claimed as live-tested against a deployed server.

| Dashboard operation | Synapse Admin API path | Method | Notes |
|---|---|---:|---|
| List local users | /_synapse/admin/v2/users | GET | from, limit, name or user_id, guests, admins, deactivated, locked, order_by, dir. locked introduced in 1.93. |
| User details | /_synapse/admin/v2/users/{user_id} | GET | displayname, admin, deactivated, shadow_banned, locked, timestamps. |
| Create/update | /_synapse/admin/v2/users/{user_id} | PUT | password, displayname, admin, deactivated, locked. |
| Reset password | /_synapse/admin/v1/reset_password/{user_id} | POST | new_password, optional logout_devices; disabled when MAS integration is enabled. |
| Deactivate | /_synapse/admin/v1/deactivate/{user_id} | POST | optional erase; erasure is more destructive than deactivation. |
| Suspend/unsuspend | /_synapse/admin/v1/suspend/{user_id} | PUT | suspend true/false; distinct from locked. |
| Shadow ban/unban | /_synapse/admin/v1/users/{user_id}/shadow_ban | POST/DELETE | local users only. |
| Read rate-limit override | /_synapse/admin/v1/users/{user_id}/override_ratelimit | GET | empty object if no override. |
| Set rate-limit override | same | POST | messages_per_second, burst_count. |
| Delete rate-limit override | same | DELETE | removes custom override. |
| List devices | /_synapse/admin/v2/users/{user_id}/devices | GET | device ID, display name, last IP, user agent, last seen. |
| Delete device | /_synapse/admin/v2/users/{user_id}/devices/{device_id} | DELETE | invalidates associated access token. |
| Joined rooms | /_synapse/admin/v1/users/{user_id}/joined_rooms | GET | joined room IDs and total. |
| Whois | /_synapse/admin/v1/whois/{user_id} | GET | device/session connections with IP, user agent, last_seen. |

## Confirmed from documentation

- Admin endpoints require a server-admin access token in Authorization: Bearer.
- V2 user-list pagination uses next_token as the next request's from value; preserve it as an opaque token.
- name searches user ID localparts or display names; user_id searches full user IDs and is ignored when name is supplied.
- locked filter and field were introduced in Synapse 1.93.
- Device listing and deletion are V2 endpoints.
- Whois includes session connection IPs and timestamps.
- Password reset and administrator-management APIs may be unavailable when Matrix Authentication Service integration is enabled.
- Deactivation does not delete uploaded media or rate-limit overrides.

## Not claimed as live-verified

No real Synapse URL/token or installed version was supplied in this task. This report does not claim requests were made against the user's server, MAS is disabled, or every endpoint is enabled by the deployment configuration.
