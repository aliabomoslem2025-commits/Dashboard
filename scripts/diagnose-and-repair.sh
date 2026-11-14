#!/usr/bin/env bash
#
# Diagnose and repair the Laravel/Go Matrix Dashboard deployment.
# Run from the repository root: sudo bash scripts/diagnose-and-repair.sh
#
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

MIGRATION="2026_11_14_094720_add_is_active_to_users_table.php"
HOST_MIGRATION="web/database/migrations/$MIGRATION"
CONTAINER_MIGRATION="/var/www/html/database/migrations/$MIGRATION"
RESET_ADMIN_PASSWORD=0
SKIP_GO_TESTS=0
TEMP_REDIS=""
TEMP_REDIS_CREATED=0
PASS=0
WARN=0
FAIL=0

for arg in "$@"; do
  case "$arg" in
    --reset-admin-password) RESET_ADMIN_PASSWORD=1 ;;
    --skip-go-tests) SKIP_GO_TESTS=1 ;;
    -h|--help)
      cat <<'HELP'
Usage: sudo bash scripts/diagnose-and-repair.sh [options]

Options:
  --reset-admin-password  Explicitly reset the configured admin password
                          to INITIAL_ADMIN_PASSWORD from the Compose environment.
  --skip-go-tests         Skip Go unit/integration tests.
  -h, --help              Show this help.

The script never prints secrets, never touches the production Redis database,
and never deletes or recreates PostgreSQL data.
HELP
      exit 0
      ;;
    *) echo "Unknown option: $arg" >&2; exit 2 ;;
  esac
done

if [[ $EUID -ne 0 ]]; then
  echo "Run with sudo: sudo bash scripts/diagnose-and-repair.sh" >&2
  exit 2
fi

dc() { docker compose "$@"; }
ok() { echo "[OK] $*"; PASS=$((PASS + 1)); }
warn() { echo "[WARN] $*"; WARN=$((WARN + 1)); }
fail() { echo "[FAIL] $*"; FAIL=$((FAIL + 1)); }
section() { printf '\n========== %s ==========\n' "$*"; }

cleanup() {
  if [[ "$TEMP_REDIS_CREATED" -eq 1 && -n "$TEMP_REDIS" ]]; then
    docker rm -f "$TEMP_REDIS" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

section "Docker and service health"
if ! docker compose version >/dev/null 2>&1; then
  echo "Docker Compose is unavailable." >&2
  exit 1
fi
ok "Docker Compose is available."

for service in web mx-core postgres redis; do
  if dc ps --status running --services | grep -Fxq "$service"; then
    ok "$service is running"
  else
    warn "$service is not running"
  fi
done

WEB_ID="$(dc ps -q web)"
if [[ -z "$WEB_ID" ]]; then
  fail "Could not find the web container. Start the stack before running repairs."
  exit 1
fi

section "Migration source and PHP syntax"
mkdir -p web/database/migrations
if [[ ! -f "$HOST_MIGRATION" ]]; then
  cat > "$HOST_MIGRATION" <<'PHP'
<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        if (!Schema::hasColumn('users', 'is_active')) {
            Schema::table('users', function (Blueprint $table) {
                $table->boolean('is_active')->default(true);
            });
        }
    }

    public function down(): void
    {
        if (Schema::hasColumn('users', 'is_active')) {
            Schema::table('users', function (Blueprint $table) {
                $table->dropColumn('is_active');
            });
        }
    }
};
PHP
  ok "Created the missing additive migration on the host."
else
  if ! grep -q "boolean('is_active')" "$HOST_MIGRATION" ||
     ! grep -q "hasColumn('users', 'is_active')" "$HOST_MIGRATION"; then
    fail "Migration exists but does not match the expected safe migration. It was not overwritten."
    exit 1
  fi
  ok "Migration source exists."
fi

section "Copy migration into the running container"
if ! docker cp "$HOST_MIGRATION" "$WEB_ID:$CONTAINER_MIGRATION"; then
  fail "Could not copy the migration into the web container."
  exit 1
fi
if ! dc exec -T web php -l "database/migrations/$MIGRATION"; then
  fail "Migration syntax check inside the container failed."
  exit 1
fi
ok "Migration is present and syntactically valid on host and in container."

section "Database schema repair"
if ! dc exec -T web php artisan migrate \
  --path="database/migrations/$MIGRATION" --force --no-interaction; then
  fail "The targeted is_active migration failed. No other migration was requested."
  exit 1
fi

COLUMN_CHECK="$(dc exec -T web php artisan tinker --execute='
echo Illuminate\Support\Facades\Schema::hasColumn("users", "is_active")
    ? "IS_ACTIVE_COLUMN_PRESENT"
    : "IS_ACTIVE_COLUMN_MISSING";
' 2>&1)" || {
  echo "$COLUMN_CHECK"
  fail "Could not verify the users.is_active column."
  exit 1
}
if grep -q "IS_ACTIVE_COLUMN_PRESENT" <<<"$COLUMN_CHECK"; then
  ok "users.is_active exists."
else
  fail "users.is_active is still missing."
  exit 1
fi

section "Configured administrator account"
ADMIN_CODE="$(cat <<'PHP'
$email = env('INITIAL_ADMIN_EMAIL');
if (!$email) {
    echo 'ADMIN_EMAIL_NOT_CONFIGURED';
    return;
}
$user = App\Models\User::where('email', $email)->first();
if (!$user) {
    echo 'ADMIN_ACCOUNT_NOT_FOUND';
    return;
}
$role = Illuminate\Support\Facades\DB::table('roles')
    ->where('name', 'super_admin')->first();
