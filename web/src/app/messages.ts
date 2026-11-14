// The sentence to show for what went wrong (docs/api/errors.md "How to show
// an error"): the server's code through errorText, or the app's own texts
// for what has no code (no network, an app too old for the server).
import { errorText } from '../api/error-text.ts';
import { vi } from '../i18n/vi.ts';
import { networkError } from '../net/http.ts';
import type { State } from '../state/store.ts';

export function stateErrorText(s: State): string {
  if (s.connection.state === 'stopped' && s.connection.reason === 'outdated') return vi.outdated;
  const e = s.error;
  return e ? errorText(e.code, e.situation, e.retryAfter) : '';
}

export function httpErrorText(code: string, retryAfter?: number): string {
  return code === networkError ? vi.network : errorText(code, undefined, retryAfter);
}
