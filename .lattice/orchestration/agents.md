# Agents — brillio run

Active table is overwritten each dispatch tick (Lattice + `c11 tree` are ground truth).

## Active

| Role | Ticket | Panel ref | Area ref | Branch | Worktree | Phase | Last seen | Spawned at |
|---|---|---|---|---|---|---|---|---|
| Orchestrator | all | panel:6 73809EE8-EE77-4F1E-8EE2-8E8B05308939 | area:6 D887B46C-F2FA-4097-ABB0-B453319B8C6B | master | /Users/forrest/Code/brillio | dispatch | 2026-10-14T03:13Z | 2026-10-14T00:45Z |
| Master Validator | all | panel:13 AFD95ABD-DC58-4303-BEBE-8591321113A7 | area:6 D887B46C-F2FA-4097-ABB0-B453319B8C6B | — | /Users/forrest/Code/brillio | audit loop | 2026-10-14T03:13Z | 2026-10-14T01:20Z |
| Delegator | BC-14 | panel:37 7D3FCE66-1BFA-4B9C-942B-2E145C875007 | area:9 29806806-32FF-427B-A78B-FA87D1CB29FD | feat/BC-14-branding-pass | /Users/forrest/Code/brillio-worktrees.noindex/bc-14-branding-pass | boot | 2026-10-14T04:40Z | 2026-10-14T04:40Z |
| Delegator | BC-16 | panel:34 F7120790-4B51-404C-8E80-5E147875C427 | area:9 29806806-32FF-427B-A78B-FA87D1CB29FD | feat/BC-16-extensions-doc | /Users/forrest/Code/brillio-worktrees.noindex/bc-16-extensions-doc | boot | 2026-10-14T04:40Z | 2026-10-14T04:40Z |
| Delegator | BC-17 | panel:38 8287B289-E303-41D5-B16F-682E5419EAE8 | area:11 C4DECFF8-1150-4E6E-9C17-277B9378CA7B | feat/BC-17-release-dist | /Users/forrest/Code/brillio-worktrees.noindex/bc-17-release-dist | boot | 2026-10-14T05:18Z | 2026-10-14T05:18Z |

### Archived (run history)

