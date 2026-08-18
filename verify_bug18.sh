#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
PROJECT="gogo-verify-18-$$"
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

last_id=""
for i in $(seq 1 10); do
  job_body='{"name":"ev-'"$i"'-'"$$"'","nodes":[{"ID":"a","Kind":"validate","Param":"hello"}],"edges":[]}'
  if ! curl -fsS -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' -d "$job_body" >/tmp/gogo-job-$$.json; then
    echo "[EXPECT] submit 201"
    echo "[ACTUAL] submit $i failed"
    exit 2
  fi
  last_id="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("id",""))' /tmp/gogo-job-$$.json)"
  if [[ -z "$last_id" ]]; then
    echo "[EXPECT] submit returns id"
    echo "[ACTUAL] empty id on submit $i"
    exit 2
  fi
done

if ! curl -fsS "$BASE/v1/events" >/tmp/gogo-events-$$.json; then
  echo "[EXPECT] GET /v1/events 200"
  echo "[ACTUAL] events request failed"
  exit 2
fi

found="$(python3 -c 'import json,sys
last=sys.argv[1]
data=json.load(open(sys.argv[2]))
if not isinstance(data, list):
    print("0"); raise SystemExit
print("1" if any(isinstance(e, dict) and e.get("Job")==last for e in data) else "0")
' "$last_id" /tmp/gogo-events-$$.json)"

echo "[EXPECT] GET /v1/events contains the last submitted job"
echo "[ACTUAL] last_id=${last_id} found=${found}"
if [[ "$found" == "1" ]]; then
  exit 0
fi
exit 1
