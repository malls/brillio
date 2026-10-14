import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiRequestError } from '../api/errors';
import { cities, internalErrorBody, jsonResponse } from '../test/fixtures';
import { useCities } from './useCities';

const fetchMock = vi.fn<typeof fetch>();
const consoleError = vi.fn<(...args: unknown[]) => void>();

beforeEach(() => {
  fetchMock.mockReset();
  consoleError.mockReset();
  vi.stubGlobal('fetch', fetchMock);
  vi.spyOn(console, 'error').mockImplementation(consoleError);
});

describe('useCities', () => {
  it('starts empty and loads the cities once from /api/cities', async () => {
    fetchMock.mockImplementation(() => Promise.resolve(jsonResponse({ cities })));

    const { result, rerender } = renderHook(() => useCities());

    expect(result.current).toEqual([]);
    await waitFor(() => {
      expect(result.current).toEqual(cities);
    });
    rerender();
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock.mock.calls[0]?.[0]).toBe('/api/cities');
    expect(consoleError).not.toHaveBeenCalled();
  });

  it('stays empty and logs once when loading fails', async () => {
    fetchMock.mockImplementation(() =>
      Promise.resolve(jsonResponse(internalErrorBody, { status: 500 })),
    );

    const { result } = renderHook(() => useCities());

    await waitFor(() => {
      expect(consoleError).toHaveBeenCalledTimes(1);
    });
    expect(consoleError.mock.calls[0]?.[0]).toBe('Failed to load cities');
    expect(consoleError.mock.calls[0]?.[1]).toBeInstanceOf(ApiRequestError);
    expect(result.current).toEqual([]);
  });

  it('under StrictMode does not log the aborted first request', async () => {
    fetchMock.mockImplementation(
      (_input, init) =>
        new Promise<Response>((resolve, reject) => {
          init?.signal?.addEventListener('abort', () => {
            reject(new DOMException('The operation was aborted.', 'AbortError'));
          });
          setTimeout(() => {
            resolve(jsonResponse({ cities }));
          }, 0);
        }),
    );

    const { result } = renderHook(() => useCities(), { reactStrictMode: true });

    await waitFor(() => {
      expect(result.current).toEqual(cities);
    });
    await act(() => new Promise<void>((resolve) => setTimeout(resolve, 20)));
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(fetchMock.mock.calls[0]?.[1]?.signal?.aborted).toBe(true);
    expect(consoleError).not.toHaveBeenCalled();
  });
});
