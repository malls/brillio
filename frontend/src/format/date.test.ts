import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { formatListedDate } from './date';

// A US time zone: there, new Date('2026-08-29') would read as Aug 28. The formatter must not care.
beforeAll(() => {
  vi.stubEnv('TZ', 'America/Los_Angeles');
});

afterAll(() => {
  vi.unstubAllEnvs();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('formatListedDate', () => {
  it('runs where parsing the date with Date would shift the day', () => {
    expect(new Date('2026-08-29').getDate()).toBe(28);
  });

  it('omits the year in the current year', () => {
    expect(formatListedDate('2026-08-29', 2026)).toBe('Listed Aug 29');
  });

  it.each([2025, 2027])('adds the year when the current year is %d', (currentYear) => {
    expect(formatListedDate('2026-08-29', currentYear)).toBe('Listed Aug 29, 2026');
  });

  it('drops the leading zero of the day', () => {
    expect(formatListedDate('2026-09-02', 2026)).toBe('Listed Sep 2');
  });

  it('keeps Jan 1 on its own day and year', () => {
    expect(formatListedDate('2026-01-01', 2026)).toBe('Listed Jan 1');
    expect(formatListedDate('2026-01-01', 2027)).toBe('Listed Jan 1, 2026');
  });

  it.each([
    ['01', 'Jan'],
    ['02', 'Feb'],
    ['03', 'Mar'],
    ['04', 'Apr'],
    ['05', 'May'],
    ['06', 'Jun'],
    ['07', 'Jul'],
    ['08', 'Aug'],
    ['09', 'Sep'],
    ['10', 'Oct'],
    ['11', 'Nov'],
    ['12', 'Dec'],
  ])('names month %s %s', (month, name) => {
    expect(formatListedDate(`2026-${month}-15`, 2026)).toBe(`Listed ${name} 15`);
  });

  it.each(['garbage', '2026-13-01', '2026-00-10', '2026-08-00', '2026-08-32', '2026-8-29', ''])(
    'shows %j raw',
    (raw) => {
      expect(formatListedDate(raw, 2026)).toBe(`Listed ${raw}`);
    },
  );

  it('defaults the current year to the local clock', () => {
    vi.useFakeTimers({ toFake: ['Date'], now: new Date(2026, 9, 13) });

    expect(formatListedDate('2026-08-29')).toBe('Listed Aug 29');
    expect(formatListedDate('2025-12-31')).toBe('Listed Dec 31, 2025');
  });
});
