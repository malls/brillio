import { useCallback, useEffect, useMemo, useState } from 'react';
import { buildSearchQuery, searchListingsByQuery } from '../api/client';
import { ApiRequestError, toApiError } from '../api/errors';
import type { ApiError, SearchParams, SearchResponse, ValidationError } from '../api/types';

// The fetch lifecycle (docs/SPEC.md section 7.2).
export type SearchState =
  | { status: 'idle' }
  // requestedPage is set only with previous, when the request differs from previous's only by
  // page: a page change, the page asked for (SPEC section 7.2).
  | { status: 'loading'; previous?: SearchResponse; requestedPage?: number }
  | { status: 'success'; data: SearchResponse } // totalItems > 0; items may be empty on a page past the end
  | { status: 'empty'; data: SearchResponse } // totalItems === 0
  | { status: 'invalid'; error: ValidationError; previous?: SearchResponse }
  | { status: 'error'; error: ApiError; previous?: SearchResponse };

export interface ListingSearch {
  state: SearchState;
  // Re-runs the current params (Try again, BRANDING.md section 5.9). The only way to refetch
  // unchanged params.
  retry: () => void;
}

type Outcome = { kind: 'data'; data: SearchResponse } | { kind: 'error'; error: ApiError };

interface FetchState {
  // The outcome of the last request that completed, keyed by its request key.
  settled: { key: string; outcome: Outcome } | null;
  // The latest 2xx response, success or empty (the list kept on screen while loading), and the
  // search key it was fetched with.
  lastGood: { searchKey: string; data: SearchResponse } | null;
}

const INITIAL: FetchState = { settled: null, lastGood: null };

// Identifies a request apart from its page: the attempt and the query without page. Two requests
// with the same search key differ only by page, so a retry or a pageSize change is never a page
// change.
function toSearchKey(attempt: number, params: SearchParams): string {
  const rest: SearchParams = { ...params };
  delete rest.page;
  return `${attempt}|${buildSearchQuery(rest)}`;
}

// Loading is derived, never set: the state is loading until the request for the current key has
// settled. Returning to a key whose result is held shows that result while it refetches. A load
// whose search key matches the list on screen is a page change and carries the page asked for.
function deriveState(
  { settled, lastGood }: FetchState,
  requestKey: string,
  searchKey: string,
  page: number,
): SearchState {
  const previous = lastGood === null ? {} : { previous: lastGood.data };
  if (settled?.key !== requestKey) {
    const paging = lastGood?.searchKey === searchKey ? { requestedPage: page } : {};
    return { status: 'loading', ...previous, ...paging };
  }
  const { outcome } = settled;
  if (outcome.kind === 'data') {
    return outcome.data.page.totalItems === 0
      ? { status: 'empty', data: outcome.data }
      : { status: 'success', data: outcome.data };
  }
  if (outcome.error.kind === 'validation') {
    return { status: 'invalid', error: outcome.error, ...previous };
  }
  return { status: 'error', error: outcome.error, ...previous };
}

// Fetches listings for params, aborting the previous request whenever params change, on retry
// and on unmount; an aborted request never updates state. Params are compared by their
// serialized query, so an equal new object does not refetch. A request that differs from the
// list on screen only by page loads as a page change. http and network failures are logged here
// with the request ID (SPEC section 7.5); callers must not log them again.
export function useListingSearch(params: SearchParams): ListingSearch {
  const query = buildSearchQuery(params);
  const [attempt, setAttempt] = useState(0);
  const [fetchState, setFetchState] = useState(INITIAL);
  const requestKey = `${attempt}|${query}`;
  const searchKey = toSearchKey(attempt, params);
  const page = params.page ?? 1;

  useEffect(() => {
    const controller = new AbortController();
    searchListingsByQuery(query, controller.signal).then(
      (data) => {
        if (controller.signal.aborted) {
          return;
        }
        setFetchState({
          settled: { key: requestKey, outcome: { kind: 'data', data } },
          lastGood: { searchKey, data },
        });
      },
      (err: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        const error = toApiError(err);
        if (error.kind !== 'validation') {
          const requestId = err instanceof ApiRequestError ? err.requestId : null;
          console.error('Listing search failed', { ...error, requestId });
        }
        setFetchState((prev) => ({
          settled: { key: requestKey, outcome: { kind: 'error', error } },
          lastGood: prev.lastGood,
        }));
      },
    );
    return () => {
      controller.abort();
    };
  }, [query, requestKey, searchKey]);

  const retry = useCallback(() => {
    setAttempt((n) => n + 1);
  }, []);

  const state = useMemo(
    () => deriveState(fetchState, requestKey, searchKey, page),
    [fetchState, requestKey, searchKey, page],
  );
  return useMemo(() => ({ state, retry }), [state, retry]);
}
