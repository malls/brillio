import styles from './Pager.module.css';

export interface PagerProps {
  /** The page on screen, 1-based. May exceed totalPages on a hand-edited URL. */
  page: number;
  totalPages: number;
  /**
   * Called with the requested page. The Pager never moves focus; its parent does. Property
   * syntax, not method syntax, so destructuring it passes typescript-eslint's unbound-method.
   */
  onChange: (page: number) => void;
}

// BRANDING.md section 5.7, SPEC sections 7.4 and 7.6. Every page number renders (no windowing).
// Disabled ends use the disabled attribute, so they leave the tab order. The current page stays
// focusable for screen-reader discovery but does nothing. On a page past the end no page is
// current and Previous goes to the last page, so recovery is one click. Hidden with no pages.
export function Pager({ page, totalPages, onChange }: PagerProps) {
  if (totalPages <= 0) {
    return null;
  }

  const pages = Array.from({ length: totalPages }, (_, i) => i + 1);

  return (
    <nav aria-label="Pagination" className={styles.pager}>
      <button
        type="button"
        className={`${styles.button} ${styles.previous}`}
        disabled={page <= 1}
        onClick={() => {
          onChange(Math.min(page - 1, totalPages));
        }}
      >
        <span aria-hidden="true">‹</span> Previous
      </button>
      <div className={styles.pages}>
        {pages.map((n) => (
          <button
            key={n}
            type="button"
            className={styles.button}
            aria-label={`Page ${n}`}
            aria-current={n === page ? 'page' : undefined}
            onClick={() => {
              if (n !== page) {
                onChange(n);
              }
            }}
          >
            {n}
          </button>
        ))}
      </div>
      <button
        type="button"
        className={`${styles.button} ${styles.next}`}
        disabled={page >= totalPages}
        onClick={() => {
          onChange(page + 1);
        }}
      >
        Next <span aria-hidden="true">›</span>
      </button>
    </nav>
  );
}
