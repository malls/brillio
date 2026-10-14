import { act, cleanup, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { Mock, MockInstance } from 'vitest';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { SearchResponse } from '../api/types';
import { RESULTS_HEADING_ID } from '../components/ListingList/ListingList';
import type { Deferred, JsonResponseInit } from '../test/fixtures';
import {
  cities,
  deferred,
  emptySearchResponse,
  internalErrorBody,
  jsonResponse,
  listingA1,
  listingA3,
  listingB7,
  malformedQueryErrorBody,
  minGtMaxErrorBody,
  newestFirstResponse,
  searchResponse,
  searchResponsePage2,
} from '../test/fixtures';
import { SearchPage } from './SearchPage';

// Integration tests for the search page (docs/SPEC.md section 8): the real hooks, client and
// components, with fetch stubbed. /api/cities always answers; every search the page issues takes
// the next entry of a per-test queue, so each test queues exactly the searches it expects. No
// StrictMode wrapper (as the hook tests), so each mount effect runs once.

const SEARCH_PATH = '/api/listings/search';
const FIELD_LABELS = [
  'Min price',
  'Max price',
  'Min bedrooms',
  'City',
  'Keyword',
  'Target budget',
  'Status',
] as const;
const MIN_GT_MAX = 'Min price must be less than max price.';
const ERROR_HEADING = "Couldn't load results.";

let fetchMock: Mock<typeof fetch>;
let consoleError: MockInstance<typeof console.error>;
let searchQueue: (() => Promise<Response>)[];
let unexpected: string[];

function urlOf(input: RequestInfo | URL): string {
  if (typeof input === 'string') {
    return input;
  }
  return input instanceof URL ? input.href : input.url;
}

function isSearch(url: string): boolean {
  return url === SEARCH_PATH || url.startsWith(`${SEARCH_PATH}?`);
}

beforeEach(() => {
  // jsdom keeps the URL from test to test within a file.
  window.history.replaceState(null, '', '/');
  searchQueue = [];
  unexpected = [];
  fetchMock = vi.fn<typeof fetch>((input) => {
    const url = urlOf(input);
    if (url === '/api/cities') {
      return Promise.resolve(jsonResponse({ cities }));
    }
    const next = isSearch(url) ? searchQueue.shift() : undefined;
    if (next === undefined) {
      unexpected.push(url);
      return Promise.reject(new Error(`unexpected fetch ${url}`));
    }
    return next();
  });
  vi.stubGlobal('fetch', fetchMock);
  consoleError = vi.spyOn(console, 'error').mockImplementation(() => undefined);
});

afterEach(() => {
  // Unmount first (setup.ts does too, but runs after this hook), so a failed check below cannot
  // leave this test's page mounted into the next test.
  cleanup();
  // Every queued search was issued, and nothing unqueued was fetched.
  expect(unexpected).toEqual([]);
  expect(searchQueue).toHaveLength(0);
  // The page logs nothing itself: the only console errors are the hook's failed-search logs (no
  // act warnings either).
  for (const [first] of consoleError.mock.calls) {
    expect(first).toBe('Listing search failed');
  }
});

// Queues the next search's answer: a JSON response now.
function reply(body: unknown, init?: JsonResponseInit) {
  searchQueue.push(() => Promise.resolve(jsonResponse(body, init)));
}

// Queues the next search's answer as a response the test settles later. The promise ignores the
// request's signal, so an aborted request is proven ignored by resolving it late.
function replyLater(): Deferred<Response> {
  const pending = deferred<Response>();
  searchQueue.push(() => pending.promise);
  return pending;
}

// Queues the next search's answer as a fetch failure (no response at all).
function failNetwork() {
  searchQueue.push(() => Promise.reject(new TypeError('Failed to fetch')));
}

function searchCalls() {
  return fetchMock.mock.calls.filter(([input]) => isSearch(urlOf(input)));
}

function searchUrls(): string[] {
  return searchCalls().map(([input]) => urlOf(input));
}

function searchSignal(index: number): AbortSignal | null | undefined {
  return searchCalls()[index]?.[1]?.signal;
}

function renderAt(url: string) {
  window.history.replaceState(null, '', url);
  const user = userEvent.setup();
  const view = render(<SearchPage />);
  return { user, container: view.container };
}

function currentUrl(): string {
  return window.location.pathname + window.location.search;
}

// The status line's left and right halves.
function statusHalves(): (string | null)[] {
  return Array.from(screen.getByRole('status').children, (half) => half.textContent);
}

function cardRanks(): (string | null)[] {
  return screen.queryAllByText(/^No\. \d+$/).map((rank) => rank.textContent);
}

function cardAddresses(): (string | null)[] {
  return screen.queryAllByRole('heading', { level: 3 }).map((heading) => heading.textContent);
}

function field(label: string): HTMLElement {
  return screen.getByLabelText(label);
}

function resultsRegion(): HTMLElement {
  return screen.getByRole('main', { name: 'Results' });
}

function pageButton(n: number): HTMLElement {
  return screen.getByRole('button', { name: `Page ${n}` });
}

async function waitForStatus(left: string, right: string) {
  await waitFor(() => {
    expect(statusHalves()).toEqual([left, right]);
  });
}

// Renders the default search (empty URL) and waits for its list.
async function renderDefaultSettled() {
  reply(newestFirstResponse);
  const view = renderAt('/');
  await waitForStatus('12 results · newest first', 'Page 1 of 2');
  // The City options arrive from their own request.
  await screen.findByRole('option', { name: 'Vienna' });
  return view;
}

// Renders `/?targetBudget=450000` and waits for page 1 (ranks 1, 4, 10 of the trimmed fixture).
async function renderBudgetSettled() {
  reply(searchResponse);
  const view = renderAt('/?targetBudget=450000');
  await waitForStatus('12 results · ranked for $450,000', 'Page 1 of 2');
  return view;
}

describe('SearchPage', () => {
  // T0
  it('lays out the five shell rows in BRANDING section 4.2 order', async () => {
    const pending = replyLater();
    const { container } = renderAt('/?targetBudget=450000');
    const shell = container.firstElementChild;
    if (shell === null) {
      throw new Error('SearchPage rendered no shell element');
    }

    const banner = screen.getByRole('banner');
    expect(within(banner).getByRole('heading', { level: 1 })).toHaveTextContent('Brillio Listings');
    expect(banner).toHaveTextContent('Northern Virginia');
    const [, wrapper] = Array.from(shell.children);
    if (wrapper === undefined) {
      throw new Error('SearchPage rendered no filter row');
    }
    // Row 2 holds the filter bar and the loading line under it.
    expect(Array.from(wrapper.children)).toEqual([
      screen.getByRole('search', { name: 'Filters' }),
      screen.getByRole('progressbar', { name: 'Loading' }),
    ]);
    // First load: no pager yet.
    expect(Array.from(shell.children)).toEqual([
      banner,
      wrapper,
      screen.getByRole('status'),
      resultsRegion(),
    ]);

    act(() => {
      pending.resolve(jsonResponse(searchResponse));
    });
    const pager = await screen.findByRole('navigation', { name: 'Pagination' });
    expect(Array.from(shell.children)).toEqual([
      banner,
      wrapper,
      screen.getByRole('status'),
      resultsRegion(),
      pager,
    ]);
    expect(resultsRegion()).toHaveAttribute('aria-live', 'polite');
    expect(resultsRegion()).toHaveAttribute('aria-busy', 'false');
    expect(resultsRegion()).toHaveAttribute('tabindex', '0');
  });

  // T1
  it('shows the loading state, then the results, status line and pager', async () => {
    const pending = replyLater();
    renderAt('/?targetBudget=450000');

    expect(statusHalves()).toEqual(['Loading', '']);
    expect(screen.getByRole('progressbar', { name: 'Loading' })).toBeInTheDocument();
    expect(resultsRegion()).toHaveAttribute('aria-busy', 'true');
    expect(screen.getByRole('button', { name: 'Search' })).toBeDisabled();
    expect(screen.queryAllByRole('article')).toHaveLength(0);
    expect(screen.queryByRole('navigation', { name: 'Pagination' })).not.toBeInTheDocument();
    expect(searchUrls()).toEqual([`${SEARCH_PATH}?targetBudget=450000`]);

    act(() => {
      pending.resolve(jsonResponse(searchResponse));
    });
    await waitForStatus('12 results · ranked for $450,000', 'Page 1 of 2');
    expect(cardRanks()).toEqual(['No. 1', 'No. 4', 'No. 10']);
    expect(cardAddresses()).toEqual([listingA1.address, listingA3.address, '42 Willow Way']);
    expect(pageButton(1)).toHaveAttribute('aria-current', 'page');
    expect(pageButton(2)).not.toHaveAttribute('aria-current');
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    expect(resultsRegion()).toHaveAttribute('aria-busy', 'false');
    expect(screen.getByRole('button', { name: 'Search' })).toBeEnabled();
    expect(consoleError).not.toHaveBeenCalled();
  });

  // T2
  it('shows No results for an empty search, and Clear filters runs the default search', async () => {
    reply(emptySearchResponse);
    const { user } = renderAt('/?city=Arlington');

    expect(await screen.findByRole('heading', { name: 'No results.' })).toBeInTheDocument();
    expect(screen.getByText('Adjust your filters.')).toBeInTheDocument();
    expect(statusHalves()).toEqual(['0 results', '']);
    expect(screen.queryByRole('navigation', { name: 'Pagination' })).not.toBeInTheDocument();
    expect(searchUrls()).toEqual([`${SEARCH_PATH}?city=Arlington`]);

    reply(newestFirstResponse);
    await user.click(screen.getByRole('button', { name: 'Clear filters' }));

    expect(currentUrl()).toBe('/');
    await waitForStatus('12 results · newest first', 'Page 1 of 2');
    expect(searchUrls()).toEqual([`${SEARCH_PATH}?city=Arlington`, SEARCH_PATH]);
    expect(screen.queryByRole('heading', { name: 'No results.' })).not.toBeInTheDocument();
    expect(cardAddresses()[0]).toBe('55 Elm Ct');
    expect(consoleError).not.toHaveBeenCalled();
  });

  // T3
  it('maps a 400 min_gt_max onto both price fields and keeps the list', async () => {
    await renderDefaultSettled();

    // A back/forward to a hand-edited URL: the bar itself never sends this pair.
    reply(minGtMaxErrorBody, { status: 400 });
    act(() => {
      window.history.pushState(null, '', '/?minPrice=500000&maxPrice=400000');
      window.dispatchEvent(new PopStateEvent('popstate'));
    });

    await waitFor(() => {
      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
    expect(searchUrls()[1]).toBe(`${SEARCH_PATH}?minPrice=500000&maxPrice=400000`);
    const minPrice = field('Min price');
    const maxPrice = field('Max price');
    expect(minPrice).toHaveAttribute('aria-invalid', 'true');
    expect(maxPrice).toHaveAttribute('aria-invalid', 'true');
    const [message, ...others] = screen.getAllByText(MIN_GT_MAX);
    expect(others).toHaveLength(0);
    expect(message?.parentElement).toBe(maxPrice.parentElement);
    expect(maxPrice).toHaveAttribute('aria-describedby', message?.id);
    expect(minPrice).toHaveAttribute('aria-describedby', message?.id);
    // The previous list and its status line stay.
    expect(statusHalves()).toEqual(['12 results · newest first', 'Page 1 of 2']);
    expect(cardAddresses()[0]).toBe('55 Elm Ct');
    expect(cardRanks()[0]).toBe('No. 1');
    expect(screen.queryByText(ERROR_HEADING)).not.toBeInTheDocument();
    expect(consoleError).not.toHaveBeenCalled();
  });

  // T4
  it('shows the error state on a 500, and Try again re-runs the same search', async () => {
    const { user } = await renderDefaultSettled();

    reply(internalErrorBody, { status: 500, requestId: 'req-500' });
    await user.selectOptions(field('City'), 'Vienna');
    await user.click(screen.getByRole('button', { name: 'Search' }));

    expect(await screen.findByRole('heading', { name: ERROR_HEADING })).toBeInTheDocument();
    expect(currentUrl()).toBe('/?city=Vienna');
    expect(screen.queryAllByRole('article')).toHaveLength(0);
    expect(statusHalves()).toEqual(['', '']);
    expect(screen.queryByRole('navigation', { name: 'Pagination' })).not.toBeInTheDocument();
    // Logged once, by the hook, with the request ID; never in the UI.
    expect(consoleError).toHaveBeenCalledTimes(1);
    expect(consoleError).toHaveBeenCalledWith(
      'Listing search failed',
      expect.objectContaining({ kind: 'http', status: 500, requestId: 'req-500' }),
    );
    expect(screen.queryByText(/internal/i)).not.toBeInTheDocument();

    reply(newestFirstResponse);
    await user.click(screen.getByRole('button', { name: 'Try again' }));

    await waitFor(() => {
      expect(cardAddresses()[0]).toBe('55 Elm Ct');
    });
    expect(searchUrls()).toEqual([
      SEARCH_PATH,
      `${SEARCH_PATH}?city=Vienna`,
      `${SEARCH_PATH}?city=Vienna`,
    ]);
    expect(screen.queryByText(ERROR_HEADING)).not.toBeInTheDocument();
    expect(statusHalves()).toEqual(['12 results · newest first', 'Page 1 of 2']);
  });

  // T5
  it('shows the error state on a network failure, and Try again recovers', async () => {
    failNetwork();
    const { user } = renderAt('/');

    expect(await screen.findByRole('heading', { name: ERROR_HEADING })).toBeInTheDocument();
    expect(statusHalves()).toEqual(['', '']);
    expect(consoleError).toHaveBeenCalledTimes(1);
    expect(consoleError).toHaveBeenCalledWith(
      'Listing search failed',
      expect.objectContaining({ kind: 'network', message: 'Failed to fetch' }),
    );

    reply(newestFirstResponse);
    await user.click(screen.getByRole('button', { name: 'Try again' }));

    await waitForStatus('12 results · newest first', 'Page 1 of 2');
    expect(searchUrls()).toEqual([SEARCH_PATH, SEARCH_PATH]);
    expect(cardRanks()[0]).toBe('No. 1');
    expect(screen.queryByText(ERROR_HEADING)).not.toBeInTheDocument();
  });

  // T6
  it('changes page: aborts the old request, keeps the list while loading, focuses the heading, continues the ranks', async () => {
    const { user } = await renderBudgetSettled();

    const page2 = replyLater();
    await user.click(pageButton(2));

    expect(currentUrl()).toBe('/?targetBudget=450000&page=2');
    expect(searchUrls()[1]).toBe(`${SEARCH_PATH}?targetBudget=450000&page=2`);
    expect(searchSignal(0)?.aborted).toBe(true);
    expect(searchSignal(1)?.aborted).toBe(false);
    expect(statusHalves()).toEqual(['Loading', 'Page 1 of 2']);
    expect(cardRanks()).toEqual(['No. 1', 'No. 4', 'No. 10']);
    // The Pager follows the URL at once.
    expect(pageButton(2)).toHaveAttribute('aria-current', 'page');
    const heading = screen.getByRole('heading', { level: 2, name: 'Results' });
    expect(heading).toHaveAttribute('id', RESULTS_HEADING_ID);
    expect(heading).toHaveFocus();

    act(() => {
      page2.resolve(jsonResponse(searchResponsePage2));
    });
    await waitForStatus('12 results · ranked for $450,000', 'Page 2 of 2');
    expect(cardRanks()).toEqual(['No. 11', 'No. 12']);
    expect(cardAddresses()).toEqual(['100 Maple Dr', '22 Birch Ln']);
    expect(heading).toHaveFocus();
    expect(consoleError).not.toHaveBeenCalled();
  });

  // T7
  it('restores page 1 on popstate', async () => {
    const { user } = await renderBudgetSettled();
    reply(searchResponsePage2);
    await user.click(pageButton(2));
    await waitForStatus('12 results · ranked for $450,000', 'Page 2 of 2');

    const page1 = replyLater();
    window.history.back();

    await waitFor(() => {
      expect(currentUrl()).toBe('/?targetBudget=450000');
    });
    await waitFor(() => {
      expect(searchUrls()).toHaveLength(3);
    });
    expect(searchUrls()[2]).toBe(`${SEARCH_PATH}?targetBudget=450000`);
    expect(statusHalves()).toEqual(['Loading', 'Page 2 of 2']);

    act(() => {
      page1.resolve(jsonResponse(searchResponse));
    });
    await waitForStatus('12 results · ranked for $450,000', 'Page 1 of 2');
    expect(cardRanks()[0]).toBe('No. 1');
    expect(pageButton(1)).toHaveAttribute('aria-current', 'page');
    expect(consoleError).not.toHaveBeenCalled();
  });

  // T8
  it('ignores a page response that arrives after the user went back', async () => {
    const { user } = await renderBudgetSettled();

    const page2 = replyLater();
    await user.click(pageButton(2));
    // Back to page 1, whose result the hook still holds: shown at once and refetched.
    const refetch = replyLater();
    window.history.back();

    await waitFor(() => {
      expect(searchUrls()).toHaveLength(3);
    });
    expect(currentUrl()).toBe('/?targetBudget=450000');
    expect(searchUrls()[2]).toBe(`${SEARCH_PATH}?targetBudget=450000`);
    expect(searchSignal(1)?.aborted).toBe(true);
    expect(searchSignal(2)?.aborted).toBe(false);
    expect(statusHalves()).toEqual(['12 results · ranked for $450,000', 'Page 1 of 2']);
    expect(cardRanks()[0]).toBe('No. 1');

    // Page 2 arrives late and is read, and is ignored: page 1 stays, before the refetch lands.
    const late = jsonResponse(searchResponsePage2);
    act(() => {
      page2.resolve(late);
    });
    await waitFor(() => {
      expect(late.bodyUsed).toBe(true);
    });
    await act(async () => {});
    expect(cardRanks()[0]).toBe('No. 1');
    expect(screen.queryByText('No. 11')).not.toBeInTheDocument();
    expect(statusHalves()).toEqual(['12 results · ranked for $450,000', 'Page 1 of 2']);

    // Then the refetch lands (page 1's first two real rows).
    const refetched: SearchResponse = { ...searchResponse, items: [listingA1, listingB7] };
    act(() => {
      refetch.resolve(jsonResponse(refetched));
    });
    await waitFor(() => {
      expect(cardAddresses()).toEqual([listingA1.address, listingB7.address]);
    });
    expect(cardRanks()).toEqual(['No. 1', 'No. 2']);
    expect(screen.queryByText('No. 11')).not.toBeInTheDocument();
    expect(statusHalves()).toEqual(['12 results · ranked for $450,000', 'Page 1 of 2']);
    expect(pageButton(1)).toHaveAttribute('aria-current', 'page');
    expect(consoleError).not.toHaveBeenCalled();
  });

  // T9
  it('runs the default search on mount and loads the City options', async () => {
    reply(newestFirstResponse);
    renderAt('/');

    await waitForStatus('12 results · newest first', 'Page 1 of 2');
    expect(searchUrls()).toEqual([SEARCH_PATH]);
    expect(fetchMock.mock.calls.map(([input]) => urlOf(input))).toContain('/api/cities');
    expect(cardAddresses()[0]).toBe('55 Elm Ct');
    expect(cardRanks()).toEqual(['No. 1', 'No. 2', 'No. 3']);
    await waitFor(() => {
      expect(
        within(field('City'))
          .getAllByRole('option')
          .map((option) => option.textContent),
      ).toEqual(['Any', ...cities]);
    });
    expect(consoleError).not.toHaveBeenCalled();
  });

  // T10
  it('shows a field error no field owns in the bar alert line and keeps the list', async () => {
    const { user } = await renderDefaultSettled();

    reply(malformedQueryErrorBody, { status: 400 });
    await user.selectOptions(field('City'), 'Vienna');
    await user.click(screen.getByRole('button', { name: 'Search' }));

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Query string is not valid URL encoding.',
    );
    for (const label of FIELD_LABELS) {
      expect(field(label)).not.toHaveAttribute('aria-invalid');
    }
    expect(statusHalves()).toEqual(['12 results · newest first', 'Page 1 of 2']);
    expect(cardAddresses()[0]).toBe('55 Elm Ct');
    expect(screen.queryByText(ERROR_HEADING)).not.toBeInTheDocument();
    expect(consoleError).not.toHaveBeenCalled();
  });

  // T11
  it('wires search and Clear to the URL: page reset, pageSize kept, keyword emphasis', async () => {
    // The real `page=2&pageSize=5` (ranks 6 to 10, SPEC section 6.4), trimmed to its first two.
    reply({
      ...newestFirstResponse,
      items: [
        { ...listingA1, rank: 6, relevance: { score: 0.3536, budgetFit: null, recency: 0.3536 } },
        { ...listingB7, rank: 7, relevance: { score: 0.3376, budgetFit: null, recency: 0.3376 } },
      ],
      page: { page: 2, pageSize: 5, totalItems: 12, totalPages: 3 },
    });
    const { user, container } = renderAt('/?page=2&pageSize=5');
    await waitForStatus('12 results · newest first', 'Page 2 of 3');
    expect(cardRanks()).toEqual(['No. 6', 'No. 7']);
    expect(searchUrls()).toEqual([`${SEARCH_PATH}?page=2&pageSize=5`]);
    expect(pageButton(2)).toHaveAttribute('aria-current', 'page');

    // The real `keyword=pets&pageSize=5` page 1, trimmed to its first three rows.
    reply({
      ...newestFirstResponse,
      items: [
        ...newestFirstResponse.items.slice(0, 2),
        { ...listingA3, rank: 3, relevance: { score: 0.3789, budgetFit: null, recency: 0.3789 } },
      ],
      page: { page: 1, pageSize: 5, totalItems: 7, totalPages: 2 },
    });
    await user.type(field('Keyword'), 'pets{Enter}');

    expect(currentUrl()).toBe('/?keyword=pets&pageSize=5');
    await waitForStatus('7 results · newest first', 'Page 1 of 2');
    expect(searchUrls()[1]).toBe(`${SEARCH_PATH}?keyword=pets&pageSize=5`);
    const a3 = screen
      .getByRole('heading', { level: 3, name: listingA3.address })
      .closest('article');
    if (a3 === null) {
      throw new Error('No card for A3');
    }
    expect(Array.from(a3.querySelectorAll('em'), (em) => em.textContent)).toEqual(['pets']);
    // Every card's matches, in its own casing.
    expect(Array.from(container.querySelectorAll('main em'), (em) => em.textContent)).toEqual([
      'pets',
      'Pets',
      'pets',
    ]);

    reply({
      ...newestFirstResponse,
      page: { page: 1, pageSize: 5, totalItems: 12, totalPages: 3 },
    });
    await user.click(screen.getByRole('button', { name: 'Clear' }));

    expect(currentUrl()).toBe('/?pageSize=5');
    await waitForStatus('12 results · newest first', 'Page 1 of 3');
    expect(searchUrls()[2]).toBe(`${SEARCH_PATH}?pageSize=5`);
    expect(field('Keyword')).toHaveValue('');
    expect(container.querySelectorAll('main em')).toHaveLength(0);
    expect(consoleError).not.toHaveBeenCalled();
  });

  // T12
  it('keeps the error view while Try again loads, and shows the list once it succeeds', async () => {
    const { user } = await renderDefaultSettled();
    reply(internalErrorBody, { status: 500 });
    await user.selectOptions(field('City'), 'Vienna');
    await user.click(screen.getByRole('button', { name: 'Search' }));
    await screen.findByRole('heading', { name: ERROR_HEADING });

    const retried = replyLater();
    await user.click(screen.getByRole('button', { name: 'Try again' }));

    // The list from before the error does not come back while the retry loads.
    expect(screen.getByRole('heading', { name: ERROR_HEADING })).toBeInTheDocument();
    expect(screen.queryAllByRole('article')).toHaveLength(0);
    expect(statusHalves()).toEqual(['Loading', '']);
    expect(screen.getByRole('progressbar', { name: 'Loading' })).toBeInTheDocument();
    expect(resultsRegion()).toHaveAttribute('aria-busy', 'true');
    expect(screen.queryByRole('navigation', { name: 'Pagination' })).not.toBeInTheDocument();

    act(() => {
      retried.resolve(jsonResponse(newestFirstResponse));
    });
    await waitForStatus('12 results · newest first', 'Page 1 of 2');
    expect(cardAddresses()[0]).toBe('55 Elm Ct');
    expect(screen.queryByText(ERROR_HEADING)).not.toBeInTheDocument();
    expect(screen.getByRole('navigation', { name: 'Pagination' })).toBeInTheDocument();
    expect(searchUrls()).toEqual([
      SEARCH_PATH,
      `${SEARCH_PATH}?city=Vienna`,
      `${SEARCH_PATH}?city=Vienna`,
    ]);
    expect(consoleError).toHaveBeenCalledTimes(1);
  });

  // T12, the 400 half: a 400 after an error keeps the error view and its empty status line.
  it('keeps the error view when a 400 follows an error', async () => {
    const { user } = await renderDefaultSettled();
    reply(internalErrorBody, { status: 500 });
    await user.selectOptions(field('City'), 'Vienna');
    await user.click(screen.getByRole('button', { name: 'Search' }));
    await screen.findByRole('heading', { name: ERROR_HEADING });

    reply(malformedQueryErrorBody, { status: 400 });
    await user.selectOptions(field('City'), 'Reston');
    await user.click(screen.getByRole('button', { name: 'Search' }));

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Query string is not valid URL encoding.',
    );
    expect(screen.getByRole('heading', { name: ERROR_HEADING })).toBeInTheDocument();
    expect(screen.queryAllByRole('article')).toHaveLength(0);
    expect(statusHalves()).toEqual(['', '']);
    expect(screen.queryByRole('navigation', { name: 'Pagination' })).not.toBeInTheDocument();
    expect(searchUrls()[2]).toBe(`${SEARCH_PATH}?city=Reston`);
  });

  // T13
  it('puts the results region in the Tab order after Clear, so the keyboard can scroll it', async () => {
    // SPEC section 7.6. The cards hold nothing focusable, so without the region's own Tab stop
    // a browser with no focusable scrollers (WebKit) goes from Clear straight to the pager.
    const { user } = await renderDefaultSettled();
    const clear = screen.getByRole('button', { name: 'Clear' });
    act(() => {
      clear.focus();
    });
    expect(clear).toHaveFocus();

    await user.tab();
    expect(resultsRegion()).toHaveFocus();

    // Previous is disabled on page 1, so the next stop is Page 1.
    await user.tab();
    expect(pageButton(1)).toHaveFocus();
    expect(consoleError).not.toHaveBeenCalled();
  });
});
