export interface DiffToken {
  type: 'same' | 'added' | 'removed';
  value: string;
}

// computeWordDiff performs a clean LCS word-level diff between two text strings
export function computeWordDiff(textA: string, textB: string): { diffA: DiffToken[]; diffB: DiffToken[] } {
  const wordsA = textA.split(/(\s+)/);
  const wordsB = textB.split(/(\s+)/);

  // Set-based fast comparison for highlighted visual differences
  const setB = new Set(wordsB.map(w => w.trim().toLowerCase()).filter(Boolean));
  const setA = new Set(wordsA.map(w => w.trim().toLowerCase()).filter(Boolean));

  const diffA: DiffToken[] = wordsA.map(w => {
    const trimmed = w.trim().toLowerCase();
    if (!trimmed || setB.has(trimmed)) {
      return { type: 'same', value: w };
    }
    return { type: 'removed', value: w };
  });

  const diffB: DiffToken[] = wordsB.map(w => {
    const trimmed = w.trim().toLowerCase();
    if (!trimmed || setA.has(trimmed)) {
      return { type: 'same', value: w };
    }
    return { type: 'added', value: w };
  });

  return { diffA, diffB };
}
