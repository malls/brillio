import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { NoResults } from './NoResults';

describe('NoResults', () => {
  it('renders the no-results copy verbatim', () => {
    const { container } = render(<NoResults onClear={vi.fn()} />);

    expect(screen.getByRole('heading', { level: 2, name: 'No results.' })).toBeInTheDocument();
    expect(screen.getByText('Adjust your filters.')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Clear filters' })).toBeInTheDocument();
    expect(container.textContent).toBe('No results.Adjust your filters.Clear filters');
  });

  it('calls onClear once on click', async () => {
    const user = userEvent.setup();
    const onClear = vi.fn();
    render(<NoResults onClear={onClear} />);

    await user.click(screen.getByRole('button', { name: 'Clear filters' }));

    expect(onClear).toHaveBeenCalledTimes(1);
    expect(onClear).toHaveBeenCalledWith();
  });

  it('calls onClear from the keyboard', async () => {
    // Keyboard-only use works end to end (docs/SPEC.md section 7.6).
    const user = userEvent.setup();
    const onClear = vi.fn();
    render(<NoResults onClear={onClear} />);

    await user.tab();
    expect(screen.getByRole('button', { name: 'Clear filters' })).toHaveFocus();
    await user.keyboard('{Enter}');
    expect(onClear).toHaveBeenCalledTimes(1);
    await user.keyboard(' ');
    expect(onClear).toHaveBeenCalledTimes(2);
  });

  it('renders a type=button', () => {
    render(<NoResults onClear={vi.fn()} />);

    expect(screen.getByRole('button', { name: 'Clear filters' })).toHaveAttribute('type', 'button');
  });
});
