import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { Pager } from './Pager';

function renderPager(page: number, totalPages: number) {
  const onChange = vi.fn<(page: number) => void>();
  const view = render(<Pager page={page} totalPages={totalPages} onChange={onChange} />);
  return { ...view, onChange };
}

const previous = () => screen.getByRole('button', { name: 'Previous' });
const next = () => screen.getByRole('button', { name: 'Next' });
const pageButton = (n: number) => screen.getByRole('button', { name: `Page ${n}` });

describe('Pager', () => {
  it('renders Previous, every page number and Next, in order, with text glyphs', () => {
    renderPager(1, 3);

    const nav = screen.getByRole('navigation', { name: 'Pagination' });
    const buttons = within(nav).getAllByRole('button');
    expect(buttons).toEqual([previous(), pageButton(1), pageButton(2), pageButton(3), next()]);
    expect(buttons.map((b) => b.textContent)).toEqual(['‹ Previous', '1', '2', '3', 'Next ›']);
    expect(buttons.every((b) => b.getAttribute('type') === 'button')).toBe(true);
    expect(nav.querySelector('svg, img')).toBeNull();
  });

  it('marks only the current page with aria-current', () => {
    renderPager(2, 3);

    expect(pageButton(2)).toHaveAttribute('aria-current', 'page');
    expect(pageButton(1)).not.toHaveAttribute('aria-current');
    expect(pageButton(3)).not.toHaveAttribute('aria-current');
    expect(previous()).not.toHaveAttribute('aria-current');
    expect(next()).not.toHaveAttribute('aria-current');
  });

  it('disables Previous on the first page and Next on the last', () => {
    const first = renderPager(1, 3);
    expect(previous()).toBeDisabled();
    expect(next()).toBeEnabled();
    first.unmount();

    const last = renderPager(3, 3);
    expect(previous()).toBeEnabled();
    expect(next()).toBeDisabled();
    last.unmount();

    renderPager(1, 1);
    expect(previous()).toBeDisabled();
    expect(next()).toBeDisabled();
  });

  it('keeps a disabled Previous out of the tab order', async () => {
    const user = userEvent.setup();
    renderPager(1, 2);

    await user.tab();
    expect(pageButton(1)).toHaveFocus();
    await user.tab();
    expect(pageButton(2)).toHaveFocus();
    await user.tab();
    expect(next()).toHaveFocus();
  });

  it('keeps a disabled Next out of the tab order', async () => {
    const user = userEvent.setup();
    renderPager(2, 2);

    await user.tab();
    expect(previous()).toHaveFocus();
    await user.tab();
    expect(pageButton(1)).toHaveFocus();
    await user.tab();
    expect(pageButton(2)).toHaveFocus();
    await user.tab();
    expect(document.body).toHaveFocus();
  });

  it('asks for the next page', async () => {
    const user = userEvent.setup();
    const { onChange } = renderPager(2, 3);

    await user.click(next());
    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith(3);
  });

  it('asks for the previous page', async () => {
    const user = userEvent.setup();
    const { onChange } = renderPager(2, 3);

    await user.click(previous());
    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith(1);
  });

  it('asks for a page by its number', async () => {
    const user = userEvent.setup();
    const { onChange } = renderPager(2, 3);

    await user.click(pageButton(3));
    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith(3);
  });

  it('works from the keyboard alone', async () => {
    const user = userEvent.setup();
    const { onChange } = renderPager(2, 3);

    next().focus();
    await user.keyboard('{Enter}');
    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith(3);

    pageButton(1).focus();
    await user.keyboard(' ');
    expect(onChange).toHaveBeenCalledTimes(2);
    expect(onChange).toHaveBeenLastCalledWith(1);
  });

  it('does nothing for the current page or a disabled end', async () => {
    const user = userEvent.setup();
    const { onChange } = renderPager(1, 2);

    await user.click(pageButton(1));
    await user.click(previous());
    expect(onChange).not.toHaveBeenCalled();
  });

  it('offers the last page from a page past the end', async () => {
    const user = userEvent.setup();
    const { onChange } = renderPager(3, 2);

    expect(screen.getAllByRole('button').filter((b) => b.hasAttribute('aria-current'))).toEqual([]);
    expect(next()).toBeDisabled();
    expect(previous()).toBeEnabled();

    await user.click(previous());
    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith(2);
  });

  it('renders nothing when there are no pages', () => {
    const { container } = renderPager(1, 0);

    expect(container).toBeEmptyDOMElement();
    expect(screen.queryByRole('navigation')).not.toBeInTheDocument();
  });
});
