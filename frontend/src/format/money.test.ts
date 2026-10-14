import { describe, expect, it } from 'vitest';
import { formatPrice, formatWholeNumber } from './money';

describe('formatPrice', () => {
  it.each([
    [450000, '$450,000'],
    [0, '$0'],
    [980, '$980'],
    [1234567, '$1,234,567'],
    [399000.4, '$399,000'],
    [399999.5, '$400,000'],
  ])('formats %d as %s', (amount, expected) => {
    expect(formatPrice(amount)).toBe(expected);
  });
});

describe('formatWholeNumber', () => {
  it.each([
    [1450, '1,450'],
    [980, '980'],
    [12, '12'],
    [1000000, '1,000,000'],
  ])('formats %d as %s', (n, expected) => {
    expect(formatWholeNumber(n)).toBe(expected);
  });
});
