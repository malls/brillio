# BC-5 Validation: Search API against a running server

- Ticket: BC-5 (task_01M4VXK2E02WC334GKAWFQ5XDG), Search API: strict query parsing, validation with field codes, pipeline handler, cities endpoint
- Validator: agent:opus-5.5-validator
- Worktree: `/Users/forrest/Code/brillio-worktrees/bc-5-search-api`, branch `feat/BC-5-search-api`
- HEAD: `1871b0127d1e942ca509a41e63a21be9493c4ec5` (confirmed with `git rev-parse HEAD`; working tree clean before and after; `backend/bin/` is git-ignored)
- Date/time: 2026-10-14, requests run 03:47:16-03:48:15 UTC (host local 2026-10-13 23:47 EDT). Server `scoring.asOf` = 2026-10-14 (UTC date).
- Contract: docs/SPEC.md §5.2-§5.5, §6.2, §10; plan `.lattice/plans/task_01M4VXK2E02WC334GKAWFQ5XDG.md` Validation section; validation-plan rows 13-20, 28, 31 (port 18005 used instead of 18099).

## Server

Pre-check: `lsof -nP -iTCP:18005 -sTCP:LISTEN` returned nothing (exit 1).

```
cd backend && go build -o bin/api ./cmd/api
# from the worktree root, in the background, PID 82064:
PORT=18005 DB_PATH=:memory: LISTINGS_FILE=./data/sample_listings.json ./backend/bin/api > <scratchpad>/bc5-api.log 2>&1 &
curl -fsS localhost:18005/healthz   # {"status":"ok"} on the first try
```

Startup log (first two lines):

```
{"time":"2026-10-13T23:46:59.649383-04:00","level":"INFO","msg":"seeded 12 listings","count":12,"listings_file":"./data/sample_listings.json","db_path":":memory:"}
{"time":"2026-10-13T23:46:59.649836-04:00","level":"INFO","msg":"listening","addr":"[::]:18005","log_level":"INFO","cors_origins":["http://localhost:5173"],"shutdown_timeout":"10s"}
```

The server logged 41 request lines and 0 ERROR lines during the run.

## A. SPEC §10 smoke queries

#### A1 GET /api/cities

```
curl -sS -D - 'http://localhost:18005/api/cities'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `3a8e141378788a4c26c044a0f0302d9e`
- Body:

```json
{"cities":["Chantilly","Fairfax","Manassas","Reston","Springfield","Vienna"]}
```
- Expected: 200, body exactly `{"cities":["Chantilly","Fairfax","Manassas","Reston","Springfield","Vienna"]}`
- Result: **PASS**

#### A2 targetBudget=450000&city=Springfield

```
curl -sS -D - 'http://localhost:18005/api/listings/search?targetBudget=450000&city=Springfield'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `9a07a7137a8a6853bb6667068e6bda4b`
- Body (condensed: each item reduced to `[rank, id, status, listedDate, price, score, budgetFit, recency]`; page, scoring and appliedFilters verbatim; full body saved, 1879 bytes):

```json
{"items":[[1,"A1","active","2026-08-29",450000,0.7382,1,0.3455],[2,"B7","active","2026-08-27",452000,0.7266,0.9911,0.3299],[3,"A2","active","2026-09-02",525000,0.5516,0.6667,0.3789],[4,"B8","active","2026-08-30",527500,0.5348,0.6556,0.3536]],"page":{"page":1,"pageSize":10,"totalItems":4,"totalPages":1},"scoring":{"targetBudget":450000,"budgetWeight":0.6,"recencyWeight":0.4,"recencyHalfLifeDays":30,"asOf":"2026-10-14"},"appliedFilters":{"minPrice":null,"maxPrice":null,"minBedrooms":null,"city":"Springfield","keyword":null,"status":["active","pending","sold"]}}
```
- Note: Byte-identical to the row 19 request below (`cmp`); numbers checked in E.
- Expected: 200, non-empty items, first item `rank: 1` (A1)
- Result: **PASS**

#### A3 minPrice=5&maxPrice=1

```
curl -sS -D - 'http://localhost:18005/api/listings/search?minPrice=5&maxPrice=1'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `6b01d3e1686e75b01f2fc799b85a5656`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"maxPrice","code":"min_gt_max","message":"Min price must be less than max price."}]}}
```
- Expected: 400 `validation_error`, one field error `maxPrice` / `min_gt_max`
- Result: **PASS**

## B. Ticket extras

#### B1 status=pending

```
curl -sS -D - 'http://localhost:18005/api/listings/search?status=pending'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `bd90626ef5edf86f092eaa2c89a2efdb`
- Body (condensed: each item reduced to `[rank, id, status, listedDate, price, score, budgetFit, recency]`; page, scoring and appliedFilters verbatim; full body saved, 683 bytes):

```json
{"items":[[1,"A7","pending","2026-07-28",540000,0.1649,null,0.1649]],"page":{"page":1,"pageSize":10,"totalItems":1,"totalPages":1},"scoring":{"targetBudget":null,"budgetWeight":0,"recencyWeight":1,"recencyHalfLifeDays":30,"asOf":"2026-10-14"},"appliedFilters":{"minPrice":null,"maxPrice":null,"minBedrooms":null,"city":null,"keyword":null,"status":["pending"]}}
```
- Expected: 200, only A7, appliedFilters.status `["pending"]`
- Result: **PASS**

#### B2 keyword=%25 (encoded percent)

```
curl -sS -D - 'http://localhost:18005/api/listings/search?keyword=%25'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `28efd5fe3b2d3096ebb6d332787702cb`
- Body:

```json
{"items":[],"page":{"page":1,"pageSize":10,"totalItems":0,"totalPages":0},"scoring":{"targetBudget":null,"budgetWeight":0,"recencyWeight":1,"recencyHalfLifeDays":30,"asOf":"2026-10-14"},"appliedFilters":{"minPrice":null,"maxPrice":null,"minBedrooms":null,"city":null,"keyword":"%","status":["active","pending","sold"]}}
```
- Expected: 200, 0 items, totalItems 0, keyword echoed as `"%"` (percent is literal, not a wildcard)
- Result: **PASS**

