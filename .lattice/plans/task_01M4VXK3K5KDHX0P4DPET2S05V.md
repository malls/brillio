# BC-14: Branding conformance pass: audit UI against BRANDING.md and the demo, fix deviations

Audit the running UI against BRANDING.md and the reference demo, and fix every deviation that is a frontend bug rather than a design change. Design: BRANDING.md in full, with special attention to §2.2 (the five oxblood jobs), §3.3 (type rules), §4 (shell and breakpoints), §5 (every component), §6 (motion), §7 (copy and formatting), §8 (don'ts). Reference: docs/branding-demo.html. Spec: docs/SPEC.md §13 for known engineering deltas.

Why: three component tasks and a wiring task were built in parallel by different agents; this pass is the one place the whole surface is checked as one system by fresh eyes before the README and packaging.

Scope:
- With api and web running, compare each component to the demo at a width above 760px and at or below 760px.
- Checklist, each recorded pass or fixed in a task comment: oxblood appears only in the five §2.2 jobs; no radius except the rank pill; no shadows, gradients, translucency, icons, images; Bodoni never below 18px; tabular figures where digits sit in columns; uppercase only in labels and the wordmark; every §7.2 string verbatim; every §7.1 format; the document never scrolls; pinned header, filter bar, status line and pager; loading line is the only animation and is static under reduced motion; disabled looks use ink-3 and stone, not opacity; keyboard-only pass through search, pagination and Try again; `grep -rP "[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}]"` over the repo finds nothing.
- Fix small CSS, markup and copy deviations in place. Anything that would require changing BRANDING.md itself (for example the Status field row, SPEC §13.1) is not changed here: raise it with `lattice needs-human` naming the section.

Acceptance:
- A checklist comment on the task listing every BRANDING section with pass or fixed.
- Screenshots at both breakpoints for the default search, the no-results state, the invalid-input state and the error state, attached to the task.
- `make check` still green.

Validation: the screenshots and the keyboard pass are the evidence; attach with --role validation.
