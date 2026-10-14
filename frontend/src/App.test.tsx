import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { App } from './App';

describe('App', () => {
  it('renders the five shell rows in BRANDING order with their roles and names', () => {
    const { container } = render(<App />);
    const shell = container.firstElementChild;
    if (shell === null) {
      throw new Error('App rendered no shell element');
    }

    const results = screen.getByRole('main', { name: 'Results' });
    expect(Array.from(shell.children)).toEqual([
      screen.getByRole('banner'),
      screen.getByRole('search', { name: 'Filters' }),
      screen.getByRole('status'),
      results,
      screen.getByRole('navigation', { name: 'Pagination' }),
    ]);
    expect(results).toHaveAttribute('aria-live', 'polite');
  });
});
