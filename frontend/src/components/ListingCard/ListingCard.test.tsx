import { render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { listingA1, listingA3, listingA7, listingB7 } from '../../test/fixtures';
import { ListingCard } from './ListingCard';

// True when `a` comes before `b` in document order.
function precedes(a: Node, b: Node): boolean {
  return (a.compareDocumentPosition(b) & Node.DOCUMENT_POSITION_FOLLOWING) !== 0;
}

describe('ListingCard', () => {
  // The listed date drops the year inside the current year (BRANDING.md section 7.1); pin the
  // clock to the fixtures' as-of date.
  beforeEach(() => {
    vi.useFakeTimers({ toFake: ['Date'], now: new Date(2026, 9, 13) });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders the address verbatim as a level 3 heading', () => {
    const { unmount } = render(<ListingCard item={listingA1} keywordTokens={[]} />);
    expect(
      screen.getByRole('heading', { level: 3, name: '123 Main St, Apt 4B' }),
    ).toBeInTheDocument();
    unmount();

    // The feed does not normalize addresses and neither does the card.
    render(<ListingCard item={listingB7} keywordTokens={[]} />);
    expect(
      screen.getByRole('heading', { level: 3, name: '123 Main Street, Unit 4B' }),
    ).toBeInTheDocument();
  });

  it('renders every meta item in BRANDING order', () => {
    const { unmount } = render(<ListingCard item={listingA1} keywordTokens={[]} />);
    // docs/branding-demo.html line 137.
    expect(
      screen.getByText('Springfield, VA 22150 · 2 bd · 1.5 ba · 980 sqft · Listed Aug 29 · MLS A'),
    ).toBeInTheDocument();
    unmount();

    render(<ListingCard item={listingB7} keywordTokens={[]} />);
    expect(screen.getByText(/ · Listed Aug 27 · MLS B$/).textContent).toBe(
      'Springfield, VA 22150 · 2 bd · 1.5 ba · 980 sqft · Listed Aug 27 · MLS B',
    );
  });

  it('shows PENDING first for A7', () => {
    render(<ListingCard item={listingA7} keywordTokens={[]} />);

    const label = screen.getByText('PENDING');
    expect(label.parentElement?.textContent).toBe(
      'PENDING · Chantilly, VA 20151 · 4 bd · 2.5 ba · 1,950 sqft · Listed Jul 28 · MLS A',
    );
    expect(label.previousSibling).toBeNull();
  });

  it('shows no status label for an active listing', () => {
    const { container } = render(<ListingCard item={listingA1} keywordTokens={[]} />);

    expect(container.textContent).not.toMatch(/ACTIVE|PENDING|SOLD/);
    const meta = screen.getByText(/^Springfield, VA 22150/);
    expect(meta.textContent?.startsWith('Springfield')).toBe(true);
    expect(meta.children).toHaveLength(0);
  });

  it('emphasizes Pets and pets case-insensitively and leaves the rest untouched', () => {
    const { container, unmount } = render(
      <ListingCard item={listingB7} keywordTokens={['pets']} />,
    );
    const emphasized = container.querySelectorAll('em');
    expect(emphasized).toHaveLength(1);
    expect(emphasized[0]?.textContent).toBe('Pets');
    expect(emphasized[0]?.parentElement?.textContent).toBe(listingB7.description);
    unmount();

    const a3 = render(<ListingCard item={listingA3} keywordTokens={['pets']} />);
    const a3Emphasized = a3.container.querySelectorAll('em');
    expect(a3Emphasized).toHaveLength(1);
    expect(a3Emphasized[0]?.textContent).toBe('pets');
    expect(a3Emphasized[0]?.parentElement?.textContent).toBe(listingA3.description);
  });

  it('renders no emphasis without keyword tokens', () => {
    const { container } = render(<ListingCard item={listingB7} keywordTokens={[]} />);

    expect(container.querySelector('em')).toBeNull();
    expect(screen.getByText(listingB7.description)).toBeInTheDocument();
  });

  it('renders the price over the score pill', () => {
    render(<ListingCard item={listingA1} keywordTokens={[]} />);

    const price = screen.getByText('$450,000');
    const rank = screen.getByText('No. 1');
    const score = screen.getByText('0.74');
    expect(precedes(price, rank)).toBe(true);
    expect(precedes(rank, score)).toBe(true);
    expect(price.parentElement?.contains(rank)).toBe(true);
  });

  it('renders the city monogram', () => {
    render(<ListingCard item={listingA1} keywordTokens={[]} />);

    const monogram = screen.getByText('Sp');
    expect(monogram).toHaveAttribute('aria-hidden', 'true');
    expect(precedes(monogram, screen.getByRole('heading', { level: 3 }))).toBe(true);
  });

  it('keeps the price block between the meta line and the description', () => {
    // Stands in for the 760px media query, which jsdom does not apply: the reflow is by grid
    // areas, so DOM order is the mobile visual order.
    render(<ListingCard item={listingA1} keywordTokens={[]} />);

    const address = screen.getByRole('heading', { level: 3 });
    const meta = screen.getByText(/^Springfield, VA 22150/);
    const price = screen.getByText('$450,000');
    const description = screen.getByText(listingA1.description);
    expect(precedes(address, meta)).toBe(true);
    expect(precedes(meta, price)).toBe(true);
    expect(precedes(price, description)).toBe(true);
  });

  it('renders no image, icon, link or button', () => {
    const { container } = render(<ListingCard item={listingA7} keywordTokens={['pets']} />);

    expect(container.querySelector('img, svg, picture, i, a, button')).toBeNull();
    expect(screen.queryByRole('link')).not.toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
    expect(screen.queryByRole('img')).not.toBeInTheDocument();
  });
});
