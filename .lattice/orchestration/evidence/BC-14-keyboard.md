# BC-14 keyboard-only pass (V6)

- HEAD: `10133f83f5eb6fdb034280a4e80aea413f8e6854` (branch `feat/BC-14-branding-pass`), worktree clean before and after.
- Run: 2026-10-14 05:41Z, validator `agent:opus-5.5-validator`.
- Browser: headless Chrome 155.0.8059.40 (`--headless=new`, CDP port 19214), viewport 1280x800, DPR 2,
  `Emulation.setFocusEmulationEnabled` on. App at http://localhost:15114/ (vite, proxy to the API on 18014).
- Input: trusted events only. Keys go through `Input.dispatchKeyEvent` (Tab 9, Enter 13 with text `\r`,
  PageDown 34, ArrowDown 40, Shift modifier 8); text through `Input.insertText`. No `el.click()`, no value setters.
  `Runtime.evaluate` was used only to read state after each key (and once to blur the initial focus).
- The API was stopped and restarted from the driver (SIGTERM to the listener on 18014, then a fresh
  `./backend/bin/api` with the same env).
- Driver: `<scratchpad>/validator/keyboard.mjs` (outside the worktree).

Columns: key, then `document.activeElement` (role and accessible name), URL, status line halves (left / right),
`main.scrollTop`, `window.scrollY`, and the focus look (computed underline for fields, outline for everything else).

