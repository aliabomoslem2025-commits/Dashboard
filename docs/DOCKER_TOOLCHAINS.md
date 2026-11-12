# Docker-only language toolchains

The project can be built, tested, and maintained without installing Go, Node.js,
PHP, or Composer on the host. Docker Engine and the Docker Compose plugin are
the only required development platform dependencies (plus ordinary shell
utilities used by scripts).

## Toolchain commands

Run these commands from the repository root:

```bash
# Go compiler, formatter, dependency manager, and tests
docker compose -f docker-compose.tools.yml run --rm go-toolchain version
docker compose -f docker-compose.tools.yml run --rm go-toolchain test ./...

# Node.js and npm
docker compose -f docker-compose.tools.yml run --rm node-toolchain install
docker compose -f docker-compose.tools.yml run --rm node-toolchain run build

# PHP 8.3 with the project's required extensions
docker compose -f docker-compose.tools.yml run --rm php-toolchain -v
docker compose -f docker-compose.tools.yml run --rm php-toolchain -l artisan

# Composer
docker compose -f docker-compose.tools.yml run --rm composer-toolchain validate --no-check-publish
docker compose -f docker-compose.tools.yml run --rm composer-toolchain install
```

The named Docker volumes retain Go module/build caches, npm dependencies/cache,
and Composer cache between runs. Source files remain on the host and are
mounted into short-lived toolchain containers.

## Production build

The regular `docker-compose.yml` already uses multi-stage Docker builds:
Go builds in the `golang` stage; Composer resolves PHP dependencies in its
own image; Node builds frontend assets in its own image; the final web runtime
contains PHP-FPM and Nginx but not the Node or Go compilers.

```bash
docker compose build
docker compose up -d
```

Do not install language toolchains on the host to build or test this project.
