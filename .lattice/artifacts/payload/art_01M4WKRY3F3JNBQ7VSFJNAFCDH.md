# BC-17 validation: release packaging (make dist, verify-dist)

- Ticket: BC-17 "Release packaging: make dist zip and fresh-directory verification"
- Validator: agent:opus-5.5-validator
- Worktree: /Users/forrest/Code/brillio-worktrees.noindex/bc-17-release-dist, branch feat/BC-17-release-dist
- HEAD: aa2d63acb6e828067531128056d153611cd30ba2 ("docs(readme): plain wording for the Linux root-owned files note (BC-17)")
- master contained: 7fb72bf3c32b2c41dced454e1e4b4d205c399a55 (`git merge-base --is-ancestor master HEAD` exit 0; master tip still 7fb72bf)
- Tree before and after: `git -c core.excludesFile=/dev/null status --porcelain` empty
- Date: 2026-10-14, 07:03:41Z (preflight) to 07:06:01Z (final state), UTC
- Host: Darwin 27.0.0 arm64; Docker server 29.8.1; Docker Compose v5.5.1; go1.27.1 darwin/arm64; node v24.21.0
- TMPDIR=/var/folders/_8/mbf7vqfs70339k03bpbr5qkw0000gq/T/

Operator environment, exported in the shell that ran `make dist`, verify-dist and every negative check
(`env | grep -E 'DOCKER_DEFAULT_PLATFORM|DOCKER_BUILDKIT|COMPOSE_DOCKER_CLI_BUILD'`):

```
COMPOSE_DOCKER_CLI_BUILD=0
DOCKER_BUILDKIT=0
DOCKER_DEFAULT_PLATFORM=linux/amd64
```

verify-dist unsets them: the `make smoke` build inside the run printed BuildKit output (`#1 [internal] load local bake
definitions`, `#9 resolving provenance for metadata file`), which the legacy builder (DOCKER_BUILDKIT=0) does not produce.

Preflight (before any Compose run): `curl localhost:8080` rc 7, `curl localhost:5173` rc 7,
`docker ps -a --filter label=com.docker.compose.project=brillio -q` empty.

## Overall: PASS

| Item | Result |
|---|---|
| make dist | PASS |
| A3 no forbidden content | PASS |
| A1 verify-dist end to end | PASS (13 checks, exit 0, 66 s) |
| A2 README sufficient | PASS (11 of 11 needles present) |
| N1 emoji | PASS (exit 1 at emoji step, no container) |
| N2 .lattice / .db / .pdf | PASS (exit 1 at inspection, all three) |
| N3 port 8080 held | PASS (exit 2, no temp dir, listener killed, rc 7 after) |
| Final state | PASS (nothing left) |

## make dist: PASS

Command: `rm -f dist/brillio-listings-*.zip; make dist` (dist/ was empty after the rm). Exit 0.

```
sh scripts/dist.sh
dist: source   HEAD aa2d63a docs(readme): plain wording for the Linux root-owned files note (BC-17)
dist: wrote    dist/brillio-listings-20261014.zip
dist: size     329238 bytes (322 KiB)
dist: files    135 (8 root, 46 backend/, 1 data/, 3 docs/, 74 frontend/, 3 scripts/)
dist: excluded 368 tracked files (366 .lattice/, 1 Coding_Test_Question_Paper.pdf, 1 Instructions.pdf)
dist: check it with: scripts/verify-dist.sh dist/brillio-listings-20261014.zip
```

- Zip: `dist/brillio-listings-20261014.zip`, `wc -c` 329238 bytes (322 KiB), 135 files.
- No dirty-tree warning was printed (clean tree).

`unzip -l dist/brillio-listings-20261014.zip | tail -1` (169 entries = 135 files + 34 directory entries):

```
   930416                     169 files
```

`unzip -l ... | head -15` (the second line is the zip comment git archive writes: the HEAD commit):