#### B3 pageSize=0

```
curl -sS -D - 'http://localhost:18005/api/listings/search?pageSize=0'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `5ff249759cc8427d58b8d3b4b7d6cc48`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"pageSize","code":"out_of_range","message":"Page size must be between 1 and 100."}]}}
```
- Expected: 400, `pageSize` / `out_of_range`, message `Page size must be between 1 and 100.`
- Result: **PASS**

#### B4 minprice=1 (lowercase typo)

```
curl -sS -D - 'http://localhost:18005/api/listings/search?minprice=1'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `a9a7de0999890f34a621e1122ff65a26`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"minprice","code":"unknown_param","message":"Unknown parameter."}]}}
```
- Expected: 400, `minprice` / `unknown_param`, message `Unknown parameter.`
- Result: **PASS**

## C. Validation-plan rows (port 18005)

Row 13 (cities exact body) is request A1 above: 200, body byte-equal to `{"cities":["Chantilly","Fairfax","Manassas","Reston","Springfield","Vienna"]}`. **PASS**

### Row 14: field validation, one request per code

#### C14a minPrice=-1

```
curl -sS -D - 'http://localhost:18005/api/listings/search?minPrice=-1'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `842d3a04143a2f047a4f725abbc85402`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"minPrice","code":"negative","message":"Min price must be 0 or more."}]}}
```
- Expected: 400 `validation_error` / `Invalid search parameters.` with fields `minPrice`/`negative`
- Result: **PASS**

#### C14b minPrice=abc

```
curl -sS -D - 'http://localhost:18005/api/listings/search?minPrice=abc'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `0469ef589200de9b220e5559ddecaf99`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"minPrice","code":"not_a_number","message":"Min price must be a number."}]}}
```
- Expected: 400 `validation_error` / `Invalid search parameters.` with fields `minPrice`/`not_a_number`
- Result: **PASS**

#### C14c minBedrooms=2.5

```
curl -sS -D - 'http://localhost:18005/api/listings/search?minBedrooms=2.5'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `3f1c22e233b40415790c0402d406dafd`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"minBedrooms","code":"not_an_integer","message":"Min bedrooms must be a whole number."}]}}
```
- Expected: 400 `validation_error` / `Invalid search parameters.` with fields `minBedrooms`/`not_an_integer`
- Result: **PASS**

#### C14d targetBudget=0

```
curl -sS -D - 'http://localhost:18005/api/listings/search?targetBudget=0'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `b108f6485be6e586768ab32145dd2453`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"targetBudget","code":"not_positive","message":"Target budget must be more than $0."}]}}
```
- Expected: 400 `validation_error` / `Invalid search parameters.` with fields `targetBudget`/`not_positive`
- Result: **PASS**

#### C14e status=foo

```
curl -sS -D - 'http://localhost:18005/api/listings/search?status=foo'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `d8592c029f858acde9857749f6842d74`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"status","code":"invalid_enum","message":"Status must be active, pending or sold."}]}}
```
- Expected: 400 `validation_error` / `Invalid search parameters.` with fields `status`/`invalid_enum`
- Result: **PASS**

#### C14f page=0

```
curl -sS -D - 'http://localhost:18005/api/listings/search?page=0'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `9cedfd3da94be700499722f8724db52a`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"page","code":"out_of_range","message":"Page must be 1 or more."}]}}
```
- Expected: 400 `validation_error` / `Invalid search parameters.` with fields `page`/`out_of_range`
- Result: **PASS**

#### C14g pageSize=0

```
curl -sS -D - 'http://localhost:18005/api/listings/search?pageSize=0'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `b1f5cb33348162a412a121a5e4185ed1`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"pageSize","code":"out_of_range","message":"Page size must be between 1 and 100."}]}}
```
- Expected: 400 `validation_error` / `Invalid search parameters.` with fields `pageSize`/`out_of_range`
- Result: **PASS**

#### C14h pageSize=101

```
curl -sS -D - 'http://localhost:18005/api/listings/search?pageSize=101'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `587dd3d2f6c0ae52d24ec82e94c76eb5`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"pageSize","code":"out_of_range","message":"Page size must be between 1 and 100."}]}}
```
- Expected: 400 `validation_error` / `Invalid search parameters.` with fields `pageSize`/`out_of_range`
- Result: **PASS**

#### C14i city=<101 chars>

```
curl -sS -D - 'http://localhost:18005/api/listings/search?city=<101 x 'a'>'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `b734b48eef6d8700d2dd1e788da06ada`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"city","code":"too_long","message":"City must be 100 characters or fewer."}]}}
```
- Expected: 400 `validation_error` / `Invalid search parameters.` with fields `city`/`too_long`
- Result: **PASS**

#### C14j keyword=<201 chars>

```
curl -sS -D - 'http://localhost:18005/api/listings/search?keyword=<201 x 'k'>'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `8ec81d75e2debec9486ea6584c042ef8`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"keyword","code":"too_long","message":"Keyword must be 200 characters or fewer."}]}}
```
- Expected: 400 `validation_error` / `Invalid search parameters.` with fields `keyword`/`too_long`
- Result: **PASS**

#### C14k minPrice=5&maxPrice=1

```
curl -sS -D - 'http://localhost:18005/api/listings/search?minPrice=5&maxPrice=1'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `7dab7965d8ac7c22bfb54cee7f939f2e`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"maxPrice","code":"min_gt_max","message":"Min price must be less than max price."}]}}
```
- Expected: 400 `validation_error` / `Invalid search parameters.` with fields `maxPrice`/`min_gt_max`
- Result: **PASS**

#### C14l minprice=1

