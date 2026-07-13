#!/usr/bin/env bash
# FAIRRIDE — one-command backend bring-up (Sprint B0).
#
# Starts Postgres + Redis (Docker), applies migrations + dev seed data,
# builds every backend service the ride flow actually needs, and launches
# them all in the background with unique ports, logging each to its own
# file under backend/logs/. Idempotent — safe to re-run any time.
#
# Usage:
#   ./scripts/start-all.sh          # start everything
#   ./scripts/start-all.sh --stop   # stop everything this script started
#
# After it prints "READY", point the Rider/Driver APKs at:
#   http://<this-machine's-LAN-IP>:8080
#
# Logs:   backend/logs/<service>.log
# PIDs:   backend/logs/pids/<service>.pid
set -uo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
INFRA_DIR="$ROOT_DIR/infra/docker"
BIN_DIR="$BACKEND_DIR/bin"
LOG_DIR="$BACKEND_DIR/logs"
PID_DIR="$LOG_DIR/pids"

DB_URL="postgres://fairride:fairride_dev_secret@localhost:5432/fairride"
REDIS_ADDR="localhost:6379"
REDIS_PASSWORD="fairride_redis_secret"
# Must be >= 32 bytes — identity/infrastructure/jwt validates this and the
# gateway panics at startup otherwise.
JWT_ACCESS_SECRET="dev_access_secret_change_me_please_32b"
JWT_REFRESH_SECRET="dev_refresh_secret_change_me_please_32b"

# Services needed for one real ride, and the unique ports each gets so they
# never collide (every service defaults to :50051/:8080 otherwise).
# Format: name:grpc_port:http_port
SERVICES=(
  "identity:50051:8001"
  "user:50052:8002"
  "driver:50053:8003"
  "trip:50054:8004"
  "dispatch:50055:8005"
  "pricing:50056:8006"
  "review:50057:8007"
  "booking:50060:8010"
)

mkdir -p "$BIN_DIR" "$LOG_DIR" "$PID_DIR"

