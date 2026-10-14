import styles from './LoadingLine.module.css';

export interface LoadingLineProps {
  active: boolean;
}

// BRANDING.md sections 5.9 and 6, SPEC section 7.6: the 2px loading line under the filter bar.
// Keep it mounted: inactive, it keeps the same 2px box (so nothing shifts) but has no role, no
// name and no segment, so assistive technology never meets an idle progressbar. Active, it is an
// indeterminate progressbar (no aria-valuenow) named with the section 7.2 string "Loading".
export function LoadingLine({ active }: LoadingLineProps) {
  if (!active) {
    return <div className={styles.track} aria-hidden="true" />;
  }
  return (
    <div className={`${styles.track} ${styles.active}`} role="progressbar" aria-label="Loading">
      <div className={styles.segment} />
    </div>
  );
}
