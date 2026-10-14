# BC-17: Release packaging: make dist zip and fresh-directory verification

Contract: docs/SPEC.md §12 (zip minus .git, node_modules, *.db, .lattice, the two PDFs; a README a stranger follows to run
`docker compose up` and `make check`; verified from a fresh directory; no emoji, BRANDING.md §8). Ticket adds dist/ and OS cruft.

## Scope
- `make dist` -> `dist/brillio-listings-YYYYMMDD.zip`, built with `git archive` of HEAD (immune to the host core.excludesFile).
- `scripts/verify-dist.sh <zip>`: inspect, extract into a fresh `mktemp -d`, emoji scan, run the README's commands, tear down.
- Smallest fixes for whatever the real run surfaces. Dry probe (git archive into scratch, no Docker) found nothing missing:
  compose.yaml only bind-mounts ./backend, ./data, ./frontend, all tracked; nothing at build or run time reads .lattice or the PDFs.

## Files (worktree-relative)
- `scripts/dist.sh` (new, mode 100755, ~45 LOC). `scripts/verify-dist.sh` (new, 100755, ~170 LOC incl. inline perl).
- `Makefile`: BC-17 block only; the stub becomes `dist:` + TAB `sh scripts/dist.sh`, plus a 2-line comment (same shape as `smoke`).
- `README.md` (BC-15 owns it, minimal edit, layout block only): add `dist` to the Makefile line; add two lines
  `scripts/dist.sh  make dist: the release zip` and `scripts/verify-dist.sh  checks a release zip from a fresh directory`.
- SPEC edit: §3 -- add the two scripts to the layout (new files this ticket creates); §12 -- name `scripts/verify-dist.sh` as the
  verifier. No other contract change.
- Estimate ~225 LOC total. Commit both scripts as 100755 and confirm with `git -c core.excludesFile=/dev/null status`.