| # | Step | Key | Focus | URL | Status | main.scrollTop | scrollY | Focus look / note |
|---|---|---|---|---|---|---|---|---|
| 1 | 1 | (start) | body "" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | - |
| 2 | 1 | Tab | textbox "Min price" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 3 | 1 | Tab | textbox "Max price" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 4 | 1 | Tab | textbox "Min bedrooms" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 5 | 1 | Tab | combobox "City" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 6 | 1 | Tab | textbox "Keyword" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 7 | 1 | Tab | textbox "Target budget" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 8 | 1 | Tab | combobox "Status" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 9 | 1 | Tab | button "Search" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 2px (:focus-visible) |
| 10 | 1 | Tab | button "Clear" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 11 | 1 | Tab | main "Results" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset -2px (:focus-visible) |
|  | R4 | results stop | tagName=MAIN, getAttribute('tabindex') === '0' -> true, outline-color=rgb(122, 31, 43), :focus-visible=true | | | | | |
| 12 | 1 | Tab | button "Page 1" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 13 | 1 | Tab | button "Page 2" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 14 | 1 | Tab | button "Next ›" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 15 | 2 | Shift+Tab | button "Page 2" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 16 | 2 | Shift+Tab | button "Page 1" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 17 | 2 | Shift+Tab | main "Results" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset -2px (:focus-visible) |
| 18 | 2 | Shift+Tab | button "Clear" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 19 | 2 | Shift+Tab | button "Search" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 2px (:focus-visible) |
| 20 | 2 | Shift+Tab | combobox "Status" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 21 | 2 | Shift+Tab | textbox "Target budget" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 22 | 2 | type "450000" | textbox "Target budget" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 23 | 2 | Enter | textbox "Target budget" | /?targetBudget=450000 | 12 results · ranked for $450,000 / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 24 | 3 | Tab | combobox "Status" | /?targetBudget=450000 | 12 results · ranked for $450,000 / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 25 | 3 | Tab | button "Search" | /?targetBudget=450000 | 12 results · ranked for $450,000 / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 2px (:focus-visible) |
| 26 | 3 | Tab | button "Clear" | /?targetBudget=450000 | 12 results · ranked for $450,000 / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 27 | 3 | Tab | main "Results" | /?targetBudget=450000 | 12 results · ranked for $450,000 / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset -2px (:focus-visible) |
| 28 | 3 | PageDown | main "Results" | /?targetBudget=450000 | 12 results · ranked for $450,000 / Page 1 of 2 | 148 | 0 | ring 2px solid rgb(122, 31, 43) offset -2px (:focus-visible) |
| 29 | 3 | (after smooth scroll) | main "Results" | /?targetBudget=450000 | 12 results · ranked for $450,000 / Page 1 of 2 | 457 | 0 | ring 2px solid rgb(122, 31, 43) offset -2px (:focus-visible) |
| 30 | 3 | ArrowDown | main "Results" | /?targetBudget=450000 | 12 results · ranked for $450,000 / Page 1 of 2 | 469 | 0 | ring 2px solid rgb(122, 31, 43) offset -2px (:focus-visible) |
| 31 | 3 | (after ArrowDown) | main "Results" | /?targetBudget=450000 | 12 results · ranked for $450,000 / Page 1 of 2 | 497 | 0 | ring 2px solid rgb(122, 31, 43) offset -2px (:focus-visible) |
| 32 | 4 | Tab | button "Page 1" | /?targetBudget=450000 | 12 results · ranked for $450,000 / Page 1 of 2 | 497 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 33 | 4 | Tab | button "Page 2" | /?targetBudget=450000 | 12 results · ranked for $450,000 / Page 1 of 2 | 497 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 34 | 4 | Enter | heading "Results #results-heading" | /?targetBudget=450000&page=2 | 12 results · ranked for $450,000 / Page 2 of 2 | 0 | 0 | no ring; first rank No. 11 |
| 35 | 4 | Tab | button "‹ Previous" | /?targetBudget=450000&page=2 | 12 results · ranked for $450,000 / Page 2 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 36 | 4 | Tab | button "Page 1" | /?targetBudget=450000&page=2 | 12 results · ranked for $450,000 / Page 2 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 37 | 4 | Tab | button "Page 2" | /?targetBudget=450000&page=2 | 12 results · ranked for $450,000 / Page 2 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 38 | 4 | Tab | textbox "Min price" | /?targetBudget=450000&page=2 | 12 results · ranked for $450,000 / Page 2 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 39 | 4 | Shift+Tab | body "" | /?targetBudget=450000&page=2 | 12 results · ranked for $450,000 / Page 2 of 2 | 0 | 0 | - |
| 40 | 4 | Shift+Tab | button "Page 2" | /?targetBudget=450000&page=2 | 12 results · ranked for $450,000 / Page 2 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 41 | 4 | Shift+Tab | button "Page 1" | /?targetBudget=450000&page=2 | 12 results · ranked for $450,000 / Page 2 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 42 | 4 | Shift+Tab | button "‹ Previous" | /?targetBudget=450000&page=2 | 12 results · ranked for $450,000 / Page 2 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 43 | 4 | Enter | heading "Results #results-heading" | /?targetBudget=450000 | 12 results · ranked for $450,000 / Page 1 of 2 | 0 | 0 | no ring; first rank No. 1 |
| 44 | 5 | Shift+Tab | main "Results" | /?targetBudget=450000 | 12 results · ranked for $450,000 / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset -2px (:focus-visible) |
| 45 | 5 | Shift+Tab | button "Clear" | /?targetBudget=450000 | 12 results · ranked for $450,000 / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 46 | 5 | Enter | button "Clear" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible); Clear |
|  | 5 | (API stopped, pid 94515) | | | | | | |
| 47 | 5 | Shift+Tab | button "Search" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 2px (:focus-visible) |
| 48 | 5 | Shift+Tab | combobox "Status" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 49 | 5 | Shift+Tab | textbox "Target budget" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 50 | 5 | Shift+Tab | textbox "Keyword" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 51 | 5 | type "pets" | textbox "Keyword" | / | 12 results · newest first / Page 1 of 2 | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 52 | 5 | Enter | textbox "Keyword" | /?keyword=pets | (empty) / (empty) | 0 | 0 | underline 2px rgb(122, 31, 43); main: Couldn't load results. | Try again |
|  | 5 | (API restarted, pid 97782) | | | | | | |
| 53 | 5 | Tab | textbox "Target budget" | /?keyword=pets | (empty) / (empty) | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 54 | 5 | Tab | combobox "Status" | /?keyword=pets | (empty) / (empty) | 0 | 0 | underline 2px rgb(122, 31, 43) |
| 55 | 5 | Tab | button "Search" | /?keyword=pets | (empty) / (empty) | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 2px (:focus-visible) |
| 56 | 5 | Tab | button "Clear" | /?keyword=pets | (empty) / (empty) | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 57 | 5 | Tab | main "Results" | /?keyword=pets | (empty) / (empty) | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset -2px (:focus-visible) |
| 58 | 5 | Tab | button "Try again" | /?keyword=pets | (empty) / (empty) | 0 | 0 | ring 2px solid rgb(122, 31, 43) offset 3px (:focus-visible) |
| 59 | 5 | Enter | body "" | /?keyword=pets | 7 results · newest first / Page 1 of 1 | 0 | 0 | -; Try again |

