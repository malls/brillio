import { describe, expect, it } from 'vitest';
import type { FieldError } from '../../api/types';
import type { Drafts, FieldName, Parsed } from './filterBarModel';
import {
  EMPTY_DRAFTS,
  cityChoice,
  displayValue,
  draftsFrom,
  filterKey,
  mapServerErrors,
  normalizeDraft,
  parseNumber,
  resolveMarks,
  toSearchParams,
  validate,
} from './filterBarModel';

function drafts(partial: Partial<Drafts>): Drafts {
  return { ...EMPTY_DRAFTS, ...partial };
}

const EMPTY: Parsed = { kind: 'empty' };
const INVALID: Parsed = { kind: 'invalid' };
const number = (value: number): Parsed => ({ kind: 'number', value });

describe('parseNumber', () => {
  it.each<[string, Parsed]>([
    ['400000', number(400000)],
    ['$400,000', number(400000)],
    [' 400000 ', number(400000)],
    ['5.', number(5)],
    ['.5', number(0.5)],
    ['2.0', number(2)],
    ['-0', number(0)],
    ['-5', number(-5)],
    ['', EMPTY],
    ['   ', EMPTY],
    ['abc', INVALID],
    ['1e5', INVALID],
    ['+5', INVALID],
    ['4 00', INVALID],
    ['$', INVALID],
    ['-', INVALID],
    ['.', INVALID],
    ['9'.repeat(400), INVALID],
  ])('reads %j as %j', (text, expected) => {
    const parsed = parseNumber(text);
    expect(parsed).toEqual(expected);
    if (parsed.kind === 'number') {
      expect(Object.is(parsed.value, -0)).toBe(false);
    }
  });
});

describe('validate', () => {
  it('finds nothing in empty drafts', () => {
    expect(validate(EMPTY_DRAFTS)).toEqual({ messages: {}, conflict: false });
  });

  it.each<[Partial<Drafts>, Partial<Record<FieldName, string>>]>([
    [{ minPrice: 'abc' }, { minPrice: 'Min price must be a number.' }],
    [{ minPrice: '-5' }, { minPrice: 'Min price must be 0 or more.' }],
    [{ maxPrice: 'abc' }, { maxPrice: 'Max price must be a number.' }],
    [{ maxPrice: '-1' }, { maxPrice: 'Max price must be 0 or more.' }],
    [{ minBedrooms: 'two' }, { minBedrooms: 'Min bedrooms must be a number.' }],
    [{ minBedrooms: '-1' }, { minBedrooms: 'Min bedrooms must be 0 or more.' }],
    [{ minBedrooms: '2.5' }, { minBedrooms: 'Min bedrooms must be a whole number.' }],
    // The API checks whole before the sign, so a negative fraction reads the same on both sides.
    [{ minBedrooms: '-1.5' }, { minBedrooms: 'Min bedrooms must be a whole number.' }],
    [{ targetBudget: 'lots' }, { targetBudget: 'Target budget must be a number.' }],
    [{ targetBudget: '0' }, { targetBudget: 'Target budget must be more than $0.' }],
    [{ targetBudget: '-5' }, { targetBudget: 'Target budget must be more than $0.' }],
  ])('%j gives %j', (partial, messages) => {
    expect(validate(drafts(partial))).toEqual({ messages, conflict: false });
  });

  it('accepts the boundaries', () => {
    expect(
      validate(drafts({ minPrice: '0', maxPrice: '0', minBedrooms: '0', targetBudget: '0.01' })),
    ).toEqual({ messages: {}, conflict: false });
  });

  it('flags min above max, but not equal prices', () => {
    expect(validate(drafts({ minPrice: '550000', maxPrice: '$400,000' }))).toEqual({
      messages: {},
      conflict: true,
    });
    expect(validate(drafts({ minPrice: '400000', maxPrice: '400000' })).conflict).toBe(false);
  });

  it('skips the min/max conflict when either price has its own error', () => {
    expect(validate(drafts({ minPrice: '5', maxPrice: '-1' }))).toEqual({
      messages: { maxPrice: 'Max price must be 0 or more.' },
      conflict: false,
    });
    expect(validate(drafts({ minPrice: 'abc', maxPrice: '1' })).conflict).toBe(false);
  });

  it('has no rules for city, keyword or status', () => {
    expect(validate(drafts({ city: 'x'.repeat(500), keyword: '%', status: 'bogus' }))).toEqual({
      messages: {},
      conflict: false,
    });
  });
});

