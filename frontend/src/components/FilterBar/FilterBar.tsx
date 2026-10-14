import { useId, useState } from 'react';
import type { ChangeEvent, FormEvent, KeyboardEvent, ReactElement } from 'react';
import type { FieldError, SearchParams } from '../../api/types';
import { LISTING_STATUSES } from '../../api/types';
import styles from './FilterBar.module.css';
import type { Drafts, FieldName } from './filterBarModel';
import {
  COPY,
  EMPTY_DRAFTS,
  FIELD_NAMES,
  FIELDS,
  cityChoice,
  displayValue,
  draftsFrom,
  filterKey,
  hasClientErrors,
  isNumericField,
  mapServerErrors,
  normalizeDraft,
  resolveMarks,
  toSearchParams,
  validate,
} from './filterBarModel';

export interface FilterBarProps {
  // The applied filters (the URL state). The bar re-seeds its drafts when the filters change.
  value: SearchParams;
  // The City select's options; [] until loaded and on failure.
  cities: string[];
  loading: boolean;
  // The API's field errors for the applied filters (SPEC section 5.5).
  serverErrors?: FieldError[] | undefined;
  onSearch: (next: SearchParams) => void;
  onClear: () => void;
}

interface Local {
  // The filterKey the drafts were seeded from; a different key means the URL moved on.
  key: string;
  drafts: Drafts;
  // Fields whose client errors are shown: after blur, an Enter attempt, or a non-empty seed.
  revealed: ReadonlySet<FieldName>;
}

interface Dismissals {
  // The serverErrors content these apply to; new content means nothing is dismissed.
  key: string;
  fields: ReadonlySet<FieldName>;
  // Clear hides every server mark, the unmapped line included.
  all: boolean;
}

const NONE: ReadonlySet<FieldName> = new Set();
const ALL: ReadonlySet<FieldName> = new Set(FIELD_NAMES);

function seed(key: string, value: SearchParams): Local {
  const drafts = draftsFrom(value);
  return { key, drafts, revealed: new Set(FIELD_NAMES.filter((field) => drafts[field] !== '')) };
}

function noDismissals(key: string): Dismissals {
  return { key, fields: NONE, all: false };
}

function isPlainEnter(event: KeyboardEvent): boolean {
  return (
    event.key === 'Enter' &&
    !event.shiftKey &&
    !event.altKey &&
    !event.ctrlKey &&
    !event.metaKey &&
    !event.nativeEvent.isComposing
  );
}

