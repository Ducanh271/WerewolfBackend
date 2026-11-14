// errorText picks the sentence for an error code, as docs/api/errors.md "How
// to show an error" says: the row of the situation the app knows from the
// message it sent, else the row without one; `%s` of pin_locked is the
// seconds left; an unknown code shows the shared sentence.
import { errorTable } from './errors.ts';

export function errorText(code: string, situation?: string, retryAfterMs?: number): string {
  const entry = errorTable.codes[code];
  if (!entry) return errorTable.fallback;
  const text = (situation && entry.situations?.[situation]) || entry.text;
  if (!text.includes('%s')) return text;
  return text.replace('%s', String(Math.max(1, Math.ceil((retryAfterMs ?? 0) / 1000))));
}
