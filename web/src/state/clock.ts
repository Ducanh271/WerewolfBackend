// The countdown of every phase and night step, on the server's clock
// (docs/api/secrecy.md "Time comes from the server"): `deadline` is on the
// server's clock, and Store.clockOffset says how far this phone's clock is
// from it, so phones whose clocks disagree still end a step together.
import type { State } from './store.ts';

/** The server's clock now, in Unix ms, from this phone's clock `localNow`. */
export function serverNow(state: Pick<State, 'clockOffset'>, localNow: number = Date.now()): number {
  return localNow + (state.clockOffset ?? 0);
}

/** Milliseconds left before `deadline` (0 once it has passed); undefined when nothing is timed. */
export function remaining(
  state: Pick<State, 'deadline' | 'clockOffset'>,
  localNow: number = Date.now(),
): number | undefined {
  if (state.deadline === undefined) return undefined;
  return Math.max(0, state.deadline - serverNow(state, localNow));
}
