// Keyword emphasis for the card description (docs/SPEC.md section 7.4, BRANDING.md section 5.4):
// each case-insensitive match of a keyword token is set at weight 600. Pure functions; none of
// them throws.

// The URL `keyword` split into tokens the way the server splits it (SPEC section 5.3,
// domain.Filter.KeywordTokens): on runs of whitespace, dropping empty strings. The card's
// `keywordTokens` prop is this value.
export function splitKeyword(keyword: string | undefined): string[] {
  if (keyword === undefined) {
    return [];
  }
  return keyword.split(/\s+/).filter((token) => token !== '');
}

// A run of the description: `match` runs are emphasized, the rest is plain text.
export interface EmphasisSegment {
  text: string;
  match: boolean;
}

// The ES2023 lib has no RegExp.escape: every regex metacharacter matches literally.
function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

// Splits `text` into plain and matched runs. Every occurrence of every token is matched
// case-insensitively and keeps the text's own casing; overlapping and adjacent matches merge into
// one run; token order does not matter; blank tokens are ignored. Joining the runs' text gives
// back `text` exactly. Empty text gives no runs; no tokens give one plain run.
export function emphasisSegments(text: string, tokens: readonly string[]): EmphasisSegment[] {
  if (text === '') {
    return [];
  }
  const marked = new Array<boolean>(text.length).fill(false);
  for (const raw of tokens) {
    const token = raw.trim();
    if (token === '') {
      continue;
    }
    const pattern = new RegExp(escapeRegExp(token), 'gi');
    for (let found = pattern.exec(text); found !== null; found = pattern.exec(text)) {
      marked.fill(true, found.index, found.index + found[0].length);
      // Restart one unit past this match's start so overlapping occurrences are found too.
      pattern.lastIndex = found.index + 1;
    }
  }
  const segments: EmphasisSegment[] = [];
  let start = 0;
  for (let end = 1; end <= text.length; end += 1) {
    if (end === text.length || marked[end] !== marked[start]) {
      segments.push({ text: text.slice(start, end), match: marked[start] === true });
      start = end;
    }
  }
  return segments;
}
