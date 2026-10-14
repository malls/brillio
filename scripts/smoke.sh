#!/bin/sh
# make smoke (docs/SPEC.md section 10): bring the Compose stack up, wait for both
# services to turn healthy, assert the API contract (SPEC 5.1-5.5) directly and
# through the Vite proxy, then tear the stack down (docker compose down -v) on
# pass, on fail and on interrupt. Exits 0 only when every assertion passes.
#
# POSIX sh. Needs docker (with compose) and curl; no jq.
set -eu

cd "$(dirname "$0")/.."

# Run native even when the shell exports DOCKER_DEFAULT_PLATFORM (see compose.yaml).
unset DOCKER_DEFAULT_PLATFORM

API=http://localhost:8080
WEB=http://localhost:5173

die() {
  printf 'smoke: %s\n' "$*" >&2
  exit 1
}

# diagnose prints the stack state and recent logs, before teardown erases them.
diagnose() {
  docker compose ps -a || true
  docker compose logs --no-color --tail=60 || true
}

# --- 1. Preflight. No trap yet, so a stack the user already runs is never torn down.
command -v curl >/dev/null 2>&1 || die 'curl is not installed'
command -v docker >/dev/null 2>&1 || die 'docker is not installed'
docker info >/dev/null 2>&1 || die 'docker is not running (docker info failed)'
docker compose version >/dev/null 2>&1 || die 'docker compose is not available'

existing=$(docker compose ps -a -q) || die 'docker compose ps failed'
[ -z "$existing" ] || die 'a brillio stack already exists; run make down first'

# Every curl call starts with -q (skip ~/.curlrc) and passes --noproxy '*', so a
# proxy in the environment or the user's curl config never sees localhost.
# curl exit 7 is "connection refused": nothing listens there, so the port is free.
for p in 8080 5173; do
  rc=0
  curl -q --noproxy '*' -s -o /dev/null --max-time 2 "http://localhost:$p/" || rc=$?
  [ "$rc" -eq 7 ] || die "port $p is in use"
done

# --- 2. Teardown trap: runs on pass, on fail and on interrupt.
tmp=$(mktemp -d)

cleanup() {
  status=$?
  trap - EXIT
  trap '' INT TERM HUP
  echo 'teardown: docker compose down -v'
  if ! docker compose down -v; then
    echo 'FAIL  teardown'
    [ "$status" -ne 0 ] || status=1
  fi
  rm -rf "$tmp"
  exit "$status"
}

trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

# --- 3. Bring-up.
echo 'smoke: docker compose up -d --build --wait'
if ! docker compose up -d --build --wait --wait-timeout 300 --quiet-pull; then
  echo 'FAIL  stack did not become healthy'
  diagnose
  exit 1
fi

# wait_for URL: the host-side check that the published port answers.
wait_for() {
  i=0
  until curl -q --noproxy '*' -fs -o /dev/null --max-time 2 "$1"; do
    i=$((i + 1))
    if [ "$i" -ge 30 ]; then
      printf 'FAIL  %s did not answer from the host\n' "$1"
      diagnose
      exit 1
    fi
    sleep 1
  done
}

wait_for "$API/healthz"
wait_for "$WEB/"

# --- 4. Assertions. Every one runs; none relies on errexit.
pass=0
fail=0
code=000

# fetch URL: the body lands in $tmp/body (always present, maybe empty), the
# status in $code (000 when the request itself failed).
fetch() {
  rm -f "$tmp/body"
  code=$(curl -q --noproxy '*' -s --max-time 10 -o "$tmp/body" -w '%{http_code}' "$1") || code=000
  [ -f "$tmp/body" ] || : >"$tmp/body"
}

# compact prints the body with all whitespace removed, for JSON-shape matches.
compact() {
  tr -d '[:space:]' <"$tmp/body"
}

ok() {
  pass=$((pass + 1))
  echo "PASS  $1"
}

bad() {
  fail=$((fail + 1))
  excerpt=$(dd if="$tmp/body" bs=200 count=1 2>/dev/null | tr '\n\r' '  ')
  # printf, not echo: the excerpt is response body text, and some echos (dash)
  # expand backslash escapes in it.
  printf 'FAIL  %s (HTTP %s: %s)\n' "$1" "$code" "$excerpt"
}

# expect NAME CODE MODE NEEDLE: pass when the status is CODE and the body
# contains NEEDLE, matched on the raw body (MODE raw) or the compacted body
# (MODE compact).
expect() {
  if [ "$code" = "$2" ]; then
    if [ "$3" = compact ]; then
      if compact | grep -qF -- "$4"; then
        ok "$1"
        return 0
      fi
    elif grep -qF -- "$4" "$tmp/body"; then
      ok "$1"
      return 0
    fi
  fi
  bad "$1"
}

fetch "$API/healthz"
expect 'a1 api /healthz is ok' 200 compact '"status":"ok"'

fetch "$API/api/cities"
expect 'a2 api /api/cities lists Springfield' 200 raw '"Springfield"'

fetch "$API/api/listings/search?targetBudget=450000&city=Springfield"
expect 'a3 api search returns items' 200 compact '"items":[{'

# The first "rank" in the document belongs to items[0], whatever the field order.
first_rank=$(compact | tr '{' ',' | tr ',' '\n' | sed -n 's/^"rank":\([0-9][0-9]*\).*$/\1/p' | head -n 1)
if [ "$code" = 200 ] && [ "$first_rank" = 1 ]; then
  ok 'a4 api search ranks the first item 1'
else
  bad "a4 api search ranks the first item 1 (got rank ${first_rank:-none})"
fi

fetch "$API/api/listings/search?minPrice=5&maxPrice=1"
expect 'a5 api search rejects minPrice > maxPrice' 400 compact '"code":"min_gt_max"'

fetch "$WEB/"
expect 'a6 web serves the app shell' 200 raw '<title>Brillio Listings</title>'

fetch "$WEB/healthz"
expect 'a7 web proxies /healthz to api' 200 compact '"status":"ok"'

fetch "$WEB/api/cities"
expect 'a8 web proxies /api/cities to api' 200 raw '"Springfield"'

# --- 5. Result. Diagnostics go out before the EXIT trap tears the stack down.
# With no failures the script ends here with status 0. There is no trailing
# `exit 0`: with one, ShellCheck 0.11 misreports cleanup as never invoked.
echo "smoke: $pass passed, $fail failed"
if [ "$fail" -ne 0 ]; then
  diagnose
  exit 1
fi
