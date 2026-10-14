import type { FieldError, ListingResult, SearchResponse } from '../api/types';

// Shared test fixtures. Listing rows are real records from data/sample_listings.json (BRANDING.md
// section 8: no placeholder data); rank and relevance come from the full 12-listing ranking of
// docs/SPEC.md sections 6.2-6.3 with targetBudget=450000 as of 2026-10-13. Every export is
// typed, so `tsc -b` fails when src/api/types.ts drifts. Tests spread these; never mutate them.
// Tickets that add fixtures do so in a labelled block at the end of the file.

// A1, Springfield: rank 1 of 12.
export const listingA1: ListingResult = {
  rank: 1,
  id: 'A1',
  source: 'MLS_A',
  address: '123 Main St, Apt 4B',
  city: 'Springfield',
  state: 'VA',
  zip: '22150',
  price: 450000,
  bedrooms: 2,
  bathrooms: 1.5,
  sqft: 980,
  latitude: 38.7893,
  longitude: -77.1873,
  listedDate: '2026-08-29',
  status: 'active',
  description: 'Bright top-floor condo near shops and transit. Pet friendly.',
  relevance: { score: 0.7414, budgetFit: 1, recency: 0.3536 },
};

// A3, Fairfax: rank 4 of 12; a whole number of bathrooms ('1 ba').
export const listingA3: ListingResult = {
  rank: 4,
  id: 'A3',
  source: 'MLS_A',
  address: '789 Pine Rd',
  city: 'Fairfax',
  state: 'VA',
  zip: '22030',
  price: 399000,
  bedrooms: 2,
  bathrooms: 1,
  sqft: 850,
  latitude: 38.8462,
  longitude: -77.3064,
  listedDate: '2026-09-01',
  status: 'active',
  description: 'Cozy starter home, no pets.',
  relevance: { score: 0.6836, budgetFit: 0.8867, recency: 0.3789 },
};

// A7, Chantilly: rank 10 of 12; the one pending listing.
export const listingA7: ListingResult = {
  rank: 10,
  id: 'A7',
  source: 'MLS_A',
  address: '42 Willow Way',
  city: 'Chantilly',
  state: 'VA',
  zip: '20151',
  price: 540000,
  bedrooms: 4,
  bathrooms: 2.5,
  sqft: 1950,
  latitude: 38.8909,
  longitude: -77.4316,
  listedDate: '2026-07-28',
  status: 'pending',
  description: 'Corner lot, recently painted, no pets due to HOA.',
  relevance: { score: 0.4275, budgetFit: 0.6, recency: 0.1688 },
};

// Page 1 of the real search `targetBudget=450000` (12 results, 2 pages). The metadata is the
// real page's; `items` is trimmed to three of its ten rows, so never compare items.length with
// page.pageSize or page.totalItems.
export const searchResponse: SearchResponse = {
  items: [listingA1, listingA3, listingA7],
  page: { page: 1, pageSize: 10, totalItems: 12, totalPages: 2 },
  scoring: {
    targetBudget: 450000,
    budgetWeight: 0.6,
    recencyWeight: 0.4,
    recencyHalfLifeDays: 30,
    asOf: '2026-10-13',
  },
  appliedFilters: {
    minPrice: null,
    maxPrice: null,
    minBedrooms: null,
    city: null,
    keyword: null,
    status: ['active', 'pending', 'sold'],
  },
};

// The real search `city=Arlington`: no listings, no budget (SPEC section 5.4's no-budget scoring).
export const emptySearchResponse: SearchResponse = {
  items: [],
  page: { page: 1, pageSize: 10, totalItems: 0, totalPages: 0 },
  scoring: {
    targetBudget: null,
    budgetWeight: 0,
    recencyWeight: 1,
    recencyHalfLifeDays: 30,
    asOf: '2026-10-13',
  },
  appliedFilters: {
    minPrice: null,
    maxPrice: null,
    minBedrooms: null,
    city: 'Arlington',
    keyword: null,
    status: ['active', 'pending', 'sold'],
  },
};

// GET /api/cities (SPEC section 5.2).
export const cities: string[] = [
  'Chantilly',
  'Fairfax',
  'Manassas',
  'Reston',
  'Springfield',
  'Vienna',
];

// The error envelope (SPEC section 5.5).
export interface ErrorBody {
  error: { code: string; message: string; fields?: FieldError[] };
}

// SPEC section 5.5's example 400 body, verbatim.
export const validationErrorBody: ErrorBody = {
  error: {
    code: 'validation_error',
    message: 'Invalid search parameters.',
    fields: [
      { field: 'maxPrice', code: 'min_gt_max', message: 'Min price must be less than max price.' },
      { field: 'pageSize', code: 'out_of_range', message: 'Page size must be between 1 and 100.' },
      { field: 'minprice', code: 'unknown_param', message: 'Unknown parameter.' },
    ],
  },
};

// The API's 500 body (backend/internal/httpapi/respond.go).
export const internalErrorBody: ErrorBody = {
  error: { code: 'internal_error', message: 'Internal server error.' },
};

export interface JsonResponseInit {
  status?: number;
  requestId?: string;
}

// A JSON response as the API sends it: JSON content type and an X-Request-ID header.
export function jsonResponse(body: unknown, init: JsonResponseInit = {}): Response {
  return new Response(JSON.stringify(body), {
    status: init.status ?? 200,
    headers: {
      'Content-Type': 'application/json; charset=utf-8',
      'X-Request-ID': init.requestId ?? 'test-request-id',
    },
  });
}

export interface Deferred<T> {
  promise: Promise<T>;
  resolve: (value: T) => void;
  reject: (error: Error) => void;
}

