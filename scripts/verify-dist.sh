#!/bin/sh
# scripts/verify-dist.sh <zip> (docs/SPEC.md section 12): check a release zip
# the way a stranger would use it. Inspect its entries, extract it into a fresh
# temporary directory, scan every file and path name for emoji (BRANDING.md
# section 8), then follow the README from the extracted copy: make up and the
# README's curl examples, make down, make smoke and make check. The stack is
# torn down and the copy removed on pass, on fail and on interrupt. It reads
# only the zip, never the checkout it lives in.
#
# Exit status: 0 pass; 1 a check failed; 2 refused (usage, a missing tool,
# Docker not running, port 8080 or 5173 in use, a brillio stack exists);
# 129, 130, 143 or 141 on HUP, INT, TERM or PIPE.
#
# POSIX sh. Needs docker (with compose), curl, unzip, perl and make, and for
# make check go, node, npm and golangci-lint (it runs natively, as the
# README says). Set TMPDIR to a directory Docker can bind-mount if the default
# temporary directory is not shared with Docker.
set -eu

# A stranger's shell exports none of these. DOCKER_DEFAULT_PLATFORM,
# DOCKER_BUILDKIT and COMPOSE_DOCKER_CLI_BUILD change how the stack builds and
# runs (see compose.yaml); COMPOSE_FILE and COMPOSE_PROJECT_NAME would point
# Compose at another file or project.
unset DOCKER_DEFAULT_PLATFORM DOCKER_BUILDKIT COMPOSE_DOCKER_CLI_BUILD COMPOSE_FILE COMPOSE_PROJECT_NAME

API=http://localhost:8080
WEB=http://localhost:5173
# Seconds for make up to answer on both ports: api's 180s start_period, then
# web's 180s start_period (web waits for a healthy api), plus a margin.
UP_DEADLINE=400

# The README's API examples, verbatim. Each must appear in the README, and the
# running stack must answer with it (compared with whitespace removed).
HEALTH_CMD="curl -s $API/healthz"
HEALTH_BODY='{"status":"ok"}'
CITIES_CMD="curl -s $API/api/cities"
CITIES_BODY='{"cities":["Chantilly","Fairfax","Manassas","Reston","Springfield","Vienna"]}'
SEARCH_URL="$API/api/listings/search?targetBudget=450000&city=Springfield"

# Entries that must never ship: .git, .lattice, node_modules, dist and
# __MACOSX folders, SQLite files, any PDF, macOS cruft, absolute paths and
# parent-directory paths.
FORBIDDEN='(^|/)(\.git|\.lattice|node_modules|dist|__MACOSX)(/|$)|\.db$|\.db-[^/]*$|\.[Pp][Dd][Ff]$|(^|/)\.DS_Store$|(^|/)\._|^/|(^|/)\.\.(/|$)'

refuse() {
  printf 'verify-dist: %s\n' "$*" >&2
  exit 2
}

pass=0
fail=0
ok() {
  pass=$((pass + 1))
  printf 'PASS  %s\n' "$1"
}
bad() {
  fail=$((fail + 1))
  printf 'FAIL  %s\n' "$1"
}

# port_free PORT: curl exit 7 is "connection refused": nothing listens there,
# whichever user owns the listener.
port_free() {
  rc=0
  curl -q --noproxy '*' -s -o /dev/null --max-time 2 "http://localhost:$1/" || rc=$?
  [ "$rc" -eq 7 ]
}

# --- 1. Usage and preflight. No trap yet: nothing to clean up, and a stack this
# script did not start is never torn down.
[ "$#" -eq 1 ] || refuse 'usage: scripts/verify-dist.sh <zip>'
case $1 in
  *.zip) ;;
  *) refuse "not a .zip file: $1" ;;
esac
[ -f "$1" ] || refuse "no such file: $1"
zip=$(cd "$(dirname "$1")" && pwd -P)/$(basename "$1")
bytes=$(($(wc -c <"$zip")))

missing=
for t in docker curl unzip perl make go node npm golangci-lint; do
  command -v "$t" >/dev/null 2>&1 || missing="$missing $t"
done
[ -z "$missing" ] || refuse "not installed:$missing (make check runs natively; see the README's tool table)"
docker info >/dev/null 2>&1 || refuse 'docker is not running (docker info failed)'
docker compose version >/dev/null 2>&1 || refuse 'docker compose is not available'
for p in 8080 5173; do
  port_free "$p" || refuse "port $p is in use"
done
existing=$(docker ps -a -q --filter label=com.docker.compose.project=brillio) ||
  refuse 'docker ps failed'
