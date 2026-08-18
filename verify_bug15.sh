#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
PROJECT="gogo-verify-15-$$"
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
wait_health() {
  local ready=0
  for _ in $(seq 1 60); do
    if curl -fsS "$BASE/health" | grep -q '"status":"ok"'; then
      ready=1
      break
    fi
    sleep 1
  done
  [[ "$ready" == "1" ]]
}

if ! wait_health; then
  echo "[EXPECT] health 200"
  echo "[ACTUAL] health never became ready"
  exit 2
fi

job_body='{"name":"keep-cancel-'"$$"'","nodes":[{"id":"a","kind":"validate","param":"hello"}],"edges":[]}'
if ! curl -fsS -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' -d "$job_body" >/tmp/gogo-job-15-$$.json; then
  echo "[EXPECT] submit 201"
  echo "[ACTUAL] submit failed"
  exit 2
fi

id="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("id",""))' "/tmp/gogo-job-15-$$.json")"
if [[ -z "$id" ]]; then
  echo "[EXPECT] submit returned id"
  echo "[ACTUAL] empty id"
  exit 2
fi

if ! curl -fsS -X POST "$BASE/v1/jobs/$id/cancel" >/tmp/gogo-cancel-15-$$.json; then
  echo "[EXPECT] cancel 200"
  echo "[ACTUAL] cancel failed"
  exit 2
fi

if ! curl -fsS "$BASE/v1/jobs/$id" >/tmp/gogo-get1-15-$$.json; then
  echo "[EXPECT] GET after cancel 200"
  echo "[ACTUAL] GET failed"
  exit 2
fi

pre="$(python3 - <<'PY' "/tmp/gogo-get1-15-$$.json"
import json,sys
j=json.load(open(sys.argv[1]))
print(j.get("Status") or j.get("status") or "")
PY
)"
if [[ "$pre" != "canceled" ]]; then
  echo "[EXPECT] job canceled before restart"
  echo "[ACTUAL] status=${pre}"
  exit 2
fi

if ! docker compose -p "$PROJECT" -f "$COMPOSE_FILE" -f "$OVERRIDE" restart; then
  echo "[EXPECT] compose restart succeeded"
  echo "[ACTUAL] compose restart failed"
  exit 2
fi

if ! wait_health; then
  echo "[EXPECT] health 200 after restart"
  echo "[ACTUAL] health never became ready"
  exit 2
fi

if ! curl -fsS "$BASE/v1/jobs/$id" >/tmp/gogo-get2-15-$$.json; then
  echo "[EXPECT] GET after restart 200"
  echo "[ACTUAL] GET failed"
  exit 2
fi

status="$(python3 - <<'PY' "/tmp/gogo-get2-15-$$.json"
import json,sys
j=json.load(open(sys.argv[1]))
print(j.get("Status") or j.get("status") or "")
PY
)"

echo "[EXPECT] job status still canceled after compose restart"
echo "[ACTUAL] status=${status}"
if [[ "$status" == "pending" ]]; then
  exit 1
fi
if [[ "$status" == "canceled" ]]; then
  exit 0
fi
echo "[EXPECT] job status pending or canceled"
echo "[ACTUAL] unexpected status=${status}"
exit 2
