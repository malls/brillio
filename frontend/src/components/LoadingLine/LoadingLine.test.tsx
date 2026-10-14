import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { LoadingLine } from './LoadingLine';
import styles from './LoadingLine.module.css';

// The app tsconfig loads no Node types, so node:fs is typed locally and loaded through Vitest.
interface NodeFs {
  readFileSync(path: string, encoding: 'utf8'): string;
}
const fs = await vi.importActual<NodeFs>('node:fs');

// The shipped stylesheet's source. jsdom evaluates no media queries and Vitest does not process
// CSS, so the motion cases below verify the stylesheet text; the visual behavior is proven in
// BC-12's browser validation. The module URL goes through a variable because Vite rewrites the
// literal pattern new URL('...', import.meta.url) into a served asset URL (http://localhost/...).
const moduleUrl = import.meta.url;
const css = fs.readFileSync(
  decodeURIComponent(new URL('LoadingLine.module.css', moduleUrl).pathname),
  'utf8',
);

// The scoped class name Vitest gives a CSS module class.
function cls(name: string): string {
  const value = styles[name];
  if (value === undefined) {
    throw new Error(`LoadingLine.module.css has no class ${name}`);
  }
  return value;
}

// The declarations of the first top-level rule for selector, e.g. rule('.segment').
function rule(selector: string): string {
  const escaped = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const match = new RegExp(`(?:^|\\n)${escaped}\\s*\\{([^}]*)\\}`).exec(css);
  if (match?.[1] === undefined) {
    throw new Error(`LoadingLine.module.css has no rule ${selector}`);
  }
  return match[1];
}

describe('LoadingLine', () => {
  it('is a progressbar named Loading with a segment on its track while active', () => {
    const { container } = render(<LoadingLine active />);

    const bar = screen.getByRole('progressbar', { name: 'Loading' });
    expect(container.firstElementChild).toBe(bar);
    expect(bar).toHaveClass(cls('track'), cls('active'));
    expect(bar).not.toHaveAttribute('aria-valuenow');
    expect(bar.children).toHaveLength(1);
    expect(bar.firstElementChild).toHaveClass(cls('segment'));
  });

  it('keeps the same empty track, with no role or segment, while inactive', () => {
    const { container } = render(<LoadingLine active={false} />);

    expect(screen.queryByRole('progressbar', { hidden: true })).not.toBeInTheDocument();
    const track = container.firstElementChild;
    expect(track).toHaveClass(cls('track'));
    expect(track).not.toHaveClass(cls('active'));
    expect(track).toHaveAttribute('aria-hidden', 'true');
    expect(track).not.toHaveAttribute('role');
    expect(track).toBeEmptyDOMElement();
  });

  it('slides a 40% oxblood segment along a 2px line track, 1.2s linear, looping', () => {
    expect(css.length).toBeGreaterThan(0);

    expect(rule('.track')).toMatch(/height:\s*2px;/);
    expect(rule('.active')).toMatch(/background:\s*var\(--line\);/);
    const segment = rule('.segment');
    expect(segment).toMatch(/width:\s*40%;/);
    expect(segment).toMatch(/background:\s*var\(--oxblood\);/);
    expect(segment).toMatch(/animation:\s*slide\s+1\.2s\s+linear\s+infinite;/);
    expect(css).toMatch(
      /@keyframes\s+slide\s*\{\s*from\s*\{\s*transform:\s*translateX\(-100%\);\s*\}\s*to\s*\{\s*transform:\s*translateX\(250%\);\s*\}\s*\}/,
    );
  });

  it('is a static full-width line under prefers-reduced-motion', () => {
    expect(css.length).toBeGreaterThan(0);

    const reduced =
      /@media\s*\(prefers-reduced-motion:\s*reduce\)\s*\{\s*\.segment\s*\{([^}]*)\}\s*\}/.exec(css);
    const declarations = reduced?.[1];
    expect(declarations).toBeDefined();
    expect(declarations).toMatch(/width:\s*100%;/);
    expect(declarations).toMatch(/animation:\s*none;/);
  });

  it('uses oxblood only for the segment and no hex colors', () => {
    expect(css.length).toBeGreaterThan(0);

    expect(css.match(/var\(--oxblood\)/g)).toHaveLength(1);
    expect(rule('.segment')).toContain('var(--oxblood)');
    expect(css).not.toMatch(/#[0-9a-fA-F]{3,8}\b/);
  });
});
