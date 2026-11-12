#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; RESET='\033[0m'
info() { printf "${GREEN}[INFO]${RESET} %s\n" "$*"; }
warn() { printf "${YELLOW}[WARN]${RESET} %s\n" "$*"; }
die()  { printf "${RED}[ERROR]${RESET} %s\n" "$*" >&2; exit 1; }

for cmd in docker openssl awk; do
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
  awk -F= -v key="$1" '$1 == key { sub(/^[^=]*=/, ""); gsub(/^["\\047]|["\\047]$/, ""); print; exit }' .env
}

set_env() {
  local key="$1" value="$2"
  KEY="$key" VALUE="$value" awk '
    BEGIN { key = ENVIRON["KEY"]; value = ENVIRON["VALUE"]; found = 0 }
    index($0, key "=") == 1 {
      print key "=" value
      found = 1
      next
    }
    { print }
    END { if (!found) print key "=" value }
  ' .env > .env.tmp
  mv .env.tmp .env
}

