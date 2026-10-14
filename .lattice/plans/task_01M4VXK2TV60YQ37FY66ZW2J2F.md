# BC-8: Frontend core: API client and types, URL state and search hooks, BRANDING formatters

The frontend's non-visual core: API types and client, URL state hook, search hook with abort and previous-results memory, cities hook, and the BRANDING formatters. Spec: docs/SPEC.md §7.1, §7.2, §7.3, §8 (frontend hooks and formatters). Design: BRANDING.md §7.1 formatting table.

Why: components (BC-9, BC-10, BC-11) must be pure props-in renderers; everything stateful and every string-formatting rule lives here with its own tests, so three component tasks can run in parallel.

Scope:
- src/api/types.ts exactly as §7.3. src/api/errors.ts with ApiError union and a `parseErrorResponse(res)` helper. src/api/client.ts: `searchListings(params, signal?)` builds the query string (undefined omitted, status repeated, numbers as plain strings), `getCities(signal?)`; non-2xx bodies parse into ApiError; fetch TypeError and AbortError map to kind 'network' (the hook ignores aborts).
- src/hooks/useSearchParamsState.ts per §7.1: parse window.location.search into SearchParams on mount (numbers parsed, invalid values dropped), `setSearch(next)` pushes the new query string without page, `setPage(n)` pushes with only page changed, `clear()` pushes an empty query, popstate re-reads; pageSize preserved if present.
- src/hooks/useListingSearch.ts per §7.2 returning SearchState; AbortController per request; previous data retained during loading and on invalid/error; the state 'empty' when totalItems is 0; 'invalid' for validation errors.
- src/hooks/useCities.ts: load once, return string[] (empty on failure, error to console.error).
- src/format/money.ts, date.ts, listing.ts implementing every row of BRANDING §7.1: price `$450,000`; `2 bd`; `1 ba` / `1.5 ba`; `1,450 sqft`; `Listed Aug 29` with `, YYYY` when not the current year (year injected for tests); source `MLS_A` to `MLS A`; status label uppercase only when not active; score two decimals; rank `No. 1`; `1 result` / `12 results`; `Page 1 of 2`; monogram initials `Sp`.
- Tests per §8: client query building and error parsing (400 validation with fields, 500 http, TypeError network, abort ignored); both hooks under jsdom with history and popstate; every formatter row.

Acceptance:
- All tests green; `make lint` green.
- No new runtime dependencies.
- Types compile against a fixture built from the real sample data (src/test/fixtures.ts with three listings and one full SearchResponse).

Validation: N/A until BC-12 wires a page. Record a one-line justification with --role validation and cite the test output.