```
Archive:  dist/brillio-listings-20261014.zip
aa2d63acb6e828067531128056d153611cd30ba2
  Length      Date    Time    Name
---------  ---------- -----   ----
        0  10-14-2026 03:00   brillio-listings-20261014/
      646  10-14-2026 03:00   brillio-listings-20261014/.editorconfig
      764  10-14-2026 03:00   brillio-listings-20261014/.gitignore
    21084  10-14-2026 03:00   brillio-listings-20261014/BRANDING.md
    27237  10-14-2026 03:00   brillio-listings-20261014/CLAUDE.md
     2318  10-14-2026 03:00   brillio-listings-20261014/Makefile
    16979  10-14-2026 03:00   brillio-listings-20261014/README.md
    27063  10-14-2026 03:00   brillio-listings-20261014/agents.md
        0  10-14-2026 03:00   brillio-listings-20261014/backend/
     1412  10-14-2026 03:00   brillio-listings-20261014/backend/.air.toml
       99  10-14-2026 03:00   brillio-listings-20261014/backend/.dockerignore
```

Per-top-level file counts (`unzip -Z1 <zip> | grep -v '/$' | awk -F/ '{print ($3==""?"(root)":$2"/")}' | sort | uniq -c`),
matching the make dist line:

```
   8 (root)
  46 backend/
   1 data/
   3 docs/
  74 frontend/
   3 scripts/
```

One top-level folder only: `brillio-listings-20261014`. Root files: .editorconfig, .gitignore, BRANDING.md, CLAUDE.md,
Makefile, README.md, agents.md, compose.yaml. data/ holds only `sample_listings.json`; scripts/ holds dist.sh, smoke.sh,
verify-dist.sh; docs/ holds EXTENSIONS.md, SPEC.md, branding-demo.html.

Cross-check: the zip's file list against `git ls-tree -r --name-only HEAD` minus `.lattice/` differs only by the two
excluded PDFs (`Coding_Test_Question_Paper.pdf`, `Instructions.pdf`), so the zip is exactly the tracked tree minus the
SPEC section 12 exclusions.

## A3 no .lattice, .git, node_modules, databases or PDFs: PASS

`unzip -Z1 dist/brillio-listings-20261014.zip | grep -Ei '(^|/)(\.git|\.lattice|node_modules|dist|__MACOSX)(/|$)|\.db$|\.db-|\.pdf$|(^|/)\.DS_Store$'`
printed nothing, grep exit 1.

Required entries present (`grep -qxF brillio-listings-20261014/<f>`): CLAUDE.md, agents.md (operator ruling: both ship),
README.md, compose.yaml, Makefile, scripts/verify-dist.sh, scripts/smoke.sh, scripts/dist.sh.

## A1 make dist then scripts/verify-dist.sh passes end to end: PASS

Command (operator variables exported, see header):
`scripts/verify-dist.sh dist/brillio-listings-20261014.zip > final-verify-dist.log 2>&1; echo "exit=$?"`

Result: `exit=0`, wall time 66 s (07:04:16Z to 07:05:22Z). Docker image layers were cached on this host; `make down -v`
dropped the Go volumes between runs and `make check` ran a fresh native `npm ci` (353 packages) in the extracted copy.

Every PASS line and stage marker, in order (183-line log, read in full):

