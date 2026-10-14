// Vitest setup (vite.config.ts setupFiles). Tests import describe/it/expect from 'vitest'
// (no globals), so RTL cannot register its own cleanup; unmount after every test here.
import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

afterEach(() => {
  cleanup();
});
