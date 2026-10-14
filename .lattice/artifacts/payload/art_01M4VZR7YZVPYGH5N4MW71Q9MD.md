# BC-1 validation — HEAD f4f8a0bd10cfdb2cee08566d55ffcfcb55f08417

Justification (N/A, no running system): BC-1 adds only repo scaffolding (ignore rules, editorconfig, Makefile stubs, README stub, data move) and no runnable service, so validation is the observed repository state below.

Result: PASS

| Check | Command | Observed | Result |
|---|---|---|---|
| one commit | `git log --oneline` | `f4f8a0b chore: bootstrap repository` (single line); `git rev-parse HEAD` = `f4f8a0bd10cfdb2cee08566d55ffcfcb55f08417` | PASS |
| clean outside .lattice | `git status --porcelain -- . ':!.lattice'` | (empty), exit=0 | PASS |
| gate fails loudly | `make check 2>&1; echo exit=$?` | `not yet implemented: lint-api (lands in BC-2)` / `make: *** [lint-api] Error 1` / exit=2 | PASS |
| data moved | `ls data/ sample_listings.json 2>&1; echo exit=$?` | `data/sample_listings.json` listed (4814 bytes); `ls: sample_listings.json: No such file or directory`; exit=1 | PASS |

Notes: the only working-tree changes are under `.lattice/` (live board updates to tracked task/event files plus untracked artifacts), which the pathspec excludes by design. The `ls` output appears in long format because the session's shell hook rewrites `ls`; the listing content is unaffected.

## Raw output

```
$ cd /Users/forrest/Code/brillio && git log --oneline; echo "---"; git rev-parse HEAD
f4f8a0b chore: bootstrap repository
---
f4f8a0bd10cfdb2cee08566d55ffcfcb55f08417
```

```
$ cd /Users/forrest/Code/brillio && git status --porcelain -- . ':!.lattice'; echo "exit=$?"
exit=0
```

```
$ cd /Users/forrest/Code/brillio && make check 2>&1; echo exit=$?
not yet implemented: lint-api (lands in BC-2)
make: *** [lint-api] Error 1

exit=2
```

```
$ cd /Users/forrest/Code/brillio && ls data/ sample_listings.json 2>&1; echo exit=$?
data/:
total 16
-rw-r--r--@ 1 forrest  staff  4814 Oct 13 19:32 sample_listings.json
ls: sample_listings.json: No such file or directory
exit=1
```
