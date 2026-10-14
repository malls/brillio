import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { listingA1 } from '../../test/fixtures';
import { ScorePill } from './ScorePill';

describe('ScorePill', () => {
  it('renders No. 11 and the score to two decimals', () => {
    // Rank 11 is page 2's first rank (BRANDING.md section 5.6).
    const { container } = render(<ScorePill rank={11} score={0.8731} />);

    expect(screen.getByText('No. 11')).toBeInTheDocument();
    expect(screen.getByText('0.87')).toBeInTheDocument();
    expect(container.textContent).toBe('No. 110.87');
  });

  it("renders a fixture's rank and score", () => {
    render(<ScorePill rank={listingA1.rank} score={listingA1.relevance.score} />);

    expect(screen.getByText('No. 1')).toBeInTheDocument();
    expect(screen.getByText('0.74')).toBeInTheDocument();
  });

  it('puts the rank pill before the score', () => {
    render(<ScorePill rank={1} score={0.7414} />);

    const pill = screen.getByText('No. 1');
    const score = screen.getByText('0.74');
    expect(pill.parentElement).toBe(score.parentElement);
    expect(pill.compareDocumentPosition(score) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });
});
