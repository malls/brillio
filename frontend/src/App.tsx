import { Header } from './components/Header/Header';
import shell from './styles/shell.module.css';

// BC-7 placeholder for the pinned app shell (BRANDING.md section 4.2). The empty regions carry
// accessible names only; BC-12 replaces this body with <SearchPage />.
export function App() {
  return (
    <div className={shell.shell}>
      <Header scope="Northern Virginia" />
      <div className={shell.row} role="search" aria-label="Filters" />
      <div className={shell.row} role="status" />
      <main className={shell.results} aria-label="Results" aria-live="polite" />
      <nav className={shell.row} aria-label="Pagination" />
    </div>
  );
}
