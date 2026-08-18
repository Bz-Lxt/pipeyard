#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
PROJECT="gogo-verify-12-$$"
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

job_body='{"name":"wal-'"$$"'","nodes":[{"id":"a","kind":"validate","param":"hello"}],"edges":[]}'
if ! curl -fsS -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' -d "$job_body" >/tmp/gogo-job-$$.json; then
  echo "[EXPECT] submit 201"
  echo "[ACTUAL] submit failed"
  exit 2
fi

if ! size_line="$(docker compose -p "$PROJECT" -f "$COMPOSE_FILE" -f "$OVERRIDE" exec -T yard wc -c /data/journal.wal)"; then
  echo "[EXPECT] wc journal succeeded"
  echo "[ACTUAL] wc failed"
  exit 2
fi
size="$(echo "$size_line" | awk '{print $1}')"

echo "[EXPECT] /data/journal.wal size > 0 after submit"
echo "[ACTUAL] wc=${size_line}"
if [[ "$size" == "0" ]]; then
  exit 1
fi
if ! [[ "$size" =~ ^[0-9]+$ ]]; then
  echo "[EXPECT] numeric byte count"
  echo "[ACTUAL] ${size_line}"
  exit 2
fi
exit 0