describe('toSearchParams', () => {
  it('omits every empty field', () => {
    expect(toSearchParams(EMPTY_DRAFTS)).toStrictEqual({});
    expect(toSearchParams(drafts({ city: '  ', keyword: ' ' }))).toStrictEqual({});
  });

  it('parses numbers, trims text and wraps the status', () => {
    expect(
      toSearchParams({
        minPrice: '$400,000',
        maxPrice: '550000',
        minBedrooms: '2.0',
        city: ' Fairfax ',
        keyword: '  pets allowed ',
        targetBudget: '450000',
        status: 'sold',
      }),
    ).toStrictEqual({
      minPrice: 400000,
      maxPrice: 550000,
      minBedrooms: 2,
      city: 'Fairfax',
      keyword: 'pets allowed',
      targetBudget: 450000,
      status: ['sold'],
    });
  });
});

describe('draftsFrom', () => {
  it('writes numbers as plain digits and picks a single status', () => {
    expect(
      draftsFrom({
        minPrice: 400000,
        maxPrice: 1e21,
        minBedrooms: 2,
        city: 'Vienna',
        keyword: 'yard',
        targetBudget: 450000.5,
        status: ['pending'],
        page: 3,
        pageSize: 20,
      }),
    ).toEqual({
      minPrice: '400000',
      maxPrice: '1000000000000000000000',
      minBedrooms: '2',
      city: 'Vienna',
      keyword: 'yard',
      targetBudget: '450000.5',
      status: 'pending',
    });
  });

  it('writes tiny fractions without an exponent, so they seed a valid draft', () => {
    const seeded = draftsFrom({ minPrice: 1e-7, maxPrice: 1.5e-7, targetBudget: 2.5e-7 });
    expect(seeded).toEqual(
      drafts({ minPrice: '0.0000001', maxPrice: '0.00000015', targetBudget: '0.00000025' }),
    );
    expect(validate(seeded)).toEqual({ messages: {}, conflict: false });
    expect(draftsFrom({ minPrice: -1e-7 }).minPrice).toBe('-0.0000001');
    // Too many digits for 20 decimals: kept as is (and so marked) rather than rounded.
    expect(draftsFrom({ minPrice: 1.2345678901234566e-7 }).minPrice).toBe('1.2345678901234566e-7');
  });

  it('shows Any for several statuses and empties non-finite numbers', () => {
    expect(draftsFrom({ status: ['active', 'sold'], minPrice: NaN })).toEqual(EMPTY_DRAFTS);
    expect(draftsFrom({ status: ['sold', 'sold'] }).status).toBe('sold');
  });
});

describe('filterKey', () => {
  it('ignores page, pageSize and status order', () => {
    expect(filterKey({ city: 'Vienna', page: 3, pageSize: 20 })).toBe(
      filterKey({ city: 'Vienna' }),
    );
    expect(filterKey({ status: ['sold', 'active'] })).toBe(
      filterKey({ status: ['active', 'sold'] }),
    );
    expect(filterKey({ minPrice: 1 })).not.toBe(filterKey({ minPrice: 2 }));
  });
});

describe('displayValue and normalizeDraft', () => {
  it('formats a blurred whole money draft and leaves everything else as typed', () => {
    expect(displayValue('minPrice', '400000', false)).toBe('$400,000');
    expect(displayValue('targetBudget', '0', false)).toBe('$0');
    expect(displayValue('minPrice', '400000', true)).toBe('400000');
    expect(displayValue('maxPrice', '400000.5', false)).toBe('400000.5');
    expect(displayValue('maxPrice', '-5', false)).toBe('-5');
    expect(displayValue('maxPrice', 'abc', false)).toBe('abc');
    expect(displayValue('minBedrooms', '2', false)).toBe('2');
  });

  it('turns a readable numeric draft into plain digits on blur', () => {
    expect(normalizeDraft('minPrice', '$400,000')).toBe('400000');
    expect(normalizeDraft('minBedrooms', '2.0')).toBe('2');
    expect(normalizeDraft('targetBudget', ' -0 ')).toBe('0');
    expect(normalizeDraft('maxPrice', '1' + '0'.repeat(21))).toBe('1' + '0'.repeat(21));
    expect(normalizeDraft('maxPrice', '0.0000001')).toBe('0.0000001');
    expect(normalizeDraft('maxPrice', '$0.00000010')).toBe('0.0000001');
    expect(normalizeDraft('maxPrice', '0.00000012345678901234567')).toBe(
      '0.00000012345678901234567',
    );
    expect(normalizeDraft('minPrice', 'abc')).toBe('abc');
    expect(normalizeDraft('keyword', ' 2.0 ')).toBe(' 2.0 ');
  });
});

describe('cityChoice', () => {
  const cities = ['Fairfax', 'Vienna'];

  it('matches a listed city case-insensitively', () => {
    expect(cityChoice(' vienna ', cities)).toEqual({ value: 'Vienna', extra: undefined });
    expect(cityChoice('', cities)).toEqual({ value: '', extra: undefined });
  });

  it('offers an unlisted city as an extra option', () => {
    expect(cityChoice('Arlington', cities)).toEqual({ value: 'Arlington', extra: 'Arlington' });
    expect(cityChoice('Vienna', [])).toEqual({ value: 'Vienna', extra: 'Vienna' });
  });
});

