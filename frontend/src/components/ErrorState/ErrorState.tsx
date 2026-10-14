import styles from './ErrorState.module.css';

export interface ErrorStateProps {
  /** Re-runs the last search (useListingSearch's retry, docs/SPEC.md section 7.2). */
  onRetry: () => void;
}

// BRANDING.md section 5.9 Error, copy verbatim from section 7.2: the state heading and the
// `Try again` text button. No error code or message: the error goes to the console, not the UI.
export function ErrorState({ onRetry }: ErrorStateProps) {
  return (
    <div className={styles.root}>
      <h2 className={styles.heading}>Couldn't load results.</h2>
      <button type="button" className={styles.action} onClick={() => onRetry()}>
        Try again
      </button>
    </div>
  );
}
