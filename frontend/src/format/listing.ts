import type { ListingResult, ListingStatus, SearchResponse } from '../api/types';
import { formatListedDate } from './date';
import { formatPrice, formatWholeNumber } from './money';

// Listing, score and status-line formatting (BRANDING.md sections 5.3-5.7 and 7.1). Pure
// functions; none of them throws.

// Separates the card's meta items (BRANDING.md section 5.4) and the parts of the status line's
// left half (section 5.3).
export const META_SEPARATOR = ' · ';

// 2 -> '2 bd'.
export function formatBedrooms(n: number): string {
  return `${n} bd`;
}

// Drops '.0' and keeps '.5': 1 -> '1 ba', 1.5 -> '1.5 ba'.
export function formatBathrooms(n: number): string {
  return `${n} ba`;
}

// 1450 -> '1,450 sqft'.
export function formatSqft(n: number): string {
  return `${formatWholeNumber(n)} sqft`;
}

// Underscore to space: 'MLS_A' -> 'MLS A'.
export function formatSource(source: string): string {
  return source.replaceAll('_', ' ');
}

// The Label-style status, shown only when the listing is not active: 'PENDING', 'SOLD', or null.
export function formatStatusLabel(status: ListingStatus): string | null {
  return status === 'active' ? null : status.toUpperCase();
}

// 'Springfield, VA 22150'.
export function formatLocation(listing: Pick<ListingResult, 'city' | 'state' | 'zip'>): string {
  return `${listing.city}, ${listing.state} ${listing.zip}`;
}

// The card's meta items in BRANDING.md section 5.4 order, without the status label (the card
// sets it in Label style ahead of these): location, bedrooms, bathrooms, area, listed date,
// source. Join with META_SEPARATOR.
export function formatListingMeta(item: ListingResult, currentYear?: number): string[] {
  return [
    formatLocation(item),
    formatBedrooms(item.bedrooms),
    formatBathrooms(item.bathrooms),
    formatSqft(item.sqft),
    formatListedDate(item.listedDate, currentYear),
    formatSource(item.source),
  ];
}

// Two decimals, for display only: 0.7414 -> '0.74'.
export function formatScore(score: number): string {
  return score.toFixed(2);
}

// 1 -> 'No. 1'.
export function formatRank(rank: number): string {
  return `No. ${rank}`;
}

// The city's first two letters, first uppercase, second lowercase: 'Springfield' -> 'Sp'.
export function formatMonogram(city: string): string {
  const name = city.trim();
  return name.charAt(0).toUpperCase() + name.charAt(1).toLowerCase();
}

// 1 -> '1 result', 12 -> '12 results', 0 -> '0 results'.
export function formatResultCount(n: number): string {
  return `${formatWholeNumber(n)} ${n === 1 ? 'result' : 'results'}`;
}

// How the results are ordered: 'ranked for $450,000', or 'newest first' without a budget.
export function formatOrdering(targetBudget: number | null): string {
  return targetBudget === null ? 'newest first' : `ranked for ${formatPrice(targetBudget)}`;
}

// The status line's left half for a 2xx response: '12 results · ranked for $450,000',
// '12 results · newest first', or '0 results'.
export function formatResultSummary(data: SearchResponse): string {
  const { totalItems } = data.page;
  if (totalItems === 0) {
    return formatResultCount(0);
  }
  return formatResultCount(totalItems) + META_SEPARATOR + formatOrdering(data.scoring.targetBudget);
}

// The status line's right half: 'Page 1 of 2', or '' when there are no pages.
export function formatPageOf(page: number, totalPages: number): string {
  return totalPages > 0 ? `Page ${page} of ${totalPages}` : '';
}