## make dist (scripts/dist.sh, POSIX sh, set -eu)
1. `cd "$(dirname "$0")/.."`; die unless `git rev-parse --show-toplevel` equals `pwd -P` ("make dist builds from a git
   checkout"; so running it inside an extracted zip fails cleanly). Needs git and unzip.
2. `name=brillio-listings-$(date +%Y%m%d)`. One top-level folder `$name/` inside the zip (unzip makes one folder, not a spray).
3. Dirty tree: warn, do not fail. `git -c core.excludesFile=/dev/null status --porcelain` non-empty -> print
   `dist: warning: the zip is HEAD <sha>; these uncommitted changes are NOT in it:` plus up to 20 paths. The zip is a pure
   function of HEAD, so failing would add friction without changing the artifact; final evidence requires a clean tree.
4. `git archive --format=zip --prefix="$name/" -o "dist/.$name.zip.tmp" HEAD -- . ':(exclude).lattice'
   ':(exclude)Coding_Test_Question_Paper.pdf' ':(exclude)Instructions.pdf' ':(exclude)dist'` plus `(exclude,glob)` for
   `**/node_modules/**`, `**/*.db`, `**/*.db-*`, `**/.DS_Store`, `**/._*`, `**/__MACOSX/**`; then `mv` onto `dist/$name.zip`
   (no partial zip on failure). git archive never includes .git; the globs are belt and braces (none of it is tracked).
   Verified on this host: the pathspec form works, executables keep 0755 (unzip and ditto), other files extract 0644.
5. Print: source (`HEAD <short sha> <subject>`), path, size in bytes and KiB (`wc -c`), file count (`unzip -Z1`, minus dirs),
   files per top-level path, excluded tracked-file count (git ls-tree count minus zip count), and the verify-dist command.
   Probe today: 133 files, 319,323 bytes; per top-level: 8 root, 46 backend/, 1 data/, 3 docs/, 74 frontend/, 1 scripts/.

## What ships
Every tracked file except .lattice/ (300 files) and the two PDFs: README.md, BRANDING.md, CLAUDE.md, agents.md, Makefile,
compose.yaml, .editorconfig, .gitignore, docs/ (SPEC.md, EXTENSIONS.md, branding-demo.html), data/sample_listings.json,
backend/, frontend/, scripts/ (smoke.sh, dist.sh, verify-dist.sh). CLAUDE.md and agents.md ship: SPEC §3 lists them in the
repo layout and §12 does not exclude them (they mention a .lattice/ the zip lacks; see open question). branding-demo.html
ships (SPEC §3; the README links it). verify-dist.sh ships too: harmless, and lets a reader re-verify. No personal paths or
emails found in the shipped set.

## scripts/verify-dist.sh <zip> (POSIX sh, set -eu, mirrors smoke.sh's trap and PASS/FAIL style)
Argument: exactly one existing `.zip` file (resolved to an absolute path). An extracted directory is not accepted: extracting
is part of the test. Uses only the zip and its own code, never the checkout it lives in.
1. Env: `unset DOCKER_DEFAULT_PLATFORM DOCKER_BUILDKIT COMPOSE_DOCKER_CLI_BUILD COMPOSE_FILE COMPOSE_PROJECT_NAME` (the operator
   shell exports the first three; the last two would point Compose at another file or project).
2. Preflight, exit 2 on refusal, no trap yet: tools docker, curl, unzip, perl, make, lsof, plus go, node, npm, golangci-lint
   for `make check`; `docker info`; `docker compose version`; `lsof -nP -iTCP:8080 -sTCP:LISTEN` and `:5173` both empty
   (else `port 8080 is in use`); no container with label `com.docker.compose.project=brillio` (never tear down a stack we did
   not start). Missing Go/Node/npm/golangci-lint -> exit 2 naming them: no in-container fallback. Ticket deviation, recorded by
   lattice comment: the Compose images cannot run `make check` (no make or golangci-lint; alpine has no C toolchain for
   `-race`), a container toolchain is new infrastructure SPEC does not ask for, and it would be untested on the only machine
   that runs this. The README's Tool table already lists Go, Node and golangci-lint for `make check`.
3. Zip inspection (`unzip -Z1`), FAIL -> exit 1: exactly one top-level folder; `README.md`, `compose.yaml`, `Makefile`,
   `scripts/smoke.sh` under it; no entry matching ERE `(^|/)(\.git|\.lattice|node_modules|dist|__MACOSX)(/|$)|\.db$|\.db-[^/]*$`
   `|\.[Pp][Dd][Ff]$|(^|/)\.DS_Store$|(^|/)\._|^/|(^|/)\.\.(/|$)` (any PDF fails here, per the acceptance wording).
4. `work=$(mktemp -d "${TMPDIR:-/tmp}/brillio-verify.XXXXXX")`, then install the trap; `unzip -q "$zip" -d "$work"`.
   Docker Desktop shares /var/folders and /private by default; override with `TMPDIR=...` if file sharing is customised.
5. Emoji scan of the extracted tree before anything runs (so it scans exactly the zip): see the Emoji section.
6. README check: `grep -F` each command the verifier runs in `$root/README.md`: `make up`, `make down`, `make smoke`,
   `make check`, `curl -s http://localhost:8080/healthz`, `# {"status":"ok"}`, `curl -s http://localhost:8080/api/cities`,
   the documented cities line, and the Springfield search URL. A README that stops documenting a step fails the verifier.
7. Run path, README commands only, from `cd "$root"`:
   - `make up` ("```sh make up```", "foreground; Ctrl-C stops it"): run as `make up </dev/null >"$work/make-up.log" 2>&1 &`
     (the verifier stands in for the user's terminal); set `stack=1`.
   - Wait (deadline 300 s, matching smoke's `--wait-timeout`; `date +%s` loop, sleep 2, no `timeout`): curl
     `http://localhost:8080/healthz` and `http://localhost:5173/` ("Then open http://localhost:5173") until both answer 2xx;
     if `kill -0 $up_pid` fails first, FAIL immediately. Every curl is `curl -q --noproxy '*' -s --max-time 10`.
   - The README's API examples: healthz body equals `{"status":"ok"}`; cities body equals
     `{"cities":["Chantilly","Fairfax","Manassas","Reston","Springfield","Vienna"]}`; the Springfield search is 200 with
     `"totalItems":4` ("returns four listings"; rank order is not asserted, it is date-sensitive).
   - `make down`, then wait for `make up` to exit (60 s, then kill) and for both ports to be free (30 s), so `make smoke`'s
     port preflight cannot race Docker's port release.
   - `make smoke` ("Compose stack up, HTTP checks, stack down"): the BC-13 a1..a8 assertions, from the zip's own smoke.sh.
   - `make check` ("lint and test both halves"), natively. The empty frontend/node_modules mount point left by Compose does
     not count as installed (Makefile sentinel), so `npm ci` runs on the host first.
8. Teardown `cleanup` (EXIT trap; INT/TERM/HUP/PIPE -> exit 130/143/129/141, same shape as smoke.sh): `set +e`; on failure
   first print `docker compose ps -a`, `logs --tail=60` and the make-up log tail; if `stack=1`, `(cd "$root" && make down)`
   (idempotent after smoke's own teardown; failure turns a pass into exit 1); kill and wait any live `make up`;
   `rm -rf "$work"` (guarded: non-empty and under the mktemp base). Background jobs ignore SIGINT in a non-interactive sh,
   so Ctrl-C reaches the trap, which owns the teardown.
Exit codes: 0 pass; 1 any check failed; 2 refused (usage, missing tool, Docker down, port held, brillio stack exists);
129/130/143 on signals. Output: `verify-dist: <step>` progress lines, one `PASS  <check>` / `FAIL  <check>` per check, make up
log on failure only, smoke and check output streamed, last line `verify-dist: PASS (<n> checks) <zip> <bytes> bytes` or
`verify-dist: FAIL (<n> failed)`.

## Emoji definition (perl 5.34.1 on this host, Unicode 13.0)
Emoji = `[\p{Extended_Pictographic}\p{Emoji_Presentation}\x{FE0F}\x{20E3}\x{E0020}-\x{E007F}]` (pictographs, default-emoji
presentation incl. regional-indicator flags, the emoji variation selector, the keycap combiner, tag characters), minus the
text symbols `\x{00A9}\x{00AE}\x{2122}` and the arrows block `\x{2190}-\x{21FF}` when they stand alone.
- Not `\p{Emoji}`: it matches ASCII 0-9, # and *. A text symbol followed by U+FE0F (U+2195 U+FE0F, or the keycap
  1 U+FE0F U+20E3) is still caught by the FE0F and 20E3 terms.
- Not flagged (not ExtPict): U+2192 (rightwards arrow), box drawing U+2500-257F, U+25BE, U+2713 (check mark), U+00A7,
  U+00B7, U+2014, U+2264, U+2265. Flagged: U+2605 (star), U+2714 (heavy check mark).
- Unicode 13 ExtPict pre-reserves U+1F000-1FAFF and U+1FC00-1FFFD, so later emoji are caught by this older perl.
- Implementation: inline perl, File::Find over `$root`, one process. Each file is read raw; empty files skipped; `-B` files
  skipped and listed as `SKIP  binary <path>` (the zip has none today); lines decoded with `Encode::decode("UTF-8", ...)`
  (bad bytes become U+FFFD, never a match). Also scans every relative path name. Reports `emoji: <file>:<line> U+XXXX`.
- Self-test before scanning: U+1F600 must match and U+2192 must not, else exit 2 (a perl without the properties must never pass).
- Probe result on the current tree (prototype, all 133 extracted files): 0 hits. Raw ExtPict alone hits only U+2195 (up
  down arrow) at CLAUDE.md:46 and agents.md:41 (the Lattice status diagram), excluded by the arrows rule. Injected U+1F600,
  U+2705, U+2195 U+FE0F, the keycap 1 U+FE0F U+20E3, U+2605 and the flag U+1F1FA U+1F1F8 were all caught.

## Tests (mapped to acceptance)
- A1 `make dist` then verify passes: T1 `make dist` exits 0 and prints name, size, counts; T2 `scripts/verify-dist.sh
  dist/<zip>` exits 0 with the PASS line, run from the operator shell with the three DOCKER/COMPOSE vars still exported.
- A2 README sufficient: T3 the README-check step passes; T4 review confirms every run-path step is a quoted README command and
  the script reads nothing from the checkout (only `$zip` and `$work`).
- A3 no forbidden content: T5 the zip-inspection PASS; T6 `unzip -Z1 <zip> | grep -E '<forbidden ERE>'` empty in the evidence.
- Negative (scratch copies; `zip -q copy.zip <prefix>/x` adds an entry): N1 a file holding U+1F600 -> exit 1 at the emoji step,
  no container started; N2 one zip each with `<prefix>/.lattice/x`, `<prefix>/data/listings.db`, `<prefix>/x.pdf` -> exit 1 at
  inspection; N3 port held (`python3 -m http.server 8080 --bind 127.0.0.1 &`) -> exit 2, no temp dir created, listener killed.
  Should: N4 Ctrl-C (`kill -INT`) during the health wait -> no brillio containers, no `brillio-verify.*` dir left;
  N5 an untracked non-ignored file -> `make dist` warns and still builds; N6 `make dist` inside an extracted copy fails cleanly.
- Lint: `sh -n` and `dash -n` both scripts; `shellcheck -s sh` (0.11.0 installed) on dist.sh, verify-dist.sh, smoke.sh;
  `make -n dist`; `make check` on the branch stays green.

## Validation (delegator; only after BC-14, BC-15, BC-16 land, per the press-ahead rule)
Merge final master into the branch; clean tree; ports 8080/5173 free (`lsof -nP -iTCP:<p> -sTCP:LISTEN`) and no brillio
containers. Run `make dist`, `unzip -l <zip> | tail -1`, the T6 grep, `scripts/verify-dist.sh <zip> 2>&1 | tee` a scratch log,
N1-N3. Attach with `--role validation`: verify-dist log, zip name and byte size, `unzip -l` summary, negative-check outputs, and
a final lsof plus `docker ps -a` showing nothing left. Expect ~8-12 min (two stack starts, each re-running npm ci and the Go
module download because `make down -v` drops the volumes, then native npm ci, lint and race tests).

## Risks
- Docker file sharing for $TMPDIR (default shared on this host; failure is loud; `TMPDIR=` override).
- Port release lag after `make down` (wait-for-free step). Background `make up` lifetime (trap kills and waits it).
- README edits collide with BC-15 rework: keep to the layout block; rebase on final master before evidence.
- New files hidden by the host global gitignore: none expected (`*.sh`), checked with core.excludesFile=/dev/null.
- Final zip is built from final master: same tree as the branch head minus .lattice, so the orchestrator rebuilds after landing.

## Open question (human, non-blocking; default applies)
Ship CLAUDE.md and agents.md? Default yes (SPEC §3 and §12). They are Lattice agent instructions that say `.lattice/` exists,
which the zip leaves out. Excluding them would be `SPEC edit: §12`.

## Plan-Review Cycle 1 Resolutions (AUTHORITATIVE — overrides earlier text on conflict)
Review: art_01M4WEMR2QYGFJGWXE9M1KYK3E (fable sub-agent), PASS-WITH-NITS. Every finding is obvious; all applied.
1. Major, dist/ missing (make dist step 4): `mkdir -p dist` before `git archive -o dist/.$name.zip.tmp`.
2. Minor, reaping `make up` under set -e (verify step 7 bullet 4, step 8): after `make down`, poll `kill -0 "$up_pid"` against a
   60 s deadline (kill it if still alive), then `wait "$up_pid" || true`. Same reap in cleanup. Never a bare `wait` under errexit.
3. Minor, body comparisons (step 7 bullet 3): bodies end with a newline (respond.go writeJSON). Compare after whitespace
   compaction (`tr -d '[:space:]'`, the same `compact` idea as smoke.sh) on both sides; the Springfield check greps the compacted
   body for `"totalItems":4`.
4. Minor, bring-up deadline (step 7 bullet 2): one named variable `UP_DEADLINE=400` seconds (api 180 s start_period + web 180 s
   start_period + margin). smoke.sh (BC-13's file) is not touched.
5. Minor, dirty-tree warning (make dist step 3): run `git -c core.excludesFile=/dev/null status --porcelain -- . ':(exclude).lattice'`
   so the always-moving board never triggers it. Evidence bar: clean outside `.lattice/`.
6. Minor, Risks line 131: read "merge final master", never rebase.
7. Nit, lsof under errexit (verify step 2): `out=$(lsof -nP -iTCP:8080 -sTCP:LISTEN 2>/dev/null || true)`; refuse when non-empty.
8. Nit, Ctrl-C (step 8): correct the reasoning in the script comment: `sh`/`make` inherit SIG_IGN for INT in the background job,
   but `docker compose` re-enables SIGINT, so on Ctrl-C compose may start its own graceful stop alongside the trap's `make down`.
   Outcome unchanged (down is idempotent, exit 130). On the interrupt path a non-zero `make down` does not change the exit code;
   N4 tolerates compose's stop output.
9. Nit, TMPDIR trailing slash (step 4): `base=${TMPDIR:-/tmp}; base=${base%/}; work=$(mktemp -d "$base/brillio-verify.XXXXXX")`;
   the rm -rf guard checks that `$work` starts with `$base/brillio-verify.`.
Open question (complex, sent to the human via needs-human; non-blocking for implementation): ship CLAUDE.md and agents.md or
exclude them with a SPEC §12 edit. Implement with the default (ship). Keep the exclusion pathspecs in one list in scripts/dist.sh
so the answer is a one-line change plus the §12 sentence; if the answer is "exclude", also drop the arrows/text-symbol carve-outs
only if the stricter scan then still passes on the shipped tree (otherwise keep them).
