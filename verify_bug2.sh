#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
PROJECT="gogo-verify-2-$$"
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

job_body='{"name":"clip-'"$$"'","nodes":[{"id":"a","kind":"validate","param":"hello"},{"id":"b","kind":"validate","param":"hello"},{"id":"c","kind":"validate","param":"hello"}],"edges":[]}'
if ! curl -fsS -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' -d "$job_body" >/tmp/gogo-job-$$.json; then
  echo "[EXPECT] submit 201"
  echo "[ACTUAL] submit failed"
  exit 2
fi

id="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("id",""))' "/tmp/gogo-job-$$.json")"
if [[ -z "$id" ]]; then
  echo "[EXPECT] submit returned id"
  echo "[ACTUAL] empty id"
  exit 2
fi

tick1="$(curl -sS -o /tmp/gogo-tick1-$$.json -w "%{http_code}" -X POST "$BASE/v1/tick?n=2")"
tick2="$(curl -sS -o /tmp/gogo-tick2-$$.json -w "%{http_code}" -X POST "$BASE/v1/tick?n=2")"
got="$(curl -sS -o /tmp/gogo-get-$$.json -w "%{http_code}" "$BASE/v1/jobs/${id}")"
status="$(python3 -c 'import json,sys; j=json.load(open(sys.argv[1])); print(j.get("Status") or j.get("status") or "")' "/tmp/gogo-get-$$.json" 2>/dev/null || true)"
nodes="$(python3 -c 'import json,sys
j=json.load(open(sys.argv[1]))
ns=j.get("Nodes") or j.get("nodes") or []
print(",".join("%s:%s"%(n.get("ID") or n.get("id"), n.get("Status") or n.get("status")) for n in ns))' "/tmp/gogo-get-$$.json" 2>/dev/null || true)"

echo "[EXPECT] POST /v1/tick?n=2 twice return 200 and job succeeded with a,b,c scheduled"
echo "[ACTUAL] tick1=${tick1} tick2=${tick2} get=${got} status=${status} nodes=${nodes}"

if [[ "$tick1" != "200" || "$tick2" != "200" ]]; then
  exit 1
fi
if [[ "$status" != "succeeded" ]]; then
  exit 1
fi
if [[ "$nodes" != *"a:succeeded"* || "$nodes" != *"b:succeeded"* || "$nodes" != *"c:succeeded"* ]]; then
  exit 1
fi
exit 0
