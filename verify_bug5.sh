#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
PROJECT="gogo-verify-5-$$"
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
    environment:
      PIPEYARD_MAX_JOBS: "2"
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

submit() {
  local name="$1"
  curl -sS -o "/tmp/gogo-job-${name}-$$.json" -w "%{http_code}" \
    -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' \
    -d "{\"name\":\"${name}\",\"nodes\":[{\"id\":\"a\",\"kind\":\"validate\",\"param\":\"hello\"}],\"edges\":[]}"
}

c1="$(submit "q1-$$")"
c2="$(submit "q2-$$")"
c3="$(submit "q3-$$")"

if [[ "$c1" != "201" || "$c2" != "201" ]]; then
  echo "[EXPECT] first two submits 201"
  echo "[ACTUAL] ${c1} ${c2} ${c3}"
  exit 2
fi

if [[ "$c3" != "409" ]]; then
  echo "[EXPECT] third submit is 409 when max jobs is 2"
  echo "[ACTUAL] ${c1} ${c2} ${c3}"
  exit 1
fi
exit 0
