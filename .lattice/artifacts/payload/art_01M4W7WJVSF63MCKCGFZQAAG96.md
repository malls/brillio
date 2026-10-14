Plan review (fresh-context Opus fallback; headless skipped, plan > 250 lines). Verdict: APPROVE-WITH-CHANGES.
[Major] .action copied BC-9 .clear with +tracking; must use the Orchestrator shared text-button recipe (no letter-spacing).
[Major] Ticket acceptance asks for a visual check vs docs/branding-demo.html; jsdom cannot provide it — add a throwaway preview + screenshots or get sign-off to defer.
[Minor] .monogram missing font-family var(--font-display).
[Minor] Ticket says "visually clear heading"; plan uses a visually hidden h2 — record the interpretation.
[Minor] B7 is not the only capital Pets (A4 has "Pets welcome.").
[Minor] No node_modules in worktree; run npm ci before prettier.
[Minor] D8 reason wrong (demo .msgbox centers vertically); copy rules, do not import across folders.
[Nit] ErrorState missing keyboard and type=button tests.
[Nit] 0.8731 attributed to SPEC 5.4 (BC-5 corrects it).
[Nit] helper keywordTokens collides with prop name; rename.
[Nit] text-transform uppercase redundant on statusLabel.
Passed: formatter signatures, every token exists, expected test strings, prop shapes, lint claims, D1, D2 grid, emphasis algorithm, scope and size.