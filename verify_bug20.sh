#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
PROJECT="gogo-verify-20-$$"
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

job_body='{"name":"split-'"$$"'","nodes":[{"ID":"a","Kind":"validate","Param":"bravo,hello"},{"ID":"b","Kind":"fanout","Param":","}],"edges":[{"From":"a","To":"b"}]}'
if ! curl -fsS -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' -d "$job_body" >/tmp/gogo-job-$$.json; then
  echo "[EXPECT] submit 201"
  echo "[ACTUAL] submit failed"
  exit 2
fi
job_id="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("id",""))' /tmp/gogo-job-$$.json)"
if [[ -z "$job_id" ]]; then
  echo "[EXPECT] submit returns id"
  echo "[ACTUAL] empty id"
  exit 2
fi

for i in 1 2; do
  if ! curl -fsS -X POST "$BASE/v1/tick" >/tmp/gogo-tick-$$.json; then
    echo "[EXPECT] tick $i 200"
    echo "[ACTUAL] tick $i failed"
    exit 2
  fi
done

if ! curl -fsS "$BASE/v1/jobs/${job_id}" >/tmp/gogo-get-$$.json; then
  echo "[EXPECT] GET job 200"
  echo "[ACTUAL] get failed"
  exit 2
fi

items="$(python3 -c 'import json,sys
j=json.load(open(sys.argv[1]))
items=[]
for n in j.get("Nodes") or []:
    if n.get("ID")=="b":
        art=n.get("Artifact") or {}
        items=art.get("Items") or []
        break
print(",".join(items))
' /tmp/gogo-get-$$.json)"

echo "[EXPECT] fanout items contain bravo and hello"
echo "[ACTUAL] items=${items}"
if [[ "$items" == *bravo* && "$items" == *hello* ]]; then
  exit 0
fi
exit 1
