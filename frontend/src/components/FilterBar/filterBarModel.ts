import { buildSearchQuery } from '../../api/client';
import type { FieldError, ListingStatus, SearchParams } from '../../api/types';
import { LISTING_STATUSES } from '../../api/types';
import { formatPrice } from '../../format/money';

// The FilterBar's copy and rules (docs/SPEC.md section 7.4, BRANDING.md sections 5.2 and 7.2).
// Pure helpers for FilterBar.tsx; nothing outside this folder imports them. Every string the bar
// shows lives in this file.

type FieldKind = 'money' | 'count' | 'city' | 'text' | 'status';

interface FieldSpec {
  name: string;
  label: string;
  placeholder: string;
  kind: FieldKind;
}

// The seven fields in DOM order. Labels and placeholders are BRANDING.md section 7.2 verbatim;
// Status is SPEC section 7.4's seventh field.
export const FIELDS = [
  { name: 'minPrice', label: 'Min price', placeholder: 'Any', kind: 'money' },
  { name: 'maxPrice', label: 'Max price', placeholder: 'Any', kind: 'money' },
  { name: 'minBedrooms', label: 'Min bedrooms', placeholder: 'Any', kind: 'count' },
  { name: 'city', label: 'City', placeholder: 'Any', kind: 'city' },
  { name: 'keyword', label: 'Keyword', placeholder: 'Search descriptions', kind: 'text' },
  { name: 'targetBudget', label: 'Target budget', placeholder: 'Any', kind: 'money' },
  { name: 'status', label: 'Status', placeholder: 'Any', kind: 'status' },
] as const satisfies readonly FieldSpec[];

export type FieldName = (typeof FIELDS)[number]['name'];

// What each field holds while the user edits: the text as typed or chosen.
export type Drafts = Record<FieldName, string>;

export const FIELD_NAMES: readonly FieldName[] = FIELDS.map((field) => field.name);

export const COPY = {
  formName: 'Filters',
  search: 'Search',
  clear: 'Clear',
  // The first option of both selects.
  any: 'Any',
  statuses: {
    active: 'Active',
    pending: 'Pending',
    sold: 'Sold',
  } satisfies Record<ListingStatus, string>,
  // BRANDING.md section 7.2 where it has a string; otherwise the API's own sentence for the same
  // rule (backend/internal/httpapi/query.go), so a value reads the same whichever side catches it.
  messages: {
    minPriceNotANumber: 'Min price must be a number.',
    minPriceNegative: 'Min price must be 0 or more.',
    maxPriceNotANumber: 'Max price must be a number.',
    maxPriceNegative: 'Max price must be 0 or more.',
    minGtMax: 'Min price must be less than max price.',
    minBedroomsNotANumber: 'Min bedrooms must be a number.',
    minBedroomsNotWhole: 'Min bedrooms must be a whole number.',
    minBedroomsNegative: 'Min bedrooms must be 0 or more.',
    targetBudgetNotANumber: 'Target budget must be a number.',
    targetBudgetNotPositive: 'Target budget must be more than $0.',
  },
} as const;

const MONEY_FIELDS: ReadonlySet<FieldName> = new Set(['minPrice', 'maxPrice', 'targetBudget']);
const NUMERIC_FIELDS: ReadonlySet<FieldName> = new Set([...MONEY_FIELDS, 'minBedrooms']);

export function isNumericField(field: FieldName): boolean {
  return NUMERIC_FIELDS.has(field);
}

function isFieldName(name: string): name is FieldName {
  return (FIELD_NAMES as readonly string[]).includes(name);
}

function isListingStatus(value: string): value is ListingStatus {
  return (LISTING_STATUSES as readonly string[]).includes(value);
}

export type Parsed = { kind: 'empty' } | { kind: 'invalid' } | { kind: 'number'; value: number };

// Deliberately stricter than the API's strconv.ParseFloat: no exponent, no '+', no hex, no
// inner spaces. Those are not how anyone types a price, and rejecting them keeps the field's
// text and the number it sends obviously the same.
const NUMBER_TEXT = /^-?(\d+\.?\d*|\.\d+)$/;

// Reads a numeric field's text. `$` and `,` are dropped so `$400,000` reads as 400000; -0 is 0.
export function parseNumber(text: string): Parsed {
  const trimmed = text.trim();
  if (trimmed === '') {
    return { kind: 'empty' };
  }
  const bare = trimmed.replace(/[$,]/g, '');
  const value = NUMBER_TEXT.test(bare) ? Number(bare) : NaN;
  if (!Number.isFinite(value)) {
    return { kind: 'invalid' };
  }
  return { kind: 'number', value: value === 0 ? 0 : value };
}

// A number as plain digits that parseNumber reads back: String(1e21) is '1e+21' and String(1e-7)
// is '1e-7', and the parser takes no exponent. A fraction that 20 decimals cannot write exactly
// keeps its exponent, so it reads as not a number rather than as a different number.
function numberText(n: number): string {
  const text = String(n);
  if (!text.includes('e')) {
    return text;
  }
  if (Number.isInteger(n)) {
    return BigInt(n).toString();
  }
  const digits = n.toLocaleString('en-US', { useGrouping: false, maximumFractionDigits: 20 });
  return Number(digits) === n ? digits : text;
}