# ─── --stop ────────────────────────────────────────────────────────────────
if [[ "${1:-}" == "--stop" ]]; then
  echo "Stopping FAIRRIDE backend services..."
  for f in "$PID_DIR"/*.pid; do
    [[ -e "$f" ]] || continue
    pid="$(cat "$f")"
    name="$(basename "$f" .pid)"
    if kill "$pid" 2>/dev/null; then
      echo "  stopped $name (pid $pid)"
    fi
    rm -f "$f"
  done
  echo "Done. Infra (postgres/redis) left running — stop with:"
  echo "  cd infra/docker && docker compose down"
  exit 0
fi

fail() { echo "FATAL: $1" >&2; exit 1; }

# ─── 1. Docker infra ─────────────────────────────────────────────────────────
echo "== 1/7 Docker infra =="
docker info >/dev/null 2>&1 || fail "Docker is not running. Start Docker Desktop first."

cd "$INFRA_DIR"
[[ -f .env ]] || cp .env.example .env
# Kafka's pinned image (bitnami/kafka:3.7) is no longer resolvable on Docker
# Hub and nothing in the ride flow actually uses Kafka — only postgres/redis
# are started. See Sprint T2/B0 reports.
docker compose --env-file .env up -d postgres redis || fail "docker compose up failed"

echo "Waiting for Postgres to become healthy..."
for i in $(seq 1 30); do
  status="$(docker inspect --format '{{.State.Health.Status}}' fairride_postgres 2>/dev/null || echo "starting")"
  [[ "$status" == "healthy" ]] && break
  sleep 1
done
[[ "$status" == "healthy" ]] || fail "Postgres did not become healthy in time"
echo "Postgres: healthy. Redis: $(docker exec fairride_redis redis-cli -a "$REDIS_PASSWORD" ping 2>/dev/null)"

# ─── 2. Migrations + seed data (idempotent) ──────────────────────────────────
echo "== 2/7 Database migrations + seed =="
for f in "$BACKEND_DIR"/migrations/*.up.sql; do
  docker exec -i fairride_postgres psql -U fairride -d fairride < "$f" >/dev/null || fail "migration failed: $f"
done
docker exec -i fairride_postgres psql -U fairride -d fairride < "$ROOT_DIR/scripts/seed_dev.sql" >/dev/null \
  || fail "seed_dev.sql failed"
echo "Schema + dev seed data (Rider +84900000001 / Driver +84900000002) applied."

# ─── 3. Build services ────────────────────────────────────────────────────────
echo "== 3/7 Building services =="
cd "$BACKEND_DIR"
for entry in "${SERVICES[@]}" "gateway::"; do
  name="${entry%%:*}"
  echo "  building $name..."
  go build -o "$BIN_DIR/$name.exe" "./services/$name/cmd/server" || fail "build failed: $name"
done

# ─── 4. Launch services ──────────────────────────────────────────────────────
echo "== 4/7 Starting services =="

start_service() {
  local name="$1"; shift
  "$BIN_DIR/$name.exe" >"$LOG_DIR/$name.log" 2>&1 &
  echo $! > "$PID_DIR/$name.pid"
  echo "  $name started (pid $!)"
}

GRPC_ADDR=:50051 HTTP_ADDR=:8001 DATABASE_URL="$DB_URL" \
  start_service identity

GRPC_ADDR=:50052 HTTP_ADDR=:8002 DATABASE_URL="$DB_URL" \
  start_service user

GRPC_ADDR=:50053 HTTP_ADDR=:8003 DATABASE_URL="$DB_URL" \
  REDIS_ADDR="$REDIS_ADDR" REDIS_PASSWORD="$REDIS_PASSWORD" \
  start_service driver

GRPC_ADDR=:50054 HTTP_ADDR=:8004 DATABASE_URL="$DB_URL" \
  start_service trip

GRPC_ADDR=:50055 HTTP_ADDR=:8005 DATABASE_URL="$DB_URL" \
  REDIS_ADDR="$REDIS_ADDR" REDIS_PASSWORD="$REDIS_PASSWORD" \
  start_service dispatch

GRPC_ADDR=:50056 HTTP_ADDR=:8006 \
  start_service pricing

GRPC_ADDR=:50057 HTTP_ADDR=:8007 DATABASE_URL="$DB_URL" \
  start_service review

GRPC_ADDR=:50060 HTTP_ADDR=:8010 DATABASE_URL="$DB_URL" \
  TRIP_ADDR=:50054 DISPATCH_ADDR=:50055 PRICING_ADDR=:50056 \
  start_service booking

sleep 2

HTTP_ADDR=:8080 \
  JWT_ACCESS_SECRET="$JWT_ACCESS_SECRET" JWT_REFRESH_SECRET="$JWT_REFRESH_SECRET" \
  BOOKING_ADDR=:50060 DRIVER_ADDR=:50053 DISPATCH_ADDR=:50055 REVIEW_ADDR=:50057 TRIP_ADDR=:50054 \
  DB_URL="$DB_URL" \
  start_service gateway

# ─── 5. Wait + verify ────────────────────────────────────────────────────────
echo "== 5/7 Verifying =="
sleep 3
if curl -sf http://localhost:8080/health >/dev/null; then
  echo "Gateway /health: OK"
else
  echo "WARNING: gateway /health did not respond — check backend/logs/gateway.log"
fi

echo "== 6/7 Status =="
for entry in "${SERVICES[@]}" "gateway:-:8080"; do
  name="${entry%%:*}"
  pid_file="$PID_DIR/$name.pid"
  if [[ -f "$pid_file" ]] && kill -0 "$(cat "$pid_file")" 2>/dev/null; then
    echo "  [UP]   $name (pid $(cat "$pid_file")) — log: backend/logs/$name.log"
  else
    echo "  [DOWN] $name — check backend/logs/$name.log"
  fi
done

LAN_IP="$(ipconfig 2>/dev/null | grep -A1 'Ethernet adapter Ethernet' | grep 'IPv4' | awk -F': ' '{print $2}' | tr -d '\r' | head -1)"
echo ""
echo "READY. Gateway: http://localhost:8080  (phones: http://${LAN_IP:-<this-machine-LAN-IP>}:8080)"

# ─── 7. Cloudflare quick tunnel ───────────────────────────────────────────────
# So the Rider/Driver APKs can reach the gateway from anywhere (real device,
# emulator, different network) via a stable-for-this-session public HTTPS
# URL instead of localhost/10.0.2.2, which only resolve on this machine.
echo "== 7/7 Cloudflare tunnel =="
CLOUDFLARED="/c/Program Files (x86)/cloudflared/cloudflared.exe"
CF_CONFIG="$HOME/.cloudflared/config.yml"
if [[ -f "$CF_CONFIG" ]]; then
  # A stale named-tunnel config.yml (from an unrelated project) silently
  # overrides --url's ingress with its own catch-all http_status:404 rule,
  # so every request through the quick tunnel 404s even though the
  # connector reports "Registered tunnel connection". Move it aside rather
  # than delete it — it may belong to another project on this machine.
  mv "$CF_CONFIG" "$CF_CONFIG.bak"
  echo "Moved stale $CF_CONFIG aside (was overriding quick-tunnel routing) -> config.yml.bak"
fi
if [[ -x "$CLOUDFLARED" ]]; then
  rm -f "$LOG_DIR/cloudflared.log"
  "$CLOUDFLARED" tunnel --url http://localhost:8080 >"$LOG_DIR/cloudflared.log" 2>&1 &
  echo $! > "$PID_DIR/cloudflared.pid"
  TUNNEL_URL=""
  for i in $(seq 1 30); do
    TUNNEL_URL="$(grep -o 'https://[a-zA-Z0-9-]*\.trycloudflare\.com' "$LOG_DIR/cloudflared.log" | head -1)"
    [[ -n "$TUNNEL_URL" ]] && break
    sleep 1
  done
  if [[ -n "$TUNNEL_URL" ]]; then
    echo "Tunnel: $TUNNEL_URL"
    for f in "$ROOT_DIR/apps/rider/lib/core/config/app_config.dart" "$ROOT_DIR/apps/driver/lib/core/config/app_config.dart"; do
      sed -i -E "s#defaultValue: 'https://[a-zA-Z0-9-]*\.trycloudflare\.com'#defaultValue: '$TUNNEL_URL'#" "$f"
      echo "  updated ${f#$ROOT_DIR/}"
    done
  else
    echo "WARNING: tunnel URL did not appear in time — check backend/logs/cloudflared.log"
  fi
else
  echo "WARNING: cloudflared not found at $CLOUDFLARED — skipping tunnel, apps must use localhost"
fi

echo "Stop everything with: ./scripts/start-all.sh --stop"
