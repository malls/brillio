import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { SearchParams, SearchResponse } from '../api/types';
import {
  deferred,
  emptySearchResponse,
  internalErrorBody,
  jsonResponse,
  searchResponse,
  searchResponsePage2,
  validationErrorBody,
} from '../test/fixtures';
import { useListingSearch } from './useListingSearch';

const fetchMock = vi.fn<typeof fetch>();
const consoleError = vi.fn<(...args: unknown[]) => void>();

beforeEach(() => {
  fetchMock.mockReset();
  consoleError.mockReset();
  vi.stubGlobal('fetch', fetchMock);
  vi.spyOn(console, 'error').mockImplementation(consoleError);
});

function respond(body: unknown, init?: { status?: number; requestId?: string }) {
  fetchMock.mockImplementationOnce(() => Promise.resolve(jsonResponse(body, init)));
}

// A fetch that honors its signal: it rejects with an AbortError once aborted, else responds on
// the next macrotask.
function respondAbortably(body: unknown) {
  fetchMock.mockImplementation(
    (_input, init) =>
      new Promise<Response>((resolve, reject) => {
        const signal = init?.signal;
        const abort = () => {
          reject(new DOMException('The operation was aborted.', 'AbortError'));
        };
        if (signal?.aborted === true) {
          abort();
          return;
        }
        signal?.addEventListener('abort', abort);
        setTimeout(() => {
          resolve(jsonResponse(body));
        }, 0);
      }),
  );
}

function signalOf(call: number): AbortSignal | null | undefined {
  return fetchMock.mock.calls[call]?.[1]?.signal;
}

function urlOf(call: number): unknown {
  return fetchMock.mock.calls[call]?.[0];
}

// Lets pending promise callbacks and timers run, inside act.
async function flush() {
  await act(() => new Promise<void>((resolve) => setTimeout(resolve, 20)));
}

function renderSearch(params: SearchParams, options: { reactStrictMode?: boolean } = {}) {
  return renderHook(({ p }: { p: SearchParams }) => useListingSearch(p), {
    initialProps: { p: params },
    ...options,
  });
}

async function renderSettled(params: SearchParams, body: SearchResponse = searchResponse) {
  respond(body);
  const view = renderSearch(params);
  await waitFor(() => {
    expect(view.result.current.state).toEqual({ status: 'success', data: body });
  });
  return view;
}