| Actor | Ticket | Outcome | Notes |
|---|---|---|---|
| agent:opus-5.5-delegator (panel:12 80350AB1-979F-45E9-A779-02167F719BBC) | BC-1 | done 01:18Z | Initial commit f4f8a0b directly on master; plan 518 lines (oversized → brief now asks for proportional plans); plan review ~7 min; review PASS-WITH-NITS (2 .gitignore nits → BC-2); validation N/A. Anomaly: closing its panel collapsed area:8 (and area:10) — placeholder rule added. Reported host global-gitignore footgun. |
| agent:opus-5.5-delegator (panel:16 E50A586D-7B7D-4FBA-89A0-918D59278F9A) | BC-2 | done 02:09Z | Merge b273a8a; head a9b7f73 (4 commits); 1 fix→re-review cycle (CORS origin minor); 2 remaining minors skipped with reasons; 82 go test runs, lint 0; plan review ~7 min (Fable 5.1); froze shared domain types at 5da9b80 and posted the Listing/DateLayout rule to BC-3/4/5. |
| agent:opus-5.5-delegator (panel:17 F93C149A-6AA9-4705-A195-09CEABBE1680) | BC-7 | done 02:27Z | Merge aa65e61; head 0d1c2a2 (merged master/BC-2 in-branch, re-reviewed exact head). Plan review timed out at 600 s → fallback sub-agent. Deviations: ESLint 9.39.5 (jsx-a11y peer), tsc -b --noEmit (SPEC §2/§8 edits). Vitest 4 tests; build 220 kB JS. |
| agent:opus-5.5-delegator (panel:18 C207406F-8A23-49A3-AA40-5069A4D11E99) | BC-3 | done 02:45Z | Merge dc35d1a; head cdec405; 66 domain tests incl. 36 §8-named subtests; 2 of 2 fix cycles used on doc fixes; no SPEC edits. Landed from board state while c11 mail recv was wedged. |
| agent:opus-5.5-delegator (panel:19 6903B98A-C658-438C-9B01-2547E56BF263) | BC-4 | done 03:04Z | Merge 1c23253; head 9daad85; 244 go tests incl. subtests; SPEC §4.2 one-sentence edit; created BC-18 (backlog) for review nits; mail sent file-first during c11 stall; worktree removed before its panel closed (close timed out, retried 03:08Z OK). |
| agent:opus-5.5-delegator (panel:20 ECFFF9F4-A172-4457-9C6E-E90C1EF1FD35) | BC-8 | done 03:10Z | Merge ccd1669; landed head 0426cb4 = reviewed 43e8824 + orchestrator master merge; 197 vitest tests; exports listed in BC-8 completion comment. |
| agent:opus-5.5-delegator (panel:27 CB8FE2DE-5324-4FA1-99A1-274F80B5EE3A) | BC-11 | done 03:41Z | Merge 0abb105; head 0e70187; +31 tests (228 FE total); text-button recipe relayed to BC-9/BC-10; created BC-19 (backlog) for the branding owner. |
| agent:opus-5.5-delegator (panel:24 2D11C0F2-6EC2-494F-8F52-856F7190DA45) | BC-5 | done 03:56Z | Merge 52c69dd; head 1871b01; 432 Go tests; proposed malformed_query (approved); waited 13 min on an approval lost to a c11 doorbell timeout; created BC-20 (backlog). |
| agent:opus-5.5-delegator (panel:26 90079513-D6DF-432E-B7C6-287034065937) | BC-10 | done 04:09Z | Merge 5eaacdc; landed head 64bad4d = reviewed fa7c08d + orchestrator master merge; 41 BC-10 tests (269 FE total); visual check via throwaway preview at 1100/375; added format/keyword.ts + fixtures listingB7. |
| agent:opus-5.5-delegator (panel:21 58816CAB-6208-40B5-B6E7-A6D371F3FD33) | BC-13 | done 04:28Z | Merge 4030996; 3 review cycles (2 fix cycles at the cap, final PASS with no findings); headless reviews died 3 times; 3 reviewer safety prompts approved by orchestrator; validation hold for BC-5 worked as designed; found the operator DOCKER_DEFAULT_PLATFORM footgun. |
| agent:opus-5.5-delegator (panel:25 93F19093-205B-4276-AD48-2ABEDE335616) | BC-9 | done 04:31Z | Merge 9de440a; landed head 1b8a2a3 = reviewed 13ae060 + orchestrator master merge; FilterBar 113 tests (382 FE total); headless reviews timed out twice (missed the 03:22Z skip-rule mail); created BC-21 (backlog). |
| agent:opus-5.5-delegator (panel:28 0EE4218B-B323-4CD7-A39F-683C8E5C4D0B) | BC-6 | done 04:44Z | Merge 269f686; head 70518b5; +1360 backend lines (openapi.yaml + embed + sync test + yaml subset test); redocly 0 errors; one headless review started just before the 04:02Z routing mail (PASS, superseded by fable PASS). |
| agent:opus-5.5-delegator (panel:32 C497296D-B8DC-4AF6-80D4-2131C2273BF6) | BC-14 | died ~04:45Z at in_planning | Its planner sub-agent ran positional `c11 close-panel panel:35`, which closed $C11_PANEL_ID = the delegator panel. Replaced by panel:37. |
| agent:opus-5.5-delegator (panel:29 C4B54149-6B16-46DE-87DF-9AE951C10B13) | BC-12 | done 05:00Z | Merge b514f8e; head bb78854; +922 lines; 396 FE tests; browser e2e 7/7 with screenshots; 2 review cycles (fable); held for BC-9 as designed; one missed doorbell (nudged). |
| agent:opus-5.5-delegator (panel:33 99BAECD2-FF68-4004-BAD1-16BFD3D8002B) | BC-15 | done 05:33Z | Merge f41fb8c; README 245 lines; native commands run, compose blocks cite BC-13 smoke (approved). |
