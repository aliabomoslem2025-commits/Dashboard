#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; RESET='\033[0m'
info() { printf "${GREEN}[INFO]${RESET} %s\n" "$*"; }
warn() { printf "${YELLOW}[WARN]${RESET} %s\n" "$*"; }
die()  { printf "${RED}[ERROR]${RESET} %s\n" "$*" >&2; exit 1; }

for cmd in docker openssl python3; do
  command -v "$cmd" >/dev/null 2>&1 || die "Required command not found: $cmd"
done
docker compose version >/dev/null 2>&1 || die "Docker Compose plugin is required."
[[ -f docker-compose.yml ]] || die "docker-compose.yml not found."
[[ -f .env.example ]] || die ".env.example not found."

if [[ ! -f .env ]]; then
  cp .env.example .env
  chmod 600 .env
  info "Created .env from .env.example."
else
  warn "Existing .env detected; preserving its values."
fi

get_env() {
  python3 - "$1" <<'PY'
import sys
from pathlib import Path
key = sys.argv[1]
for line in Path(".env").read_text().splitlines():
    if line.startswith(key + "="):
        print(line.split("=", 1)[1].strip().strip('"').strip("'"))
        break
PY
}

set_env() {
  KEY="$1" VALUE="$2" python3 <<'PY'
import os
from pathlib import Path
path = Path(".env")
key, value = os.environ["KEY"], os.environ["VALUE"]
lines = path.read_text().splitlines()
prefix = key + "="
for i, line in enumerate(lines):
    if line.startswith(prefix):
        lines[i] = prefix + value
        break
else:
    lines.append(prefix + value)
path.write_text("\n".join(lines) + "\n")
PY
}

random_secret() { openssl rand -hex 32; }

app_key="$(get_env APP_KEY || true)"
db_password="$(get_env DB_PASSWORD || true)"
core_token="$(get_env CORE_SERVICE_TOKEN || true)"
synapse_url="$(get_env SYNAPSE_BASE_URL || true)"
synapse_token="$(get_env SYNAPSE_ADMIN_TOKEN || true)"

if [[ -z "$app_key" ]]; then
  app_key="base64:$(openssl rand -base64 32 | tr -d '\n')"
  set_env APP_KEY "$app_key"
fi
if [[ -z "$db_password" || "$db_password" == "change-this-db-password" ]]; then
  set_env DB_PASSWORD "$(random_secret)"
fi
if [[ -z "$core_token" || "$core_token" == "change-this-long-random-service-token" ]]; then
  set_env CORE_SERVICE_TOKEN "$(random_secret)"
fi

if [[ -z "$synapse_url" || "$synapse_url" == "http://synapse:8008" ]]; then
  printf "Synapse URL reachable from Docker network (e.g. http://synapse:8008): "
  read -r synapse_url
fi
[[ -n "$synapse_url" ]] || die "SYNAPSE_BASE_URL is required."
set_env SYNAPSE_BASE_URL "$synapse_url"

if [[ -z "$synapse_token" || "$synapse_token" == "change-this-synapse-admin-token" ]]; then
  read -r -s -p "Synapse Admin access token: " synapse_token
  printf "\n"
fi
[[ -n "$synapse_token" ]] || die "SYNAPSE_ADMIN_TOKEN is required."
set_env SYNAPSE_ADMIN_TOKEN "$synapse_token"

chmod 600 .env

info "Validating Docker Compose configuration..."
docker compose config --quiet

info "Building dashboard images..."
docker compose build

info "Starting dashboard containers..."
docker compose up -d

info "Waiting for containers to start..."
for attempt in $(seq 1 60); do
  if docker compose ps --status running --services | grep -qx web &&
     docker compose ps --status running --services | grep -qx mx-core &&
     docker compose ps --status running --services | grep -qx postgres &&
     docker compose ps --status running --services | grep -qx redis; then
    break
  fi
  if [[ "$attempt" -eq 60 ]]; then
    warn "Not all services reached running state. Recent logs:"
    docker compose ps
    docker compose logs --tail=80
    exit 1
  fi
  sleep 2
done

printf "\n"
docker compose ps
printf "\n"
info "Dashboard startup sequence completed."
printf "Dashboard: %s\n" "$(get_env APP_URL)"
printf "\nUseful commands:\n"
printf "  docker compose logs -f\n"
printf "  docker compose ps\n"
printf "  docker compose restart\n"
printf "  docker compose down\n"
warn "Synapse must be reachable from the mx-core container. Do not use localhost unless Synapse runs inside that same container."
warn "Keep .env private. Do not commit it or share its secrets."
