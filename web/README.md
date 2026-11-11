# Matrix Synapse Dashboard

Laravel 11 + Inertia + Vue 3/Vuetify 3 UI and a Go `mx-core` service.

## First run

1. Copy the root `.env.example` to `.env`.
2. Generate a persistent Laravel key and put it in `.env`: `printf 'base64:%s\\n' "$(openssl rand -base64 32)"` (remove the line break from the generated value if your shell adds one).
3. Replace every `change-this-*` value with unique secrets. Set `SYNAPSE_BASE_URL` to the reachable Synapse URL and `SYNAPSE_ADMIN_TOKEN` to a Synapse administrator token.
4. Set `INITIAL_ADMIN_EMAIL` and a strong `INITIAL_ADMIN_PASSWORD` for the first dashboard administrator.
5. Run `docker compose up --build` from the repository root. The web container runs migrations and the initial role/admin seeder.
6. Open `http://localhost:8080/login`.

The Synapse administrator token is injected into `mx-core` only; it is deliberately not passed to the Laravel/PHP container or browser. Registration is not exposed.

## Services

- `web`: Laravel + Inertia + Vue/Vuetify
- `mx-core`: authenticated Go API; `/health` is public and `/api/v1/*` requires `CORE_SERVICE_TOKEN`
- `postgres`: dashboard data
- `redis`: available for future cache/session configuration

## Current implementation boundary

This branch currently contains the application and security foundation, session login, role storage/gates, audit-log persistence/service, Go health/version endpoints, and the initial server-version page. Synapse user/device/room administration endpoints, role-management UI, complete audit-log UI, and the full component/style-guide system remain to be implemented.
