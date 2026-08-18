#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT/docker-compose.yml}"
PROJECT="gogo-verify-1-$$"
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

job_body='{"name":"conc-'"$$"'","nodes":[{"id":"a","kind":"validate","param":"hello"}],"edges":[]}'
if ! curl -fsS -X POST "$BASE/v1/jobs" -H 'Content-Type: application/json' -d "$job_body" >/tmp/gogo-job-$$.json; then
  echo "[EXPECT] submit 201"
  echo "[ACTUAL] submit failed"
  exit 2
fi

dir="$(mktemp -d)"
for i in $(seq 1 16); do
  curl -sS -o "$dir/b$i" -w "%{http_code}" -X POST "$BASE/v1/tick" >"$dir/c$i" &
done
wait

bad=0
sample=""
for i in $(seq 1 16); do
  code="$(cat "$dir/c$i")"
  body="$(cat "$dir/b$i")"
  if [[ "$code" != "200" ]]; then
    bad=1
    sample="HTTP $code $body"
    break
  fi
  if echo "$body" | grep -Eq 'lease still held|quota'; then
    bad=1
    sample="$body"
    break
  fi
done
rm -rf "$dir"

echo "[EXPECT] all concurrent POST /v1/tick return 200 without lease/quota errors"
echo "[ACTUAL] ${sample:-all 200}"
if [[ "$bad" == "1" ]]; then
  exit 1
fi
exit 0
