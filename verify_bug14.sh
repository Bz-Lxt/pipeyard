#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
PROJECT="gogo-verify-14-$$"
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

job_body='{"name":"stale-view-'"$$"'","nodes":[{"id":"a","kind":"validate","param":"hello"},{"id":"b","kind":"map","param":"x:"}],"edges":[{"from":"a","to":"b"}]}'
if ! curl -fsS -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' -d "$job_body" >/tmp/gogo-job-14-$$.json; then
  echo "[EXPECT] submit 201"
  echo "[ACTUAL] submit failed"
  exit 2
fi

id="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("id",""))' "/tmp/gogo-job-14-$$.json")"
if [[ -z "$id" ]]; then
  echo "[EXPECT] submit returned id"
  echo "[ACTUAL] empty id"
  exit 2
fi

if ! curl -fsS "$BASE/v1/jobs/$id" >/tmp/gogo-get1-14-$$.json; then
  echo "[EXPECT] first GET 200"
  echo "[ACTUAL] first GET failed"
  exit 2
fi

if ! curl -fsS -X POST "$BASE/v1/tick" >/tmp/gogo-tick1-14-$$.json; then
  echo "[EXPECT] tick 200"
  echo "[ACTUAL] first tick failed"
  exit 2
fi
if ! curl -fsS -X POST "$BASE/v1/tick" >/tmp/gogo-tick2-14-$$.json; then
  echo "[EXPECT] tick 200"
  echo "[ACTUAL] second tick failed"
  exit 2
fi

if ! curl -fsS "$BASE/v1/jobs/$id" >/tmp/gogo-get2-14-$$.json; then
  echo "[EXPECT] second GET 200"
  echo "[ACTUAL] second GET failed"
  exit 2
fi

status="$(python3 - <<'PY' "/tmp/gogo-get2-14-$$.json"
import json,sys
j=json.load(open(sys.argv[1]))
print(j.get("Status") or j.get("status") or "")
PY
)"

echo "[EXPECT] job status succeeded after GET then two ticks"
echo "[ACTUAL] status=${status}"
if [[ "$status" == "pending" ]]; then
  exit 1
fi
if [[ "$status" == "succeeded" ]]; then
  exit 0
fi
echo "[EXPECT] job status pending or succeeded"
echo "[ACTUAL] unexpected status=${status}"
exit 2
