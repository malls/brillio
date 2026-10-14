Verdict: PASS-WITH-NITS
Reviewed HEAD: e5f83170afe5969647248ba9c0ef7306342d32a8 (contains master 7fb72bf)
Reviewer: fable sub-agent (decision 04:02Z), cycle 2, judged against the Orchestrator scope bound 06:12Z
Checks run:
- git diff master...HEAD -- . ':!.lattice' -> 5 files: scripts/dist.sh, scripts/verify-dist.sh (100755), Makefile BC-17 block only, README.md layout block only (+3/-1), docs/SPEC.md §3 (+2) and §12 (one sentence)
- sh -n, dash -n, shellcheck -s sh 0.11.0 on both scripts -> clean; make -n dist -> "sh scripts/dist.sh"
- make dist (host core.excludesFile active) -> dist/brillio-listings-20261014.zip, 328894 bytes, 135 files; "excluded 368 tracked files (366 .lattice/, 1 Coding_Test_Question_Paper.pdf, 1 Instructions.pdf)"
- unzip -Z1 -> one top-level folder; forbidden grep empty; diff vs git ls-tree HEAD minus .lattice/ is exactly the two PDFs; CLAUDE.md and agents.md present
- proc_tree unit test on a synthetic 4-process tree under bash 3.2 and dash -> exact set, deepest first, never $$ or $PPID, empty for a dead pid
- dist.sh trap pattern under bash 3.2 and dash -> temp file removed, rc 143
- verify-dist under /bin/sh with the three DOCKER/COMPOSE vars exported -> PASS (13 checks), exit 0, 65 s; same under dash -> PASS (13 checks), 66 s; nothing left
- interrupt re-test (RUN sleep 90, TERM at +5 s after make up) -> exit 143 13 s after TERM; at exit and +133 s no brillio containers, volumes, networks, compose processes, temp dirs; ports free
- N3 with the curl probe: http.server on 8080 -> exit 2 "port 8080 is in use"; raw nc -l on 5173 -> exit 2 in under 1 s; bad args -> exit 2
Cycle-1 items: 1 Major fixed (proc_tree :151-161, reap_up :168-185, cleanup :220, :228-232); 2 Minor fixed (port_free :63-67, lsof dropped); 3 fixed; 4 fixed; 141 fixed; SKIP 5 and 6 left alone; stated deviations correct.
Findings:
- [Minor] Linux host: root-owned bind-mount paths would fail the native make check and the rm -rf - scripts/verify-dist.sh:432-437, :234-236 (root cause compose.yaml web volumes) - on a Linux daemon without uid mapping, frontend/node_modules mount point and data/listings.db are created as root; native npm ci gets EACCES and rm -rf leaks. Not reproduced on macOS. Inherited from BC-13's compose design. - Route as a new task; not blocking.
- [Nit] leftovers cannot see anonymous volumes, comment overstates - scripts/verify-dist.sh:187-193 - make down -v removed it in every run; no leak. - Amend the comment to "named volumes".
- [Nit] Springfield assertion is a substring match - scripts/verify-dist.sh:405 - "totalItems":4 also matches 42. - Use '"totalItems":4,' if exact.
- [Nit] dist.sh totals can disagree with the breakdown if a tracked glob-excluded file ever existed - scripts/dist.sh:75-84, :90 - today 368 = 366+1+1. - None required.
- [Backlog] PID reuse in reap_up - scripts/verify-dist.sh:171-183 - recycled pid after make exits could be TERMed; never observed. - Check ps -o ppid= -p "$up_pid" equals $$.
- [Backlog] Signal latency for a TERM/HUP sent to the script alone - :395, :426, :433, :220 - trap runs after the foreground child returns; outcome always clean. - Optional.