```
verify-dist: inspecting /Users/forrest/Code/brillio-worktrees.noindex/bc-17-release-dist/dist/brillio-listings-20261014.zip (329238 bytes)
PASS  zip: one top-level folder, brillio-listings-20261014/
PASS  zip: holds README.md, compose.yaml, Makefile and scripts/smoke.sh
PASS  zip: no .git, .lattice, node_modules, dist, SQLite, PDF or OS cruft entries
verify-dist: extracting into /var/folders/_8/mbf7vqfs70339k03bpbr5qkw0000gq/T/brillio-verify.kUCUp5
PASS  zip: extracts into a fresh directory
verify-dist: scanning for emoji
emoji: scanned 135 files and 169 path names, 0 found
PASS  emoji: none in any file or path name (BRANDING.md section 8)
PASS  README: documents every command run below
verify-dist: make up (background; log in /var/folders/_8/mbf7vqfs70339k03bpbr5qkw0000gq/T/brillio-verify.kUCUp5/make-up.log)
verify-dist: waiting up to 400s for http://localhost:8080/healthz and http://localhost:5173/
PASS  make up: http://localhost:8080/healthz and http://localhost:5173/ answer
PASS  README: curl -s http://localhost:8080/healthz prints {"status":"ok"}
PASS  README: curl -s http://localhost:8080/api/cities prints the six cities
PASS  README: the Springfield search returns four listings
verify-dist: make down
docker compose down -v
  (brillio-web-1 and brillio-api-1 stopped and removed; volumes brillio_gobuild, brillio_gomod and network brillio_default removed)
PASS  make down
verify-dist: make smoke
PASS  make smoke
verify-dist: make check
PASS  make check
verify-dist: teardown: make down
docker compose down -v
verify-dist: PASS (13 checks) /Users/forrest/Code/brillio-worktrees.noindex/bc-17-release-dist/dist/brillio-listings-20261014.zip 329238 bytes
```

Stages confirmed to have run: zip inspection (3 checks), extraction into a fresh `mktemp -d` directory, emoji scan (135
files, 169 path names, 0 found), README check, `make up` plus health wait, the three README curls (healthz, cities,
Springfield `"totalItems":4,`), `make down`, `make smoke`, `make check`, EXIT-trap teardown. The extracted directory
`brillio-verify.kUCUp5` was removed by the trap (none left afterwards).

