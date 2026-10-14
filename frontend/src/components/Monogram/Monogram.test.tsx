import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { cities, listingA1, listingA3 } from '../../test/fixtures';
import { Monogram } from './Monogram';

describe('Monogram', () => {
  it.each([
    [listingA1.city, 'Sp'],
    ['Vienna', 'Vi'],
    [listingA3.city, 'Fa'],
  ])('renders the city initials: %s as %s', (city, initials) => {
    // Every input is a real city from GET /api/cities.
    expect(cities).toContain(city);

    const { container } = render(<Monogram city={city} />);

    expect(container.textContent).toBe(initials);
  });

  it('hides the decorative block from assistive technology', () => {
    render(<Monogram city={listingA1.city} />);

    expect(screen.getByText('Sp')).toHaveAttribute('aria-hidden', 'true');
    expect(screen.queryByRole('img')).not.toBeInTheDocument();
  });
});