describe('useListingSearch', () => {
  it('starts loading with no previous, then succeeds', async () => {
    respond(searchResponse);

    const { result } = renderSearch({ targetBudget: 450000 });

    expect(result.current.state).toEqual({ status: 'loading' });
    await waitFor(() => {
      expect(result.current.state).toEqual({ status: 'success', data: searchResponse });
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(urlOf(0)).toBe('/api/listings/search?targetBudget=450000');
  });

  it('starts loading, then reports empty when there are no results', async () => {
    respond(emptySearchResponse);

    const { result } = renderSearch({ city: 'Arlington' });

    expect(result.current.state).toEqual({ status: 'loading' });
    await waitFor(() => {
      expect(result.current.state).toEqual({ status: 'empty', data: emptySearchResponse });
    });
  });

  it('keeps the previous response while new params load', async () => {
    const { result, rerender } = await renderSettled({ targetBudget: 450000 });
    const next = deferred<Response>();
    fetchMock.mockReturnValueOnce(next.promise);

    rerender({ p: { city: 'Arlington' } });

    expect(result.current.state).toEqual({ status: 'loading', previous: searchResponse });
    expect(urlOf(1)).toBe('/api/listings/search?city=Arlington');
    act(() => {
      next.resolve(jsonResponse(emptySearchResponse));
    });
    await waitFor(() => {
      expect(result.current.state).toEqual({ status: 'empty', data: emptySearchResponse });
    });
  });

  it('aborts a request when params change and ignores its late response', async () => {
    const first = deferred<Response>();
    const second = deferred<Response>();
    fetchMock.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);

    const { result, rerender } = renderSearch({ city: 'Vienna' });
    rerender({ p: { city: 'Reston' } });

    expect(signalOf(0)?.aborted).toBe(true);
    expect(signalOf(1)?.aborted).toBe(false);
    act(() => {
      first.resolve(jsonResponse(emptySearchResponse));
    });
    await flush();
    expect(result.current.state).toEqual({ status: 'loading' });

    act(() => {
      second.resolve(jsonResponse(searchResponse));
    });
    await waitFor(() => {
      expect(result.current.state).toEqual({ status: 'success', data: searchResponse });
    });
    await flush();
    expect(result.current.state).toEqual({ status: 'success', data: searchResponse });
    expect(consoleError).not.toHaveBeenCalled();
  });

  it('does not refetch for equal params in a new object', async () => {
    const { result, rerender } = await renderSettled({
      status: ['sold', 'active'],
      city: 'Vienna',
    });
    const state = result.current.state;

    rerender({ p: { city: ' Vienna', status: ['active', 'sold'] } });
    await flush();

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(result.current.state).toBe(state);
  });

  it('reports a 400 as invalid with its fields and the previous response, unlogged', async () => {
    const { result, rerender } = await renderSettled({});
    respond(validationErrorBody, { status: 400, requestId: 'req-400' });

    rerender({ p: { minPrice: 500000, maxPrice: 1 } });

    await waitFor(() => {
      expect(result.current.state).toEqual({
        status: 'invalid',
        error: {
          kind: 'validation',
          message: 'Invalid search parameters.',
          fields: validationErrorBody.error.fields,
        },
        previous: searchResponse,
      });
    });
    expect(consoleError).not.toHaveBeenCalled();
  });

  it('reports a 500 as an http error with the previous response and logs it once', async () => {
    const { result, rerender } = await renderSettled({});
    respond(internalErrorBody, { status: 500, requestId: 'req-500' });

    rerender({ p: { city: 'Vienna' } });

    await waitFor(() => {
      expect(result.current.state).toEqual({
        status: 'error',
        error: { kind: 'http', status: 500, message: 'Internal server error.' },
        previous: searchResponse,
      });
    });
    expect(consoleError).toHaveBeenCalledTimes(1);
    expect(consoleError).toHaveBeenCalledWith('Listing search failed', {
      kind: 'http',
      status: 500,
      message: 'Internal server error.',
      requestId: 'req-500',
    });
  });

  it('reports a fetch failure as a network error and logs it', async () => {
    fetchMock.mockRejectedValueOnce(new TypeError('Failed to fetch'));

    const { result } = renderSearch({});

    await waitFor(() => {
      expect(result.current.state).toEqual({
        status: 'error',
        error: { kind: 'network', message: 'Failed to fetch' },
      });
    });
    expect(consoleError).toHaveBeenCalledTimes(1);
    expect(consoleError).toHaveBeenCalledWith('Listing search failed', {
      kind: 'network',
      message: 'Failed to fetch',
      requestId: null,
    });
  });

  it('retry refetches the same query, keeping the previous response while loading', async () => {
    const { result, rerender } = await renderSettled({});
    fetchMock.mockRejectedValueOnce(new TypeError('Failed to fetch'));
    rerender({ p: { city: 'Vienna', page: 2 } });
    await waitFor(() => {
      expect(result.current.state.status).toBe('error');
    });
    const next = deferred<Response>();
    fetchMock.mockReturnValueOnce(next.promise);

    act(() => {
      result.current.retry();
    });

    expect(result.current.state).toEqual({ status: 'loading', previous: searchResponse });
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(urlOf(2)).toBe('/api/listings/search?city=Vienna&page=2');
    expect(urlOf(2)).toBe(urlOf(1));
    act(() => {
      next.resolve(jsonResponse(searchResponse));
    });
    await waitFor(() => {
      expect(result.current.state).toEqual({ status: 'success', data: searchResponse });
    });
  });

  it('keeps retry stable across renders', async () => {
    const { result, rerender } = await renderSettled({});
    const { retry } = result.current;
    respond(emptySearchResponse);

    rerender({ p: { city: 'Arlington' } });

    expect(result.current.retry).toBe(retry);
    await waitFor(() => {
      expect(result.current.state.status).toBe('empty');
    });
  });

  it('reports a page past the end as success with no items', async () => {
    const pastTheEnd: SearchResponse = {
      ...searchResponse,
      items: [],
      page: { ...searchResponse.page, page: 3 },
    };
    respond(pastTheEnd);

    const { result } = renderSearch({ targetBudget: 450000, page: 3 });

    await waitFor(() => {
      expect(result.current.state).toEqual({ status: 'success', data: pastTheEnd });
    });
  });

  it('returning to a held query shows its result at once and refetches without loading', async () => {
    const { result, rerender } = await renderSettled({ city: 'Vienna' });
    const inFlight = deferred<Response>();
    fetchMock.mockReturnValueOnce(inFlight.promise);
    rerender({ p: { city: 'Reston' } });
    expect(result.current.state).toEqual({ status: 'loading', previous: searchResponse });
    const refreshed: SearchResponse = {
      ...searchResponse,
      scoring: { ...searchResponse.scoring, asOf: '2026-10-14' },
    };
    respond(refreshed);

    rerender({ p: { city: 'Vienna' } });

    expect(signalOf(1)?.aborted).toBe(true);
    expect(result.current.state).toEqual({ status: 'success', data: searchResponse });
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(urlOf(2)).toBe('/api/listings/search?city=Vienna');
    await waitFor(() => {
      expect(result.current.state).toEqual({ status: 'success', data: refreshed });
    });
  });

  it('marks a page-only change with the requested page and keeps the previous response', async () => {
    const { result, rerender } = await renderSettled({ targetBudget: 450000 });
    const page2 = deferred<Response>();
    fetchMock.mockReturnValueOnce(page2.promise);

    rerender({ p: { targetBudget: 450000, page: 2 } });

    expect(result.current.state).toStrictEqual({
      status: 'loading',
      previous: searchResponse,
      requestedPage: 2,
    });
    act(() => {
      page2.resolve(jsonResponse(searchResponsePage2));
    });
    await waitFor(() => {
      expect(result.current.state).toEqual({ status: 'success', data: searchResponsePage2 });
    });
    const page1 = deferred<Response>();
    fetchMock.mockReturnValueOnce(page1.promise);

    rerender({ p: { targetBudget: 450000 } });

    expect(result.current.state).toStrictEqual({
      status: 'loading',
      previous: searchResponsePage2,
      requestedPage: 1,
    });
    act(() => {
      page1.resolve(jsonResponse(searchResponse));
    });
    await waitFor(() => {
      expect(result.current.state).toEqual({ status: 'success', data: searchResponse });
    });
  });

  it('does not mark a change of filters or pageSize as a page change', async () => {
    const { result, rerender } = await renderSettled({ targetBudget: 450000 });
    const filtered = deferred<Response>();
    const resized = deferred<Response>();
    fetchMock.mockReturnValueOnce(filtered.promise).mockReturnValueOnce(resized.promise);

    rerender({ p: { targetBudget: 450000, city: 'Vienna', page: 2 } });

    expect(result.current.state).toStrictEqual({ status: 'loading', previous: searchResponse });

    rerender({ p: { targetBudget: 450000, pageSize: 5 } });

    expect(result.current.state).toStrictEqual({ status: 'loading', previous: searchResponse });
    expect(urlOf(2)).toBe('/api/listings/search?targetBudget=450000&pageSize=5');
    act(() => {
      resized.resolve(jsonResponse(searchResponse));
    });
    await waitFor(() => {
      expect(result.current.state).toEqual({ status: 'success', data: searchResponse });
    });
  });

  it('compares a page change with the response on screen, not with the params before', async () => {
    // A search is in flight over A's list when only the page changes: the request still
    // differs from A's by its filters, so it is no page change.
    const { result, rerender } = await renderSettled({ targetBudget: 450000 });
    const search = deferred<Response>();
    const page2 = deferred<Response>();
    fetchMock.mockReturnValueOnce(search.promise).mockReturnValueOnce(page2.promise);

    rerender({ p: { city: 'Vienna' } });

    expect(result.current.state).toStrictEqual({ status: 'loading', previous: searchResponse });

    rerender({ p: { city: 'Vienna', page: 2 } });

    expect(signalOf(1)?.aborted).toBe(true);
    expect(result.current.state).toStrictEqual({ status: 'loading', previous: searchResponse });
    act(() => {
      page2.resolve(jsonResponse(searchResponsePage2));
    });
    await waitFor(() => {
      expect(result.current.state).toEqual({ status: 'success', data: searchResponsePage2 });
    });
  });

  it('does not mark a retry as a page change', async () => {
    const { result, rerender } = await renderSettled({ targetBudget: 450000 });
    fetchMock.mockRejectedValueOnce(new TypeError('Failed to fetch'));
    rerender({ p: { targetBudget: 450000, page: 2 } });
    await waitFor(() => {
      expect(result.current.state.status).toBe('error');
    });
    const next = deferred<Response>();
    fetchMock.mockReturnValueOnce(next.promise);

    act(() => {
      result.current.retry();
    });

    expect(result.current.state).toStrictEqual({ status: 'loading', previous: searchResponse });
    act(() => {
      next.resolve(jsonResponse(searchResponsePage2));
    });
    await waitFor(() => {
      expect(result.current.state).toEqual({ status: 'success', data: searchResponsePage2 });
    });
  });

  it('aborts the request on unmount without updating or logging', async () => {
    respondAbortably(searchResponse);
    const { result, unmount } = renderSearch({ city: 'Vienna' });

    unmount();
    await flush();

    expect(signalOf(0)?.aborted).toBe(true);
    expect(result.current.state).toEqual({ status: 'loading' });
    expect(consoleError).not.toHaveBeenCalled();
  });

  it('under StrictMode aborts the first request unlogged and settles on the second', async () => {
    respondAbortably(searchResponse);

    const { result } = renderSearch({ city: 'Vienna' }, { reactStrictMode: true });

    await waitFor(() => {
      expect(result.current.state).toEqual({ status: 'success', data: searchResponse });
    });
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(signalOf(0)?.aborted).toBe(true);
    expect(signalOf(1)?.aborted).toBe(false);
    expect(consoleError).not.toHaveBeenCalled();
  });
});