[ -z "$existing" ] || refuse 'a brillio stack already exists; run make down in its checkout first'

# result prints the last line.
result() {
  if [ "$1" -eq 0 ] && [ "$fail" -eq 0 ]; then
    echo "verify-dist: PASS ($pass checks) $zip $bytes bytes"
  elif [ "$fail" -gt 0 ]; then
    echo "verify-dist: FAIL ($fail failed)"
  else
    echo "verify-dist: FAIL (exit $1)"
  fi
}

# --- 2. Zip inspection, from the listing alone.
echo "verify-dist: inspecting $zip ($bytes bytes)"
entries=$(unzip -Z1 "$zip" 2>/dev/null) || entries=
if [ -z "$entries" ]; then
  bad 'zip: lists its entries'
  result 1
  exit 1
fi
top=$(printf '%s\n' "$entries" | sed 's|/.*||' | sort -u)
ntop=$(($(printf '%s\n' "$top" | wc -l)))
loose=$(printf '%s\n' "$entries" | grep -v / || true)
if [ "$ntop" -eq 1 ] && [ -n "$top" ] && [ -z "$loose" ]; then
  ok "zip: one top-level folder, $top/"
else
  bad "zip: one top-level folder (found: $(printf '%s\n' "$top" | head -n 5 | tr '\n' ' '))"
fi
absent=
for f in README.md compose.yaml Makefile scripts/smoke.sh; do
  printf '%s\n' "$entries" | grep -qxF -- "$top/$f" || absent="$absent $f"
done
if [ -z "$absent" ]; then
  ok 'zip: holds README.md, compose.yaml, Makefile and scripts/smoke.sh'
else
  bad "zip: holds README.md, compose.yaml, Makefile and scripts/smoke.sh (missing:$absent)"
fi
bad_entries=$(printf '%s\n' "$entries" | grep -E -- "$FORBIDDEN" || true)
if [ -z "$bad_entries" ]; then
  ok 'zip: no .git, .lattice, node_modules, dist, SQLite, PDF or OS cruft entries'
else
  bad 'zip: no .git, .lattice, node_modules, dist, SQLite, PDF or OS cruft entries; found:'
  printf '%s\n' "$bad_entries" | sed -n '1,20s/^/        /p'
fi
if [ "$fail" -ne 0 ]; then
  result 1
  exit 1
fi

# --- 3. A fresh directory, and the teardown trap.
base=${TMPDIR:-/tmp}
base=${base%/}
work=$(mktemp -d "$base/brillio-verify.XXXXXX") || refuse "cannot create a directory under $base"
root=$work/$top
stack=0
up_pid=

# proc_tree PID: PID and every process under it, deepest first.
proc_tree() {
  ps -A -o pid= -o ppid= | awk -v root="$1" '
    { up[$1] = $2 }
    END {
      for (p in up) {
        d = 0
        for (q = p; q != root && (q in up) && d < 64; q = up[q]) d++
        if (q == root) print d, p
      }
    }' | sort -rn | awk '{ print $2 }'
}

# reap_up SECONDS: give the background make up SECONDS to exit by itself, then
# TERM it and every process under it, deepest first, and KILL what is left 10
# seconds later. Killing make alone is not enough: the docker compose client is
# its grandchild, would be reparented to PID 1 and go on building and starting
# the stack after the teardown. Never a bare wait under errexit.
reap_up() {
  [ -n "$up_pid" ] || return 0
  d=$(($(date +%s) + $1))
  while kill -0 "$up_pid" 2>/dev/null && [ "$(date +%s)" -lt "$d" ]; do
    sleep 1
  done
  if kill -0 "$up_pid" 2>/dev/null; then
    pids=$(proc_tree "$up_pid")
    for p in $pids; do kill "$p" 2>/dev/null || true; done
    d=$(($(date +%s) + 10))
    for p in $pids; do
      while kill -0 "$p" 2>/dev/null && [ "$(date +%s)" -lt "$d" ]; do sleep 1; done
      kill -s KILL "$p" 2>/dev/null || true
    done
  fi
  wait "$up_pid" 2>/dev/null || true
  up_pid=
}

# leftovers prints the ID of every brillio container, named volume and network,
# the objects that carry the compose project label. The web container's
# anonymous node_modules volume has no such label, so it is not listed; make
# down -v removes it along with the container it is attached to.
leftovers() {
  f=label=com.docker.compose.project=brillio
  docker ps -a -q --filter "$f"
  docker volume ls -q --filter "$f"
  docker network ls -q --filter "$f"
}

