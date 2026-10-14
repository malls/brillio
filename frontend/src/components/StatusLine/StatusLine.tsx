import type { SearchResponse } from '../../api/types';
import { formatPageOf, formatResultSummary } from '../../format/listing';
import type { SearchState } from '../../hooks/useListingSearch';
import styles from './StatusLine.module.css';

export interface StatusLineProps {
  /** The fetch lifecycle state (docs/SPEC.md section 7.2). */
  state: SearchState;
}

// BRANDING.md section 7.2, verbatim.
const LOADING = 'Loading';

// Left: what the list on screen is. Right: where that list sits in the pages.
function summarize(data: SearchResponse): [string, string] {
  return [formatResultSummary(data), formatPageOf(data.page.page, data.page.totalPages)];
}

// BRANDING.md sections 5.3 and 5.9, SPEC section 7.5. The left half says what the list on screen
// is (or Loading); the right half says where that list sits in the pages.
function describeState(state: SearchState): [string, string] {
  switch (state.status) {
    case 'idle':
      return ['', ''];
    case 'loading':
      // The previous list stays on screen while loading; a first load shows the status alone.
      return [
        LOADING,
        state.previous === undefined
          ? ''
          : formatPageOf(state.previous.page.page, state.previous.page.totalPages),
      ];
    case 'success':
    case 'empty':
      return summarize(state.data);
    case 'invalid':
      // The previous list stays and the status line keeps describing it; with no list on
      // screen and nothing loading, the line is as empty as idle.
      return state.previous === undefined ? ['', ''] : summarize(state.previous);
    case 'error':
      // The error state replaces the list, so no page is on screen.
      return ['', ''];
  }
}

// The status line row (BRANDING.md section 5.3): Meta 13 in ink-2, left and right halves. Both
// halves always render so the structure is stable and the right half stays right.
export function StatusLine({ state }: StatusLineProps) {
  const [left, right] = describeState(state);
  return (
    <div role="status" className={styles.status}>
      <span className={styles.half}>{left}</span>
      <span className={styles.half}>{right}</span>
    </div>
  );
}
