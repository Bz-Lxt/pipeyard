#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
PROJECT="gogo-verify-10-$$"
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

body_a='{"name":"alpha-'"$$"'","nodes":[{"id":"a","kind":"validate","param":"hello"}],"edges":[]}'
body_b='{"name":"bravo-'"$$"'","nodes":[{"id":"a","kind":"validate","param":"hello"}],"edges":[]}'

code1="$(curl -sS -o /tmp/gogo-job-a-$$.json -w "%{http_code}" -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' -d "$body_a")"
if [[ "$code1" != "201" ]]; then
  echo "[EXPECT] first submit 201"
  echo "[ACTUAL] HTTP ${code1} $(cat /tmp/gogo-job-a-$$.json)"
  exit 2
fi

code2="$(curl -sS -o /tmp/gogo-job-b-$$.json -w "%{http_code}" -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' -d "$body_b")"

echo "[EXPECT] second same-graph different-name submit returns 201"
echo "[ACTUAL] HTTP ${code2} $(cat /tmp/gogo-job-b-$$.json)"
if [[ "$code2" != "201" ]]; then
  exit 1
fi
exit 0
