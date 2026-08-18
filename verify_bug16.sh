#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
PROJECT="gogo-verify-16-$$"
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

job_body='{"name":"multi-tick-'"$$"'","nodes":[{"id":"a","kind":"validate","param":"one"},{"id":"b","kind":"validate","param":"two"},{"id":"c","kind":"validate","param":"three"}],"edges":[]}'
if ! curl -fsS -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' -d "$job_body" >/tmp/gogo-job-16-$$.json; then
  echo "[EXPECT] submit 201"
  echo "[ACTUAL] submit failed"
  exit 2
fi

id="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("id",""))' "/tmp/gogo-job-16-$$.json")"
if [[ -z "$id" ]]; then
  echo "[EXPECT] submit returned id"
  echo "[ACTUAL] empty id"
  exit 2
fi

if ! curl -fsS -X POST "$BASE/v1/tick?n=2" >/tmp/gogo-tick-16-$$.json; then
  echo "[EXPECT] POST /v1/tick?n=2 200"
  echo "[ACTUAL] tick failed"
  exit 2
fi

if ! curl -fsS "$BASE/v1/jobs/$id" >/tmp/gogo-get-16-$$.json; then
  echo "[EXPECT] GET job 200"
  echo "[ACTUAL] GET failed"
  exit 2
fi

eval "$(python3 - <<'PY' "/tmp/gogo-tick-16-$$.json" "/tmp/gogo-get-16-$$.json"
import json,sys
tick=json.load(open(sys.argv[1]))
job=json.load(open(sys.argv[2]))
ran=tick.get("ran", tick.get("Ran", -1))
status=job.get("Status") or job.get("status") or ""
nodes=job.get("Nodes") or job.get("nodes") or []
pending=0
for n in nodes:
    st=n.get("Status") or n.get("status") or ""
    if st in ("pending","ready",""):
        pending+=1
print(f"ran={ran}")
print(f"status={status}")
print(f"pending={pending}")
PY
)"

echo "[EXPECT] POST /v1/tick?n=2 ran=2 and nodes leave pending"
echo "[ACTUAL] ran=${ran} status=${status} pending=${pending}"
if [[ "${ran}" == "0" && "${pending}" == "3" ]]; then
  exit 1
fi
if [[ "${ran}" == "2" ]]; then
  exit 0
fi
if [[ "${pending}" != "3" ]]; then
  exit 0
fi
echo "[EXPECT] ran=2 or nodes advanced"
echo "[ACTUAL] ran=${ran} pending=${pending}"
exit 2