# diagnose prints the stack state, recent logs and the make up log, before the
# teardown erases them.
diagnose() {
  (cd "$root" && docker compose ps -a)
  (cd "$root" && docker compose logs --no-color --tail=60)
  if [ -f "$work/make-up.log" ]; then
    echo 'verify-dist: last 40 lines of make up:'
    tail -n 40 "$work/make-up.log"
  fi
}

# Ctrl-C reaches every process in the terminal's process group. The shell and
# make in the background make up job inherit SIGINT ignored, but docker compose
# re-enables it and starts its own stop, which reap_up gives 10 seconds. A TERM
# or HUP reaches this script alone, so reap_up stops the compose client itself.
# Either way it is gone before make down, so nothing builds the stack again.
cleanup() {
  status=$?
  set +e # an echo into a dead stdout (EPIPE) must not skip the teardown
  trap - EXIT
  trap '' INT TERM HUP PIPE
  if [ "$stack" = 1 ]; then
    if [ "$status" -ne 0 ] || [ "$fail" -ne 0 ]; then
      diagnose
    fi
    reap_up 10
    echo 'verify-dist: teardown: make down'
    if ! (cd "$root" && make down); then
      bad 'teardown: make down'
      # On a failure or interrupt path the first status stands.
      [ "$status" -ne 0 ] || status=1
    fi
    # A last look: nothing of the stack may outlive the run.
    [ -z "$(leftovers)" ] || (cd "$root" && make down)
    if [ -n "$(leftovers)" ]; then
      bad 'teardown: no brillio container, volume or network is left'
      [ "$status" -ne 0 ] || status=1
    fi
  fi
  case $work in
    "$base"/brillio-verify.?*) rm -rf "$work" ;;
  esac
  result "$status"
  exit "$status"
}

trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
trap 'exit 141' PIPE

echo "verify-dist: extracting into $work"
if unzip -q "$zip" -d "$work" && [ -f "$root/README.md" ]; then
  ok "zip: extracts into a fresh directory"
else
  bad "zip: extracts into a fresh directory"
  exit 1
fi

# --- 4. Emoji scan of exactly what was extracted, before anything runs.
# Emoji: pictographs, default-emoji presentation (regional-indicator flags
# included), the emoji variation selector U+FE0F, the keycap combiner U+20E3
# and tag characters; minus (c), (r), TM and the arrows block U+2190-21FF
# standing alone, which are text symbols. Not \p{Emoji}: that matches 0-9, #
# and *. Lines are decoded as UTF-8; a bad byte becomes U+FFFD, never a match.
echo 'verify-dist: scanning for emoji'
rc=0
perl - "$work" "$root" <<'PERL' || rc=$?
use strict;
use warnings;
use Encode ();
use File::Find ();

my ($work, $root) = @ARGV;
my $re = eval q{qr/(?![\x{A9}\x{AE}\x{2122}\x{2190}-\x{21FF}])[\p{Extended_Pictographic}\p{Emoji_Presentation}\x{FE0F}\x{20E3}\x{E0020}-\x{E007F}]/};
# A perl without these properties must never pass.
unless ($re && "\x{1F600}" =~ $re && "\x{2192}" !~ $re && "1" !~ $re) {
    print STDERR "emoji: self-test failed; this perl lacks the Unicode emoji properties\n";
    exit 2;
}

my ($files, $names, $hits, $errors) = (0, 0, 0, 0);
sub scan {
    my ($where, $bytes) = @_;
    my $text = Encode::decode('UTF-8', $bytes);
    while ($text =~ /($re)/g) {
        $hits++;
        printf "emoji: %s U+%04X\n", $where, ord $1 if $hits <= 50;
    }
}

File::Find::find({ no_chdir => 1, wanted => sub {
    my $path = $File::Find::name;
    my $rel = substr $path, length($work) + 1;
    $names++;
    scan("$rel (path name)", $rel);
    return if -l $path || !-f $path || -z $path;
    if (-B $path) {
        print "SKIP  binary $rel\n";
        return;
    }
    my $fh;
    unless (open $fh, '<:raw', $path) {
        print "emoji: cannot read $rel: $!\n";
        $errors++;
        return;
    }
    $files++;
    while (my $line = <$fh>) {
        scan("$rel:$.", $line);
    }
    close $fh;
} }, $root);

print "emoji: ... and ", $hits - 50, " more\n" if $hits > 50;
print "emoji: scanned $files files and $names path names, $hits found\n";
exit($hits || $errors ? 1 : 0);
PERL
case $rc in
  0) ok 'emoji: none in any file or path name (BRANDING.md section 8)' ;;
  1) bad 'emoji: none in any file or path name (BRANDING.md section 8)' ;;
  *) refuse 'the emoji scan could not run' ;;