// A promise settled from outside, for ordering async steps in tests (the ES2023 lib has no
// Promise.withResolvers).
export function deferred<T>(): Deferred<T> {
  let resolve: (value: T) => void = () => undefined;
  let reject: (error: Error) => void = () => undefined;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

// --- BC-10 additions ---

// B7, Springfield: rank 2 of 12 and the demo's second card (docs/branding-demo.html); its
// description carries a capital 'Pets' for the keyword-emphasis tests.
export const listingB7: ListingResult = {
  rank: 2,
  id: 'B7',
  source: 'MLS_B',
  address: '123 Main Street, Unit 4B',
  city: 'Springfield',
  state: 'VA',
  zip: '22150',
  price: 452000,
  bedrooms: 2,
  bathrooms: 1.5,
  sqft: 980,
  latitude: 38.7893,
  longitude: -77.1873,
  listedDate: '2026-08-27',
  status: 'active',
  description: 'Top floor condo, walk to shopping. Pets allowed.',
  relevance: { score: 0.7297, budgetFit: 0.9911, recency: 0.3376 },
};

// --- BC-12 additions ---

// B10, Reston: rank 11 of 12, the first card of page 2.
export const listingB10: ListingResult = {
  rank: 11,
  id: 'B10',
  source: 'MLS_B',
  address: '100 Maple Dr',
  city: 'Reston',
  state: 'VA',
  zip: '20190',
  price: 585000,
  bedrooms: 3,
  bathrooms: 2.5,
  sqft: 1900,
  latitude: 38.9601,
  longitude: -77.3499,
  listedDate: '2026-08-20',
  status: 'active',
  description: 'Townhome with 2-car garage, community pool.',
  relevance: { score: 0.3549, budgetFit: 0.4, recency: 0.2872 },
};

// A4, Reston: rank 12 of 12, the last card of page 2.
export const listingA4: ListingResult = {
  rank: 12,
  id: 'A4',
  source: 'MLS_A',
  address: '22 Birch Ln',
  city: 'Reston',
  state: 'VA',
  zip: '20190',
  price: 610000,
  bedrooms: 4,
  bathrooms: 3,
  sqft: 2100,
  latitude: 38.9586,
  longitude: -77.357,
  listedDate: '2026-09-03',
  status: 'active',
  description: 'Spacious family home near Reston Town Center. Pets welcome.',
  relevance: { score: 0.3321, budgetFit: 0.2889, recency: 0.3969 },
};

// Page 2 of the real search `targetBudget=450000`: the last two of the 12 results, ranks 11 and
// 12, so rank numbering continues across pages.
export const searchResponsePage2: SearchResponse = {
  ...searchResponse,
  items: [listingB10, listingA4],
  page: { page: 2, pageSize: 10, totalItems: 12, totalPages: 2 },
};

// Page 1 of the real search with no params (the default search on an empty URL): no budget, so
// the score is recency alone (SPEC section 5.4) and the order is newest first. Trimmed like
// searchResponse to its first three rows; rank and relevance are this search's, not the
// targetBudget=450000 ranking's.
export const newestFirstResponse: SearchResponse = {
  items: [
    {
      rank: 1,
      id: 'A5',
      source: 'MLS_A',
      address: '55 Elm Ct',
      city: 'Vienna',
      state: 'VA',
      zip: '22180',
      price: 470000,
      bedrooms: 3,
      bathrooms: 2,
      sqft: 1300,
      latitude: 38.9012,
      longitude: -77.2653,
      listedDate: '2026-09-04',
      status: 'active',
      description: 'Quiet cul-de-sac, walkable to Metro. No pets.',
      relevance: { score: 0.4061, budgetFit: null, recency: 0.4061 },
    },
    {
      ...listingA4,
      rank: 2,
      relevance: { score: 0.3969, budgetFit: null, recency: 0.3969 },
    },
    {
      rank: 3,
      id: 'A2',
      source: 'MLS_A',
      address: '456 Oak Ave',
      city: 'Springfield',
      state: 'VA',
      zip: '22150',
      price: 525000,
      bedrooms: 3,
      bathrooms: 2,
      sqft: 1450,
      latitude: 38.7791,
      longitude: -77.1901,
      listedDate: '2026-09-02',
      status: 'active',
      description: 'Updated kitchen, fenced yard, close to schools.',
      relevance: { score: 0.3878, budgetFit: null, recency: 0.3878 },
    },
  ],
  page: { page: 1, pageSize: 10, totalItems: 12, totalPages: 2 },
  scoring: {
    targetBudget: null,
    budgetWeight: 0,
    recencyWeight: 1,
    recencyHalfLifeDays: 30,
    asOf: '2026-10-13',
  },
  appliedFilters: {
    minPrice: null,
    maxPrice: null,
    minBedrooms: null,
    city: null,
    keyword: null,
    status: ['active', 'pending', 'sold'],
  },
};

// The API's 400 for `minPrice=500000&maxPrice=400000`: one min_gt_max error, on maxPrice.
export const minGtMaxErrorBody: ErrorBody = {
  error: {
    code: 'validation_error',
    message: 'Invalid search parameters.',
    fields: [
      { field: 'maxPrice', code: 'min_gt_max', message: 'Min price must be less than max price.' },
    ],
  },
};

// The API's 400 for a query string that is not valid URL encoding
// (backend/internal/httpapi/query.go): reported on the `query` pseudo-field, which no bar field
// shows.
export const malformedQueryErrorBody: ErrorBody = {
  error: {
    code: 'validation_error',
    message: 'Invalid search parameters.',
    fields: [
      {
        field: 'query',
        code: 'malformed_query',
        message: 'Query string is not valid URL encoding.',
      },
    ],
  },
};
