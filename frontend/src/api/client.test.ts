import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  cities,
  internalErrorBody,
  jsonResponse,
  searchResponse,
  validationErrorBody,
} from '../test/fixtures';
import { buildSearchQuery, getCities, searchListings, searchListingsByQuery } from './client';
import { ApiRequestError } from './errors';
import type { SearchParams } from './types';

describe('buildSearchQuery', () => {
  it('gives an empty string for no params', () => {
    expect(buildSearchQuery({})).toBe('');
  });

  it('writes every param in SPEC section 5.3 order', () => {
    const params: SearchParams = {
      pageSize: 20,
      page: 2,
      status: ['sold', 'active'],
      targetBudget: 450000,
      keyword: 'pets allowed',
      city: 'Springfield',
      minBedrooms: 2,
      maxPrice: 600000,
      minPrice: 400000,
    };

    expect(buildSearchQuery(params)).toBe(
      'minPrice=400000&maxPrice=600000&minBedrooms=2&city=Springfield&keyword=pets+allowed' +
        '&targetBudget=450000&status=active&status=sold&page=2&pageSize=20',
    );
  });

  it('repeats status once per value, deduplicated, in enum order', () => {
    expect(buildSearchQuery({ status: ['sold', 'pending', 'sold', 'active'] })).toBe(
      'status=active&status=pending&status=sold',
    );
    expect(buildSearchQuery({ status: ['sold', 'active'] })).toBe(
      buildSearchQuery({ status: ['active', 'sold'] }),
    );
  });

  it.each<[string, SearchParams]>([
    ['an empty city', { city: '' }],
    ['a whitespace-only city', { city: '  ' }],
    ['an empty keyword', { keyword: ' ' }],
    ['an empty status list', { status: [] }],
    ['a NaN price', { minPrice: Number.NaN }],
    ['an infinite budget', { targetBudget: Number.POSITIVE_INFINITY }],
    ['a negative infinite page', { page: Number.NEGATIVE_INFINITY }],
  ])('omits %s', (_name, params) => {
    expect(buildSearchQuery(params)).toBe('');
  });

  it('writes numbers as plain strings and keeps out-of-range values for the API to judge', () => {
    expect(buildSearchQuery({ targetBudget: 450000.5, minPrice: -5, minBedrooms: 0 })).toBe(
      'minPrice=-5&minBedrooms=0&targetBudget=450000.5',
    );
  });

  it('trims city and keyword and encodes them', () => {
    expect(buildSearchQuery({ city: '  Fairfax ', keyword: ' pets allowed & 50% ' })).toBe(
      'city=Fairfax&keyword=pets+allowed+%26+50%25',
    );
  });
});

