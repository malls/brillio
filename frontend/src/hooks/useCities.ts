import { useEffect, useState } from 'react';
import { getCities } from '../api/client';

// The City select's options (docs/SPEC.md section 7.2): loaded once on mount. [] until loaded
// and on failure, which is logged; the select then offers only "Any" and search still works.
export function useCities(): string[] {
  const [cities, setCities] = useState<string[]>([]);

  useEffect(() => {
    const controller = new AbortController();
    getCities(controller.signal).then(
      (list) => {
        if (!controller.signal.aborted) {
          setCities(list);
        }
      },
      (err: unknown) => {
        if (!controller.signal.aborted) {
          console.error('Failed to load cities', err);
        }
      },
    );
    return () => {
      controller.abort();
    };
  }, []);

  return cities;
}
