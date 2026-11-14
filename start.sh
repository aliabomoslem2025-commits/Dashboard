#!/usr/bin/env bash
set -Eeuo pipefail
ROOT_DIR="$(cd -- "$(dirname -- "$0")" && pwd)"
cd "$ROOT_DIR"
GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; RESET='\033[0m'
info(){ printf "%b[INFO]%b %s\n" "$GREEN" "$RESET" "$*"; }
warn(){ printf "%b[WARN]%b %s\n" "$YELLOW" "$RESET" "$*"; }
die(){ printf "%b[ERROR]%b %s\n" "$RED" "$RESET" "$*" >&2; exit 1; }
for cmd in docker openssl awk grep sed; do command -v "$cmd" >/dev/null 2>&1 || die "Required command not found: $cmd"; done
docker compose version >/dev/null 2>&1 || die "Docker Compose plugin is required."
[[ -f docker-compose.yml ]] || die "docker-compose.yml not found."
[[ -f .env.example ]] || die ".env.example not found."
if [[ ! -f .env ]]; then cp .env.example .env; chmod 600 .env; info "Created .env from .env.example."; else warn "Existing .env detected; preserving unrelated values."; fi

get_env(){
 awk -F= -v key="$1" '$1==key {sub(/^[^=]*=/,""); q=sprintf("%c",39); if (substr($0,1,1)==q || substr($0,1,1)=="\\\"") $0=substr($0,2,length($0)-2); print; exit}' .env
}
set_env(){
 local key="$1" value="$2" encoded
 encoded="$(printf "%s" "$value" | sed "s/\\\\/\\\\\\\\/g; s/'/\\\\'/g")"
 KEY="$key" VALUE="$encoded" awk '
 BEGIN { key=ENVIRON["KEY"]; value=ENVIRON["VALUE"]; q=sprintf("%c",39); found=0 }
 index($0,key"=")==1 { print key"="q value q; found=1; next }
 { print }
 END { if(!found) print key"="q value q }
 ' .env > .env.tmp
 mv .env.tmp .env
}
random_secret(){ openssl rand -hex 32; }
prompt_value(){
 local label="$1" current="$2" answer
 if [[ -n "$current" ]]; then printf "%s [مقدار فعلی حفظ می‌شود؛ برای تغییر مقدار جدید وارد کنید]: " "$label"; else printf "%s: " "$label"; fi
 read -r answer
 if [[ -n "$answer" ]]; then REPLY="$answer"; else REPLY="$current"; fi
}
prompt_secret(){
 local label="$1" current="$2" answer
 if [[ -n "$current" ]]; then printf "%s [Enter برای حفظ مقدار فعلی، مقدار جدید برای تغییر]: " "$label"; else printf "%s: " "$label"; fi
 read -r -s answer; printf "\n"
 if [[ -n "$answer" ]]; then REPLY="$answer"; else REPLY="$current"; fi
}
app_key="$(get_env APP_KEY || true)"
db_password="$(get_env DB_PASSWORD || true)"
core_token="$(get_env CORE_SERVICE_TOKEN || true)"
if [[ -z "$app_key" ]]; then app_key="base64:$(openssl rand -base64 32 | tr -d '\n')"; set_env APP_KEY "$app_key"; fi
if [[ -z "$db_password" || "$db_password" == "change-this-db-password" ]]; then set_env DB_PASSWORD "$(random_secret)"; fi
if [[ -z "$core_token" || "$core_token" == "change-this-long-random-service-token" ]]; then set_env CORE_SERVICE_TOKEN "$(random_secret)"; fi

prompt_value "آدرس داخلی Synapse (مثلاً http://synapse:8008)" "$(get_env SYNAPSE_BASE_URL || true)"; synapse_url="$REPLY"
[[ -n "$synapse_url" ]] || die "SYNAPSE_BASE_URL is required."
set_env SYNAPSE_BASE_URL "$synapse_url"
prompt_value "دامنه Matrix (مثلاً example.com، بدون @ و بدون پورت)" "$(get_env SYNAPSE_SERVER_NAME || true)"; matrix_server="$REPLY"
[[ -n "$matrix_server" ]] || die "Matrix server name is required."
[[ "$matrix_server" != *@* && "$matrix_server" != *"/"* ]] || die "Enter only the Matrix server name, e.g. example.com."
set_env SYNAPSE_SERVER_NAME "$matrix_server"

printf "\n=== حساب سوپرادمین داشبورد ===\n"
prompt_value "نام کاربری ورود به داشبورد" "$(get_env INITIAL_ADMIN_USERNAME || true)"; dashboard_user="$REPLY"
[[ "$dashboard_user" =~ ^[A-Za-z0-9._-]{3,64}$ ]] || die "Dashboard username must be 3-64 characters: letters, numbers, dot, underscore or hyphen."
prompt_secret "رمز عبور سوپرادمین داشبورد" "$(get_env INITIAL_ADMIN_PASSWORD || true)"; dashboard_password="$REPLY"
[[ ${#dashboard_password} -ge 12 ]] || die "Dashboard admin password must contain at least 12 characters."
set_env INITIAL_ADMIN_USERNAME "$dashboard_user"
set_env INITIAL_ADMIN_PASSWORD "$dashboard_password"

printf "\n=== حساب ادمین Synapse ===\n"
prompt_value "نام کاربری کوتاه Synapse (مثلاً ali، نه @ali:server.com)" "$(get_env SYNAPSE_ADMIN_USER || true)"; synapse_localpart="$REPLY"
synapse_localpart="${synapse_localpart#@}"
synapse_localpart="${synapse_localpart%%:*}"
[[ "$synapse_localpart" =~ ^[A-Za-z0-9._=/-]{1,255}$ ]] || die "Invalid Synapse username/localpart."
synapse_mxid="@${synapse_localpart}:${matrix_server}"
prompt_secret "رمز عبور حساب ادمین Synapse (${synapse_mxid})" "$(get_env SYNAPSE_ADMIN_PASSWORD || true)"; synapse_password="$REPLY"
[[ -n "$synapse_password" ]] || die "Synapse admin password is required."
set_env SYNAPSE_ADMIN_USER "$synapse_mxid"
set_env SYNAPSE_ADMIN_PASSWORD "$synapse_password"
# Clear any old static token so the backend obtains a fresh token through Matrix login.
set_env SYNAPSE_ADMIN_TOKEN ""

chmod 600 .env
info "Validating Docker Compose configuration..."
docker compose config --quiet
info "Building dashboard images..."
docker compose build
info "Starting dashboard containers..."
docker compose up -d
info "Waiting for containers to start..."
for attempt in $(seq 1 60); do
 running="$(docker compose ps --status running --services)"
 if grep -qx web <<<"$running" && grep -qx mx-core <<<"$running" && grep -qx postgres <<<"$running" && grep -qx redis <<<"$running"; then break; fi
 if [[ "$attempt" -eq 60 ]]; then warn "Not all services reached running state. Recent logs:"; docker compose ps; docker compose logs --tail=80; exit 1; fi
 sleep 2
done
printf "\n"; docker compose ps; printf "\n"
info "Dashboard startup sequence completed."
printf "Dashboard: %s\n" "$(get_env APP_URL)"
printf "Dashboard login username: %s\n" "$dashboard_user"
printf "Synapse admin Matrix ID: %s\n" "$synapse_mxid"
printf "Synapse access token: obtained automatically inside mx-core (not displayed or sent to browser).\n"
warn "Keep .env private. Never commit it or share its secrets."