describe('client', () => {
  const fetchMock = vi.fn<typeof fetch>();

  beforeEach(() => {
    fetchMock.mockReset();
    vi.stubGlobal('fetch', fetchMock);
  });

  async function rejection(promise: Promise<unknown>): Promise<ApiRequestError> {
    const err = await promise.then(
      () => null,
      (reason: unknown) => reason,
    );
    if (!(err instanceof ApiRequestError)) {
      throw new Error(`expected an ApiRequestError, got ${String(err)}`);
    }
    return err;
  }

  describe('searchListings', () => {
    it('sends the built query with the signal and resolves with the body', async () => {
      fetchMock.mockResolvedValue(jsonResponse(searchResponse));
      const controller = new AbortController();

      const data = await searchListings(
        { targetBudget: 450000, city: 'Springfield' },
        controller.signal,
      );

      expect(data).toEqual(searchResponse);
      expect(fetchMock).toHaveBeenCalledTimes(1);
      expect(fetchMock.mock.calls[0]?.[0]).toBe(
        '/api/listings/search?city=Springfield&targetBudget=450000',
      );
      expect(fetchMock.mock.calls[0]?.[1]?.signal).toBe(controller.signal);
      expect(fetchMock.mock.calls[0]?.[1]?.headers).toEqual({ Accept: 'application/json' });
    });

    it('requests the bare path for no params and passes a null signal when none is given', async () => {
      fetchMock.mockResolvedValue(jsonResponse(searchResponse));

      await searchListings({});

      expect(fetchMock.mock.calls[0]?.[0]).toBe('/api/listings/search');
      expect(fetchMock.mock.calls[0]?.[1]?.signal).toBeNull();
    });

    it('searchListingsByQuery sends the query as given', async () => {
      fetchMock.mockResolvedValue(jsonResponse(searchResponse));

      await searchListingsByQuery('status=sold&page=2');

      expect(fetchMock.mock.calls[0]?.[0]).toBe('/api/listings/search?status=sold&page=2');
    });

    it('rejects a 400 with a validation error, its fields and the request ID', async () => {
      fetchMock.mockResolvedValue(
        jsonResponse(validationErrorBody, { status: 400, requestId: 'req-400' }),
      );

      const err = await rejection(searchListings({ minPrice: 5, maxPrice: 1 }));

      expect(err.apiError).toEqual({
        kind: 'validation',
        message: 'Invalid search parameters.',
        fields: validationErrorBody.error.fields,
      });
      expect(err.requestId).toBe('req-400');
    });

    it('rejects a 500 with an http error carrying the body message', async () => {
      fetchMock.mockResolvedValue(
        jsonResponse(internalErrorBody, { status: 500, requestId: 'req-500' }),
      );

      const err = await rejection(searchListings({}));

      expect(err.apiError).toEqual({
        kind: 'http',
        status: 500,
        message: 'Internal server error.',
      });
      expect(err.requestId).toBe('req-500');
    });

    it('rejects a 404 with an http error', async () => {
      fetchMock.mockResolvedValue(
        jsonResponse(
          { error: { code: 'not_found', message: 'Resource not found.' } },
          { status: 404 },
        ),
      );

      const err = await rejection(searchListings({}));

      expect(err.apiError).toEqual({ kind: 'http', status: 404, message: 'Resource not found.' });
    });

    it('rejects the dev proxy HTML 502 with the fallback message', async () => {
      fetchMock.mockResolvedValue(
        new Response('<html><body>Bad Gateway</body></html>', {
          status: 502,
          headers: { 'Content-Type': 'text/html' },
        }),
      );

      const err = await rejection(searchListings({}));

      expect(err.apiError).toEqual({
        kind: 'http',
        status: 502,
        message: 'Request failed with status 502.',
      });
      expect(err.requestId).toBeNull();
    });

    it.each([
      ['a non-JSON body', () => new Response('not json', { status: 200 })],
      ['a body without items', () => jsonResponse({ ...searchResponse, items: null })],
      ['a body without page', () => jsonResponse({ items: [], scoring: {} })],
      ['a JSON array', () => jsonResponse([searchResponse])],
    ])('rejects a 2xx with %s as malformed', async (_name, makeResponse) => {
      fetchMock.mockResolvedValue(makeResponse());

      const err = await rejection(searchListings({}));

      expect(err.apiError).toEqual({
        kind: 'http',
        status: 200,
        message: 'Malformed response body.',
      });
    });

    it('rejects a fetch TypeError with a network error and its message', async () => {
      const cause = new TypeError('Failed to fetch');
      fetchMock.mockRejectedValue(cause);

      const err = await rejection(searchListings({}));

      expect(err.apiError).toEqual({ kind: 'network', message: 'Failed to fetch' });
      expect(err.requestId).toBeNull();
      expect(err.cause).toBe(cause);
    });

    it('rejects an abort with a network error', async () => {
      fetchMock.mockRejectedValue(new DOMException('The operation was aborted.', 'AbortError'));

      const err = await rejection(searchListings({}));

      expect(err.apiError).toEqual({ kind: 'network', message: 'Request aborted.' });
    });

    it('treats a body read cut short by an abort as an abort', async () => {
      const controller = new AbortController();
      fetchMock.mockImplementation(() => {
        controller.abort();
        return Promise.resolve(new Response('{"items": [', { status: 200 }));
      });

      const err = await rejection(searchListings({}, controller.signal));

      expect(err.apiError).toEqual({ kind: 'network', message: 'Request aborted.' });
    });
  });

  describe('getCities', () => {
    it('fetches /api/cities and resolves with the names', async () => {
      fetchMock.mockResolvedValue(jsonResponse({ cities }));

      await expect(getCities()).resolves.toEqual(cities);
      expect(fetchMock.mock.calls[0]?.[0]).toBe('/api/cities');
    });

    it('drops non-string entries', async () => {
      fetchMock.mockResolvedValue(jsonResponse({ cities: ['Fairfax', null, 3, 'Vienna'] }));

      await expect(getCities()).resolves.toEqual(['Fairfax', 'Vienna']);
    });

    it('rejects a 500 with an http error', async () => {
      fetchMock.mockResolvedValue(jsonResponse(internalErrorBody, { status: 500 }));

      const err = await rejection(getCities());

      expect(err.apiError).toEqual({
        kind: 'http',
        status: 500,
        message: 'Internal server error.',
      });
    });

    it('rejects a body of the wrong shape as malformed', async () => {
      fetchMock.mockResolvedValue(jsonResponse(cities));

      const err = await rejection(getCities());

      expect(err.apiError).toEqual({
        kind: 'http',
        status: 200,
        message: 'Malformed response body.',
      });
    });
  });
});
