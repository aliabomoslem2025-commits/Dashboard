# Matrix Synapse Dashboard

Laravel 11 + Inertia + Vue 3/Vuetify 3 UI and a Go `mx-core` service.

## First run

1. Copy the root `.env.example` to `.env`.
2. Replace every `change-this-*` value with unique secrets. Set `APP_KEY` by running `cd web && php artisan key:generate` in a local PHP environment, or generate a Laravel-compatible 32-byte base64 key.
3. Set `SYNAPSE_BASE_URL` to the reachable Synapse URL and `SYNAPSE_ADMIN_TOKEN` to a Synapse administrator token. The token is read only by `mx-core`.
4. Set `INITIAL_ADMIN_EMAIL` and a strong `INITIAL_ADMIN_PASSWORD` for the first dashboard administrator.
5. Run `docker compose up --build` from the repository root. The web container runs migrations and the initial role/admin seeder.
6. Open `http://localhost:8080/login`.

The initial admin credentials are created only when both initial-admin environment variables are set. Registration is not exposed.

## Services

- `web`: Laravel + Inertia + Vue/Vuetify
- `mx-core`: authenticated Go API; `/health` is public and `/api/v1/*` requires `CORE_SERVICE_TOKEN`
- `postgres`: dashboard data
- `redis`: reserved cache/session infrastructure

## Current implementation boundary

The current branch contains the application and security foundation, session login, role storage/gates, audit-log storage, Go health/version endpoints, and the initial server-version page. Synapse user/device/room administration endpoints, role-management UI, complete audit-log UI, and the full component/style-guide system are not yet implemented.