```
curl -sS -D - 'http://localhost:18005/api/listings/search?minprice=1'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `2f271a896028d2f9a675c626781d08ae`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"minprice","code":"unknown_param","message":"Unknown parameter."}]}}
```
- Expected: 400 `validation_error` / `Invalid search parameters.` with fields `minprice`/`unknown_param`
- Result: **PASS**

#### C14m three faults in one request

```
curl -sS -D - 'http://localhost:18005/api/listings/search?minPrice=-1&status=foo&pageSize=101'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `24e32d5ed131d3c0f5440b22531e1f30`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"minPrice","code":"negative","message":"Min price must be 0 or more."},{"field":"status","code":"invalid_enum","message":"Status must be active, pending or sold."},{"field":"pageSize","code":"out_of_range","message":"Page size must be between 1 and 100."}]}}
```
- Expected: 400 with all three field errors together, in table order: `minPrice`/`negative`, `status`/`invalid_enum`, `pageSize`/`out_of_range`
- Result: **PASS**

#### C14n keyword=100% (bare percent)

```
curl -sS -D - 'http://localhost:18005/api/listings/search?keyword=100%'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `796c440ebb44e4c9104f698b491c549e`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"query","code":"malformed_query","message":"Query string is not valid URL encoding."}]}}
```
- Expected: 400, field `query` / `malformed_query`, message `Query string is not valid URL encoding.`
- Result: **PASS**

#### C14o SPEC §5.5 combined request

```
curl -sS -D - 'http://localhost:18005/api/listings/search?minPrice=5&maxPrice=1&pageSize=0&minprice=1'
```

- Status: `HTTP/1.1 400 Bad Request`
- X-Request-ID: `e4447d235a891075dfb51a82e7df5cc6`
- Body:

```json
{"error":{"code":"validation_error","message":"Invalid search parameters.","fields":[{"field":"maxPrice","code":"min_gt_max","message":"Min price must be less than max price."},{"field":"pageSize","code":"out_of_range","message":"Page size must be between 1 and 100."},{"field":"minprice","code":"unknown_param","message":"Unknown parameter."}]}}
```
- Note: Compared with a python3 script that extracts the §5.5 JSON block from docs/SPEC.md: `parsed equal: True`; `byte-equal after compact: True` (the example after JSON compaction equals the response body byte for byte).
- Expected: 400, body equal to the SPEC §5.5 example (maxPrice/min_gt_max, pageSize/out_of_range, minprice/unknown_param, same messages, same order)
- Result: **PASS**

Field codes observed across row 14: negative, not_a_number, not_an_integer, not_positive, invalid_enum, out_of_range x3 (page, pageSize 0, pageSize 101), too_long x2 (city, keyword), min_gt_max (once, on maxPrice), unknown_param, malformed_query.

### Rows 15-20, 28, 31

#### C15 minPrice= (empty value is absent)

```
curl -sS -D - 'http://localhost:18005/api/listings/search?minPrice='
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `3140bffb584d88c99ec64e0244384c3d`
- Body (condensed: each item reduced to `[rank, id, status, listedDate, price, score, budgetFit, recency]`; page, scoring and appliedFilters verbatim; full body saved, 4066 bytes):

```json
{"items":[[1,"A5","active","2026-09-04",470000,0.3969,null,0.3969],[2,"A4","active","2026-09-03",610000,0.3878,null,0.3878],[3,"A2","active","2026-09-02",525000,0.3789,null,0.3789],[4,"A3","active","2026-09-01",399000,0.3703,null,0.3703],[5,"B8","active","2026-08-30",527500,0.3536,null,0.3536],[6,"A1","active","2026-08-29",450000,0.3455,null,0.3455],[7,"B7","active","2026-08-27",452000,0.3299,null,0.3299],[8,"B9","active","2026-08-25",399500,0.315,null,0.315],[9,"B10","active","2026-08-20",585000,0.2806,null,0.2806],[10,"B11","active","2026-08-15",465000,0.25,null,0.25]],"page":{"page":1,"pageSize":10,"totalItems":12,"totalPages":2},"scoring":{"targetBudget":null,"budgetWeight":0,"recencyWeight":1,"recencyHalfLifeDays":30,"asOf":"2026-10-14"},"appliedFilters":{"minPrice":null,"maxPrice":null,"minBedrooms":null,"city":null,"keyword":null,"status":["active","pending","sold"]}}
```
- Expected: 200, 12 total items, appliedFilters.minPrice null
- Result: **PASS**

#### C16 page=99 (past the end)

```
curl -sS -D - 'http://localhost:18005/api/listings/search?page=99'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `b7f61682d1342f0cdb3c0c0b410577d7`
- Body:

```json
{"items":[],"page":{"page":99,"pageSize":10,"totalItems":12,"totalPages":2},"scoring":{"targetBudget":null,"budgetWeight":0,"recencyWeight":1,"recencyHalfLifeDays":30,"asOf":"2026-10-14"},"appliedFilters":{"minPrice":null,"maxPrice":null,"minBedrooms":null,"city":null,"keyword":null,"status":["active","pending","sold"]}}
```
- Expected: 200, `items: []`, page.page 99, totalItems 12, totalPages 2
- Result: **PASS**

#### C17 city=Atlantis (no matches)

```
curl -sS -D - 'http://localhost:18005/api/listings/search?city=Atlantis'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `5565622323ebf7e0983716419b3f5c2e`
- Body:

```json
{"items":[],"page":{"page":1,"pageSize":10,"totalItems":0,"totalPages":0},"scoring":{"targetBudget":null,"budgetWeight":0,"recencyWeight":1,"recencyHalfLifeDays":30,"asOf":"2026-10-14"},"appliedFilters":{"minPrice":null,"maxPrice":null,"minBedrooms":null,"city":"Atlantis","keyword":null,"status":["active","pending","sold"]}}
```
- Expected: 200, totalItems 0, totalPages 0, city echoed
- Result: **PASS**

#### C18 status param (row 18)

#### C18a status=pending

```
curl -sS -D - 'http://localhost:18005/api/listings/search?status=pending'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `c328803127336bba7c3a22189d54359e`
- Body (condensed: each item reduced to `[rank, id, status, listedDate, price, score, budgetFit, recency]`; page, scoring and appliedFilters verbatim; full body saved, 683 bytes):

