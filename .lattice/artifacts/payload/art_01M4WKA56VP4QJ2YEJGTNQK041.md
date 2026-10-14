Verdict: PASS-WITH-NITS
Reviewed HEAD: 70fa0e00ed04050416cc936e402b4b460566ecb2
Reviewer: fable sub-agent (decision 04:02Z), cycle 4 doc-delta review
Scope: delta 1cbdb63..70fa0e0 on top of the cycle-3 PASS-WITH-NITS (art_01M4WJXG4BSY8Y6T2N20VEA0R9)
Checks run:
- HEAD 70fa0e0 on feat/BC-17-release-dist, contains master 7fb72bf, tree clean
- git diff --stat 1cbdb63 70fa0e0 -> README.md (+2) and scripts/verify-dist.sh (+2/-1, comment only); 0 README lines removed or changed
- the 11 README needles from scripts/verify-dist.sh:329-331 -> all present
- no new non-ASCII in added lines; sh -n, dash -n, shellcheck -s sh -> clean
- Facts checked against compose.yaml:41-45, :76-78, backend/Dockerfile.dev (no USER), Makefile:27-39 sentinel, backend/cmd/api/main.go:56-69, internal/store/store.go:54, backend/.air.toml -> the Linux note is accurate; chown quoting correct and safe
- Placement README.md:19 in "Run it", between the DOCKER_DEFAULT_PLATFORM note and "Without Docker"
- Voice (BRANDING §7 plain product) -> conforms; see Nit
Cycle-3 nit: fixed (verify-dist.sh:189-191 "make down (docker compose down -v) removes it ...")
Findings:
- [Minor] verify-dist.sh runs make check after make up, so on a Linux Engine host it hits the leftover the note describes - scripts/verify-dist.sh:384,438 - pre-existing, outside this delta, README makes no Linux claim for verify-dist; not a blocker. - New task or BC-23: run make check before make up in verify-dist, or state the macOS-only expectation.
- [Nit] sudo is not needed to clear the leftovers - README.md:19 - rm -rf frontend/node_modules data/listings.db works unprivileged; the chown as written is correct and matches the decision. - Optional.
- [Nit] parenthetical can be misread - README.md:19 - "(no Docker Desktop or rootless uid mapping)". - Optional: "(no uid mapping from Docker Desktop or rootless mode)".
- [Nit] two figurative phrases against BRANDING §7 plain voice - README.md:19 - "stays yours", "hands both back". - Optional: plainer verbs.