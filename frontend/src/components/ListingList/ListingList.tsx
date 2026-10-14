import type { ListingResult } from '../../api/types';
import { ListingCard } from '../ListingCard/ListingCard';
import styles from './ListingList.module.css';

// The id of the results heading. After a page change the page moves focus here (docs/SPEC.md
// section 7.4): document.getElementById(RESULTS_HEADING_ID)?.focus().
export const RESULTS_HEADING_ID = 'results-heading';

export interface ListingListProps {
  /** One page of results, in rank order. */
  items: ListingResult[];
  /** Passed to every card for keyword emphasis; [] for none. */
  keywordTokens: string[];
}

// BRANDING.md section 5.4: the cards in rank order, a hairline above every card but the first.
// The live region is SearchPage's <main> (SPEC section 7.6); this section is labelled by its
// heading only. Script can focus the heading, which is visually hidden: BRANDING and the demo
// show no heading above the cards.
export function ListingList({ items, keywordTokens }: ListingListProps) {
  return (
    <section className={styles.region} aria-labelledby={RESULTS_HEADING_ID}>
      <h2 id={RESULTS_HEADING_ID} className={styles.heading} tabIndex={-1}>
        Results
      </h2>
      {items.length > 0 && (
        <ol className={styles.list}>
          {items.map((item) => (
            // Ids are unique only per source (SPEC section 4), so the key carries both.
            <li key={`${item.source}:${item.id}`} className={styles.item}>
              <ListingCard item={item} keywordTokens={keywordTokens} />
            </li>
          ))}
        </ol>
      )}
    </section>
  );
}