// The drafts for a value: numbers as plain digits, text as is, and the status when exactly one
// is set (none or several show Any).
export function draftsFrom(value: SearchParams): Drafts {
  const number = (n: number | undefined) =>
    n === undefined || !Number.isFinite(n) ? '' : numberText(n);
  const statuses = LISTING_STATUSES.filter((status) => value.status?.includes(status) === true);
  return {
    minPrice: number(value.minPrice),
    maxPrice: number(value.maxPrice),
    minBedrooms: number(value.minBedrooms),
    city: value.city ?? '',
    keyword: value.keyword ?? '',
    targetBudget: number(value.targetBudget),
    status: statuses.length === 1 ? (statuses[0] ?? '') : '',
  };
}

export const EMPTY_DRAFTS: Drafts = draftsFrom({});

// Identifies the filters a value carries: the canonical query of its seven filter fields, so an
// equal new object, a page change or a pageSize change keeps the drafts.
export function filterKey(value: SearchParams): string {
  const filters: SearchParams = { ...value };
  delete filters.page;
  delete filters.pageSize;
  return buildSearchQuery(filters);
}

// What a numeric field shows: a money field blurred on a whole number of 0 or more reads
// `$400,000`; everything else shows the draft as typed, so a fraction or a bad value is never
// rewritten into something that looks different.
export function displayValue(field: FieldName, draft: string, focused: boolean): string {
  if (focused || !MONEY_FIELDS.has(field)) {
    return draft;
  }
  const parsed = parseNumber(draft);
  return parsed.kind === 'number' && Number.isInteger(parsed.value) && parsed.value >= 0
    ? formatPrice(parsed.value)
    : draft;
}

// On blur a readable numeric draft becomes plain digits (`$400,000` -> `400000`, `2.0` -> `2`),
// so focusing it again shows what will be sent. Anything else is left as typed.
export function normalizeDraft(field: FieldName, draft: string): string {
  if (!NUMERIC_FIELDS.has(field)) {
    return draft;
  }
  const parsed = parseNumber(draft);
  if (parsed.kind !== 'number') {
    return draft;
  }
  const text = numberText(parsed.value);
  return parseNumber(text).kind === 'number' ? text : draft;
}

// The City select: the listed city matching the draft case-insensitively, else the draft
// itself as an extra option after the list (a URL city before the list loads, after it fails, or
// one with no listings), so the select always shows the applied filter.
export function cityChoice(
  draft: string,
  cities: readonly string[],
): { value: string; extra: string | undefined } {
  const wanted = draft.trim().toLowerCase();
  if (wanted === '') {
    return { value: '', extra: undefined };
  }
  const listed = cities.find((city) => city.trim().toLowerCase() === wanted);
  return listed === undefined
    ? { value: draft, extra: draft }
    : { value: listed, extra: undefined };
}

export interface ClientMarks {
  // Each field's own error message.
  messages: Partial<Record<FieldName, string>>;
  // Both prices are valid and min is above max (equal is valid, SPEC section 5.3).
  conflict: boolean;
}

function priceRule(parsed: Parsed, notANumber: string, negative: string): string | undefined {
  if (parsed.kind === 'invalid') return notANumber;
  if (parsed.kind === 'number' && parsed.value < 0) return negative;
  return undefined;
}

// The API's order (query.go): a number, then whole, then 0 or more.
function bedroomsRule(parsed: Parsed): string | undefined {
  const { messages } = COPY;
  if (parsed.kind === 'invalid') return messages.minBedroomsNotANumber;
  if (parsed.kind !== 'number') return undefined;
  if (!Number.isInteger(parsed.value)) return messages.minBedroomsNotWhole;
  if (parsed.value < 0) return messages.minBedroomsNegative;
  return undefined;
}

function budgetRule(parsed: Parsed): string | undefined {
  const { messages } = COPY;
  if (parsed.kind === 'invalid') return messages.targetBudgetNotANumber;
  if (parsed.kind === 'number' && parsed.value <= 0) return messages.targetBudgetNotPositive;
  return undefined;
}

// The API's cheap rules, run client-side (SPEC section 7.4). The first failing rule per field
// wins; the min/max conflict is checked only when neither price has an error of its own. City,
// Keyword and Status have no client rule: their length and enum checks are the API's.
export function validate(drafts: Drafts): ClientMarks {
  const { messages: copy } = COPY;
  const minPrice = parseNumber(drafts.minPrice);
  const maxPrice = parseNumber(drafts.maxPrice);
  const rules: Partial<Record<FieldName, string | undefined>> = {
    minPrice: priceRule(minPrice, copy.minPriceNotANumber, copy.minPriceNegative),
    maxPrice: priceRule(maxPrice, copy.maxPriceNotANumber, copy.maxPriceNegative),
    minBedrooms: bedroomsRule(parseNumber(drafts.minBedrooms)),
    targetBudget: budgetRule(parseNumber(drafts.targetBudget)),
  };
  const messages: Partial<Record<FieldName, string>> = {};
  for (const field of FIELD_NAMES) {
    const message = rules[field];
    if (message !== undefined) {
      messages[field] = message;
    }
  }
  const conflict =
    minPrice.kind === 'number' &&
    maxPrice.kind === 'number' &&
    messages.minPrice === undefined &&
    messages.maxPrice === undefined &&
    minPrice.value > maxPrice.value;
  return { messages, conflict };
}

