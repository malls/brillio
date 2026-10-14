Plan review BC-14 (reviewer: fable sub-agent, decision 04:02Z)

Verdict: PASS-WITH-NITS
Reviewed plan: task_01M4VXK3K5KDHX0P4DPET2S05V.md @ HEAD 0b24829

The audit table covers every BRANDING section (§0-§10) and every ticket checklist item (C1-C14); 14 "pass" claims spot-checked hold against code and demo. D2-D6 are correct, minimal and inside card note 1. Validation section is executable on this machine. Two "obvious" amendments needed before in_progress; the rest are nits.

Findings:
- [Major] Two of the four handover comments are not triaged (BC-12 validation 04:57Z item 7: cards clip flush against the StatusLine row with no divider or fade, same construction as demo:42-45 -> skip-with-reason or owner question; BC-12 implementation 04:57Z: FilterBar.module.css:2-3 comment says the bar is "a direct child of the shell" while SearchPage.tsx:98-108 wraps it in a row-2 div -> one-line comment fix). Recommendation: add S9 and D7.
- [Major] D1's oxblood ring on <main> is a sixth oxblood placement beyond §2.2 job 3 ("focus ring on buttons and links"), yet C1/A3 read "pass". D1 itself is right (tabIndex=0 is engineering per SPEC §7.6; a ring-less or browser-blue focus would be worse). Recommendation: C1 and A3 become "pass + D1 ring (O12), D5 (O6)"; raise the O1-O14 needs-human before the implementer starts; keep the SPEC §7.6 sentence to the engineering fact only.
- [Minor] V6(5) expected string wrong: after step (2) targetBudget=450000 persists, so Keyword pets reads "7 results · ranked for $450,000". Fix the string or Clear first.
- [Minor] V6(1): Chrome 130+ makes scrollers Tab-reachable by default, so headless Chrome reaches <main> regardless of tabIndex. Log tagName, tabindex==="0" and outline-color rgb(122, 31, 43); state that the WebKit path is proven by the new jsdom unit test.
- [Minor] D2 misdescribes the comment at ListingList.tsx:16-18 (it says the section itself is the polite live region). Rewrite it explicitly; rename the test at ListingList.test.tsx:24.
- [Minor] C14's grep -P does not run on macOS grep. Name an executable command (perl -CSD over git ls-files minus PDFs); note binary-noise hits inside .lattice PNGs. Zero hits outside .lattice/ confirmed; commit messages clean.
- [Nit] D3: write NBSP as ' ', not an invisible literal.
- [Nit] V5: compare against a non-current page number; font-variant-numeric deliberately not compared.
- [Nit] V0: DB_PATH in the session scratchpad; /healthz returns {"status":"ok"}.
- [Nit] D1 also adds a Tab stop in no-results and error states; mention in O12.
- [Nit] D6 vs S5: keep D6, record it under §4.3/§8 as fixed (D6).

Assessment: all 14 O items are genuine owner questions; S1-S8 sound (S3 verified at 13ae060); text-button no-consolidation decision honours card note 2; gate risk low (jsx-a11y/no-noninteractive-tabindex fires on <main tabIndex={0}>, so the disable directive is consumed; reportUnusedDisableDirectives warn + --max-warnings 0 would fail an unused one). Headless Chrome over CDP for evidence with the c11 panel as cross-check is the right split. Rows 44-46 covered by V4(i), V4(a-b), V3+V7.

Spot-checks: A2 tokens byte-identical; A7 fonts URL; A8 type scale values; A12/C2/C3/A25/C12 greps; A16 focus/invalid/disabled; A13 7 columns at 1040 -> O1 framing correct; A5/O8 demo:102; A19/A23 aria; A21 pager; A22 text-button computed parity; C14/S8 counts; D2/D3 test targets; V3/V6 data (12 listings, 6 cities, 7 'pets', page 2 = No. 11-12); lint config for D1.