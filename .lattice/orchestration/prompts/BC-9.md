# Ticket card — BC-9

```
TICKET=BC-9
TASK_UUID=task_01M4VXK2YVDRXZ0BA1B8DW3CEQ
TITLE=FilterBar: seven fields, client-side validation with exact copy, server error mapping, Search and Clear
WORKTREE=/Users/forrest/Code/brillio-worktrees.noindex/bc-9-filterbar
BRANCH=feat/BC-9-filterbar
REVIEW_BASE=feat/BC-8-frontend-core   (switch to master once BC-8 lands)
ANCHOR=feat/BC-8-frontend-core @43e8824 (BC-8 at review 02:50Z, not yet landed)
PORTS=none (component tests under jsdom; no running surface until BC-12)
AREA=area:9
```

Read `/Users/forrest/Code/brillio/.lattice/orchestration/prompts/delegator-brief.md` now and follow it.

## Wave notes shared by BC-9, BC-10 and BC-11 (binding)

1. **Press-ahead anchor.** Your branch was cut from BC-8's in-review branch. Review your own diff against
   `REVIEW_BASE`. If BC-8 gains commits before it lands, `git merge feat/BC-8-frontend-core`; when it lands,
   `git merge master` and switch REVIEW_BASE to `master`. Never rebase.
2. **BC-9, BC-10 and BC-11 run in parallel** on the same anchor and own disjoint component folders. Each component is a
   folder with `Component.tsx`, `Component.module.css`, `Component.test.tsx` (SPEC §7.4). Do not touch another
   ticket's components, `App.tsx` (BC-12 wires the page) or the shell styles.
3. **Consume BC-8's helpers; never fork them.** Types from `src/api`, formatters from `src/format`, fixtures from
   `src/test/fixtures.ts`. If a formatter or fixture you need is missing, add it to BC-8's file (additive only, with a
   test) and name it in your completion comment; the Orchestrator unions these on landing.
4. **Design truth is BRANDING.md** (§2.1 tokens via `var(--…)` from `src/styles/tokens.css`, never hard-coded hex;
   §2.2 oxblood only in its five jobs; §3 type; §4.4 breakpoints at 760px; §5 component anatomy; §6 motion;
   §7.2 strings verbatim; §8 don'ts). Compare against `docs/branding-demo.html`. Do not edit BRANDING.md; a needed
   design change goes to `lattice needs-human`.
5. **Gate:** full `make check` plus `npm run build`.
6. **Validation:** N/A until BC-12 wires the page. The validator cites the vitest output listing this ticket's cases
   (and, optionally, a jsdom render snapshot description). No browser needed.

## BC-9 notes

7. Status is the seventh field (SPEC §7.4, §13.1): label `Status`, options Any / Active / Pending / Sold, placeholder
   `Any`. BRANDING §5.2's table does not list it yet; follow §5.2's seventh-field rule, do not edit BRANDING.md.
8. Expose stable label text and `aria-describedby` ids so BC-12's integration tests can find fields by label.