export function hasClientErrors(marks: ClientMarks): boolean {
  return marks.conflict || Object.keys(marks.messages).length > 0;
}

// The onSearch payload: only fields with a value, numbers parsed, city and keyword trimmed,
// status as a one-element list unless Any. Never page or pageSize: setSearch drops page and keeps
// the URL's pageSize (SPEC section 7.1). Call it only when validate found nothing.
export function toSearchParams(drafts: Drafts): SearchParams {
  const params: SearchParams = {};
  const number = (field: FieldName): number | undefined => {
    const parsed = parseNumber(drafts[field]);
    return parsed.kind === 'number' ? parsed.value : undefined;
  };
  const minPrice = number('minPrice');
  if (minPrice !== undefined) params.minPrice = minPrice;
  const maxPrice = number('maxPrice');
  if (maxPrice !== undefined) params.maxPrice = maxPrice;
  const minBedrooms = number('minBedrooms');
  if (minBedrooms !== undefined && Number.isInteger(minBedrooms)) params.minBedrooms = minBedrooms;
  const city = drafts.city.trim();
  if (city !== '') params.city = city;
  const keyword = drafts.keyword.trim();
  if (keyword !== '') params.keyword = keyword;
  const targetBudget = number('targetBudget');
  if (targetBudget !== undefined) params.targetBudget = targetBudget;
  if (isListingStatus(drafts.status)) params.status = [drafts.status];
  return params;
}

export interface ServerMarks {
  // Each field's server message (verbatim; '' still marks the field).
  messages: Partial<Record<FieldName, string>>;
  // The message of a min_gt_max error on maxPrice: shown under Max price, both prices marked.
  conflict: string | undefined;
  // Messages for params no field shows (pageSize, page, unknown params, the query itself),
  // deduplicated and joined with one space; '' when there are none.
  unmapped: string;
}

// Maps the API's field errors (SPEC section 5.5) onto the bar. The first error per field wins.
export function mapServerErrors(errors: readonly FieldError[] | undefined): ServerMarks {
  const messages: Partial<Record<FieldName, string>> = {};
  let conflict: string | undefined;
  const seen = new Set<FieldName>();
  const unmapped: string[] = [];
  for (const { field, code, message } of errors ?? []) {
    if (!isFieldName(field)) {
      if (message !== '' && !unmapped.includes(message)) {
        unmapped.push(message);
      }
      continue;
    }
    if (seen.has(field)) {
      continue;
    }
    seen.add(field);
    if (field === 'maxPrice' && code === 'min_gt_max') {
      conflict = message;
    } else {
      messages[field] = message;
    }
  }
  return { messages, conflict, unmapped: unmapped.join(' ') };
}

// How a field is marked: with a message under it ('' marks it without one), or, for Min price in
// a min/max conflict, marked and described by Max price's message.
export type FieldMark = { kind: 'message'; text: string } | { kind: 'conflict' };

// The marks on screen. Per field a revealed client error wins over an undismissed server error.
// The client conflict shows once both prices are revealed; the server's is dismissed by editing
// either price.
export function resolveMarks(
  client: ClientMarks,
  server: ServerMarks,
  revealed: ReadonlySet<FieldName>,
  dismissed: ReadonlySet<FieldName>,
): Partial<Record<FieldName, FieldMark>> {
  const clientConflict = client.conflict && revealed.has('minPrice') && revealed.has('maxPrice');
  const serverConflict =
    server.conflict !== undefined && !dismissed.has('minPrice') && !dismissed.has('maxPrice');
  const marks: Partial<Record<FieldName, FieldMark>> = {};
  let maxShowsConflict = false;
  for (const field of FIELD_NAMES) {
    const own = revealed.has(field) ? client.messages[field] : undefined;
    const fromServer = dismissed.has(field) ? undefined : server.messages[field];
    let text = own;
    if (field === 'maxPrice' && text === undefined && clientConflict) {
      text = COPY.messages.minGtMax;
      maxShowsConflict = true;
    }
    text ??= fromServer;
    if (field === 'maxPrice' && text === undefined && serverConflict) {
      text = server.conflict;
      maxShowsConflict = true;
    }
    if (text !== undefined) {
      marks[field] = { kind: 'message', text };
    }
  }
  if (maxShowsConflict && marks.minPrice === undefined) {
    marks.minPrice = { kind: 'conflict' };
  }
  return marks;
}
