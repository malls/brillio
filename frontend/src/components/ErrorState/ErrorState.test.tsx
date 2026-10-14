import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { ErrorState } from './ErrorState';

describe('ErrorState', () => {
  it('renders the error copy verbatim', () => {
    render(<ErrorState onRetry={vi.fn()} />);

    // ASCII apostrophe, as in BRANDING.md section 7.2.
    expect(
      screen.getByRole('heading', { level: 2, name: "Couldn't load results." }),
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Try again' })).toBeInTheDocument();
  });

  it('calls onRetry once on click', async () => {
    const user = userEvent.setup();
    const onRetry = vi.fn();
    render(<ErrorState onRetry={onRetry} />);

    await user.click(screen.getByRole('button', { name: 'Try again' }));

    expect(onRetry).toHaveBeenCalledTimes(1);
    expect(onRetry).toHaveBeenCalledWith();
  });

  it('calls onRetry from the keyboard', async () => {
    // Keyboard-only use works end to end (docs/SPEC.md section 7.6).
    const user = userEvent.setup();
    const onRetry = vi.fn();
    render(<ErrorState onRetry={onRetry} />);

    await user.tab();
    expect(screen.getByRole('button', { name: 'Try again' })).toHaveFocus();
    await user.keyboard('{Enter}');
    expect(onRetry).toHaveBeenCalledTimes(1);
    await user.keyboard(' ');
    expect(onRetry).toHaveBeenCalledTimes(2);
  });

  it('renders a type=button', () => {
    render(<ErrorState onRetry={vi.fn()} />);

    expect(screen.getByRole('button', { name: 'Try again' })).toHaveAttribute('type', 'button');
  });

  it('shows no error code', () => {
    const { container } = render(<ErrorState onRetry={vi.fn()} />);

    expect(container.textContent).toBe("Couldn't load results.Try again");
  });
});