describe('mapServerErrors', () => {
  it('maps fields, keeps the first error per field and the empty message', () => {
    const errors: FieldError[] = [
      {
        field: 'minBedrooms',
        code: 'not_an_integer',
        message: 'Min bedrooms must be a whole number.',
      },
      { field: 'minBedrooms', code: 'negative', message: 'Min bedrooms must be 0 or more.' },
      { field: 'keyword', code: 'too_long', message: '' },
    ];
    expect(mapServerErrors(errors)).toEqual({
      messages: { minBedrooms: 'Min bedrooms must be a whole number.', keyword: '' },
      conflict: undefined,
      unmapped: '',
    });
  });

  it('reads min_gt_max on maxPrice as the conflict and joins unmapped messages once', () => {
    const errors: FieldError[] = [
      {
        field: 'query',
        code: 'malformed_query',
        message: 'Query string is not valid URL encoding.',
      },
      { field: 'maxPrice', code: 'min_gt_max', message: 'Min price must be less than max price.' },
      { field: 'pageSize', code: 'out_of_range', message: 'Page size must be between 1 and 100.' },
      { field: 'minprice', code: 'unknown_param', message: 'Unknown parameter.' },
      { field: 'maxprice', code: 'unknown_param', message: 'Unknown parameter.' },
      { field: 'page', code: 'out_of_range', message: '' },
    ];
    expect(mapServerErrors(errors)).toEqual({
      messages: {},
      conflict: 'Min price must be less than max price.',
      unmapped:
        'Query string is not valid URL encoding. Page size must be between 1 and 100. Unknown parameter.',
    });
    expect(mapServerErrors(undefined)).toEqual({ messages: {}, conflict: undefined, unmapped: '' });
  });
});

describe('resolveMarks', () => {
  const none = new Set<FieldName>();
  const both = new Set<FieldName>(['minPrice', 'maxPrice']);
  const noServer = mapServerErrors([]);
  const serverConflict = mapServerErrors([
    { field: 'maxPrice', code: 'min_gt_max', message: 'Min price must be less than max price.' },
  ]);

  it('shows a client error only once its field is revealed', () => {
    const client = validate(drafts({ minPrice: '-5' }));
    expect(resolveMarks(client, noServer, none, none)).toEqual({});
    expect(resolveMarks(client, noServer, new Set(['minPrice']), none)).toEqual({
      minPrice: { kind: 'message', text: 'Min price must be 0 or more.' },
    });
  });

  it('puts the conflict under Max price and marks Min price, once both are revealed', () => {
    const client = validate(drafts({ minPrice: '5', maxPrice: '1' }));
    expect(resolveMarks(client, noServer, new Set(['minPrice']), none)).toEqual({});
    expect(resolveMarks(client, noServer, both, none)).toEqual({
      minPrice: { kind: 'conflict' },
      maxPrice: { kind: 'message', text: 'Min price must be less than max price.' },
    });
  });

  it('prefers a revealed client error to a server error', () => {
    const client = validate(drafts({ minBedrooms: 'x' }));
    const server = mapServerErrors([
      { field: 'minBedrooms', code: 'negative', message: 'Min bedrooms must be 0 or more.' },
    ]);
    expect(resolveMarks(client, server, none, none)).toEqual({
      minBedrooms: { kind: 'message', text: 'Min bedrooms must be 0 or more.' },
    });
    expect(resolveMarks(client, server, new Set(['minBedrooms']), none)).toEqual({
      minBedrooms: { kind: 'message', text: 'Min bedrooms must be a number.' },
    });
  });

  it('drops the server conflict when either price is dismissed', () => {
    const client = validate(EMPTY_DRAFTS);
    expect(resolveMarks(client, serverConflict, none, none)).toEqual({
      minPrice: { kind: 'conflict' },
      maxPrice: { kind: 'message', text: 'Min price must be less than max price.' },
    });
    expect(resolveMarks(client, serverConflict, none, new Set(['minPrice']))).toEqual({});
    expect(resolveMarks(client, serverConflict, none, new Set(['maxPrice']))).toEqual({});
  });

  it('keeps Min price own server message over the conflict mark', () => {
    const server = mapServerErrors([
      { field: 'minPrice', code: 'negative', message: 'Min price must be 0 or more.' },
      { field: 'maxPrice', code: 'min_gt_max', message: 'Min price must be less than max price.' },
    ]);
    expect(resolveMarks(validate(EMPTY_DRAFTS), server, none, none).minPrice).toEqual({
      kind: 'message',
      text: 'Min price must be 0 or more.',
    });
  });
});
