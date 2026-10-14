import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  emptySearchResponse,
  listingA1,
  listingA3,
  listingA7,
  searchResponse,
} from '../test/fixtures';
import {
  META_SEPARATOR,
  formatBathrooms,
  formatBedrooms,
  formatListingMeta,
  formatLocation,
  formatMonogram,
  formatOrdering,
  formatPageOf,
  formatRank,
  formatResultCount,
  formatResultSummary,
  formatScore,
  formatSource,
  formatSqft,
  formatStatusLabel,
} from './listing';

afterEach(() => {
  vi.useRealTimers();
});

// BRANDING.md section 7.1, row by row, plus the strings of sections 5.3-5.7.
describe('BRANDING section 7.1 formatters', () => {
  it('Bedrooms: integer + " bd"', () => {
    expect(formatBedrooms(2)).toBe('2 bd');
    expect(formatBedrooms(0)).toBe('0 bd');
  });

  it.each([
    [1, '1 ba'],
    [1.5, '1.5 ba'],
    [2, '2 ba'],
    [2.5, '2.5 ba'],
  ])('Bathrooms: %d as %s', (n, expected) => {
    expect(formatBathrooms(n)).toBe(expected);
  });

  it.each([
    [1450, '1,450 sqft'],
    [980, '980 sqft'],
  ])('Area: %d as %s', (n, expected) => {
    expect(formatSqft(n)).toBe(expected);
  });

  it('Source: underscore to space', () => {
    expect(formatSource('MLS_A')).toBe('MLS A');
    expect(formatSource('MLS_B')).toBe('MLS B');
  });

  it('Status: a label only when not active', () => {
    expect(formatStatusLabel('pending')).toBe('PENDING');
    expect(formatStatusLabel('sold')).toBe('SOLD');
    expect(formatStatusLabel('active')).toBeNull();
  });

  it.each([
    [0.94, '0.94'],
    [0.7414, '0.74'],
    [1, '1.00'],
    [0, '0.00'],
  ])('Score: %d as %s', (score, expected) => {
    expect(formatScore(score)).toBe(expected);
  });

  it('Rank: "No. " + integer', () => {
    expect(formatRank(1)).toBe('No. 1');
    expect(formatRank(11)).toBe('No. 11');
  });

  it.each([
    [1, '1 result'],
    [12, '12 results'],
    [0, '0 results'],
    [1200, '1,200 results'],
  ])('Counts: %d as %s', (n, expected) => {
    expect(formatResultCount(n)).toBe(expected);
  });

  it('Pages: "Page N of M", empty when there are no pages', () => {
    expect(formatPageOf(1, 2)).toBe('Page 1 of 2');
    expect(formatPageOf(1, 0)).toBe('');
  });
});

describe('listing card', () => {
  it('formats the location as City, ST ZIP', () => {
    expect(formatLocation(listingA1)).toBe('Springfield, VA 22150');
  });

  it('gives the meta items in BRANDING section 5.4 order', () => {
    expect(formatListingMeta(listingA1, 2026)).toEqual([
      'Springfield, VA 22150',
      '2 bd',
      '1.5 ba',
      '980 sqft',
      'Listed Aug 29',
      'MLS A',
    ]);
    expect(formatListingMeta(listingA3, 2026).join(META_SEPARATOR)).toBe(
      'Fairfax, VA 22030 · 2 bd · 1 ba · 850 sqft · Listed Sep 1 · MLS A',
    );
  });

  it('leaves the status label out of the meta items', () => {
    expect(formatListingMeta(listingA7, 2026)).toEqual([
      'Chantilly, VA 20151',
      '4 bd',
      '2.5 ba',
      '1,950 sqft',
      'Listed Jul 28',
      'MLS A',
    ]);
  });

  it('dates the meta items against the local clock by default', () => {
    vi.useFakeTimers({ toFake: ['Date'], now: new Date(2027, 0, 5) });

    expect(formatListingMeta(listingA1)[4]).toBe('Listed Aug 29, 2026');
  });

  it.each([
    ['Springfield', 'Sp'],
    ['Vienna', 'Vi'],
    ['Fairfax', 'Fa'],
    ['RESTON', 'Re'],
    ['  Manassas ', 'Ma'],
    ['chantilly', 'Ch'],
    ['', ''],
  ])('gives the monogram of %j as %j', (city, expected) => {
    expect(formatMonogram(city)).toBe(expected);
  });
});

describe('status line', () => {
  it('names the ordering', () => {
    expect(formatOrdering(450000)).toBe('ranked for $450,000');
    expect(formatOrdering(null)).toBe('newest first');
  });

  it('summarizes a ranked response', () => {
    expect(formatResultSummary(searchResponse)).toBe('12 results · ranked for $450,000');
  });

  it('summarizes a response without a budget', () => {
    const noBudget = {
      ...searchResponse,
      scoring: { ...searchResponse.scoring, targetBudget: null, budgetWeight: 0, recencyWeight: 1 },
    };

    expect(formatResultSummary(noBudget)).toBe('12 results · newest first');
  });

  it('summarizes an empty response as 0 results', () => {
    expect(formatResultSummary(emptySearchResponse)).toBe('0 results');
  });

  it('gives the page position of the fixtures', () => {
    expect(formatPageOf(searchResponse.page.page, searchResponse.page.totalPages)).toBe(
      'Page 1 of 2',
    );
    expect(formatPageOf(emptySearchResponse.page.page, emptySearchResponse.page.totalPages)).toBe(
      '',
    );
  });
});
