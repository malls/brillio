import { ApiRequestError, parseErrorResponse } from './errors';
import type { ApiError, ListingStatus, SearchParams, SearchResponse } from './types';
import { LISTING_STATUSES } from './types';

// The API client (docs/SPEC.md section 7.3). Every call resolves with the parsed body or rejects
// with an ApiRequestError carrying an ApiError: validation and http errors from non-2xx bodies,
// network for fetch failures and aborts.

const SEARCH_PATH = '/api/listings/search';
const CITIES_PATH = '/api/cities';
const MALFORMED_MESSAGE = 'Malformed response body.';

function aborted(): ApiError {
  return { kind: 'network', message: 'Request aborted.' };
}

function appendNumber(query: URLSearchParams, name: string, value: number | undefined): void {
  // String(NaN) is "NaN", which Go's ParseFloat accepts, so non-finite numbers are omitted.
  if (value !== undefined && Number.isFinite(value)) {
    query.append(name, String(value));
  }
}

function appendText(query: URLSearchParams, name: string, value: string | undefined): void {
  const trimmed = value?.trim() ?? '';
  if (trimmed !== '') {
    query.append(name, trimmed);
  }
}

function appendStatuses(query: URLSearchParams, statuses: ListingStatus[] | undefined): void {
  // Enum order and deduplicated, so equal sets give equal strings. An empty set is absent,
  // which the API reads as all three.
  for (const status of LISTING_STATUSES) {
    if (statuses?.includes(status) === true) {
      query.append('status', status);
    }
  }
}

// The one serializer for SearchParams (SPEC sections 5.3 and 7.1). Its output is the API query,
// the pushState query and the useListingSearch equality key, so it is canonical: params in the
// SPEC section 5.3 table order, undefined, empty and non-finite values omitted, city and keyword
// trimmed, status repeated once per value. No leading '?'.
export function buildSearchQuery(params: SearchParams): string {
  const query = new URLSearchParams();
  appendNumber(query, 'minPrice', params.minPrice);
  appendNumber(query, 'maxPrice', params.maxPrice);
  appendNumber(query, 'minBedrooms', params.minBedrooms);
  appendText(query, 'city', params.city);
  appendText(query, 'keyword', params.keyword);
  appendNumber(query, 'targetBudget', params.targetBudget);
  appendStatuses(query, params.status);
  appendNumber(query, 'page', params.page);
  appendNumber(query, 'pageSize', params.pageSize);
  return query.toString();
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function isAbortError(err: unknown): boolean {
  // Checked by name: jsdom and Node each have their own DOMException, so instanceof is unreliable.
  return isRecord(err) && err.name === 'AbortError';
}

function networkError(err: unknown, signal: AbortSignal | undefined): ApiError {
  if (signal?.aborted === true || isAbortError(err)) {
    return aborted();
  }
  const message = err instanceof Error && err.message !== '' ? err.message : 'Network error.';
  return { kind: 'network', message };
}

// Fetches url and returns its 2xx body as accepted by accept, or rejects with an
// ApiRequestError. accept returns undefined for a body of the wrong shape.
async function request<T>(
  url: string,
  signal: AbortSignal | undefined,
  accept: (body: unknown) => T | undefined,
): Promise<T> {
  let res: Response;
  try {
    res = await fetch(url, { signal: signal ?? null, headers: { Accept: 'application/json' } });
  } catch (err) {
    throw new ApiRequestError(networkError(err, signal), null, { cause: err });
  }

  const requestId = res.headers.get('X-Request-ID');
  if (!res.ok) {
    throw new ApiRequestError(await parseErrorResponse(res), requestId);
  }

  const malformed: ApiError = { kind: 'http', status: res.status, message: MALFORMED_MESSAGE };
  let body: unknown;
  try {
    body = await res.json();
  } catch (err) {
    // A body read cut short by an abort is an abort, not a malformed body.
    const apiError = signal?.aborted === true || isAbortError(err) ? aborted() : malformed;
    throw new ApiRequestError(apiError, requestId, { cause: err });
  }
  const result = accept(body);
  if (result === undefined) {
    throw new ApiRequestError(malformed, requestId);
  }
  return result;
}

// Shape guard: the parts of a search body the UI cannot render without.
function isSearchResponse(body: unknown): body is SearchResponse {
  return (
    isRecord(body) && Array.isArray(body.items) && isRecord(body.page) && isRecord(body.scoring)
  );
}

function acceptSearchResponse(body: unknown): SearchResponse | undefined {
  return isSearchResponse(body) ? body : undefined;
}

// Non-string entries are dropped.
function acceptCities(body: unknown): string[] | undefined {
  if (!isRecord(body) || !Array.isArray(body.cities)) {
    return undefined;
  }
  return body.cities.filter((city): city is string => typeof city === 'string');
}

// Searches with an already-serialized query (the output of buildSearchQuery).
export function searchListingsByQuery(
  query: string,
  signal?: AbortSignal,
): Promise<SearchResponse> {
  const url = query === '' ? SEARCH_PATH : `${SEARCH_PATH}?${query}`;
  return request(url, signal, acceptSearchResponse);
}

export function searchListings(
  params: SearchParams,
  signal?: AbortSignal,
): Promise<SearchResponse> {
  return searchListingsByQuery(buildSearchQuery(params), signal);
}

// Loads the City select's options (SPEC section 5.2).
export function getCities(signal?: AbortSignal): Promise<string[]> {
  return request(CITIES_PATH, signal, acceptCities);
}
