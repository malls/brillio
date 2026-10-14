import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { buildSearchQuery } from '../api/client';
import type { SearchParams } from '../api/types';
import { parseSearchParams, useSearchParamsState } from './useSearchParamsState';

const fullParams: SearchParams = {
  minPrice: 400000,
  maxPrice: 600000,
  minBedrooms: 2,
  city: 'Springfield',
  keyword: 'pets allowed',
  targetBudget: 450000.5,
  status: ['active', 'sold'],
  page: 2,
  pageSize: 20,
};

describe('parseSearchParams', () => {
  it.each<[string, SearchParams]>([
    ['minPrice=400000', { minPrice: 400000 }],
    ['maxPrice=600000.5', { maxPrice: 600000.5 }],
    ['minBedrooms=2', { minBedrooms: 2 }],
    ['city=Springfield', { city: 'Springfield' }],
    ['keyword=pets+allowed', { keyword: 'pets allowed' }],
    ['targetBudget=450000', { targetBudget: 450000 }],
    ['status=pending', { status: ['pending'] }],
    ['status=active&status=sold', { status: ['active', 'sold'] }],
    ['page=3', { page: 3 }],
    ['pageSize=100', { pageSize: 100 }],
  ])('reads %s', (search, expected) => {
    expect(parseSearchParams(`?${search}`)).toEqual(expected);
  });

  it('round-trips the full set through buildSearchQuery', () => {
    expect(parseSearchParams(buildSearchQuery(fullParams))).toEqual(fullParams);
  });

  it('reads a query with or without the leading "?"', () => {
    expect(parseSearchParams('city=Vienna')).toEqual({ city: 'Vienna' });
    expect(parseSearchParams('')).toEqual({});
  });

  it.each([
    'minPrice=abc',
    'minPrice=12abc',
    'minPrice=0x10',
    'minPrice=Infinity',
    'maxPrice=NaN',
    'maxPrice=%20',
    'minBedrooms=2.5',
    'targetBudget=1e400',
    'page=0',
    'page=-1',
    'page=1.5',
    'pageSize=0',
    'pageSize=101',
    'status=bogus',
    'status=ACTIVE',
    'city=',
    'city=%20',
    'keyword=',
    'minprice=500000',
  ])('drops %s', (search) => {
    expect(parseSearchParams(`?${search}`)).toEqual({});
  });

  it.each<[string, SearchParams]>([
    ['minPrice=-5', { minPrice: -5 }],
    ['minBedrooms=-1', { minBedrooms: -1 }],
    ['minPrice=500000&maxPrice=1', { minPrice: 500000, maxPrice: 1 }],
    ['targetBudget=0', { targetBudget: 0 }],
    ['minBedrooms=2.0', { minBedrooms: 2 }],
    ['minPrice=1e5', { minPrice: 100000 }],
  ])('keeps %s for the API to judge', (search, expected) => {
    expect(parseSearchParams(`?${search}`)).toEqual(expected);
  });

  it('keeps over-long text for the API to judge', () => {
    const city = 'x'.repeat(101);

    expect(parseSearchParams(`?city=${city}`)).toEqual({ city });
  });

  it.each<[string, SearchParams]>([
    ['status=active,sold', { status: ['active', 'sold'] }],
    ['status=sold,active,sold', { status: ['active', 'sold'] }],
    ['status=sold&status=pending,active', { status: ['active', 'pending', 'sold'] }],
    ['status=active,bogus', { status: ['active'] }],
    ['status=,', {}],
  ])('reads status %s', (search, expected) => {
    expect(parseSearchParams(`?${search}`)).toEqual(expected);
  });

  it('trims city and keyword', () => {
    expect(parseSearchParams('?city=%20Fairfax%20&keyword=%20no%20pets%20')).toEqual({
      city: 'Fairfax',
      keyword: 'no pets',
    });
  });

  it('takes the first value of a repeated scalar', () => {
    expect(parseSearchParams('?city=Vienna&city=Reston&page=2&page=3')).toEqual({
      city: 'Vienna',
      page: 2,
    });
  });

  it('ignores unknown params and keeps the known ones', () => {
    expect(parseSearchParams('?minprice=5&city=Reston&utm_source=mail')).toEqual({
      city: 'Reston',
    });
  });
});

