import { render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Header } from './Header';

describe('Header', () => {
  it('renders the wordmark as the level 1 heading inside the banner', () => {
    render(<Header />);

    const banner = screen.getByRole('banner');
    expect(
      within(banner).getByRole('heading', { level: 1, name: 'Brillio Listings' }),
    ).toBeInTheDocument();
  });

  it('renders no scope label when scope is absent', () => {
    render(<Header />);

    const banner = screen.getByRole('banner');
    expect(banner.textContent).toBe('Brillio Listings');
    expect(within(banner).queryByText('Northern Virginia')).not.toBeInTheDocument();
  });

  it('renders the scope label when scope is given', () => {
    render(<Header scope="Northern Virginia" />);

    const banner = screen.getByRole('banner');
    expect(within(banner).getByText('Northern Virginia')).toBeInTheDocument();
  });
});