if (!$role) {
    echo 'SUPER_ADMIN_ROLE_NOT_FOUND';
    return;
}
$user->is_active = true;
if (getenv('REPAIR_RESET_ADMIN_PASSWORD') === '1') {
    $password = env('INITIAL_ADMIN_PASSWORD');
    if (!$password || $password === 'change-this-initial-password') {
        echo 'ADMIN_PASSWORD_NOT_CONFIGURED';
        return;
    }
    $user->password = $password;
}
$user->save();
$user->roles()->syncWithoutDetaching([$role->id]);
echo 'ADMIN_REPAIR_OK';
PHP
)"

ADMIN_OUTPUT="$(dc exec -e REPAIR_RESET_ADMIN_PASSWORD="$RESET_ADMIN_PASSWORD" -T web php artisan tinker --execute="$ADMIN_CODE" 2>&1)" || {
  echo "$ADMIN_OUTPUT"
  fail "Administrator repair command failed."
  exit 1
}

if grep -q "ADMIN_EMAIL_NOT_CONFIGURED" <<<"$ADMIN_OUTPUT"; then
  fail "INITIAL_ADMIN_EMAIL is not configured in the web container."
elif grep -q "ADMIN_ACCOUNT_NOT_FOUND" <<<"$ADMIN_OUTPUT"; then
  fail "Configured administrator does not exist. Check INITIAL_ADMIN_EMAIL and run the documented seeder if this is a first setup."
elif grep -q "SUPER_ADMIN_ROLE_NOT_FOUND" <<<"$ADMIN_OUTPUT"; then
  fail "super_admin role is missing. Check database migrations and seeder."
elif grep -q "ADMIN_PASSWORD_NOT_CONFIGURED" <<<"$ADMIN_OUTPUT"; then
  fail "Password reset was requested, but INITIAL_ADMIN_PASSWORD is missing or still the placeholder."
elif grep -q "ADMIN_REPAIR_OK" <<<"$ADMIN_OUTPUT"; then
  ok "Configured administrator is active and has the super_admin role."
  if [[ "$RESET_ADMIN_PASSWORD" -eq 1 ]]; then
    ok "Administrator password was explicitly reset from the Compose environment."
  else
    ok "Existing administrator password was left unchanged."
  fi
else
  echo "$ADMIN_OUTPUT"
  fail "Unexpected administrator repair result."
  exit 1
fi

VERIFY_CODE='
$email = env("INITIAL_ADMIN_EMAIL");
$user = App\Models\User::where("email", $email)->with("roles")->first();
if (!$user) {
    echo "ADMIN_VERIFY_FAILED";
} else {
    echo "ADMIN_ACTIVE=" . ($user->is_active ? "yes" : "no") . PHP_EOL;
    echo "SUPER_ADMIN=" . ($user->roles->contains("name", "super_admin") ? "yes" : "no") . PHP_EOL;
}
'
VERIFY_OUTPUT="$(dc exec -T web php artisan tinker --execute="$VERIFY_CODE" 2>&1)" || {
  echo "$VERIFY_OUTPUT"
  fail "Administrator verification failed."
  exit 1
}
echo "$VERIFY_OUTPUT"
if grep -q "ADMIN_ACTIVE=yes" <<<"$VERIFY_OUTPUT" &&
   grep -q "SUPER_ADMIN=yes" <<<"$VERIFY_OUTPUT"; then
  ok "Administrator verification passed."
else
  fail "Administrator is not fully configured."
fi

section "Go tests with isolated Redis"
if [[ "$SKIP_GO_TESTS" -eq 1 ]]; then
  warn "Go tests were skipped by request."
else
  NETWORK="$(docker inspect "$WEB_ID" --format '{{range $name, $cfg := .NetworkSettings.Networks}}{{println $name}}{{end}}' | head -n1)"
  if [[ -z "$NETWORK" ]]; then
    warn "Could not determine the web container Docker network; Go tests skipped."
  else
    TEMP_REDIS="dashboard-mediagate-test-redis-$$"
    if docker run -d --rm --name "$TEMP_REDIS" --network "$NETWORK" redis:7-alpine >/dev/null; then
      TEMP_REDIS_CREATED=1
      READY=0
      for _ in $(seq 1 20); do
        if docker exec "$TEMP_REDIS" redis-cli ping 2>/dev/null | grep -q PONG; then
          READY=1
          break
        fi
        sleep 1
      done

      if [[ "$READY" -eq 1 ]]; then
        ok "Temporary isolated Redis is ready."
        if docker run --rm \
          --network "$NETWORK" \
          -e REDIS_TEST_ADDR="$TEMP_REDIS:6379" \
          -v "$ROOT/core:/src:ro" \
          -w /src golang:1.22-alpine \
          sh -c 'go mod download && go test -v ./internal/mediagate ./internal/synapse'; then
          ok "Go mediagate and synapse tests passed."
        else
          fail "One or more Go tests failed."
        fi
      else
        fail "Temporary Redis did not become ready."
      fi
    else
      fail "Could not start temporary isolated Redis."
    fi
  fi
fi

section "Login endpoint"
if command -v curl >/dev/null 2>&1; then
  HTTP_CODE="$(curl -sS -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/login || true)"
  echo "HTTP status: $HTTP_CODE"
  if [[ "$HTTP_CODE" == "200" ]]; then
    ok "Login endpoint returned HTTP 200."
  else
    warn "Login endpoint did not return HTTP 200 (check WEB_PORT and web logs)."
  fi
else
  warn "curl is not installed; endpoint check skipped."
fi

section "Summary"
echo "Passed:   $PASS"
echo "Warnings: $WARN"
echo "Failed:   $FAIL"
if [[ "$FAIL" -eq 0 ]]; then
  echo "RESULT: Completed without detected failures."
else
  echo "RESULT: Repair is incomplete; review the failures above."
  exit 1
fi
