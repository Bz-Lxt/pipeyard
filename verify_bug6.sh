#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
PROJECT="gogo-verify-6-$$"
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

job1='{"name":"cp1-'"$$"'","nodes":[{"id":"a","kind":"validate","param":"hello"}],"edges":[]}'
c1="$(curl -sS -o /tmp/gogo-job1-$$.json -w "%{http_code}" -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' -d "$job1")"
if [[ "$c1" != "201" ]]; then
  echo "[EXPECT] first submit 201"
  echo "[ACTUAL] ${c1}"
  exit 2
fi

cp="$(curl -sS -o /tmp/gogo-cp-$$.json -w "%{http_code}" -X POST "$BASE/v1/checkpoint")"
if [[ "$cp" != "200" ]]; then
  echo "[EXPECT] checkpoint 200"
  echo "[ACTUAL] ${cp}"
  exit 2
fi

job2='{"name":"cp2-'"$$"'","nodes":[{"id":"a","kind":"validate","param":"hello"}],"edges":[]}'
c2="$(curl -sS -o /tmp/gogo-job2-$$.json -w "%{http_code}" -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' -d "$job2")"

echo "[EXPECT] submit after checkpoint is 201"
echo "[ACTUAL] first=${c1} checkpoint=${cp} second=${c2}"

if [[ "$c2" != "201" ]]; then
  exit 1
fi
exit 0
