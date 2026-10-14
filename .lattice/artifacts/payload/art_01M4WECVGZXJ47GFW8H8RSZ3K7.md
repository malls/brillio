Code review BC-14 (reviewer: fable sub-agent, decision 04:02Z)

Verdict: PASS-WITH-NITS
Reviewed HEAD: 10133f83f5eb6fdb034280a4e80aea413f8e6854
Base: master 0b24829a014593929e29b45a8e77f3a9949116b6
Gate: make check rc=0 (398 FE tests / 23 files; eslint --max-warnings 0, tsc -b, prettier --check clean; backend 5 packages ok), npm run build rc=0. git status clean afterwards.

Findings:
- [Nit] Hover guard differs textually across the four text buttons: NoResults.module.css:47 and ErrorState.module.css:39 use `.action:hover`; FilterBar.module.css:197 `.clear:hover:not(:disabled)`; Pager.module.css:49 `.button:hover:not(:disabled, [aria-current='page'])`. No computed difference (those two buttons are never disabled); pre-existing, outside D4. Recommendation: none required; fold into BC-19 if the recipe changes.
- [Nit] D3 has no overflow fallback for a long unbreakable location item: ListingCard.tsx:21 makes each item one unbreakable run; ListingCard.module.css:37-45 sets no overflow-wrap. 375px text column ~251px, longest seed location ~130px, so nothing overflows today. Recommendation: no change for BC-14; validator V4(k) covers seed data.
- [Nit] Commit subjects of 91 and 87 chars (1543178, 10133f8). No length rule in the ticket. Recommendation: none.

Verified: eslint directive consumed (rule reports on the tabIndex JSXAttribute; --report-unused-disable-directives-severity error clean); T13 fails without tabIndex (jsdom has no focusable scrollers); D4 computed parity by static reasoning (button { font: inherit } -> body var(--font-ui); 13px/1/500 tokens; reset properties inherit initial values); no literal NBSP in diff, NBSP = ' '; separator breaks only after the dot; formatters and StatusLine untouched; no emoji, trailer on all three commits, no .lattice/ paths; SPEC §7.6 deviation commented; scope = 13 files all named in plan §3/§7; commit split matches plan §7.

Plan conformance: D1-D7 all implemented as planned (D1 SearchPage.tsx:110-117, shell.module.css:29-32, SPEC:485-486; D2 ListingList.tsx:17-22 + test :24-30; D3 ListingCard.tsx + tests; D4 token longhands + cursor default; D5 LoadingLine.module.css:20-23; D6 global.css:23-24; D7 FilterBar.module.css:2-4).

Ready for in_validation; no rework needed.