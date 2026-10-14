# BC-5: Search API: strict query parsing, validation with field codes, pipeline handler, cities endpoint

The search endpoint end to end: strict query parsing, collected validation errors with field codes, the pipeline in one readable function, response mapping, plus the cities endpoint. Spec: docs/SPEC.md §5.2, §5.3, §5.4, §5.5, §6, §8 (httpapi list). Messages for field errors come from BRANDING.md §7.2 where a string exists.

Why: this is where "handle invalid input deliberately" is demonstrated. Unknown params are rejected, all problems are returned together, a page past the end is a 200, and a city with no matches is a 200. The pipeline function is the thing the interviewer reads when adding a requirement, so it must show its stages plainly.

Scope (package internal/httpapi):
- query.go: `ParseSearchQuery(url.Values) (SearchQuery, []FieldError)`. Allowlist of param names; anything else yields `unknown_param`. Empty value = absent. Numbers via strconv.ParseFloat; integers must be whole (`not_an_integer`). Rules and codes per §5.3/§5.5: negative, not_positive, out_of_range (pageSize 1..100, page >= 1), too_long (city 100, keyword 200), invalid_enum (status), min_gt_max reported once on maxPrice. Keyword split on whitespace into lowercase tokens. Status accepts repeats and comma-separated values; absent means all three. Defaults page 1, pageSize 10.
- search.go: handler for GET /api/listings/search implementing §6 in one function with a comment marking the dedupe slot. Dependencies: a Searcher interface (BC-4 store), a `Now func() time.Time` (main injects time.Now, tests inject a fixed clock). Maps domain output to the §5.4 JSON: rank, relevance rounded to four decimals, budgetFit null without budget, scoring block (weights 0/1 without budget, asOf = now UTC YYYY-MM-DD), appliedFilters with nulls.
- cities.go: GET /api/cities per §5.2.
- Tests per §8 httpapi list with httptest and a seeded :memory: store: every rule returns 400 with the right code and all errors collected; unknown param; page past end is 200 with empty items and correct metadata; empty city is 200 with totalItems 0; happy path for targetBudget=450000&city=Springfield with a fixed clock asserts rank order and the exact relevance numbers computed from §6.2; status filter; keyword with %; X-Request-ID present on every response.

Acceptance:
- Response bodies match §5.4 and §5.5 field for field (a test unmarshals into a struct with DisallowUnknownFields both ways).
- `make check` green.
- The pipeline function reads top to bottom as parse, search, [dedupe slot], rank, paginate, respond.

Validation: against the running API, curl the §10 smoke queries plus status=pending, keyword=%25 (encoded percent), pageSize=0, and minprice=1 (lowercase typo); paste the responses with --role validation.