## Results against plan V6 (with Resolutions R3 and R4)

- PASS: R4 results stop: MAIN, tabindex "0", outline rgb(122, 31, 43) ({"tagName":"MAIN","tabindex0":true,"outlineColor":"rgb(122, 31, 43)","focusVisible":true})
- PASS: (1) Tab order Min price ... Clear, results region, Page 1, Page 2, Next; Previous skipped (Min price > Max price > Min bedrooms > City > Keyword > Target budget > Status > Search > Clear > Results > Page 1 > Page 2 > Next ›)
- PASS: (2) URL ?targetBudget=450000 and status "12 results · ranked for $450,000" (/?targetBudget=450000 | 12 results · ranked for $450,000)
- PASS: (3) PageDown on the focused results region raises main.scrollTop, scrollY stays 0 (PageDown -> 457, ArrowDown -> 497, scrollY 0)
- PASS: (4) Page 2: URL page=2, "Page 2 of 2", focus on #results-heading, first rank No. 11 (/?targetBudget=450000&page=2 | Page 2 of 2 | Results #results-heading | No. 11)
- PASS: (4) From the heading, Tab reaches Previous, then Page 1, Page 2; Next (disabled) is skipped (‹ Previous > Page 1 > Page 2 > textbox Min price)
- PASS: (4) Enter on Previous returns to page 1 (/?targetBudget=450000 | Page 1 of 2 | focus Results #results-heading)
- PASS: (5) Clear by keyboard: URL /, "12 results · newest first" (/ | 12 results · newest first)
- PASS: (5) API down: error state "Couldn't load results.", status line empty (/?keyword=pets | status "" / "")
- PASS: (5) Try again after restart: "7 results · newest first" (/?keyword=pets | 7 results · newest first)

All 10 checks PASS.

## R4: attributing the results stop to D1

Headless Chrome 130+ makes a scroll container Tab-reachable even without `tabindex`, so the Chrome pass alone
cannot prove D1. At the results stop (row 11 and the R4 line) the focused element is `MAIN`,
`getAttribute('tabindex') === '0'` is true, and the computed outline is `2px solid rgb(122, 31, 43)` at offset
`-2px` with `:focus-visible` true (the D1 rule `.results:focus-visible`). The WebKit path (no auto-focusable
scrollers, which is the case D1 fixes) is proven by the SearchPage jsdom unit test **T13**, "puts the results
region in the Tab order after Clear, so the keyboard can scroll it" (frontend/src/pages/SearchPage.test.tsx:629).
It passes at this HEAD: a targeted `npx vitest run src/pages/SearchPage.test.tsx -t "puts the results region in
the Tab order"` reported 1 passed, 14 skipped, and the full `make check` reported 23/23 files and 398/398 tests
passed. The c11 WKWebView panel cannot send trusted keys (`browser.input_keyboard is not supported on
WKWebView`, BC-12 evidence), so it was used as a visual cross-check only.

## Observations (not failures of V6)

- Row 59: after Enter on **Try again**, the ErrorState (and the focused button) unmounts when the 2xx settles, so
  focus falls to `<body>`. A follow-up probe shows the same for **Clear filters** in the no-results state
  (focus on BODY after Enter, URL `/`, `12 results · newest first`). Keyboard use still works (the next Tab
  starts again at Min price), and SPEC section 7.4 only names focus handling for page changes. Raised for the
  delegator; not a BRANDING item.
- Row 38: on page 2, Tab from Page 2 skips the disabled Next and leaves the document (headless Chrome then
  wraps to Min price). Shift+Tab from outside (row 39 shows `body`) returns to Page 2.
- Row 28-29: PageDown is a smooth scroll (148 right after the key, 457 after it settles); ArrowDown adds 40.
  `window.scrollY` is 0 throughout.