```json
{"items":[[1,"A7","pending","2026-07-28",540000,0.1649,null,0.1649]],"page":{"page":1,"pageSize":10,"totalItems":1,"totalPages":1},"scoring":{"targetBudget":null,"budgetWeight":0,"recencyWeight":1,"recencyHalfLifeDays":30,"asOf":"2026-10-14"},"appliedFilters":{"minPrice":null,"maxPrice":null,"minBedrooms":null,"city":null,"keyword":null,"status":["pending"]}}
```
- Expected: 200, only A7
- Result: **PASS**

#### C18b status=active,pending

```
curl -sS -D - 'http://localhost:18005/api/listings/search?status=active,pending'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `769e705556d6f8ec0968e9e322386730`
- Body (condensed: each item reduced to `[rank, id, status, listedDate, price, score, budgetFit, recency]`; page, scoring and appliedFilters verbatim; full body saved, 4059 bytes):

```json
{"items":[[1,"A5","active","2026-09-04",470000,0.3969,null,0.3969],[2,"A4","active","2026-09-03",610000,0.3878,null,0.3878],[3,"A2","active","2026-09-02",525000,0.3789,null,0.3789],[4,"A3","active","2026-09-01",399000,0.3703,null,0.3703],[5,"B8","active","2026-08-30",527500,0.3536,null,0.3536],[6,"A1","active","2026-08-29",450000,0.3455,null,0.3455],[7,"B7","active","2026-08-27",452000,0.3299,null,0.3299],[8,"B9","active","2026-08-25",399500,0.315,null,0.315],[9,"B10","active","2026-08-20",585000,0.2806,null,0.2806],[10,"B11","active","2026-08-15",465000,0.25,null,0.25]],"page":{"page":1,"pageSize":10,"totalItems":12,"totalPages":2},"scoring":{"targetBudget":null,"budgetWeight":0,"recencyWeight":1,"recencyHalfLifeDays":30,"asOf":"2026-10-14"},"appliedFilters":{"minPrice":null,"maxPrice":null,"minBedrooms":null,"city":null,"keyword":null,"status":["active","pending"]}}
```
- Expected: 200, totalItems 12 (11 active + 1 pending), status echoed `["active","pending"]`
- Result: **PASS**

#### C18c status=active&status=pending

```
curl -sS -D - 'http://localhost:18005/api/listings/search?status=active&status=pending'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `529e271a366efa9bb1063cc4fdbda4a3`
- Note: `cmp C18_comma.body C18_repeat.body` -> identical (exit 0); `diff <(jq -S . a) <(jq -S . b)` -> empty. Only the X-Request-ID header differs.
- Expected: 200, body identical to C18b
- Result: **PASS**

Supplementary (all 12 on one page): `?status=active,pending&pageSize=12` and `?status=active&status=pending&pageSize=12` -> both `HTTP/1.1 200 OK` (X-Request-ID `c0e0aa3feb69d3f8d3f1a14ed8ce2faf` / `ff55828c771383904c010ccef4fd1eeb`), bodies byte-identical: True; page `{"page":1,"pageSize":12,"totalItems":12,"totalPages":1}`; items: `A5:active A4:active A2:active A3:active B8:active A1:active B7:active B9:active B10:active B11:active A6:active A7:pending` (11 active + A7 pending).

#### C19 targetBudget=450000&city=Springfield (row 19)

```
curl -sS -D - 'http://localhost:18005/api/listings/search?targetBudget=450000&city=Springfield'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `1ef9b24d7de88e61e0c02c0fb12fa363`
- Body (full, pretty-printed):

```json
{
  "items": [
    {
      "rank": 1,
      "id": "A1",
      "source": "MLS_A",
      "address": "123 Main St, Apt 4B",
      "city": "Springfield",
      "state": "VA",
      "zip": "22150",
      "price": 450000,
      "bedrooms": 2,
      "bathrooms": 1.5,
      "sqft": 980,
      "latitude": 38.7893,
      "longitude": -77.1873,
      "listedDate": "2026-08-29",
      "status": "active",
      "description": "Bright top-floor condo near shops and transit. Pet friendly.",
      "relevance": {
        "score": 0.7382,
        "budgetFit": 1,
        "recency": 0.3455
      }
    },
    {
      "rank": 2,
      "id": "B7",
      "source": "MLS_B",
      "address": "123 Main Street, Unit 4B",
      "city": "Springfield",
      "state": "VA",
      "zip": "22150",
      "price": 452000,
      "bedrooms": 2,
      "bathrooms": 1.5,
      "sqft": 980,
      "latitude": 38.7893,
      "longitude": -77.1873,
      "listedDate": "2026-08-27",
      "status": "active",
      "description": "Top floor condo, walk to shopping. Pets allowed.",
      "relevance": {
        "score": 0.7266,
        "budgetFit": 0.9911,
        "recency": 0.3299
      }
    },
    {
      "rank": 3,
      "id": "A2",
      "source": "MLS_A",
      "address": "456 Oak Ave",
      "city": "Springfield",
      "state": "VA",
      "zip": "22150",
      "price": 525000,
      "bedrooms": 3,
      "bathrooms": 2,
      "sqft": 1450,
      "latitude": 38.7791,
      "longitude": -77.1901,
      "listedDate": "2026-09-02",
      "status": "active",
      "description": "Updated kitchen, fenced yard, close to schools.",
      "relevance": {
        "score": 0.5516,
        "budgetFit": 0.6667,
        "recency": 0.3789
      }
    },
    {
      "rank": 4,
      "id": "B8",
      "source": "MLS_B",
      "address": "456 Oak Avenue",
      "city": "Springfield",
      "state": "VA",
      "zip": "22151",
      "price": 527500,
      "bedrooms": 3,
      "bathrooms": 2,
      "sqft": 1450,
      "latitude": 38.7791,
      "longitude": -77.1901,
      "listedDate": "2026-08-30",
      "status": "active",
      "description": "Renovated kitchen, fenced backyard, near schools.",
      "relevance": {
        "score": 0.5348,
        "budgetFit": 0.6556,
        "recency": 0.3536
      }
    }
  ],
  "page": {
    "page": 1,
    "pageSize": 10,
    "totalItems": 4,
    "totalPages": 1
  },
  "scoring": {
    "targetBudget": 450000,
    "budgetWeight": 0.6,
    "recencyWeight": 0.4,
    "recencyHalfLifeDays": 30,
    "asOf": "2026-10-14"
  },
  "appliedFilters": {
    "minPrice": null,
    "maxPrice": null,
    "minBedrooms": null,
    "city": "Springfield",
    "keyword": null,
    "status": [
      "active",
      "pending",
      "sold"
    ]
  }
}
```
- Note: Key sets and order of the top level, item (17 keys), relevance, page, scoring and appliedFilters equal the SPEC §5.4 example (python3 comparison against the §5.4 block). `budgetFit` 1.0 is encoded by Go as `1` (same JSON number; noted in the plan).
- Expected: 200, 4 items A1, B7, A2, B8 with ranks 1-4; scoring 450000/0.6/0.4/30/asOf; appliedFilters city `Springfield`, minPrice/maxPrice/minBedrooms/keyword null, status all three; scores equal §6.2 at asOf (see E)
- Result: **PASS**

#### C20 pageSize=12, no targetBudget (row 20)

```
curl -sS -D - 'http://localhost:18005/api/listings/search?pageSize=12'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `7631a9ecf5eb6cc00ccbaec358340865`
- Body (condensed: each item reduced to `[rank, id, status, listedDate, price, score, budgetFit, recency]`; page, scoring and appliedFilters verbatim; full body saved, 4823 bytes):

