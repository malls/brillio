# BC-6: OpenAPI 3.1 document, embedded and served at /api/openapi.yaml, with param sync test

Write the OpenAPI 3.1 document for the API and serve it from the binary. Spec: docs/SPEC.md §5.7 and §5.1–5.6.

Why: senior bar. It also gives the interviewer a precise contract to read, and the sync test keeps the handler's param allowlist and the document from drifting.

Scope:
- backend/openapi.yaml, OpenAPI 3.1: paths /healthz, /api/cities, /api/listings/search, /api/openapi.yaml. Components: Listing, ListingResult (with rank and relevance), PageMeta, ScoringMeta, SearchResponse, ErrorEnvelope, FieldError with the code enum from §5.5. Every query param with type, constraints, default and a one-line description; document the unknown-parameter rule and the page-past-end and empty-city behavior in the operation description. Examples lifted from §5.4 and §5.5.
- Embed with embed.FS and serve at GET /api/openapi.yaml with Content-Type application/yaml.
- Tests: handler returns 200 and the body parses as YAML; a sync test that reads the allowlist of param names from the search handler and asserts each appears under the search operation's parameters, and vice versa.
- Run one external validator once (for example `npx @redocly/cli lint backend/openapi.yaml`) and record its output in a task comment; do not add it to make check.

Acceptance:
- curl /api/openapi.yaml returns the document.
- Validator reports no errors.
- Sync test green; `make check` green.

Validation: the curl and the validator output, attached with --role validation.
