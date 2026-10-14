import styles from './NoResults.module.css';

export interface NoResultsProps {
  /** Clears every filter (docs/SPEC.md section 7.1 Clear). */
  onClear: () => void;
}

// BRANDING.md section 5.9 No results, copy verbatim from section 7.2: the state heading, the body
// line and the `Clear filters` text button.
export function NoResults({ onClear }: NoResultsProps) {
  return (
    <div className={styles.root}>
      <h2 className={styles.heading}>No results.</h2>
      <p className={styles.body}>Adjust your filters.</p>
      <button type="button" className={styles.action} onClick={() => onClear()}>
        Clear filters
      </button>
    </div>
  );
}