```json
{"items":[[1,"A5","active","2026-09-04",470000,0.3969,null,0.3969],[2,"A4","active","2026-09-03",610000,0.3878,null,0.3878],[3,"A2","active","2026-09-02",525000,0.3789,null,0.3789],[4,"A3","active","2026-09-01",399000,0.3703,null,0.3703],[5,"B8","active","2026-08-30",527500,0.3536,null,0.3536],[6,"A1","active","2026-08-29",450000,0.3455,null,0.3455],[7,"B7","active","2026-08-27",452000,0.3299,null,0.3299],[8,"B9","active","2026-08-25",399500,0.315,null,0.315],[9,"B10","active","2026-08-20",585000,0.2806,null,0.2806],[10,"B11","active","2026-08-15",465000,0.25,null,0.25],[11,"A6","active","2026-08-10",415000,0.2227,null,0.2227],[12,"A7","pending","2026-07-28",540000,0.1649,null,0.1649]],"page":{"page":1,"pageSize":12,"totalItems":12,"totalPages":1},"scoring":{"targetBudget":null,"budgetWeight":0,"recencyWeight":1,"recencyHalfLifeDays":30,"asOf":"2026-10-14"},"appliedFilters":{"minPrice":null,"maxPrice":null,"minBedrooms":null,"city":null,"keyword":null,"status":["active","pending","sold"]}}
```
- Note: Order observed: A5, A4, A2, A3, B8, A1, B7, B9, B10, B11, A6, A7; listedDate strictly descending.
- Expected: 200, scoring.targetBudget null, weights 0/1, every budgetFit null, newest first (A5, A4, A2, A3, ...); score equals recency
- Result: **PASS**

#### C28a keyword=%25 (row 28)

```
curl -sS -D - 'http://localhost:18005/api/listings/search?keyword=%25'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `7011dcb602979b1239140b7e00545ae9`
- Body:

```json
{"items":[],"page":{"page":1,"pageSize":10,"totalItems":0,"totalPages":0},"scoring":{"targetBudget":null,"budgetWeight":0,"recencyWeight":1,"recencyHalfLifeDays":30,"asOf":"2026-10-14"},"appliedFilters":{"minPrice":null,"maxPrice":null,"minBedrooms":null,"city":null,"keyword":"%","status":["active","pending","sold"]}}
```
- Expected: 200 with 0 items
- Result: **PASS**

#### C28b keyword=_ (row 28)

```
curl -sS -D - 'http://localhost:18005/api/listings/search?keyword=_'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `2fdb1d9ac6f7d23c6f31e72945a52012`
- Body:

```json
{"items":[],"page":{"page":1,"pageSize":10,"totalItems":0,"totalPages":0},"scoring":{"targetBudget":null,"budgetWeight":0,"recencyWeight":1,"recencyHalfLifeDays":30,"asOf":"2026-10-14"},"appliedFilters":{"minPrice":null,"maxPrice":null,"minBedrooms":null,"city":null,"keyword":"_","status":["active","pending","sold"]}}
```
- Expected: 200 with 0 items
- Result: **PASS**

#### C31 pageSize=3&page=2 (row 31)

```
curl -sS -D - 'http://localhost:18005/api/listings/search?pageSize=3&page=2'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `c0d723bc04489d74c27431c2e4265bdc`
- Body (condensed: each item reduced to `[rank, id, status, listedDate, price, score, budgetFit, recency]`; page, scoring and appliedFilters verbatim; full body saved, 1446 bytes):

```json
{"items":[[4,"A3","active","2026-09-01",399000,0.3703,null,0.3703],[5,"B8","active","2026-08-30",527500,0.3536,null,0.3536],[6,"A1","active","2026-08-29",450000,0.3455,null,0.3455]],"page":{"page":2,"pageSize":3,"totalItems":12,"totalPages":4},"scoring":{"targetBudget":null,"budgetWeight":0,"recencyWeight":1,"recencyHalfLifeDays":30,"asOf":"2026-10-14"},"appliedFilters":{"minPrice":null,"maxPrice":null,"minBedrooms":null,"city":null,"keyword":null,"status":["active","pending","sold"]}}
```
- Expected: 200, ranks 4, 5, 6 (A3, B8, A1), totalPages 4
- Result: **PASS**

## D. Headers and methods

#### D1 inbound X-Request-ID echoed (search)

```
curl -sS -D - -H 'X-Request-ID: bc5-check' 'http://localhost:18005/api/listings/search?city=Reston'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `bc5-check`
- Expected: response header `X-Request-ID: bc5-check`
- Result: **PASS**

