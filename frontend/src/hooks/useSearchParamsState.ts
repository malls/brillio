import { useCallback, useEffect, useMemo, useState } from 'react';
import { buildSearchQuery } from '../api/client';
import type { ListingStatus, SearchParams } from '../api/types';
import { LISTING_STATUSES } from '../api/types';

// URL state (docs/SPEC.md section 7.1): the query string is the search. Param names are SPEC
// section 5.3's, and buildSearchQuery writes them, so a pushed URL and the API call it triggers
// carry the same query.

const NUMBER = /^[+-]?(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?$/;

// First value wins for a repeated scalar. Empty and whitespace-only values are absent (SPEC
// section 5.3).
function readText(query: URLSearchParams, name: string): string | undefined {
  const value = query.get(name)?.trim() ?? '';
  return value === '' ? undefined : value;
}

function readNumber(query: URLSearchParams, name: string): number | undefined {
  const raw = readText(query, name);
  if (raw === undefined || !NUMBER.test(raw)) {
    return undefined;
  }
  const value = Number(raw);
  return Number.isFinite(value) ? value : undefined;
}

function readInteger(
  query: URLSearchParams,
  name: string,
  min = -Infinity,
  max = Infinity,
): number | undefined {
  const value = readNumber(query, name);
  return value !== undefined && Number.isInteger(value) && value >= min && value <= max
    ? value
    : undefined;
}

// status may be repeated or comma-separated (SPEC section 5.3). Unknown values are dropped and
// the rest kept, deduplicated, in enum order.
function readStatuses(query: URLSearchParams): ListingStatus[] | undefined {
  const values = new Set(
    query
      .getAll('status')
      .flatMap((value) => value.split(','))
      .map((value) => value.trim()),
  );
  const statuses = LISTING_STATUSES.filter((status) => values.has(status));
  return statuses.length === 0 ? undefined : statuses;
}

// Reads SearchParams from a query string (with or without the leading '?').
//
// The UI can only show the server's verdict on a value it has a field for, so this drops what
// no field could display: values of the wrong type (non-numeric or non-finite numbers,
// non-integer minBedrooms, page or pageSize, unknown status values, empty values), page below 1
// and pageSize outside 1..100. Range errors on visible fields (minPrice=-5, min above max,
// targetBudget=0, over-long text) are kept, so the API's 400 marks the field: the API is the
// authority (BRANDING.md section 5.9). Unknown params are dropped deliberately for the same
// reason: no field can display their unknown_param error. They vanish on the next push.
export function parseSearchParams(search: string): SearchParams {
  const query = new URLSearchParams(search);
  const params: SearchParams = {};

  const minPrice = readNumber(query, 'minPrice');
  if (minPrice !== undefined) params.minPrice = minPrice;
  const maxPrice = readNumber(query, 'maxPrice');
  if (maxPrice !== undefined) params.maxPrice = maxPrice;
  const minBedrooms = readInteger(query, 'minBedrooms');
  if (minBedrooms !== undefined) params.minBedrooms = minBedrooms;
  const city = readText(query, 'city');
  if (city !== undefined) params.city = city;
  const keyword = readText(query, 'keyword');
  if (keyword !== undefined) params.keyword = keyword;
  const targetBudget = readNumber(query, 'targetBudget');
  if (targetBudget !== undefined) params.targetBudget = targetBudget;
  const status = readStatuses(query);
  if (status !== undefined) params.status = status;
  const page = readInteger(query, 'page', 1);
  if (page !== undefined) params.page = page;
  const pageSize = readInteger(query, 'pageSize', 1, 100);
  if (pageSize !== undefined) params.pageSize = pageSize;

  return params;
}

export interface SearchParamsState {
  params: SearchParams;
  // Submits the filter bar: writes next with page omitted (= 1), keeping pageSize.
  setSearch: (next: SearchParams) => void;
  // Changes only page; 1 is written as no page. A non-integer or a page below 1 does nothing.
  setPage: (page: number) => void;
  // Writes an empty query, except for pageSize, which is preserved like everywhere else.
  clear: () => void;
}

function currentParams(): SearchParams {
  return parseSearchParams(window.location.search);
}

export function useSearchParamsState(): SearchParamsState {
  const [params, setParams] = useState(currentParams);

  useEffect(() => {
    const onPopState = () => {
      setParams(currentParams());
    };
    window.addEventListener('popstate', onPopState);
    return () => {
      window.removeEventListener('popstate', onPopState);
    };
  }, []);

  // Writers read the live URL, so they never go stale and keep a stable identity. A write equal
  // to the current query adds no history entry. pushState fires no popstate, so the new params
  // are set here.
  const navigate = useCallback((next: SearchParams) => {
    const query = buildSearchQuery(next);
    const { pathname, search, hash } = window.location;
    if (query === search.replace(/^\?/, '')) {
      return;
    }
    // The path is always written out: a relative '' URL would keep the old query.
    window.history.pushState(
      null,
      '',
      query === '' ? pathname + hash : `${pathname}?${query}${hash}`,
    );
    setParams(parseSearchParams(query));
  }, []);

  const setSearch = useCallback(
    (next: SearchParams) => {
      const copy: SearchParams = { ...next };
      delete copy.page;
      const { pageSize } = currentParams();
      if (copy.pageSize === undefined && pageSize !== undefined) {
        copy.pageSize = pageSize;
      }
      navigate(copy);
    },
    [navigate],
  );

  const setPage = useCallback(
    (page: number) => {
      if (!Number.isInteger(page) || page < 1) {
        return;
      }
      const copy = currentParams();
      if (page === 1) {
        delete copy.page;
      } else {
        copy.page = page;
      }
      navigate(copy);
    },
    [navigate],
  );

  const clear = useCallback(() => {
    const { pageSize } = currentParams();
    navigate(pageSize === undefined ? {} : { pageSize });
  }, [navigate]);

  return useMemo(
    () => ({ params, setSearch, setPage, clear }),
    [params, setSearch, setPage, clear],
  );
}
