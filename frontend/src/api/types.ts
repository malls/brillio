// API types. Mirrors docs/SPEC.md sections 5.3-5.5 exactly (the block in SPEC section 7.3);
// change SPEC first, then this file.

export type ListingStatus = 'active' | 'pending' | 'sold';

export interface SearchParams {
  minPrice?: number;
  maxPrice?: number;
  minBedrooms?: number;
  city?: string;
  keyword?: string;
  targetBudget?: number;
  status?: ListingStatus[];
  page?: number;
  pageSize?: number;
}

export interface ListingResult {
  rank: number;
  id: string;
  source: string;
  address: string;
  city: string;
  state: string;
  zip: string;
  price: number;
  bedrooms: number;
  bathrooms: number;
  sqft: number;
  latitude: number;
  longitude: number;
  listedDate: string;
  status: ListingStatus;
  description: string;
  relevance: { score: number; budgetFit: number | null; recency: number };
}

export interface PageMeta {
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
}

export interface ScoringMeta {
  targetBudget: number | null;
  budgetWeight: number;
  recencyWeight: number;
  recencyHalfLifeDays: number;
  asOf: string;
}

export interface SearchResponse {
  items: ListingResult[];
  page: PageMeta;
  scoring: ScoringMeta;
  appliedFilters: Record<string, unknown>;
}

export interface FieldError {
  field: string;
  code: string;
  message: string;
}

export type ValidationError = { kind: 'validation'; message: string; fields: FieldError[] };

export type ApiError =
  | ValidationError
  | { kind: 'http'; status: number; message: string }
  | { kind: 'network'; message: string };

// Every listing status in the API's enum order (SPEC section 5.3). The URL parser, the query
// serializer and the Status select all read this list.
export const LISTING_STATUSES = [
  'active',
  'pending',
  'sold',
] as const satisfies readonly ListingStatus[];
