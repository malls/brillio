// Money and whole-number formatting (BRANDING.md section 7.1). Pinned to en-US so output never
// depends on the browser's locale.

const wholeNumber = new Intl.NumberFormat('en-US', {
  minimumFractionDigits: 0,
  maximumFractionDigits: 0,
});

// Thousands separators, rounded to a whole number: 1450 -> '1,450'.
export function formatWholeNumber(n: number): string {
  return wholeNumber.format(n);
}

// '$' plus thousands separators, no cents: 450000 -> '$450,000'.
export function formatPrice(amount: number): string {
  return `$${formatWholeNumber(amount)}`;
}
