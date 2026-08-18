#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
PROJECT="gogo-verify-7-$$"
PORT="$(python3 -c 'import socket;s=socket.socket();s.bind(("127.0.0.1",0));print(s.getsockname()[1]);s.close()')"
OVERRIDE="$ROOT/.compose-verify-$$.yml"

cleanup() {
  docker compose -p "$PROJECT" -f "$COMPOSE_FILE" -f "$OVERRIDE" down -v >/dev/null 2>&1 || true
  rm -f "$OVERRIDE"
}
trap cleanup EXIT

cat > "$OVERRIDE" <<EOF
services:
  yard:
    ports:
      - "${PORT}:8080"
EOF

if ! docker compose -p "$PROJECT" -f "$COMPOSE_FILE" -f "$OVERRIDE" up -d --wait --build; then
  echo "[EXPECT] docker compose up succeeded"
  echo "[ACTUAL] compose up failed"
  exit 2
fi

BASE="http://127.0.0.1:${PORT}"
ok=0
for _ in $(seq 1 60); do
  if curl -fsS "$BASE/health" | grep -q '"status":"ok"'; then
    ok=1
    break
  fi
  sleep 1
done
if [[ "$ok" != "1" ]]; then
  echo "[EXPECT] health 200"
  echo "[ACTUAL] health never became ready"
  exit 2
fi

job_body='{"name":"keep-'"$$"'","nodes":[{"id":"a","kind":"validate","param":"hello"}],"edges":[]}'
if ! curl -fsS -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' -d "$job_body" >/tmp/gogo-job-$$.json; then
  echo "[EXPECT] submit 201"
  echo "[ACTUAL] submit failed"
  exit 2
fi

if ! docker compose -p "$PROJECT" -f "$COMPOSE_FILE" -f "$OVERRIDE" exec -T yard sh -c 'printf PYAR >> /data/journal.wal'; then
  echo "[EXPECT] append journal tail succeeded"
  echo "[ACTUAL] append failed"
  exit 2
fi

if ! docker compose -p "$PROJECT" -f "$COMPOSE_FILE" -f "$OVERRIDE" restart; then
  echo "[EXPECT] compose restart succeeded"
  echo "[ACTUAL] restart failed"
  exit 2
fi

ok=0
for _ in $(seq 1 60); do
  if curl -fsS "$BASE/health" | grep -q '"status":"ok"'; then
    ok=1
    break
  fi
  sleep 1
done
if [[ "$ok" != "1" ]]; then
  echo "[EXPECT] health 200 after restart"
  echo "[ACTUAL] health never became ready"
  exit 2
fi

if ! jobs="$(curl -fsS "$BASE/v1/jobs")"; then
  echo "[EXPECT] GET /v1/jobs 200"
  echo "[ACTUAL] list failed"
  exit 2
fi

count="$(python3 -c 'import json,sys; d=json.loads(sys.argv[1]); print(len(d) if isinstance(d,list) else -1)' "$jobs")"
has_ghost=0
if echo "$jobs" | grep -q '"Name":"ghost"'; then
  has_ghost=1
fi

echo "[EXPECT] after truncated journal restart, only the original job remains"
echo "[ACTUAL] jobs=${count} ghost=${has_ghost} body=${jobs}"
if [[ "$has_ghost" == "1" || "$count" -ge 2 ]]; then
  exit 1
fi
exit 0
