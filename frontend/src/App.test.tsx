import { render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cities, jsonResponse, newestFirstResponse } from './test/fixtures';
import { App } from './App';

describe('App', () => {
  beforeEach(() => {
    window.history.replaceState(null, '', '/');
    vi.stubGlobal(
      'fetch',
      vi.fn<typeof fetch>((input) =>
        Promise.resolve(jsonResponse(input === '/api/cities' ? { cities } : newestFirstResponse)),
      ),
    );
  });

  it('renders the search page and its default search', async () => {
    render(<App />);

    expect(screen.getByRole('heading', { level: 1, name: 'Brillio Listings' })).toBeInTheDocument();
    expect(screen.getByRole('search', { name: 'Filters' })).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByRole('status')).toHaveTextContent('12 results · newest first');
    });
    expect(screen.getByRole('main', { name: 'Results' })).toBeInTheDocument();
    expect(screen.getByRole('navigation', { name: 'Pagination' })).toBeInTheDocument();
  });
});
