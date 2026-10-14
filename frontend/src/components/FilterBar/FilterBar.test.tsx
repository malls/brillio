import { fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { FieldError, SearchParams } from '../../api/types';
import { cities, validationErrorBody } from '../../test/fixtures';
import type { FilterBarProps } from './FilterBar';
import { FilterBar } from './FilterBar';

const LABELS = [
  'Min price',
  'Max price',
  'Min bedrooms',
  'City',
  'Keyword',
  'Target budget',
  'Status',
] as const;
const MONEY_LABELS = ['Min price', 'Max price', 'Target budget'] as const;

const MIN_GT_MAX = 'Min price must be less than max price.';

function setup(props: Partial<FilterBarProps> = {}) {
  const onSearch = vi.fn<(next: SearchParams) => void>();
  const onClear = vi.fn<() => void>();
  const base: FilterBarProps = { value: {}, cities, loading: false, onSearch, onClear, ...props };
  const user = userEvent.setup();
  const view = render(<FilterBar {...base} />);
  const rerender = (next: Partial<FilterBarProps>) => {
    view.rerender(<FilterBar {...base} {...next} />);
  };
  return { user, onSearch, onClear, rerender, container: view.container };
}

function field(label: string): HTMLInputElement | HTMLSelectElement {
  const element = screen.getByLabelText(label);
  if (!(element instanceof HTMLInputElement || element instanceof HTMLSelectElement)) {
    throw new Error(`${label} is not a form control`);
  }
  return element;
}

function searchButton(): HTMLElement {
  return screen.getByRole('button', { name: 'Search' });
}

function optionTexts(select: HTMLElement): string[] {
  return within(select)
    .getAllByRole('option')
    .map((option) => option.textContent);
}

// The element a control's aria-describedby points at.
function description(control: HTMLElement): HTMLElement | null {
  const id = control.getAttribute('aria-describedby');
  return id === null ? null : document.getElementById(id);
}

function expectMarked(label: string, message: string) {
  const control = field(label);
  expect(control).toHaveAttribute('aria-invalid', 'true');
  expect(description(control)).toHaveTextContent(message);
}

function expectUnmarked(label: string) {
  const control = field(label);
  expect(control).not.toHaveAttribute('aria-invalid');
  expect(control).not.toHaveAttribute('aria-describedby');
}

const filledValue: SearchParams = {
  minPrice: 400000,
  maxPrice: 550000,
  minBedrooms: 2,
  city: 'Fairfax',
  keyword: 'pets',
  targetBudget: 450000,
  status: ['pending'],
};

describe('FilterBar copy and structure', () => {
  it('renders seven labelled fields in order with the verbatim labels', () => {
    setup();

    const form = screen.getByRole('search', { name: 'Filters' });
    const labels = Array.from(form.querySelectorAll('label')).map((label) => label.textContent);
    expect(labels).toEqual(LABELS);
    const controls = Array.from(form.querySelectorAll('input, select'));
    expect(controls).toEqual(LABELS.map((label) => field(label)));
  });

  it('uses the BRANDING placeholders on inputs and Any first in both selects', () => {
    setup();

    for (const label of ['Min price', 'Max price', 'Min bedrooms', 'Target budget']) {
      expect(field(label)).toHaveAttribute('placeholder', 'Any');
    }
    expect(field('Keyword')).toHaveAttribute('placeholder', 'Search descriptions');
    for (const label of ['City', 'Status']) {
      const select = field(label);
      expect(select.tagName).toBe('SELECT');
      expect(select).toHaveValue('');
      expect(within(select).getAllByRole('option')[0]).toHaveTextContent('Any');
    }
  });

  it('lists Any then the cities prop in City', () => {
    setup();

    expect(optionTexts(field('City'))).toEqual(['Any', ...cities]);
  });

  it.each([
    ['the list', cities],
    ['no list yet', []],
  ])('shows a URL city missing from %s as an extra selected option', (_name, list) => {
    setup({ value: { city: 'Arlington' }, cities: list });

    expect(optionTexts(field('City'))).toEqual(['Any', ...list, 'Arlington']);
    expect(field('City')).toHaveValue('Arlington');
  });

  it('offers Any, Active, Pending and Sold in Status', () => {
    setup();

    const status = field('Status');
    expect(optionTexts(status)).toEqual(['Any', 'Active', 'Pending', 'Sold']);
    expect(
      within(status)
        .getAllByRole<HTMLOptionElement>('option')
        .map((option) => option.value),
    ).toEqual(['', 'active', 'pending', 'sold']);
  });

  it('is a search landmark named Filters with Search (submit) and Clear (button) and no emoji', () => {
    const { container } = setup({ value: filledValue });

    expect(screen.getByRole('search', { name: 'Filters' }).tagName).toBe('FORM');
    expect(searchButton()).toHaveAttribute('type', 'submit');
    expect(screen.getByRole('button', { name: 'Clear' })).toHaveAttribute('type', 'button');
    expect(screen.getAllByRole('button')).toHaveLength(2);
    expect(container.textContent).not.toMatch(/\p{Extended_Pictographic}/u);
  });
});

describe('FilterBar seeding', () => {
  it('seeds every field from value, money formatted, single status selected', () => {
    setup({ value: filledValue });

    expect(field('Min price')).toHaveValue('$400,000');
    expect(field('Max price')).toHaveValue('$550,000');
    expect(field('Min bedrooms')).toHaveValue('2');
    expect(field('City')).toHaveValue('Fairfax');
    expect(field('Keyword')).toHaveValue('pets');
    expect(field('Target budget')).toHaveValue('$450,000');
    expect(field('Status')).toHaveValue('pending');
    expect(searchButton()).toBeEnabled();
  });

  it('re-seeds on a different value and keeps typing on an equal object or a page change', async () => {
    const { user, rerender } = setup({ value: filledValue });

    await user.clear(field('Keyword'));
    await user.type(field('Keyword'), 'yard');
    rerender({ value: { ...filledValue } });
    rerender({ value: { ...filledValue, page: 3 } });
    rerender({ value: { ...filledValue, pageSize: 20, status: ['pending', 'pending'] } });
    expect(field('Keyword')).toHaveValue('yard');

    rerender({ value: { ...filledValue, city: 'Vienna' } });
    expect(field('Keyword')).toHaveValue('pets');
    expect(field('City')).toHaveValue('Vienna');

    rerender({ value: {} });
    for (const label of LABELS) {
      expect(field(label)).toHaveValue('');
    }
  });

  it('shows the filters, not old typing, on returning to earlier filters (A, B, A)', async () => {
    const { user, rerender } = setup({ value: filledValue });

    await user.type(field('Keyword'), ' yard');
    await user.click(searchButton());
    rerender({ value: { ...filledValue, keyword: 'pets yard' } });
    rerender({ value: filledValue });
    expect(field('Keyword')).toHaveValue('pets');
  });

  it('shows Any for several statuses and omits status on submit', async () => {
    const { user, onSearch } = setup({ value: { status: ['active', 'pending'] } });

    expect(field('Status')).toHaveValue('');
    await user.click(searchButton());
    expect(onSearch).toHaveBeenCalledTimes(1);
    expect(onSearch.mock.calls[0]?.[0]).toStrictEqual({});
  });

  it('marks invalid values seeded from value at once', () => {
    setup({ value: { minPrice: 550000, maxPrice: 400000, minBedrooms: -1, targetBudget: 0 } });

    expectMarked('Max price', MIN_GT_MAX);
    expect(field('Min price')).toHaveAttribute('aria-invalid', 'true');
    expectMarked('Min bedrooms', 'Min bedrooms must be 0 or more.');
    expectMarked('Target budget', 'Target budget must be more than $0.');
    expect(searchButton()).toBeDisabled();
  });
});

describe('FilterBar money and number display', () => {
  it.each(MONEY_LABELS)('%s shows digits while focused and $400,000 on blur', async (label) => {
    const { user } = setup();

    await user.click(field(label));
    await user.keyboard('400000');
    expect(field(label)).toHaveValue('400000');
    await user.tab();
    expect(field(label)).toHaveValue('$400,000');
    await user.click(field(label));
    expect(field(label)).toHaveValue('400000');
  });

  it('accepts a typed $400,000 and submits 400000', async () => {
    const { user, onSearch } = setup();

    await user.type(field('Min price'), '$400,000');
    await user.tab();
    expect(field('Min price')).toHaveValue('$400,000');
    await user.click(field('Min price'));
    expect(field('Min price')).toHaveValue('400000');
    await user.click(searchButton());
    expect(onSearch.mock.calls[0]?.[0]).toStrictEqual({ minPrice: 400000 });
  });

  it('never formats Min bedrooms as money', async () => {
    const { user } = setup();

    await user.type(field('Min bedrooms'), '2.0');
    await user.tab();
    expect(field('Min bedrooms')).toHaveValue('2');
  });

  it('leaves a fraction or an invalid money value as typed on blur', async () => {
    const { user } = setup();

    await user.type(field('Max price'), '400000.5');
    await user.type(field('Target budget'), 'abc');
    await user.tab();
    expect(field('Max price')).toHaveValue('400000.5');
    expect(field('Target budget')).toHaveValue('abc');
  });
});

describe('FilterBar client rules', () => {
  it('marks both prices for min above max, with one message inside Max price', async () => {
    const { user, container } = setup();

    await user.type(field('Min price'), '550000');
    await user.type(field('Max price'), '400000');
    await user.tab();

    const messages = within(container).getAllByText(MIN_GT_MAX);
    expect(messages).toHaveLength(1);
    const [message] = messages;
    expect(field('Max price').parentElement).toContainElement(message ?? null);
    expect(field('Min price').parentElement).not.toContainElement(message ?? null);
    expect(field('Min price')).toHaveAttribute('aria-invalid', 'true');
    expect(field('Max price')).toHaveAttribute('aria-invalid', 'true');
    expect(description(field('Min price'))).toBe(message);
    expect(description(field('Max price'))).toBe(message);
    expect(searchButton()).toBeDisabled();
  });

  it('accepts equal prices', async () => {
    const { user, onSearch } = setup();

    await user.type(field('Min price'), '400000');
    await user.type(field('Max price'), '400000');
    await user.tab();
    expectUnmarked('Min price');
    expectUnmarked('Max price');
    await user.click(searchButton());
    expect(onSearch.mock.calls[0]?.[0]).toStrictEqual({ minPrice: 400000, maxPrice: 400000 });
  });

  it.each([
    ['Min price', 'Min price must be 0 or more.'],
    ['Max price', 'Max price must be 0 or more.'],
  ])('a negative %s reads "%s" under that field', async (label, message) => {
    const { user } = setup();

    await user.type(field(label), '-5');
    await user.tab();
    expectMarked(label, message);
    expect(field(label).parentElement).toHaveTextContent(message);
  });

  it.each([
    ['Min price', 'Min price must be a number.'],
    ['Max price', 'Max price must be a number.'],
    ['Min bedrooms', 'Min bedrooms must be a number.'],
    ['Target budget', 'Target budget must be a number.'],
  ])('a non-numeric %s reads "%s"', async (label, message) => {
    const { user } = setup();

    await user.type(field(label), 'abc');
    await user.tab();
    expectMarked(label, message);
  });

  it.each([
    ['-1', 'Min bedrooms must be 0 or more.'],
    ['2.5', 'Min bedrooms must be a whole number.'],
  ])('Min bedrooms %s reads "%s"', async (text, message) => {
    const { user } = setup();

    await user.type(field('Min bedrooms'), text);
    await user.tab();
    expectMarked('Min bedrooms', message);
  });

  it.each(['0', '-5'])(
    'Target budget %s reads "Target budget must be more than $0."',
    async (text) => {
      const { user } = setup();

      await user.type(field('Target budget'), text);
      await user.tab();
      expectMarked('Target budget', 'Target budget must be more than $0.');
    },
  );

  it('shows the message on blur, not while typing, yet disables Search at once', async () => {
    const { user } = setup();

    await user.type(field('Min price'), '-5');
    expect(screen.queryByText('Min price must be 0 or more.')).not.toBeInTheDocument();
    expectUnmarked('Min price');
    expect(searchButton()).toBeDisabled();

    await user.tab();
    expectMarked('Min price', 'Min price must be 0 or more.');
  });

  it('reveals every message on Enter in an invalid bar and sends nothing', async () => {
    const { user, onSearch } = setup();

    await user.type(field('Min price'), 'abc');
    await user.type(field('Target budget'), '0');
    await user.click(field('Keyword'));
    await user.keyboard('{Enter}');

    expectMarked('Min price', 'Min price must be a number.');
    expectMarked('Target budget', 'Target budget must be more than $0.');
    expect(onSearch).not.toHaveBeenCalled();
  });

  it('reveals the messages when the disabled Search is pressed', async () => {
    const { user } = setup();

    await user.type(field('Min bedrooms'), '-1');
    expectUnmarked('Min bedrooms');
    fireEvent.pointerDown(searchButton());
    expectMarked('Min bedrooms', 'Min bedrooms must be 0 or more.');
  });

  it('reveals nothing while Clear is pressed, so Clear stays put until the click lands', async () => {
    const { user, onClear } = setup();

    await user.type(field('Min bedrooms'), '-1');
    const clear = screen.getByRole('button', { name: 'Clear' });
    await user.pointer({ keys: '[MouseLeft>]', target: clear });
    expectUnmarked('Min bedrooms');
    expect(field('Min bedrooms')).toHaveFocus();
    expect(screen.queryByText('Min bedrooms must be 0 or more.')).not.toBeInTheDocument();

    await user.pointer({ keys: '[/MouseLeft]', target: clear });
    expect(onClear).toHaveBeenCalledTimes(1);
    expect(field('Min bedrooms')).toHaveValue('');
    expectUnmarked('Min bedrooms');
  });

  it('removes the mark and enables Search once the value is corrected', async () => {
    const { user } = setup();

    await user.type(field('Min price'), '-5');
    await user.tab();
    expectMarked('Min price', 'Min price must be 0 or more.');

    await user.clear(field('Min price'));
    await user.type(field('Min price'), '5');
    expectUnmarked('Min price');
    expect(screen.queryByText('Min price must be 0 or more.')).not.toBeInTheDocument();
    expect(searchButton()).toBeEnabled();
  });
});

describe('FilterBar server errors', () => {
  const fixtureFields: FieldError[] = [
    ...(validationErrorBody.error.fields ?? []),
    { field: 'query', code: 'malformed_query', message: 'Query string is not valid URL encoding.' },
  ];

  it('maps the fixture 400: min_gt_max on both prices, the rest on one line under the bar', () => {
    setup({ serverErrors: fixtureFields });

    expectMarked('Max price', MIN_GT_MAX);
    expect(field('Min price')).toHaveAttribute('aria-invalid', 'true');
    expect(description(field('Min price'))).toBe(description(field('Max price')));
    const alert = screen.getByRole('alert');
    expect(alert).toHaveTextContent(
      'Page size must be between 1 and 100. Unknown parameter. Query string is not valid URL encoding.',
    );
    expect(screen.getByRole('search').lastElementChild).toBe(alert);
    for (const label of ['Min bedrooms', 'City', 'Keyword', 'Target budget', 'Status']) {
      expectUnmarked(label);
    }
    expect(searchButton()).toBeEnabled();
  });

  it.each([
    ['minPrice', 'Min price', 'negative', 'Min price must be 0 or more.'],
    ['maxPrice', 'Max price', 'not_a_number', 'Max price must be a number.'],
    ['minBedrooms', 'Min bedrooms', 'not_an_integer', 'Min bedrooms must be a whole number.'],
    ['city', 'City', 'too_long', 'City must be 100 characters or fewer.'],
    ['keyword', 'Keyword', 'too_long', 'Keyword must be 200 characters or fewer.'],
    ['targetBudget', 'Target budget', 'not_positive', 'Target budget must be more than $0.'],
    ['status', 'Status', 'invalid_enum', 'Status must be active, pending or sold.'],
  ])('marks %s with the server message', (name, label, code, message) => {
    setup({ serverErrors: [{ field: name, code, message }] });

    expectMarked(label, message);
    expect(field(label).closest('div')).toHaveTextContent(message);
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    expect(searchButton()).toBeEnabled();
  });

  it('marks a field for an empty server message without describing it', () => {
    setup({ serverErrors: [{ field: 'keyword', code: 'too_long', message: '' }] });

    expect(field('Keyword')).toHaveAttribute('aria-invalid', 'true');
    expect(field('Keyword')).not.toHaveAttribute('aria-describedby');
  });

  it('clears a server mark on edit; Min price clears min_gt_max; new errors restore marks', async () => {
    const serverErrors: FieldError[] = [
      { field: 'maxPrice', code: 'min_gt_max', message: MIN_GT_MAX },
      { field: 'keyword', code: 'too_long', message: 'Keyword must be 200 characters or fewer.' },
    ];
    const { user, rerender } = setup({ serverErrors });

    await user.type(field('Min price'), '1');
    expectUnmarked('Min price');
    expectUnmarked('Max price');
    expectMarked('Keyword', 'Keyword must be 200 characters or fewer.');

    await user.type(field('Keyword'), 'x');
    expectUnmarked('Keyword');

    rerender({ serverErrors: [...serverErrors] });
    expectUnmarked('Max price');
    expectUnmarked('Keyword');

    rerender({ serverErrors: undefined });
    rerender({ serverErrors });
    expectMarked('Max price', MIN_GT_MAX);
    expectMarked('Keyword', 'Keyword must be 200 characters or fewer.');
  });

  it('restores a dismissed server mark after a submit that gets the same errors', async () => {
    const serverErrors: FieldError[] = [
      { field: 'city', code: 'too_long', message: 'City must be 100 characters or fewer.' },
    ];
    const { user, onSearch, rerender } = setup({ serverErrors });

    await user.selectOptions(field('City'), 'Reston');
    expectUnmarked('City');
    await user.click(searchButton());
    expect(onSearch).toHaveBeenCalledTimes(1);
    rerender({ serverErrors: [...serverErrors] });
    expectMarked('City', 'City must be 100 characters or fewer.');
  });
});

describe('FilterBar submit', () => {
  it('sends parsed numbers, the trimmed city and keyword and the status, nothing else', async () => {
    const { user, onSearch } = setup({ value: { city: '  reston ' } });

    expect(field('City')).toHaveValue('Reston');
    await user.type(field('Min price'), '$400,000');
    await user.type(field('Max price'), '550000');
    await user.type(field('Min bedrooms'), '2');
    await user.type(field('Keyword'), '  pets allowed  ');
    await user.type(field('Target budget'), '450000');
    await user.selectOptions(field('Status'), 'Pending');
    await user.click(searchButton());

    expect(onSearch).toHaveBeenCalledTimes(1);
    expect(onSearch.mock.calls[0]?.[0]).toStrictEqual({
      minPrice: 400000,
      maxPrice: 550000,
      minBedrooms: 2,
      city: 'Reston',
      keyword: 'pets allowed',
      targetBudget: 450000,
      status: ['pending'],
    });
  });

  it('sends a chosen city', async () => {
    const { user, onSearch } = setup();

    await user.selectOptions(field('City'), 'Fairfax');
    await user.click(searchButton());
    expect(onSearch.mock.calls[0]?.[0]).toStrictEqual({ city: 'Fairfax' });
  });

  it('omits empty fields, Any status, page and pageSize', async () => {
    const { user, onSearch } = setup({ value: { page: 3, pageSize: 20 } });

    await user.click(searchButton());
    expect(onSearch.mock.calls[0]?.[0]).toStrictEqual({});
  });

  it.each(LABELS)('Enter in %s submits exactly once', async (label) => {
    const { user, onSearch } = setup({ value: filledValue });

    await user.click(field(label));
    await user.keyboard('{Enter}');
    expect(onSearch).toHaveBeenCalledTimes(1);
    expect(onSearch.mock.calls[0]?.[0]).toStrictEqual(filledValue);
  });

  it('ignores the Enter that ends an IME composition', () => {
    const { onSearch } = setup({ value: filledValue });

    // WebKit sends the committing Enter after compositionend, with isComposing false.
    fireEvent.keyDown(field('Keyword'), { key: 'Enter', keyCode: 229 });
    fireEvent.keyDown(field('Keyword'), { key: 'Enter', keyCode: 13, isComposing: true });
    expect(onSearch).not.toHaveBeenCalled();
    fireEvent.keyDown(field('Keyword'), { key: 'Enter', keyCode: 13 });
    expect(onSearch).toHaveBeenCalledTimes(1);
  });

  it('while loading disables Search, keeps fields enabled and sends nothing on Enter', async () => {
    const { user, onSearch } = setup({ value: filledValue, loading: true });

    expect(searchButton()).toBeDisabled();
    for (const label of LABELS) {
      expect(field(label)).toBeEnabled();
    }
    expect(screen.getByRole('button', { name: 'Clear' })).toBeEnabled();
    await user.type(field('Keyword'), ' yard{Enter}');
    expect(field('Keyword')).toHaveValue('pets yard');
    await user.click(searchButton());
    expect(onSearch).not.toHaveBeenCalled();
  });
});

describe('FilterBar Clear', () => {
  it('calls onClear and empties drafts and field marks at once, keeping the unmapped line', async () => {
    const { user, onClear, onSearch, rerender } = setup({
      value: filledValue,
      serverErrors: validationErrorBody.error.fields,
    });

    await user.clear(field('Min price'));
    await user.type(field('Min price'), '-5');
    await user.tab();
    expectMarked('Min price', 'Min price must be 0 or more.');
    const unmapped = 'Page size must be between 1 and 100. Unknown parameter.';
    expect(screen.getByRole('alert')).toHaveTextContent(unmapped);

    await user.click(screen.getByRole('button', { name: 'Clear' }));

    expect(onClear).toHaveBeenCalledTimes(1);
    expect(onSearch).not.toHaveBeenCalled();
    for (const label of LABELS) {
      expect(field(label)).toHaveValue('');
      expectUnmarked(label);
    }
    // Clear keeps pageSize (SPEC section 7.1), so the 400 behind the line may still apply.
    expect(screen.getByRole('alert')).toHaveTextContent(unmapped);
    expect(searchButton()).toBeEnabled();

    rerender({ serverErrors: undefined });
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    rerender({
      serverErrors: [{ field: 'page', code: 'out_of_range', message: 'Page is too large.' }],
    });
    expect(screen.getByRole('alert')).toHaveTextContent('Page is too large.');
  });

  it('empties typed drafts when value is already empty', async () => {
    const { user, onClear, rerender } = setup();

    await user.type(field('Keyword'), 'yard');
    await user.selectOptions(field('Status'), 'Sold');
    await user.click(screen.getByRole('button', { name: 'Clear' }));
    rerender({ value: {} });

    expect(onClear).toHaveBeenCalledTimes(1);
    expect(field('Keyword')).toHaveValue('');
    expect(field('Status')).toHaveValue('');
  });
});

describe('FilterBar ids', () => {
  it('leaves valid fields without aria-invalid or aria-describedby', () => {
    setup({ value: filledValue });

    for (const label of LABELS) {
      expectUnmarked(label);
    }
  });

  it('gives two instances distinct ids with the same labels', () => {
    setup();
    setup();

    for (const label of LABELS) {
      const [first, second] = screen.getAllByLabelText(label);
      expect(first?.id).toBeTruthy();
      expect(second?.id).toBeTruthy();
      expect(first?.id).not.toBe(second?.id);
    }
  });
});
