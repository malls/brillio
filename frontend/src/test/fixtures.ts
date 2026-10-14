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