// The filter bar (SPEC section 7.4, BRANDING.md section 5.2). Typing edits local drafts only;
// Search or Enter in any field sends them through onSearch.
export function FilterBar({
  value,
  cities,
  loading,
  serverErrors,
  onSearch,
  onClear,
}: FilterBarProps) {
  const baseId = useId();
  const key = filterKey(value);
  const [storedLocal, setLocal] = useState(() => seed(key, value));
  const serverKey = JSON.stringify(serverErrors ?? []);
  const [storedDismissals, setDismissals] = useState(() => noDismissals(serverKey));
  const [focused, setFocused] = useState<FieldName | null>(null);

  // New filters re-seed the drafts and new server errors reset the dismissals, on the render
  // that brings them (React's "adjusting state when a prop changes", no effect). The reset is
  // stored, not only derived, so returning to earlier filters or errors (A, B, A) starts fresh.
  const local = storedLocal.key === key ? storedLocal : seed(key, value);
  if (storedLocal !== local) {
    setLocal(local);
  }
  const dismissals =
    storedDismissals.key === serverKey ? storedDismissals : noDismissals(serverKey);
  if (storedDismissals !== dismissals) {
    setDismissals(dismissals);
  }

  const city = cityChoice(local.drafts.city, cities);
  // What the bar shows, and so what Search sends.
  const drafts: Drafts = { ...local.drafts, city: city.value };
  const client = validate(drafts);
  const invalid = hasClientErrors(client);
  const canSubmit = !loading && !invalid;
  const server = mapServerErrors(serverErrors);
  const marks = resolveMarks(
    client,
    server,
    local.revealed,
    dismissals.all ? ALL : dismissals.fields,
  );
  const notice = dismissals.all ? '' : server.unmapped;

  // Every update starts from the drafts on screen and records the key they belong to.
  const updateLocal = (change: (current: Local) => Local) => {
    setLocal((prev) => ({ ...change(prev.key === key ? prev : seed(key, value)), key }));
  };

  const dismiss = (field: FieldName) => {
    setDismissals((prev) => {
      const current = prev.key === serverKey ? prev : noDismissals(serverKey);
      return { ...current, fields: new Set([...current.fields, field]) };
    });
  };

  const revealAll = () => {
    updateLocal((current) => ({ ...current, revealed: ALL }));
  };

  const edit = (field: FieldName, text: string) => {
    updateLocal((current) => ({ ...current, drafts: { ...current.drafts, [field]: text } }));
    dismiss(field);
  };

  const blur = (field: FieldName) => {
    setFocused(null);
    updateLocal((current) => ({
      ...current,
      drafts: { ...current.drafts, [field]: normalizeDraft(field, current.drafts[field]) },
      revealed: new Set([...current.revealed, field]),
    }));
  };

  const submit = () => {
    if (loading) {
      return;
    }
    if (invalid) {
      revealAll();
      return;
    }
    // The API is asked again, so its marks may come back.
    setDismissals(noDismissals(serverKey));
    onSearch(toSearchParams(drafts));
  };

  const clear = () => {
    setLocal({ key, drafts: EMPTY_DRAFTS, revealed: NONE });
    setDismissals({ key: serverKey, fields: NONE, all: true });
    onClear();
  };

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    submit();
  };

  // Enter is handled on the fields: browsers skip implicit submission while Search is disabled,
  // and selects never submit implicitly.
  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement | HTMLSelectElement>) => {
    if (isPlainEnter(event)) {
      event.preventDefault();
      submit();
    }
  };

  // A press on a disabled Search explains itself (best effort: not every browser reports it).
  const handleActionsPointerDown = () => {
    if (!loading && invalid) {
      revealAll();
    }
  };

  const maxPriceMessageId = `${baseId}-maxPrice-msg`;
  const maxPriceMark = marks.maxPrice;
  const maxPriceHasMessage = maxPriceMark?.kind === 'message' && maxPriceMark.text !== '';

  return (
    <form className={styles.bar} role="search" aria-label={COPY.formName} onSubmit={handleSubmit}>
      {FIELDS.map(({ name, label, placeholder, kind }) => {
        const id = `${baseId}-${name}`;
        const messageId = `${id}-msg`;
        const mark = marks[name];
        const message = mark?.kind === 'message' && mark.text !== '' ? mark.text : undefined;
        let describedBy: string | undefined;
        if (message !== undefined) {
          describedBy = messageId;
        } else if (mark?.kind === 'conflict' && maxPriceHasMessage) {
          describedBy = maxPriceMessageId;
        }
        const shared = {
          id,
          className: styles.control,
          'aria-invalid': mark === undefined ? undefined : true,
          'aria-describedby': describedBy,
          onFocus: () => {
            setFocused(name);
          },
          onBlur: () => {
            blur(name);
          },
          onKeyDown: handleKeyDown,
          onChange: (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
            edit(name, event.target.value);
          },
        };

        let control: ReactElement;
        if (kind === 'city') {
          control = (
            <span className={styles.select}>
              <select {...shared} value={drafts.city}>
                <option value="">{COPY.any}</option>
                {cities.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
                {city.extra !== undefined && <option value={city.extra}>{city.extra}</option>}
              </select>
            </span>
          );
        } else if (kind === 'status') {
          control = (
            <span className={styles.select}>
              <select {...shared} value={drafts.status}>
                <option value="">{COPY.any}</option>
                {LISTING_STATUSES.map((status) => (
                  <option key={status} value={status}>
                    {COPY.statuses[status]}
                  </option>
                ))}
              </select>
            </span>
          );
        } else {
          control = (
            <input
              {...shared}
              type="text"
              inputMode={kind === 'money' ? 'decimal' : kind === 'count' ? 'numeric' : undefined}
              autoComplete="off"
              placeholder={placeholder}
              value={
                isNumericField(name)
                  ? displayValue(name, drafts[name], focused === name)
                  : drafts[name]
              }
            />
          );
        }

        return (
          <div key={name} className={styles.field}>
            <label className={styles.label} htmlFor={id}>
              {label}
            </label>
            {control}
            {message !== undefined && (
              <p id={messageId} className={styles.message}>
                {message}
              </p>
            )}
          </div>
        );
      })}
      {/* Pointer only: Enter and blur already reveal messages from the keyboard. */}
      <div className={styles.actions} onPointerDown={handleActionsPointerDown}>
        <button type="submit" className={styles.search} disabled={!canSubmit}>
          {COPY.search}
        </button>
        <button type="button" className={styles.clear} onClick={clear}>
          {COPY.clear}
        </button>
      </div>
      {notice !== '' && (
        <p role="alert" className={styles.notice}>
          {notice}
        </p>
      )}
    </form>
  );
}
