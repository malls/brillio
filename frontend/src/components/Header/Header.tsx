import styles from './Header.module.css';

export interface HeaderProps {
  /** Optional scope label at the right, e.g. "Northern Virginia" (BRANDING.md section 5.1). */
  scope?: string;
}

// The wordmark is the page's h1. Its DOM text is "Brillio Listings" and CSS uppercases it,
// so screen readers read words, not letters.
export function Header({ scope }: HeaderProps) {
  return (
    <header className={styles.header}>
      <h1 className={styles.wordmark}>Brillio Listings</h1>
      {scope !== undefined && <span className={styles.scope}>{scope}</span>}
    </header>
  );
}
