# BC-16: Extension playbook: how each likely live-session requirement lands (docs/EXTENSIONS.md)

Spike: write docs/EXTENSIONS.md, the playbook for the requirements the interviewer is likely to add live. Spec: docs/SPEC.md §9. Design hooks: BRANDING.md §5 ("where a new requirement goes" notes in each component) and §10.

Why: the live session is explicitly about extending a design already built, narrating the reasoning. A rehearsed playbook with the exact seams turns a surprise into a demo. This is also where the owner rehearses before the session.

Scope (docs/EXTENSIONS.md):
- One entry per §9 item: the requirement in one line; the exact files and functions to touch (verified to exist by grep at writing time); the API change (params, response fields) and the UI change (which BRANDING §5 slot); the tests to add; a minutes estimate. Keep each entry under 25 lines.
- A rehearsal section working three of them through as mini-plans the owner can narrate aloud: cross-source dedupe (stage in the §6 pipeline, key choice, which record wins, the `2 sources` meta item), geo radius (bounding box in SQL, haversine in Go, two new params and validation), sort options (one change in domain ordering, a select in the status line per BRANDING §5.7's precedent).
- A short "what I would say" paragraph for the two trade-offs most likely to be questioned: scoring weights and SQL-versus-Go ranking.

Acceptance:
- Every referenced symbol exists in the code (a reviewer greps each one).
- Plain product voice, no emoji, under 300 lines.

Validation: N/A (documentation). Record a one-line justification with --role validation.
