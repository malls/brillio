Verdict: FAIL (implementation-level)
Reviewed HEAD: 12ba0dfb7f31a1ff210f4823ef5410eff20c5e0b
Reviewer: fable sub-agent (decision 04:02Z)
Checks run:
- git diff master...HEAD -- . ':!.lattice' -> 5 files: scripts/dist.sh (new, 100755), scripts/verify-dist.sh (new, 100755), Makefile (BC-17 block only), README.md (layout block only, +3/-1), docs/SPEC.md (§3 two lines, §12 one sentence); no non-ASCII beyond SPEC's own section sign
- sh -n / dash -n both scripts -> clean; shellcheck -s sh (0.11.0) -> clean; make -n dist -> "sh scripts/dist.sh"
- make dist -> dist/brillio-listings-20261014.zip, 328187 bytes, 135 files; scripts keep 0755; run with the host core.excludesFile active
- unzip -Z1 vs git ls-tree HEAD minus .lattice/ -> identical except the two PDFs; forbidden grep empty; CLAUDE.md and agents.md present (operator ruling); one top-level folder
- scripts/verify-dist.sh end to end with the three DOCKER/COMPOSE vars exported -> PASS (13 checks), exit 0, 72 s; nothing left behind
- Emoji scanner on 22 injected files -> all emoji caught (incl. tag, keycap, flag, ZWJ, after invalid bytes, NUL file, file and dir names); text symbols passed; binary SKIP; symlink skipped
- Negative zips: .lattice/x, data/listings.db, x.pdf -> exit 1 at inspection; loose README beside folder -> exit 1; emoji file -> exit 1 at emoji step, no container; port 8080 held -> exit 2; make dist in extracted copy -> exit 1
- TERM 0.5 s after make up (cached build) -> exit 143 in 3 s, clean
- TERM during a 90 s image build (scratch zip with RUN sleep 90) -> exit 143 after 62 s with `docker compose up --build` orphaned (ppid 1); containers, volumes and network then created; next verify-dist refuses (exit 2)
- Process-tree probe -> kill -TERM <make> ends make and sh; grandchild reparented to PID 1 keeps running
- Plan Cycle-1 resolutions 1-9 and operator ruling -> all present
- Acceptance A1/A2/A3 -> yes
Findings:
- [Major] Interrupt during a long image build leaves an orphaned compose client that rebuilds the stack after teardown - scripts/verify-dist.sh:174-196 (cleanup order) and :145-157 (reap_up) - kill "$up_pid" reaches make and its sh -c, but docker compose up --build is a grandchild reparented to PID 1; a TERM/HUP/kill -INT to the script while the api image builds > ~60 s (first run on a fresh machine) leaves an orphan that creates containers/volumes/network from deleted bind-mount sources; later runs refuse with exit 2. - Signal the recipe's descendants first (pgrep -P or a ps -A -o pid=,ppid= walk), call reap_up with a short deadline before make down, then make down, then rm -rf; or loop make down while brillio containers exist until make exits, bounded by UP_DEADLINE. Re-test with RUN sleep 90 + TERM at +5 s.
- [Minor] lsof port checks are user-scoped on Linux - scripts/verify-dist.sh:79-82, :335-339 - a root-owned listener passes the preflight; no false PASS (make up fails later). - Use the curl exit-7 probe smoke.sh:41-45 already uses, or document the limitation.
- [Nit] Excluded-files line counts .lattice - scripts/dist.sh:78 - "excluded 368 tracked files" reads as if much of the repo was dropped. - Print ".lattice/ (N) and the two handout PDFs" or count per pathspec.
- [Nit] No cleanup of the temp zip on interrupt - scripts/dist.sh:57-63 - optional trap rm -f "$tmp".
- [Nit] Native-tool refusal precedes the zip-only checks - scripts/verify-dist.sh:73-76 - plan-conformant; a Docker-only machine would get 9 of 13 checks if the tool check moved before make check.
- [Nit] Emoji definition passes dingbat marks (U+2713/U+2717, not Emoji in Unicode; none shipped); header exit list omits 141 (:10-12 vs :202); an emoji in a directory name is reported once per descendant (:248). - None required.