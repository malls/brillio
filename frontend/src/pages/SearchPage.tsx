import { useState } from 'react';
import type { ReactNode } from 'react';
import type { SearchResponse } from '../api/types';
import { ErrorState } from '../components/ErrorState/ErrorState';
import { FilterBar } from '../components/FilterBar/FilterBar';
import { Header } from '../components/Header/Header';
import { ListingList, RESULTS_HEADING_ID } from '../components/ListingList/ListingList';
import { LoadingLine } from '../components/LoadingLine/LoadingLine';
import { NoResults } from '../components/NoResults/NoResults';
import { Pager } from '../components/Pager/Pager';
import { StatusLine } from '../components/StatusLine/StatusLine';
import { splitKeyword } from '../format/keyword';
import { useCities } from '../hooks/useCities';
import type { SearchState } from '../hooks/useListingSearch';
import { useListingSearch } from '../hooks/useListingSearch';
import { useSearchParamsState } from '../hooks/useSearchParamsState';
import shell from '../styles/shell.module.css';

// The response whose list is on screen: the previous one while loading or after a 400 (the list
// stays, BRANDING.md section 5.9), the new one once a 2xx settles, none on error.
function shownResponse(state: SearchState): SearchResponse | undefined {
  switch (state.status) {
    case 'loading':
    case 'invalid':
      return state.previous;
    case 'success':
    case 'empty':
      return state.data;
    case 'idle':
    case 'error':
      return undefined;
  }
}

// The state with no list on screen, for the status line while the error view is held.
function withoutPrevious(state: SearchState): SearchState {
  switch (state.status) {
    case 'loading':
      return { status: 'loading' };
    case 'invalid':
      return { status: 'invalid', error: state.error };
    case 'error':
      return { status: 'error', error: state.error };
    default:
      return state;
  }
}

// The search page (docs/SPEC.md sections 7.1-7.6): the pinned app shell of BRANDING.md section
// 4.2, wired to the URL state and the fetch lifecycle. The URL is the search: the filter bar and
// the pager write it, and every change fetches.
export function SearchPage() {
  const { params, setSearch, setPage, clear } = useSearchParamsState();
  const { state, retry } = useListingSearch(params);
  const cities = useCities();
  const loading = state.status === 'loading';

  // Once a search fails, the error view stays until a 2xx settles: a Try again (loading) or a
  // 400 (invalid) must not bring back the list from before the error (SPEC section 7.5). Stored
  // from render to render with React's "adjusting state when a prop changes" pattern.
  const [errorHeld, setErrorHeld] = useState(false);
  let holdError = errorHeld;
  if (state.status === 'error') {
    holdError = true;
  } else if (state.status === 'success' || state.status === 'empty') {
    holdError = false;
  }
  if (holdError !== errorHeld) {
    setErrorHeld(holdError);
  }

  const shown = holdError ? undefined : shownResponse(state);
  // Emphasis follows the URL's keyword (SPEC section 7.4).
  const keywordTokens = splitKeyword(params.keyword);

  // After a page change focus moves to the results heading (SPEC section 7.4). The list stays
  // mounted while the next page loads, so the heading keeps focus, and focusing it scrolls the
  // results back to the top.
  const changePage = (page: number) => {
    setPage(page);
    document.getElementById(RESULTS_HEADING_ID)?.focus();
  };

  // With no response yet (first load) the results region is empty: the status line alone
  // (BRANDING.md section 5.9).
  let results: ReactNode = null;
  if (holdError) {
    results = <ErrorState onRetry={retry} />;
  } else if (shown?.page.totalItems === 0) {
    results = <NoResults onClear={clear} />;
  } else if (shown !== undefined) {
    results = <ListingList items={shown.items} keywordTokens={keywordTokens} />;
  }

  return (
    <div className={shell.shell}>
      <Header scope="Northern Virginia" />
      <div>
        <FilterBar
          value={params}
          cities={cities}
          loading={loading}
          serverErrors={state.status === 'invalid' ? state.error.fields : undefined}
          onSearch={setSearch}
          onClear={clear}
        />
        <LoadingLine active={loading} />
      </div>
      <StatusLine state={holdError ? withoutPrevious(state) : state} />
      <main className={shell.results} aria-label="Results" aria-live="polite" aria-busy={loading}>
        {results}
      </main>
      {shown !== undefined && (
        <Pager page={params.page ?? 1} totalPages={shown.page.totalPages} onChange={changePage} />
      )}
    </div>
  );
}
