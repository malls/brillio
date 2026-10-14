# BC-17: Release packaging: make dist zip and fresh-directory verification

Produce the zip the recruiter receives and prove it works from a fresh directory using only the README. Spec: docs/SPEC.md §12; Instructions.pdf ("send us a zip of your project, code plus a short README").

Why: the last mile. The zip must run on a machine that has only Docker, and the README must be sufficient on its own.

Scope:
- `make dist`: builds dist/brillio-listings-YYYYMMDD.zip from a clean tree excluding .git, node_modules, *.db, .lattice, dist, the two handout PDFs, and OS cruft. Prints the size and the file list summary.
- scripts/verify-dist.sh: extracts the zip into a fresh temporary directory, runs `docker compose up -d`, waits for health, runs the smoke checks from BC-13, runs `make check` (natively if Go and Node are present, otherwise inside the containers), tears down, and cleans up. Fails on any emoji codepoint in the extracted tree.
- Fix whatever the fresh-directory run surfaces (a missing file in the zip, a README step that assumes local state, a path that only worked in the original checkout).

Acceptance:
- `make dist` then `scripts/verify-dist.sh dist/<zip>` passes end to end.
- The README in the zip is sufficient: the verifier follows only its commands.
- Zip contains no .lattice, .git, node_modules, databases or PDFs.

Validation: the verify-dist output is the evidence; attach with --role validation along with the zip size.
