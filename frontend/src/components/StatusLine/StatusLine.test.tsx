import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { ApiError, SearchResponse, ValidationError } from '../../api/types';
import type { SearchState } from '../../hooks/useListingSearch';
import {
  emptySearchResponse,
  listingA7,
  searchResponse,
  validationErrorBody,
} from '../../test/fixtures';
import { StatusLine } from './StatusLine';

// The real search with no params: the same 12 results and 2 pages, ordered newest first
// (SPEC section 5.4's no-budget scoring).
const noBudgetResponse: SearchResponse = {
  ...searchResponse,
  scoring: { ...searchResponse.scoring, targetBudget: null, budgetWeight: 0, recencyWeight: 1 },
};

// The real search `targetBudget=450000&city=Chantilly`: A7 is the only match.
const singleResponse: SearchResponse = {
  ...searchResponse,
  items: [{ ...listingA7, rank: 1 }],
  page: { ...searchResponse.page, totalItems: 1, totalPages: 1 },
  appliedFilters: { ...searchResponse.appliedFilters, city: 'Chantilly' },
};

// The real search `targetBudget=450000&page=3`: a page past the end (SPEC section 5.3 answers
// 200 with no items and the true metadata).
const pastEndResponse: SearchResponse = {
  ...searchResponse,
  items: [],
  page: { ...searchResponse.page, page: 3 },
};

// SPEC section 5.5's example 400, as the client maps it.
const invalid: ValidationError = {
  kind: 'validation',
  message: validationErrorBody.error.message,
  fields: validationErrorBody.error.fields ?? [],
};

const failure: ApiError = { kind: 'network', message: 'Request failed.' };

// The text of the status line's left and right halves.
function halves(state: SearchState): [string | null, string | null] {
  render(<StatusLine state={state} />);
  const [left, right] = Array.from(screen.getByRole('status').children);
  return [left?.textContent ?? null, right?.textContent ?? null];
}

describe('StatusLine', () => {
  it('is empty when idle', () => {
    expect(halves({ status: 'idle' })).toEqual(['', '']);
  });

  it('reads Loading alone on a first load', () => {
    expect(halves({ status: 'loading' })).toEqual(['Loading', '']);
  });

  it('reads Loading and keeps the previous page while a search loads over a list', () => {
    expect(halves({ status: 'loading', previous: searchResponse })).toEqual([
      'Loading',
      'Page 1 of 2',
    ]);
  });

  it('keeps the count and order and shows the requested page during a page change', () => {
    expect(halves({ status: 'loading', previous: searchResponse, requestedPage: 2 })).toEqual([
      '12 results · ranked for $450,000',
      'Page 2 of 2',
    ]);
  });

  it('keeps newest first during a page change without a target budget', () => {
    expect(halves({ status: 'loading', previous: noBudgetResponse, requestedPage: 2 })).toEqual([
      '12 results · newest first',
      'Page 2 of 2',
    ]);
  });

  it('shows a requested page past the end', () => {
    expect(halves({ status: 'loading', previous: searchResponse, requestedPage: 3 })).toEqual([
      '12 results · ranked for $450,000',
      'Page 3 of 2',
    ]);
  });

  it('keeps 0 results during a page change over an empty list', () => {
    expect(halves({ status: 'loading', previous: emptySearchResponse, requestedPage: 2 })).toEqual([
      '0 results',
      '',
    ]);
  });

  it('summarizes a ranked result', () => {
    expect(halves({ status: 'success', data: searchResponse })).toEqual([
      '12 results · ranked for $450,000',
      'Page 1 of 2',
    ]);
  });

  it('says newest first without a target budget', () => {
    expect(halves({ status: 'success', data: noBudgetResponse })).toEqual([
      '12 results · newest first',
      'Page 1 of 2',
    ]);
  });

  it('uses the singular for one result', () => {
    expect(halves({ status: 'success', data: singleResponse })).toEqual([
      '1 result · ranked for $450,000',
      'Page 1 of 1',
    ]);
  });

  it('reports the true page on a page past the end', () => {
    expect(halves({ status: 'success', data: pastEndResponse })).toEqual([
      '12 results · ranked for $450,000',
      'Page 3 of 2',
    ]);
  });

  it('reads 0 results with no page when empty', () => {
    expect(halves({ status: 'empty', data: emptySearchResponse })).toEqual(['0 results', '']);
  });

  it('keeps describing the previous list on invalid input', () => {
    expect(halves({ status: 'invalid', error: invalid, previous: searchResponse })).toEqual([
      '12 results · ranked for $450,000',
      'Page 1 of 2',
    ]);
  });

  it('keeps describing a previous empty list on invalid input', () => {
    expect(halves({ status: 'invalid', error: invalid, previous: emptySearchResponse })).toEqual([
      '0 results',
      '',
    ]);
  });

  it('is empty on invalid input with no list on screen', () => {
    expect(halves({ status: 'invalid', error: invalid })).toEqual(['', '']);
  });

  it('is empty on error over a previous list', () => {
    expect(halves({ status: 'error', error: failure, previous: searchResponse })).toEqual(['', '']);
  });

  it('is empty on error with no previous list', () => {
    expect(halves({ status: 'error', error: failure })).toEqual(['', '']);
  });

  it('renders one status region with exactly two halves', () => {
    render(<StatusLine state={{ status: 'success', data: searchResponse }} />);

    const regions = screen.getAllByRole('status');
    expect(regions).toHaveLength(1);
    expect(regions[0]?.children).toHaveLength(2);
    expect(regions[0]?.textContent).toBe('12 results · ranked for $450,000Page 1 of 2');
  });
});