#### D1b inbound X-Request-ID echoed (cities)

```
curl -sS -D - -H 'X-Request-ID: bc5-check' 'http://localhost:18005/api/cities'
```

- Status: `HTTP/1.1 200 OK`
- X-Request-ID: `bc5-check`
- Expected: response header `X-Request-ID: bc5-check`
- Result: **PASS**

#### D2 POST /api/listings/search

```
curl -sS -D - -X POST 'http://localhost:18005/api/listings/search'
```

```
HTTP/1.1 405 Method Not Allowed
Allow: GET, HEAD
Content-Type: application/json; charset=utf-8
Vary: Origin
X-Request-Id: 887d275aea9d78948576d3cbbe7d122c
Date: Wed, 14 Oct 2026 03:47:16 GMT
Content-Length: 72

{"error":{"code":"method_not_allowed","message":"Method not allowed."}}
```

- Expected: 405 `method_not_allowed` with an `Allow` header and X-Request-ID
- Result: **PASS**

#### D3 Supplementary: CORS preflight and unknown route

```
curl -sS -D - -X OPTIONS -H 'Origin: http://localhost:5173' -H 'Access-Control-Request-Method: GET' -H 'Access-Control-Request-Headers: X-Request-ID' 'http://localhost:18005/api/listings/search'
```

```
HTTP/1.1 204 No Content
Access-Control-Allow-Headers: Content-Type, X-Request-ID
Access-Control-Allow-Methods: GET, OPTIONS
Access-Control-Allow-Origin: http://localhost:5173
Access-Control-Expose-Headers: X-Request-ID
Vary: Origin
X-Request-Id: 766b2e49bf693f0f056dc98ea0a3fa4b
Date: Wed, 14 Oct 2026 03:48:15 GMT
```

```
curl -sS -D - 'http://localhost:18005/api/nope'
```

```
HTTP/1.1 404 Not Found
Content-Type: application/json; charset=utf-8
Vary: Origin
X-Request-Id: 59183fb95023520aafd4c10babaf2d97
Date: Wed, 14 Oct 2026 03:48:15 GMT
Content-Length: 63

{"error":{"code":"not_found","message":"Resource not found."}}
```

- Expected: preflight 204 with allow-origin/methods/headers and expose X-Request-ID (§5.6); unknown route 404 `not_found`; both carry X-Request-ID
- Result: **PASS**

#### D4 X-Request-ID on every response

| response | status | X-Request-ID |
|---|---|---|
| A1_cities | 200 | `3a8e141378788a4c26c044a0f0302d9e` |
| A2_springfield | 200 | `9a07a7137a8a6853bb6667068e6bda4b` |
| A3_minmax | 400 | `6b01d3e1686e75b01f2fc799b85a5656` |
| B1_pending | 200 | `bd90626ef5edf86f092eaa2c89a2efdb` |
| B2_kwpct | 200 | `28efd5fe3b2d3096ebb6d332787702cb` |
| B3_ps0 | 400 | `5ff249759cc8427d58b8d3b4b7d6cc48` |
| B4_minprice | 400 | `a9a7de0999890f34a621e1122ff65a26` |
| C14_city101 | 400 | `b734b48eef6d8700d2dd1e788da06ada` |
| C14_kw100pct | 400 | `796c440ebb44e4c9104f698b491c549e` |
| C14_kw201 | 400 | `8ec81d75e2debec9486ea6584c042ef8` |
| C14_minBed_25 | 400 | `3f1c22e233b40415790c0402d406dafd` |
| C14_minPrice_abc | 400 | `0469ef589200de9b220e5559ddecaf99` |
| C14_minPrice_neg | 400 | `842d3a04143a2f047a4f725abbc85402` |
| C14_minmax | 400 | `7dab7965d8ac7c22bfb54cee7f939f2e` |
| C14_minprice | 400 | `2f271a896028d2f9a675c626781d08ae` |
| C14_page0 | 400 | `9cedfd3da94be700499722f8724db52a` |
| C14_ps0 | 400 | `b1f5cb33348162a412a121a5e4185ed1` |
| C14_ps101 | 400 | `587dd3d2f6c0ae52d24ec82e94c76eb5` |
| C14_spec55 | 400 | `e4447d235a891075dfb51a82e7df5cc6` |
| C14_status_foo | 400 | `d8592c029f858acde9857749f6842d74` |
| C14_tb0 | 400 | `b108f6485be6e586768ab32145dd2453` |
| C14_three | 400 | `24e32d5ed131d3c0f5440b22531e1f30` |
| C15_minPrice_empty | 200 | `3140bffb584d88c99ec64e0244384c3d` |
| C16_page99 | 200 | `b7f61682d1342f0cdb3c0c0b410577d7` |
| C17_atlantis | 200 | `5565622323ebf7e0983716419b3f5c2e` |
| C18_comma | 200 | `769e705556d6f8ec0968e9e322386730` |
| C18_comma12 | 200 | `c0e0aa3feb69d3f8d3f1a14ed8ce2faf` |
| C18_pending | 200 | `c328803127336bba7c3a22189d54359e` |
| C18_repeat | 200 | `529e271a366efa9bb1063cc4fdbda4a3` |
| C18_repeat12 | 200 | `ff55828c771383904c010ccef4fd1eeb` |
| C19_springfield | 200 | `1ef9b24d7de88e61e0c02c0fb12fa363` |
| C20_ps12 | 200 | `7631a9ecf5eb6cc00ccbaec358340865` |
| C28_pct | 200 | `7011dcb602979b1239140b7e00545ae9` |
| C28_us | 200 | `2fdb1d9ac6f7d23c6f31e72945a52012` |
| C31_ps3p2 | 200 | `c0d723bc04489d74c27431c2e4265bdc` |
| D1_reqid | 200 | `bc5-check` |
| D2_post | 405 | `887d275aea9d78948576d3cbbe7d122c` |
| D3_cities_reqid | 200 | `bc5-check` |
| D4_preflight | 204 | `766b2e49bf693f0f056dc98ea0a3fa4b` |
| D5_404 | 404 | `59183fb95023520aafd4c10babaf2d97` |

