// Store holds what the server last told this phone, and tells the screens
// when it changes (docs/api/README.md "Server → client"):
// - each `view` message replaces the view, `deadline`, `speaker` and `people`
//   together; a field the message leaves out is gone, never kept from before;
// - `events` with `reset` replaces the whole list; without it, `from` is the
//   length of the list already held and the events continue it.
// Unknown event types stay in the list: the server counts them in `from`.
// Each view's `serverTime` gives the phone's clock offset from the server's,
// the clock `deadline` is on (README "Server → client"; state/clock.ts).
import type { Server } from '../api/types.ts';

/** An error the server sent, with the situation the app knows from what it sent (docs/api/errors.md). */
export interface ShownError {
  code: string;
  situation?: string;
  /** Milliseconds; only with pin_locked. */
  retryAfter?: number;
  /** Goes up with every error, so the same code twice is still news. */
  seq: number;
}

/** Why the phone stopped reconnecting and went back to the join screen (README "Joining and rejoining"). */
export type StopReason =
  | 'replaced' // replaced_by_other_device: another phone took this player over
  | 'kicked' // kicked_by_host
  | 'badToken' // a token rejoin got bad_credentials
  | 'roomGone' // a token rejoin got room_not_found
  | 'left' // the player sent leave
  | 'outdated'; // the server speaks a newer protocol (README "Versioning")

export type Connection =
  | { state: 'idle' } // no socket, nothing to rejoin
  | { state: 'connecting' } // socket opening, or join/rejoin sent
  | { state: 'joined' }
  | { state: 'waiting'; retryInMs: number } // dropped, rejoining after a delay
  | { state: 'stopped'; reason: StopReason };

export interface State {
  connection: Connection;
  room?: string;
  /** This phone's public player id, from welcome. */
  player?: string;
  /** Shown once after a join, then cleared by the screen (docs/api/screens/lobby.md). Never stored. */
  pin?: string;
  view?: Server.View;
  deadline?: number;
  /**
   * Server clock minus this phone's clock, in ms: the largest offset seen
   * since the last welcome. Each sample is low by the time its view took on
   * the network, so the largest is the closest; a new socket starts again.
   */
  clockOffset?: number;
  speaker: boolean;
  people?: Server.Person[];
  events: readonly Server.Event[];
  error?: ShownError;
}

/** What applying a message did; `gap` means events were missed and the phone must rejoin. */
export type Applied = 'ok' | 'gap';

export type Listener = (state: State) => void;

const empty: State = { connection: { state: 'idle' }, speaker: false, events: [] };

export class Store {
  private current: State = empty;
  private readonly listeners = new Set<Listener>();
  private errorSeq = 0;
  private readonly now: () => number;

  /** `now` is this phone's clock in Unix ms (a test passes its own). */
  constructor(now: () => number = Date.now) {
    this.now = now;
  }

  get state(): State {
    return this.current;
  }

  /** Calls `fn` on every change; returns the function that stops it. */
  subscribe(fn: Listener): () => void {
    this.listeners.add(fn);
    return () => this.listeners.delete(fn);
  }

  /** Applies one server message. `situation` names what the error answers (errors.md "Situation"). */
  apply(msg: Server.ServerMessage, situation?: string): Applied {
    switch (msg.type) {
      case 'welcome':
        this.set({ player: msg.player, clockOffset: undefined, ...(msg.pin !== undefined && { pin: msg.pin }) });
        return 'ok';
      case 'view': {
        const sample = typeof msg.serverTime === 'number' ? msg.serverTime - this.now() : undefined;
        const held = this.current.clockOffset;
        const clockOffset = sample === undefined || (held !== undefined && held >= sample) ? held : sample;
        this.set({ view: msg.view, deadline: msg.deadline, speaker: msg.speaker === true, people: msg.people, clockOffset });
        return 'ok';
      }
      case 'events':
        return this.applyEvents(msg);
      case 'error':
        this.set({
          error: {
            code: msg.code,
            seq: ++this.errorSeq,
            ...(situation !== undefined && { situation }),
            ...(msg.retryAfter !== undefined && { retryAfter: msg.retryAfter }),
          },
        });
        return 'ok';
      default: {
        // A newer server may add message types; ignore them (README "Versioning").
        const unknown: never = msg;
        void unknown;
        return 'ok';
      }
    }
  }

  setConnection(connection: Connection): void {
    this.set({ connection });
  }

  /** Starts a room: everything from the previous one is dropped. */
  enterRoom(room: string): void {
    this.current = { ...empty, connection: this.current.connection, room };
    this.emit();
  }

  /**
   * Back at the join screen: the game is dropped so nothing of it stays on
   * screen (a kicked or replaced phone keeps no secret); the last error and
   * the connection stay for the join screen to explain why.
   */
  leaveRoom(): void {
    const { connection, error } = this.current;
    this.current = { ...empty, connection, ...(error && { error }) };
    this.emit();
  }

  clearPin(): void {
    this.set({ pin: undefined });
  }

  clearError(): void {
    this.set({ error: undefined });
  }

  private applyEvents(msg: Server.EventsMessage): Applied {
    if (msg.reset) {
      this.set({ events: msg.events });
      return 'ok';
    }
    const held = this.current.events;
    if (msg.from > held.length) return 'gap';
    // from < length: the phone already holds part of it; the server's copy wins.
    this.set({ events: [...held.slice(0, msg.from), ...msg.events] });
    return 'ok';
  }

  private set(patch: Partial<State>): void {
    this.current = { ...this.current, ...patch };
    this.emit();
  }

  private emit(): void {
    for (const fn of this.listeners) fn(this.current);
  }
}
