import { Fragment } from 'react';
import type { ListingResult } from '../../api/types';
import { emphasisSegments } from '../../format/keyword';
import { META_SEPARATOR, formatListingMeta, formatStatusLabel } from '../../format/listing';
import { formatPrice } from '../../format/money';
import { Monogram } from '../Monogram/Monogram';
import { ScorePill } from '../ScorePill/ScorePill';
import styles from './ListingCard.module.css';

export interface ListingCardProps {
  item: ListingResult;
  /** The URL keyword split into tokens (splitKeyword in src/format/keyword.ts); [] for none. */
  keywordTokens: string[];
}

// BRANDING.md section 5.4: monogram, text column, price block. A pure renderer over one
// ListingResult: no link, no button, no hover state (cards are inert in v1, section 10).
//
// The DOM order (monogram, address, meta, price block, description) is the reading order at
// every width. The card is one flat grid, so at 760px and below (section 4.4) the price block
// moves under the meta line by grid areas alone, without duplicating markup.
export function ListingCard({ item, keywordTokens }: ListingCardProps) {
  const statusLabel = formatStatusLabel(item.status);
  return (
    <article className={styles.card}>
      <div className={styles.monogramSlot}>
        <Monogram city={item.city} />
      </div>
      <h3 className={styles.address}>{item.address}</h3>
      <p className={styles.meta}>
        {statusLabel !== null && (
          <>
            <span className={styles.statusLabel}>{statusLabel}</span>
            {META_SEPARATOR}
          </>
        )}
        {formatListingMeta(item).join(META_SEPARATOR)}
      </p>
      <div className={styles.priceBlock}>
        <p className={styles.price}>{formatPrice(item.price)}</p>
        <ScorePill rank={item.rank} score={item.relevance.score} />
      </div>
      <p className={styles.description}>
        {emphasisSegments(item.description, keywordTokens).map((segment, index) => (
          <Fragment key={index}>
            {segment.match ? <em className={styles.match}>{segment.text}</em> : segment.text}
          </Fragment>
        ))}
      </p>
    </article>
  );
}
