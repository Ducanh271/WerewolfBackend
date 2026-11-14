// Where the app starts and where it goes when the room is over for this
// phone (docs/api/README.md "Joining and rejoining"):
// - /j/{code} with a saved token rejoins at once; without one it checks the
//   code (GET /join/{code}) and opens the join form;
// - "/" rejoins the room the phone was playing in, if its token is saved;
// - after a stop: `leave` goes home; an expired room found by that automatic
//   rejoin goes home without an error (the player did not ask for anything);
//   anything else opens the join form of that room, with its reason.
import type { StopReason } from '../state/store.ts';
import { codeFromPath } from '../net/current.ts';
import type { Route } from './screen.ts';

export interface Boot {
  route: Route;
  /** Rejoin this room with its saved token. */
  rejoin?: string;
  /** Check this code with GET /join/{code}. */
  check?: string;
}

export function bootPlan(path: string, hasToken: (room: string) => boolean, current: string | undefined): Boot {
  const code = codeFromPath(path);
  if (code) return hasToken(code) ? { route: { name: 'join', code }, rejoin: code } : { route: { name: 'join', code }, check: code };
  if (current && hasToken(current)) return { route: { name: 'home' }, rejoin: current };
  return { route: { name: 'home' } };
}

export function afterStop(reason: StopReason, room: string | undefined, automatic: boolean): { route: Route; clearError: boolean } {
  switch (reason) {
    case 'left':
    case 'outdated':
      return { route: { name: 'home' }, clearError: false };
    case 'roomGone':
      return automatic ? { route: { name: 'home' }, clearError: true } : { route: { name: 'join', ...(room && { code: room }) }, clearError: false };
    case 'kicked':
    case 'replaced':
    case 'badToken':
      return { route: { name: 'join', ...(room && { code: room }) }, clearError: false };
  }
}
