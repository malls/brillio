import { formatRank, formatScore } from '../../format/listing';
import styles from './ScorePill.module.css';

export interface ScorePillProps {
  /** Position in the whole ranked result set, so page 2 begins at 11 (BRANDING.md section 5.6). */
  rank: number;
  /** The API's relevance score in [0, 1]; rounded to two decimals for display only. */
  score: number;
}

// BRANDING.md section 5.6: the `No. 1` pill followed by the score to two decimals.
export function ScorePill({ rank, score }: ScorePillProps) {
  return (
    <span className={styles.score}>
      <span className={styles.rank}>{formatRank(rank)}</span>
      <span>{formatScore(score)}</span>
    </span>
  );
}
