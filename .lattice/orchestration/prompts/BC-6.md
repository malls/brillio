# Ticket card — BC-6

```
TICKET=BC-6
TASK_UUID=task_01M4VXK2J03TXW4B4D9KGWQ145
TITLE=OpenAPI 3.1 document, embedded and served at /api/openapi.yaml, with param sync test
WORKTREE=/Users/forrest/Code/brillio-worktrees.noindex/bc-6-openapi
BRANCH=feat/BC-6-openapi
REVIEW_BASE=master   (review only after BC-5 has landed — note 1)
ANCHOR=feat/BC-5-search-api @75530c0 (BC-5 in review) + master @0e6730f merged at creation
PORTS=API 18006 (PORT=18006, DB_PATH=:memory:)
AREA=area:9
```

Read `/Users/forrest/Code/brillio/.lattice/orchestration/prompts/delegator-brief.md` now and follow it.

## BC-6 notes (binding)

1. **Anchor.** Your branch was cut from BC-5's in-review branch with current master merged in, so a diff against either
   one alone shows foreign files. If BC-5 gains commits, `git merge feat/BC-5-search-api`. **Do not enter `review` until
   BC-5 has landed on master** (`git log --oneline master | grep BC-5`); if you get there first, mail the Orchestrator
   `HOLD BC-6 waiting for BC-5` and wait. Then `git merge master`, re-run the gate, and review against `master`. Never rebase.
2. **Contract as BC-5 left it.** Read BC-5's SPEC §5.3/§5.5 edits: the field-error code enum includes the new
   `malformed_query` (pseudo-field `query`, run-state decision 03:00Z) and the §5.4 example score is the corrected
   value. The document's code enum and examples must match the SPEC text exactly.
3. The sync test reads the param allowlist from BC-5's handler (BC-5 kept it readable) and asserts set equality with the
   search operation's parameters, both directions. Register `GET /api/openapi.yaml` in `NewRouter` with one line.
4. `go:embed` cannot reach a parent directory: place the embedding Go file where `openapi.yaml` lives (SPEC §3 puts it at
   `backend/openapi.yaml`) or move the yaml next to the embedding package and update SPEC §3 in-branch with a comment.
5. Run the external validator once (`npx --yes @redocly/cli@latest lint backend/openapi.yaml`) and paste its output in a
   `lattice comment`; do not add it to `make check`.
6. **Gate:** full `make check`. **Validation:** `curl -si :18006/api/openapi.yaml` (200, `application/yaml`) plus the validator
   output, attached with `--role validation`.