describe('useSearchParamsState', () => {
  beforeEach(() => {
    window.history.replaceState(null, '', '/');
  });

  function render(url = '/') {
    window.history.replaceState(null, '', url);
    return renderHook(() => useSearchParamsState());
  }

  it('reads the URL on mount', () => {
    const { result } = render(`/?${buildSearchQuery(fullParams)}`);

    expect(result.current.params).toEqual(fullParams);
  });

  it('reads an empty URL as no params', () => {
    const { result } = render();

    expect(result.current.params).toEqual({});
  });

  it('setSearch pushes the new filters without page', () => {
    const { result } = render('/?city=Vienna&page=3');
    const before = window.history.length;

    act(() => {
      result.current.setSearch({ city: 'Reston', targetBudget: 450000, page: 4 });
    });

    expect(window.location.search).toBe('?city=Reston&targetBudget=450000');
    expect(window.history.length).toBe(before + 1);
    expect(result.current.params).toEqual({ city: 'Reston', targetBudget: 450000 });
  });

  it('keeps pageSize through setSearch, setPage and clear', () => {
    const { result } = render('/?city=Vienna&pageSize=5');

    act(() => {
      result.current.setSearch({ city: 'Reston' });
    });
    expect(window.location.search).toBe('?city=Reston&pageSize=5');

    act(() => {
      result.current.setPage(2);
    });
    expect(window.location.search).toBe('?city=Reston&page=2&pageSize=5');

    act(() => {
      result.current.clear();
    });
    expect(window.location.search).toBe('?pageSize=5');
    expect(result.current.params).toEqual({ pageSize: 5 });
  });

  it('setSearch writes a pageSize it is given', () => {
    const { result } = render('/?pageSize=5');

    act(() => {
      result.current.setSearch({ pageSize: 20 });
    });

    expect(window.location.search).toBe('?pageSize=20');
  });

  it('setPage changes only page', () => {
    const { result } = render('/?city=Vienna&status=sold');

    act(() => {
      result.current.setPage(2);
    });

    expect(window.location.search).toBe('?city=Vienna&status=sold&page=2');
    expect(result.current.params).toEqual({ city: 'Vienna', status: ['sold'], page: 2 });
  });

  it('setPage(1) removes page', () => {
    const { result } = render('/?city=Vienna&page=3');

    act(() => {
      result.current.setPage(1);
    });

    expect(window.location.search).toBe('?city=Vienna');
    expect(result.current.params).toEqual({ city: 'Vienna' });
  });

  it.each([0, -1, 1.5, Number.NaN])('setPage(%d) does nothing', (page) => {
    const { result } = render('/?city=Vienna&page=3');
    const before = window.history.length;

    act(() => {
      result.current.setPage(page);
    });

    expect(window.location.search).toBe('?city=Vienna&page=3');
    expect(window.history.length).toBe(before);
  });

  it('clear writes an empty query', () => {
    const { result } = render('/?city=Vienna&page=3');

    act(() => {
      result.current.clear();
    });

    expect(window.location.search).toBe('');
    expect(window.location.pathname).toBe('/');
    expect(result.current.params).toEqual({});
  });

  it('follows popstate', () => {
    const { result } = render('/?city=Vienna');
    window.history.pushState(null, '', '/?city=Reston&page=2');

    act(() => {
      window.dispatchEvent(new PopStateEvent('popstate'));
    });

    expect(result.current.params).toEqual({ city: 'Reston', page: 2 });
  });

  it('stops following popstate after unmount', () => {
    const { result, unmount } = render('/?city=Vienna');
    unmount();
    window.history.pushState(null, '', '/?city=Reston');

    window.dispatchEvent(new PopStateEvent('popstate'));

    expect(result.current.params).toEqual({ city: 'Vienna' });
  });

  it('pushes nothing for a search equal to the current one', () => {
    const { result } = render('/?city=Vienna&status=active&status=sold');
    const before = window.history.length;
    const params = result.current.params;

    act(() => {
      result.current.setSearch({ status: ['sold', 'active'], city: ' Vienna ' });
    });

    expect(window.history.length).toBe(before);
    expect(result.current.params).toBe(params);
  });

  it('keeps params and writers stable across a rerender', () => {
    const { result, rerender } = render('/?city=Vienna');
    const first = result.current;

    rerender();

    expect(result.current).toBe(first);
    expect(result.current.params).toBe(first.params);
  });
});
