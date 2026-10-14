import { describe, expect, it } from 'vitest';
import { listingA1, listingA3, listingB7 } from '../test/fixtures';
import { type EmphasisSegment, emphasisSegments, splitKeyword } from './keyword';

// The matched runs' text, in order.
function matches(segments: EmphasisSegment[]): string[] {
  return segments.filter((segment) => segment.match).map((segment) => segment.text);
}

// Joining every run's text gives back the input.
function joined(segments: EmphasisSegment[]): string {
  return segments.map((segment) => segment.text).join('');
}

describe('splitKeyword', () => {
  it('splitKeyword splits on whitespace like the server', () => {
    expect(splitKeyword('  pets  allowed ')).toEqual(['pets', 'allowed']);
    expect(splitKeyword('pets\tallowed\ncondo')).toEqual(['pets', 'allowed', 'condo']);
    expect(splitKeyword('')).toEqual([]);
    expect(splitKeyword('   ')).toEqual([]);
    expect(splitKeyword(undefined)).toEqual([]);
  });
});

describe('emphasisSegments', () => {
  it('returns one plain segment without tokens', () => {
    expect(emphasisSegments(listingA1.description, [])).toEqual([
      { text: listingA1.description, match: false },
    ]);
  });

  it('returns no segments for empty text', () => {
    expect(emphasisSegments('', [])).toEqual([]);
    expect(emphasisSegments('', ['pets'])).toEqual([]);
  });

  it('matches case-insensitively and keeps the original casing', () => {
    const b7 = emphasisSegments(listingB7.description, ['pets']);
    expect(matches(b7)).toEqual(['Pets']);
    expect(b7).toEqual([
      { text: 'Top floor condo, walk to shopping. ', match: false },
      { text: 'Pets', match: true },
      { text: ' allowed.', match: false },
    ]);

    const a3 = emphasisSegments(listingA3.description, ['PETS']);
    expect(matches(a3)).toEqual(['pets']);
    expect(joined(a3)).toBe(listingA3.description);
  });

  it('marks every occurrence', () => {
    const segments = emphasisSegments('pets, pets', ['pets']);
    expect(segments).toEqual([
      { text: 'pets', match: true },
      { text: ', ', match: false },
      { text: 'pets', match: true },
    ]);
  });

  it('handles several tokens in any order', () => {
    const forward = emphasisSegments(listingB7.description, ['condo', 'pets']);
    const backward = emphasisSegments(listingB7.description, ['pets', 'condo']);
    expect(forward).toEqual(backward);
    expect(matches(forward)).toEqual(['condo', 'Pets']);
    expect(joined(forward)).toBe(listingB7.description);
  });

  it('merges overlapping and adjacent matches into one run', () => {
    expect(emphasisSegments('Pets allowed', ['pet', 'pets'])).toEqual([
      { text: 'Pets', match: true },
      { text: ' allowed', match: false },
    ]);
    // Overlapping: 'top-fl' and 'floor' share 'fl'.
    expect(matches(emphasisSegments(listingA1.description, ['top-fl', 'floor']))).toEqual([
      'top-floor',
    ]);
    // Adjacent: 'top' ends where '-floor' starts.
    expect(matches(emphasisSegments(listingA1.description, ['top', '-floor']))).toEqual([
      'top-floor',
    ]);
    // Overlapping occurrences of one token.
    expect(emphasisSegments('aaa', ['aa'])).toEqual([{ text: 'aaa', match: true }]);
  });

  it('treats regex metacharacters literally', () => {
    const periods = emphasisSegments(listingA1.description, ['.']);
    expect(matches(periods)).toEqual(['.', '.']);
    expect(joined(periods)).toBe(listingA1.description);

    // '.' is not a wildcard: 'P.ts' does not match 'Pets'.
    expect(matches(emphasisSegments(listingB7.description, ['P.ts']))).toEqual([]);

    for (const token of ['(', ')', '$', '\\', '[', '*', '+', '?', '^', '{', '|']) {
      expect(() => emphasisSegments(listingA1.description, [token])).not.toThrow();
      expect(emphasisSegments(listingA1.description, [token])).toEqual([
        { text: listingA1.description, match: false },
      ]);
    }
    expect(matches(emphasisSegments('cost (est.) $5', ['(est.)', '$5']))).toEqual(['(est.)', '$5']);
  });

  it('ignores blank tokens', () => {
    expect(emphasisSegments(listingA1.description, ['', '  '])).toEqual([
      { text: listingA1.description, match: false },
    ]);
  });
});