Every one of the 40 responses (200, 204, 400, 404, 405) carries X-Request-ID. Result: **PASS**

## E. Relevance numbers

### E1 Live server, recomputed from SPEC §6.2 at the response's asOf

`scoring.asOf` read from the row 19 response: `2026-10-14`. A python3 script recomputed every item from `data/sample_listings.json` (age = whole days listedDate -> asOf; recency = 0.5^(age/30); budgetFit = price/target when under or at budget, max(0, 1 - 2*(price-target)/target) when over; score = 0.6*fit + 0.4*recency with a budget, recency without; rounded to 4 decimals) and compared with the responses for row 19, row 20 (all 12), row 31 and status=pending. It also checks each item's 15 listing fields equal the seed record.

```
-- C19_springfield.body  asOf=2026-10-14 targetBudget=450000
rank  1 A1   age  46d  fit 1.000000 rec 0.345478 score 0.738191 | expected {'score': 0.7382, 'budgetFit': 1.0, 'recency': 0.3455} | got {'score': 0.7382, 'budgetFit': 1, 'recency': 0.3455} | MATCH | seedfields equal
rank  2 B7   age  48d  fit 0.991111 rec 0.329877 score 0.726617 | expected {'score': 0.7266, 'budgetFit': 0.9911, 'recency': 0.3299} | got {'score': 0.7266, 'budgetFit': 0.9911, 'recency': 0.3299} | MATCH | seedfields equal
rank  3 A2   age  42d  fit 0.666667 rec 0.378929 score 0.551572 | expected {'score': 0.5516, 'budgetFit': 0.6667, 'recency': 0.3789} | got {'score': 0.5516, 'budgetFit': 0.6667, 'recency': 0.3789} | MATCH | seedfields equal
rank  4 B8   age  45d  fit 0.655556 rec 0.353553 score 0.534755 | expected {'score': 0.5348, 'budgetFit': 0.6556, 'recency': 0.3536} | got {'score': 0.5348, 'budgetFit': 0.6556, 'recency': 0.3536} | MATCH | seedfields equal
-- C20_ps12.body  asOf=2026-10-14 targetBudget=None
rank  1 A5   age  40d  fit null rec 0.396850 score 0.396850 | expected {'score': 0.3969, 'budgetFit': None, 'recency': 0.3969} | got {'score': 0.3969, 'budgetFit': None, 'recency': 0.3969} | MATCH | seedfields equal
rank  2 A4   age  41d  fit null rec 0.387786 score 0.387786 | expected {'score': 0.3878, 'budgetFit': None, 'recency': 0.3878} | got {'score': 0.3878, 'budgetFit': None, 'recency': 0.3878} | MATCH | seedfields equal
rank  3 A2   age  42d  fit null rec 0.378929 score 0.378929 | expected {'score': 0.3789, 'budgetFit': None, 'recency': 0.3789} | got {'score': 0.3789, 'budgetFit': None, 'recency': 0.3789} | MATCH | seedfields equal
rank  4 A3   age  43d  fit null rec 0.370274 score 0.370274 | expected {'score': 0.3703, 'budgetFit': None, 'recency': 0.3703} | got {'score': 0.3703, 'budgetFit': None, 'recency': 0.3703} | MATCH | seedfields equal
rank  5 B8   age  45d  fit null rec 0.353553 score 0.353553 | expected {'score': 0.3536, 'budgetFit': None, 'recency': 0.3536} | got {'score': 0.3536, 'budgetFit': None, 'recency': 0.3536} | MATCH | seedfields equal
rank  6 A1   age  46d  fit null rec 0.345478 score 0.345478 | expected {'score': 0.3455, 'budgetFit': None, 'recency': 0.3455} | got {'score': 0.3455, 'budgetFit': None, 'recency': 0.3455} | MATCH | seedfields equal
rank  7 B7   age  48d  fit null rec 0.329877 score 0.329877 | expected {'score': 0.3299, 'budgetFit': None, 'recency': 0.3299} | got {'score': 0.3299, 'budgetFit': None, 'recency': 0.3299} | MATCH | seedfields equal
rank  8 B9   age  50d  fit null rec 0.314980 score 0.314980 | expected {'score': 0.315, 'budgetFit': None, 'recency': 0.315} | got {'score': 0.315, 'budgetFit': None, 'recency': 0.315} | MATCH | seedfields equal
rank  9 B10  age  55d  fit null rec 0.280616 score 0.280616 | expected {'score': 0.2806, 'budgetFit': None, 'recency': 0.2806} | got {'score': 0.2806, 'budgetFit': None, 'recency': 0.2806} | MATCH | seedfields equal
rank 10 B11  age  60d  fit null rec 0.250000 score 0.250000 | expected {'score': 0.25, 'budgetFit': None, 'recency': 0.25} | got {'score': 0.25, 'budgetFit': None, 'recency': 0.25} | MATCH | seedfields equal
rank 11 A6   age  65d  fit null rec 0.222725 score 0.222725 | expected {'score': 0.2227, 'budgetFit': None, 'recency': 0.2227} | got {'score': 0.2227, 'budgetFit': None, 'recency': 0.2227} | MATCH | seedfields equal
rank 12 A7   age  78d  fit null rec 0.164938 score 0.164938 | expected {'score': 0.1649, 'budgetFit': None, 'recency': 0.1649} | got {'score': 0.1649, 'budgetFit': None, 'recency': 0.1649} | MATCH | seedfields equal
-- C31_ps3p2.body  asOf=2026-10-14 targetBudget=None
rank  4 A3   age  43d  fit null rec 0.370274 score 0.370274 | expected {'score': 0.3703, 'budgetFit': None, 'recency': 0.3703} | got {'score': 0.3703, 'budgetFit': None, 'recency': 0.3703} | MATCH | seedfields equal
rank  5 B8   age  45d  fit null rec 0.353553 score 0.353553 | expected {'score': 0.3536, 'budgetFit': None, 'recency': 0.3536} | got {'score': 0.3536, 'budgetFit': None, 'recency': 0.3536} | MATCH | seedfields equal
rank  6 A1   age  46d  fit null rec 0.345478 score 0.345478 | expected {'score': 0.3455, 'budgetFit': None, 'recency': 0.3455} | got {'score': 0.3455, 'budgetFit': None, 'recency': 0.3455} | MATCH | seedfields equal
-- B1_pending.body  asOf=2026-10-14 targetBudget=None
rank  1 A7   age  78d  fit null rec 0.164938 score 0.164938 | expected {'score': 0.1649, 'budgetFit': None, 'recency': 0.1649} | got {'score': 0.1649, 'budgetFit': None, 'recency': 0.1649} | MATCH | seedfields equal
ALL MATCH
```

