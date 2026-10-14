import { render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { listingA1, listingA3, listingA7, searchResponse } from '../../test/fixtures';
import { ListingList, RESULTS_HEADING_ID } from './ListingList';

describe('ListingList', () => {
  it('renders one card per item in rank order inside an ordered list', () => {
    render(<ListingList items={searchResponse.items} keywordTokens={[]} />);

    const list = screen.getByRole('list');
    expect(list.tagName).toBe('OL');
    const items = within(list).getAllByRole('listitem');
    expect(items).toHaveLength(3);
    expect(
      items.map((item) => within(item).getByRole('heading', { level: 3 }).textContent),
    ).toEqual([listingA1.address, listingA3.address, listingA7.address]);
    expect(items.map((item) => within(item).getByText(/^No\. \d+$/).textContent)).toEqual([
      'No. 1',
      'No. 4',
      'No. 10',
    ]);
  });

  it('labels the region with the Results heading and leaves the live region to the page', () => {
    render(<ListingList items={searchResponse.items} keywordTokens={[]} />);

    // SearchPage's <main> is the one live region (SPEC section 7.6): nested regions can
    // announce twice.
    const region = screen.getByRole('region', { name: 'Results' });
    expect(region).not.toHaveAttribute('aria-live');
    expect(region).toHaveAttribute('aria-labelledby', RESULTS_HEADING_ID);
    expect(within(region).getByRole('list')).toBeInTheDocument();
  });

  it('exposes a level 2 Results heading that script can focus', () => {
    render(<ListingList items={searchResponse.items} keywordTokens={[]} />);

    const heading = screen.getByRole('heading', { level: 2, name: 'Results' });
    expect(heading).toHaveAttribute('id', RESULTS_HEADING_ID);
    expect(heading).toHaveAttribute('tabindex', '-1');

    document.getElementById(RESULTS_HEADING_ID)?.focus();
    expect(heading).toHaveFocus();
  });

  it('passes keyword tokens to every card', () => {
    const { container } = render(
      <ListingList items={[listingA3, listingA7]} keywordTokens={['pets']} />,
    );

    const emphasized = Array.from(container.querySelectorAll('em'));
    expect(emphasized.map((em) => em.textContent)).toEqual(['pets', 'pets']);
    expect(emphasized[0]?.closest('li')).not.toBe(emphasized[1]?.closest('li'));
  });

  it('renders the heading and no list for zero items', () => {
    render(<ListingList items={[]} keywordTokens={[]} />);

    expect(screen.getByRole('region', { name: 'Results' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 2, name: 'Results' })).toBeInTheDocument();
    expect(screen.queryByRole('list')).not.toBeInTheDocument();
    expect(screen.queryByRole('article')).not.toBeInTheDocument();
  });

  it('keys cards by source and id', () => {
    // Ids are unique only per source (SPEC section 4); keying on id alone would make React log a
    // duplicate-key error and could reuse the wrong card.
    const consoleError = vi.spyOn(console, 'error');

    render(
      <ListingList items={[listingA1, { ...listingA1, source: 'MLS_B' }]} keywordTokens={[]} />,
    );

    expect(screen.getAllByRole('article')).toHaveLength(2);
    expect(screen.getByText(/ · MLS A$/)).toBeInTheDocument();
    expect(screen.getByText(/ · MLS B$/)).toBeInTheDocument();
    expect(consoleError).not.toHaveBeenCalled();
  });
});
