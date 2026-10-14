# Ticket card — BC-6

```
TICKET=BC-6
TASK_UUID=task_01M4VXK2J03TXW4B4D9KGWQ145
TITLE=OpenAPI 3.1 document, embedded and served at /api/openapi.yaml, with param sync test
WORKTREE=/Users/forrest/Code/brillio-worktrees.noindex/bc-6-openapi
BRANCH=feat/BC-6-openapi
REVIEW_BASE=__REVIEW_BASE__
ANCHOR=__ANCHOR__
PORTS=API 18006 (PORT=18006, DB_PATH=:memory:)
AREA=__AREA__
```

Read `/Users/forrest/Code/brillio/.lattice/orchestration/prompts/delegator-brief.md` now and follow it.

## BC-6 notes (binding)

1. **Anchor.** If `ANCHOR` names `feat/BC-5-search-api`, your branch was cut from BC-5's in-review branch. Review your own
   diff against `REVIEW_BASE`; if BC-5 gains commits, `git merge feat/BC-5-search-api`; when it lands, `git merge master`
   and review against `master`. Never rebase.
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