`make smoke` (the zip's own scripts/smoke.sh, `docker compose up -d --build --wait`; api and web both reported Healthy):

```
PASS  a1 api /healthz is ok
PASS  a2 api /api/cities lists Springfield
PASS  a3 api search returns items
PASS  a4 api search ranks the first item 1
PASS  a5 api search rejects minPrice > maxPrice
PASS  a6 web serves the app shell
PASS  a7 web proxies /healthz to api
PASS  a8 web proxies /api/cities to api
smoke: 8 passed, 0 failed
teardown: docker compose down -v
```

`make check` (native, in the extracted copy):

```
cd backend && golangci-lint run
0 issues.
cd frontend && npm ci
added 353 packages, and audited 354 packages in 2s
found 0 vulnerabilities
cd frontend && npm run lint && npm run typecheck && npm run format:check
> eslint . --max-warnings 0
> tsc -b --noEmit
> prettier --check .
All matched files use Prettier code style!
cd backend && go test -race ./...
ok  	brillio/backend/internal/config	1.113s
ok  	brillio/backend/internal/domain	1.133s
ok  	brillio/backend/internal/httpapi	1.677s
ok  	brillio/backend/internal/store	1.544s
ok  	brillio/backend/internal/store/seed	1.183s
cd frontend && npm test
> vitest run
 Test Files  23 passed (23)
      Tests  398 passed (398)
```

(npm printed only warnings: eslint@9.39.5 deprecated, and fsevents@2.3.3 install scripts not covered by allowScripts.)

## A2 the README in the zip is sufficient: PASS

verify-dist runs only commands its README check (scripts/verify-dist.sh lines 329-332) requires to be quoted in the
README. Each needle checked with `grep -F` against `unzip -p dist/brillio-listings-20261014.zip brillio-listings-20261014/README.md`:

| Needle | Matching lines | First line |
|---|---|---|
| `make up` | 4 | 8 |
| `Then open http://localhost:5173` | 1 | 11 |
| `make down` | 2 | 11 |
| `make smoke` | 5 | 32 |
| `make check` | 3 | 19 |
| `curl -s http://localhost:8080/healthz` | 1 | 110 |
| `# {"status":"ok"}` | 1 | 111 |
| `curl -s http://localhost:8080/api/cities` | 1 | 112 |
| `# {"cities":["Chantilly","Fairfax","Manassas","Reston","Springfield","Vienna"]}` | 1 | 113 |
| `curl -s 'http://localhost:8080/api/listings/search?targetBudget=450000&city=Springfield'` | 1 | 114 |
| `The Springfield search returns four listings` | 1 | 117 |

11 of 11 present; verify-dist's own `PASS  README: documents every command run below` agrees, and the three README curls
answered exactly as the README shows (the A1 PASS lines above). The run read only the zip: it executed from the extracted
copy under TMPDIR, and the vitest banner shows the run root
`/private/var/folders/.../brillio-verify.kUCUp5/brillio-listings-20261014/frontend`; the worktree stayed clean.

## Negative checks (scratch copies under the session scratchpad, never in the repo): PASS

Each scratch zip is a copy of the real zip plus one entry added with `zip -q`. Each run had the operator variables exported.

N1 emoji (`brillio-listings-20261014/smile.txt` containing `printf '\360\237\230\200\n'`, U+1F600): exit 1 at the emoji step,
0 brillio containers, extraction directory removed (0 `brillio-verify.*` before and after).

```
PASS  zip: one top-level folder, brillio-listings-20261014/
PASS  zip: holds README.md, compose.yaml, Makefile and scripts/smoke.sh
PASS  zip: no .git, .lattice, node_modules, dist, SQLite, PDF or OS cruft entries
verify-dist: extracting into /var/folders/_8/mbf7vqfs70339k03bpbr5qkw0000gq/T/brillio-verify.eaH3zP
PASS  zip: extracts into a fresh directory
verify-dist: scanning for emoji
emoji: brillio-listings-20261014/smile.txt:1 U+1F600
emoji: scanned 136 files and 170 path names, 1 found
FAIL  emoji: none in any file or path name (BRANDING.md section 8)
verify-dist: FAIL (1 failed)
```

N2 forbidden entries, one zip each; every run exit 1 at inspection, before extraction (no `brillio-verify.*` created),
0 brillio containers:

| Added entry | Exit | Reported |
|---|---|---|
| `brillio-listings-20261014/.lattice/x` | 1 | `FAIL  zip: no .git, .lattice, node_modules, dist, SQLite, PDF or OS cruft entries; found:` / `brillio-listings-20261014/.lattice/x` |
| `brillio-listings-20261014/data/listings.db` | 1 | same FAIL line / `brillio-listings-20261014/data/listings.db` |
| `brillio-listings-20261014/x.pdf` | 1 | same FAIL line / `brillio-listings-20261014/x.pdf` |

Each ended `verify-dist: FAIL (1 failed)`.

N3 port held: `python3 -I -m http.server 8080 --bind 127.0.0.1` (empty directory) in the background, pid 10504; it answered
HTTP 200 and `lsof -nP -iTCP:8080 -sTCP:LISTEN` showed `Python 10504 ... TCP 127.0.0.1:8080 (LISTEN)`. Then
`scripts/verify-dist.sh dist/brillio-listings-20261014.zip`:

```
verify-dist: port 8080 is in use
```

exit 2, 0 `brillio-verify.*` directories before and after, 0 brillio containers. Listener killed (`kill 10504`, wait rc 143,
`kill -0` fails); afterwards `curl localhost:8080` rc 7 and lsof shows no listener.

## Final state: PASS

At 2026-10-14T07:06:01Z:

- `docker ps -a --filter label=com.docker.compose.project=brillio`: header only, no containers
- `docker volume ls --filter label=com.docker.compose.project=brillio`: header only, no volumes
- `docker network ls --filter label=com.docker.compose.project=brillio`: header only, no networks
- `curl localhost:8080` rc 7, `curl localhost:5173` rc 7; `lsof -nP -iTCP:<port> -sTCP:LISTEN` empty for both
- `ls -d ${TMPDIR:-/tmp}/brillio-verify.*`: no matches
- `git -c core.excludesFile=/dev/null status --porcelain`: empty; HEAD still aa2d63a
- No verify-dist, http.server, make up or compose processes left running

The release zip `dist/brillio-listings-20261014.zip` (329238 bytes) stays in the gitignored dist/ directory.
