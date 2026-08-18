#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
PROJECT="gogo-verify-9-$$"
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

job_body='{"name":"clock-'"$$"'","nodes":[{"id":"a","kind":"validate","param":"hello"}],"edges":[]}'
if ! curl -fsS -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' -d "$job_body" >/tmp/gogo-job-$$.json; then
  echo "[EXPECT] submit 201"
  echo "[ACTUAL] submit failed"
  exit 2
fi

job_id="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("id",""))' "/tmp/gogo-job-$$.json")"
if [[ -z "$job_id" ]]; then
  echo "[EXPECT] submit returned id"
  echo "[ACTUAL] empty id"
  exit 2
fi

if ! curl -fsS -X POST "$BASE/v1/tick" >/tmp/gogo-tick-$$.json; then
  echo "[EXPECT] tick 200"
  echo "[ACTUAL] tick failed"
  exit 2
fi

if ! job="$(curl -fsS "$BASE/v1/jobs/${job_id}")"; then
  echo "[EXPECT] GET job 200"
  echo "[ACTUAL] get job failed"
  exit 2
fi
updated="$(python3 -c 'import json,sys; d=json.loads(sys.argv[1]); print(d.get("UpdatedAt") or d.get("updated_at") or "")' "$job")"

echo "[EXPECT] updated_at is a current Beijing wall-clock string, not 1970-01-01 08:00:00"
echo "[ACTUAL] updated_at=${updated}"
if [[ "$updated" == "1970-01-01 08:00:00" ]]; then
  exit 1
fi
if [[ -z "$updated" ]]; then
  echo "[EXPECT] updated_at present"
  echo "[ACTUAL] empty"
  exit 2
fi
exit 0