- Expected at asOf 2026-10-14: A1 0.7382, B7 0.7266, A2 0.5516, B8 0.5348; every value equal to the recomputation
- Observed: A1 0.7382, B7 0.7266, A2 0.5516, B8 0.5348
- Result: **PASS**

### E2 Fixed clock (asOf 2026-10-13)

```
cd backend && go test -race -count=1 -run 'TestSearchRankedSpringfield' -v ./internal/httpapi/
```

```
=== RUN   TestSearchRankedSpringfield
    search_test.go:413: rank 1 MLS_A/A1 price 450000 listed 2026-08-29: score 0.7414 budgetFit 1 recency 0.3536
    search_test.go:413: rank 2 MLS_B/B7 price 452000 listed 2026-08-27: score 0.7297 budgetFit 0.9911 recency 0.3376
    search_test.go:413: rank 3 MLS_A/A2 price 525000 listed 2026-09-02: score 0.5551 budgetFit 0.6667 recency 0.3878
    search_test.go:413: rank 4 MLS_B/B8 price 527500 listed 2026-08-30: score 0.5381 budgetFit 0.6556 recency 0.3618
--- PASS: TestSearchRankedSpringfield (0.03s)
PASS
ok  	brillio/backend/internal/httpapi	1.479s
```

Independent python3 recomputation at asOf 2026-10-13 (id, age, budgetFit, recency, score): A1 45 1.0 0.3536 0.7414; B7 47 0.9911 0.3376 0.7297; A2 41 0.6667 0.3878 0.5551; B8 44 0.6556 0.3618 0.5381. Matches the test log.

- Expected: test passes; A1 0.7414 at asOf 2026-10-13
- Result: **PASS**

## Summary

| item | check | result |
|---|---|---|
| A1 | GET /api/cities | PASS |
| A2 | targetBudget=450000&city=Springfield | PASS |
| A3 | minPrice=5&maxPrice=1 | PASS |
| B1 | status=pending | PASS |
| B2 | keyword=%25 (encoded percent) | PASS |
| B3 | pageSize=0 | PASS |
| B4 | minprice=1 (lowercase typo) | PASS |
| C13 | row 13 cities exact body | PASS |
| C14a | minPrice=-1 | PASS |
| C14b | minPrice=abc | PASS |
| C14c | minBedrooms=2.5 | PASS |
| C14d | targetBudget=0 | PASS |
| C14e | status=foo | PASS |
| C14f | page=0 | PASS |
| C14g | pageSize=0 | PASS |
| C14h | pageSize=101 | PASS |
| C14i | city=<101 chars> | PASS |
| C14j | keyword=<201 chars> | PASS |
| C14k | minPrice=5&maxPrice=1 | PASS |
| C14l | minprice=1 | PASS |
| C14m | three faults in one request | PASS |
| C14n | keyword=100% (bare percent) | PASS |
| C14o | SPEC §5.5 combined request | PASS |
| C15 | minPrice= (empty value is absent) | PASS |
| C16 | page=99 (past the end) | PASS |
| C17 | city=Atlantis (no matches) | PASS |
| C18a | status=pending | PASS |
| C18b | status=active,pending | PASS |
| C18c | status=active&status=pending | PASS |
| C18d | status comma vs repeat, pageSize=12, 12 items identical | PASS |
| C19 | targetBudget=450000&city=Springfield (row 19) | PASS |
| C20 | pageSize=12, no targetBudget (row 20) | PASS |
| C28a | keyword=%25 (row 28) | PASS |
| C28b | keyword=_ (row 28) | PASS |
| C31 | pageSize=3&page=2 (row 31) | PASS |
| D1 | inbound X-Request-ID echoed (search) | PASS |
| D1b | inbound X-Request-ID echoed (cities) | PASS |
| D2 | POST -> 405 with Allow | PASS |
| D3 | preflight 204, 404 envelope (supplementary) | PASS |
| D4 | X-Request-ID on every response (200/204/400/404/405) | PASS |
| E1 | live relevance numbers equal §6.2 at asOf 2026-10-14 | PASS |
| E2 | fixed-clock test, A1 0.7414 at asOf 2026-10-13 | PASS |

**Overall verdict: PASS** (42/42 checks passed).

## Teardown

- Server PID 82064 stopped with SIGTERM at 03:48:39 UTC; log ends with `shutting down` and `server stopped`; `kill -0 82064` fails (process gone).
- `lsof -nP -iTCP:18005 -sTCP:LISTEN` after shutdown: no output (exit 1), port 18005 is free.
- No source files edited, nothing committed; `git status --short` in the worktree is empty.
