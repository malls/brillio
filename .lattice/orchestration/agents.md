# Agents — brillio run

Active table is overwritten each dispatch tick (Lattice + `c11 tree` are ground truth).

## Active

| Role | Ticket | Panel ref | Area ref | Branch | Worktree | Phase | Last seen | Spawned at |
|---|---|---|---|---|---|---|---|---|
| Orchestrator | all | panel:6 73809EE8-EE77-4F1E-8EE2-8E8B05308939 | area:6 D887B46C-F2FA-4097-ABB0-B453319B8C6B | master | /Users/forrest/Code/brillio | dispatch | 2026-10-14T01:23Z | 2026-10-14T00:45Z |
| Master Validator | all | panel:13 AFD95ABD-DC58-4303-BEBE-8591321113A7 | area:6 D887B46C-F2FA-4097-ABB0-B453319B8C6B | — | /Users/forrest/Code/brillio | boot | 2026-10-14T01:23Z | 2026-10-14T01:20Z |
| Delegator | BC-2 | panel:16 E50A586D-7B7D-4FBA-89A0-918D59278F9A | area:11 C4DECFF8-1150-4E6E-9C17-277B9378CA7B | feat/BC-2-backend-scaffold | /Users/forrest/Code/brillio-worktrees/bc-2-backend-scaffold | review | 2026-10-14T01:23Z | 2026-10-14T01:22Z |
| Delegator | BC-7 | panel:17 F93C149A-6AA9-4705-A195-09CEABBE1680 | area:9 29806806-32FF-427B-A78B-FA87D1CB29FD | feat/BC-7-frontend-scaffold | /Users/forrest/Code/brillio-worktrees/bc-7-frontend-scaffold | plan-review | 2026-10-14T01:23Z | 2026-10-14T01:22Z |
| Delegator | BC-3 | panel:18 C207406F-8A23-49A3-AA40-5069A4D11E99 | area:11 C4DECFF8-1150-4E6E-9C17-277B9378CA7B | feat/BC-3-domain-ranking | /Users/forrest/Code/brillio-worktrees/bc-3-domain-ranking | boot | 2026-10-14T01:49Z | 2026-10-14T01:49Z |
| Delegator | BC-4 | panel:19 6903B98A-C658-438C-9B01-2547E56BF263 | area:9 29806806-32FF-427B-A78B-FA87D1CB29FD | feat/BC-4-sqlite-store | /Users/forrest/Code/brillio-worktrees/bc-4-sqlite-store | boot | 2026-10-14T01:49Z | 2026-10-14T01:49Z |

### Archived (run history)

| Actor | Ticket | Outcome | Notes |
|---|---|---|---|
| agent:opus-5.5-delegator (panel:12 80350AB1-979F-45E9-A779-02167F719BBC) | BC-1 | done 01:18Z | Initial commit f4f8a0b directly on master; plan 518 lines (oversized → brief now asks for proportional plans); plan review ~7 min; review PASS-WITH-NITS (2 .gitignore nits → BC-2); validation N/A. Anomaly: closing its panel collapsed area:8 (and area:10) — placeholder rule added. Reported host global-gitignore footgun. |