esac
if [ "$fail" -ne 0 ]; then
  exit 1
fi

# --- 5. The README documents every step run below, so a stranger can follow it.
absent=
for needle in 'make up' 'Then open http://localhost:5173' 'make down' 'make smoke' 'make check' \
  "$HEALTH_CMD" "# $HEALTH_BODY" "$CITIES_CMD" "# $CITIES_BODY" "curl -s '$SEARCH_URL'" \
  'The Springfield search returns four listings'; do
  grep -qF -- "$needle" "$root/README.md" || absent="$absent
        $needle"
done
if [ -z "$absent" ]; then
  ok 'README: documents every command run below'
else
  bad "README: documents every command run below; missing:$absent"
  exit 1
fi

# --- 6. Follow the README from the extracted copy.
cd "$root"

# Every curl call starts with -q (skip ~/.curlrc) and passes --noproxy '*', so a
# proxy in the environment or the user's curl config never sees localhost.
answers() {
  curl -q --noproxy '*' -fs -o /dev/null --max-time 10 "$1"
}

code=000
# fetch URL: the body lands in $work/body (always present, maybe empty), the
# status in $code (000 when the request itself failed).
fetch() {
  rm -f "$work/body"
  code=$(curl -q --noproxy '*' -s --max-time 10 -o "$work/body" -w '%{http_code}' "$1") || code=000
  [ -f "$work/body" ] || : >"$work/body"
}

# expect NAME NEEDLE MODE: pass on a 200 whose body, with whitespace removed,
# equals (MODE equals) or contains (MODE contains) NEEDLE, also compacted.
expect() {
  got=$(tr -d '[:space:]' <"$work/body")
  want=$(printf '%s' "$2" | tr -d '[:space:]')
  matched=no
  if [ "$code" = 200 ]; then
    case $3 in
      equals) if [ "$got" = "$want" ]; then matched=yes; fi ;;
      contains) case $got in *"$want"*) matched=yes ;; esac ;;
    esac
  fi
  if [ "$matched" = yes ]; then
    ok "$1"
    return 0
  fi
  excerpt=$(printf '%s' "$got" | cut -c 1-200)
  # printf, not echo: the excerpt is response body text.
  bad "$(printf '%s (HTTP %s: %s)' "$1" "$code" "$excerpt")"
}

# make up runs in the foreground until Ctrl-C; this script stands in for the
# user's terminal, so it runs in the background with its output in a log.
echo "verify-dist: make up (background; log in $work/make-up.log)"
make up </dev/null >"$work/make-up.log" 2>&1 &
up_pid=$!
stack=1

echo "verify-dist: waiting up to ${UP_DEADLINE}s for $API/healthz and $WEB/"
deadline=$(($(date +%s) + UP_DEADLINE))
until answers "$API/healthz" && answers "$WEB/"; do
  if ! kill -0 "$up_pid" 2>/dev/null; then
    bad 'make up: exited before the stack answered'
    exit 1
  fi
  if [ "$(date +%s)" -ge "$deadline" ]; then
    bad "make up: api and web answer within ${UP_DEADLINE}s"
    exit 1
  fi
  sleep 2
done
ok "make up: $API/healthz and $WEB/ answer"

fetch "$API/healthz"
expect "README: $HEALTH_CMD prints $HEALTH_BODY" "$HEALTH_BODY" equals
fetch "$API/api/cities"
expect "README: $CITIES_CMD prints the six cities" "$CITIES_BODY" equals
# Rank order is date-sensitive (recency), so only the count is asserted. The
# trailing comma makes it exact (not 42): totalPages always follows totalItems.
fetch "$SEARCH_URL"
expect 'README: the Springfield search returns four listings' '"totalItems":4,' contains

echo 'verify-dist: make down'
if make down; then
  ok 'make down'
else
  bad 'make down'
  exit 1
fi
reap_up 60
# make smoke refuses a port still held, so wait for Docker to release both.
d=$(($(date +%s) + 30))
until port_free 8080 && port_free 5173; do
  if [ "$(date +%s)" -ge "$d" ]; then
    bad 'make down: ports 8080 and 5173 are free within 30s'
    exit 1
  fi
  sleep 1
done

echo 'verify-dist: make smoke'
if make smoke; then
  ok 'make smoke'
else
  bad 'make smoke'
fi

echo 'verify-dist: make check'
if make check; then
  ok 'make check'
else
  bad 'make check'
fi

# The EXIT trap tears down and prints the result line.
[ "$fail" -eq 0 ] || exit 1
