// The host's "Giờ & âm thanh" panel: every phase time, drawn from
// view.timers alone (docs/api/screens/lobby.md "What drives the lobby"),
// within the ranges of rules §2 "Thời gian mặc định"; configureTimers sends
// only the times that changed (lobby.md "What the phone may send").
import type { Client, Server } from '../api/types.ts';

export type TimerKey = keyof Server.Timers;

export interface TimerField {
  key: TimerKey;
  min: number;
  max: number;
  step: number;
}

const s = 1000;
const min = 60 * s;
const phase = { min: 5 * s, max: 10 * min, step: 5 * s }; // rules §2: every phase 5 s – 10 min
const short = { min: 5 * s, max: 1 * min, step: 5 * s }; // announcements and the nomination lock: 5 s – 1 min

/** In the order of the rules §2 table. */
export const timerFields: readonly TimerField[] = [
  { key: 'roleView', ...phase },
  { key: 'chiefElection', ...phase },
  { key: 'nightStep', ...phase },
  { key: 'phoneCheck', ...phase },
  { key: 'loversMeet', ...phase },
  { key: 'announce', ...short },
  { key: 'day', min: 1 * min, max: 30 * min, step: 1 * min }, // day: 1 – 30 min
  { key: 'nominationLock', ...short },
  { key: 'defense', ...phase },
  { key: 'verdict', ...phase },
  { key: 'hunterShot', ...phase },
  { key: 'chiefSuccession', ...phase },
];

/** An online room has no phone check (rules §1b), so its time is not shown (giao-dien-3d.md §10). */
export function visibleTimerFields(mode: Server.View['mode']): readonly TimerField[] {
  return mode === 'online' ? timerFields.filter((f) => f.key !== 'phoneCheck') : timerFields;
}

/** One step up (+1) or down (−1), kept inside the field's range. */
export function stepTimer(field: TimerField, value: number, dir: 1 | -1): number {
  return Math.min(field.max, Math.max(field.min, value + dir * field.step));
}

/** The configureTimers command for what changed, or undefined when nothing did. */
export function timersCommand(
  current: Server.Timers,
  edited: Server.Timers,
): Client.ConfigureTimersCommand | undefined {
  const timers: Client.ConfigureTimersArgs['timers'] = {};
  let changed = false;
  for (const f of timerFields) {
    if (edited[f.key] !== current[f.key]) {
      timers[f.key] = edited[f.key];
      changed = true;
    }
  }
  return changed ? { type: 'command', name: 'configureTimers', args: { timers } } : undefined;
}

/** "mm:ss", for the panel and the countdowns. */
export function formatDuration(ms: number): string {
  const total = Math.max(0, Math.ceil(ms / 1000)); // a countdown shows 00:01 until it is really over
  const m = Math.floor(total / 60);
  const sec = total % 60;
  return `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`;
}
