// GameSocket is the phone's line to the room (docs/api/README.md
// "Connecting", "Joining and rejoining", "Limits"):
// - it opens /ws (no query string) and sends join or rejoin at once, well
//   inside the server's 10 seconds;
// - welcome's token is saved per room (TokenBook); hostKey and PIN are never
//   stored;
// - a dropped line is rejoined with the token, after a growing delay;
// - it stops, and the app goes back to the join screen, on
//   replaced_by_other_device, kicked_by_host, a token rejoin refused with
//   bad_credentials or room_not_found, after leave, and when the server
//   speaks a newer protocol.
// Every message from the server goes into the Store.
import type { Client, Server } from '../api/types.ts';
import type { Applied, StopReason, Store } from '../state/store.ts';
import { TokenBook, roomKey } from './tokens.ts';

/** The protocol this app was built for (README "Versioning"). */
export const protocolVersion = 1;
/** The server refuses bigger messages (README "Limits"). */
export const maxMessageBytes = 4096;

/** How the player enters a room (docs/api/screens/lobby.md "Join screen"). */
export type Entry =
  | { kind: 'join'; room: string; nickname: string; hostKey?: string }
  | { kind: 'pin'; room: string; nickname: string; pin: string }
  | { kind: 'token'; room: string };

/** What a joined phone may send. */
export type Outgoing = Client.Command | Client.Kick | Client.TransferHost;

/** The part of the browser WebSocket used here, so tests can stand in for it. */
export interface WebSocketLike {
  send(data: string): void;
  close(code?: number): void;
  onopen: (() => void) | null;
  onmessage: ((ev: { data: unknown }) => void) | null;
  onclose: (() => void) | null;
  onerror: (() => void) | null;
}

export interface SocketEnv {
  open(url: string): WebSocketLike;
  url: string;
  setTimeout(fn: () => void, ms: number): unknown;
  clearTimeout(handle: unknown): void;
  random(): number;
}

/** Delay before reconnect try `attempt` (0 = at once): 1 s, 2 s, 4 s … up to 30 s, ±25 %. */
export function retryDelay(attempt: number, random: () => number): number {
  if (attempt <= 0) return 0;
  const base = Math.min(30_000, 1000 * 2 ** (attempt - 1));
  return Math.round(base * (0.75 + 0.5 * random()));
}

function browserEnv(): SocketEnv {
  const scheme = location.protocol === 'https:' ? 'wss:' : 'ws:';
  return {
    open: (url) => new WebSocket(url) as unknown as WebSocketLike,
    url: `${scheme}//${location.host}/ws`,
    setTimeout: (fn, ms) => globalThis.setTimeout(fn, ms),
    clearTimeout: (h) => globalThis.clearTimeout(h as ReturnType<typeof setTimeout>),
    random: Math.random,
  };
}

/** The join or rejoin for an entry; undefined for a token rejoin with no token saved. */
function entryMessage(e: Entry, tokens: TokenBook): Client.Join | Client.RejoinWithPIN | Client.RejoinWithToken | undefined {
  switch (e.kind) {
    case 'join':
      return { type: 'join', room: e.room, nickname: e.nickname, ...(e.hostKey && { hostKey: e.hostKey }) };
    case 'pin':
      return { type: 'rejoin', room: e.room, nickname: e.nickname, pin: e.pin };
    case 'token': {
      const token = tokens.get(e.room);
      return token ? { type: 'rejoin', room: e.room, token } : undefined;
    }
  }
}

/** The situation an error answers (docs/api/errors.md "How to show an error"). */
function situationOf(e: Entry | undefined): string | undefined {
  if (e?.kind === 'join') return 'join';
  if (e?.kind === 'token') return 'token';
  return undefined;
}

export class GameSocket {
  private readonly store: Store;
  private readonly tokens: TokenBook;
  private readonly env: SocketEnv;
  private ws: WebSocketLike | undefined;
  private isOpen = false;
  private joined = false;
  private stopped = false;
  /** The join/rejoin waiting for its welcome. */
  private pending: Entry | undefined;
  private room: string | undefined;
  private attempt = 0;
  private timer: unknown;

  constructor(store: Store, tokens = new TokenBook(), env: SocketEnv = browserEnv()) {
    this.store = store;
    this.tokens = tokens;
    this.env = env;
  }

  /**
   * Enters a room. A refused entry (wrong nickname, wrong PIN …) keeps the
   * socket open, so the join screen can call this again on the same line.
   */
  enter(entry: Entry): void {
    const room = roomKey(entry.room);
    if (room !== this.room || this.joined || this.timer !== undefined) this.close();
    if (room !== this.room) this.store.enterRoom(room);
    this.attempt = 0;
    this.room = room;
    this.stopped = false;
    this.pending = { ...entry, room };
    this.store.clearError();
    if (this.ws && this.isOpen) {
      this.store.setConnection({ state: 'connecting' });
      this.sendEntry();
    }
    else if (!this.ws) this.connect();
  }

  /** Sends a command, kick or transferHost; false when not joined or too big. */
  send(msg: Outgoing): boolean {
    if (!this.joined || !this.ws) return false;
    return this.write(msg);
  }

  /** Leaves the lobby for good (README "Joining and rejoining"); the token stays for history. */
  leave(): void {
    if (this.joined && this.ws) this.write({ type: 'leave' } satisfies Client.Leave);
    this.stop('left');
  }

  /** Tries again now instead of waiting (e.g. the phone came back online). */
  retryNow(): void {
    if (this.timer === undefined) return;
    this.env.clearTimeout(this.timer);
    this.timer = undefined;
    this.connect();
  }

  private connect(): void {
    this.store.setConnection({ state: 'connecting' });
    const ws = this.env.open(this.env.url);
    this.ws = ws;
    this.isOpen = false;
    ws.onopen = () => {
      if (ws !== this.ws) return;
      this.isOpen = true;
      this.sendEntry();
    };
    ws.onmessage = (ev) => {
      if (ws === this.ws) this.receive(ev.data);
    };
    ws.onclose = () => {
      if (ws === this.ws) this.dropped();
    };
    ws.onerror = () => {}; // a close always follows
  }

  private sendEntry(): void {
    const entry = this.pending;
    if (!entry) return;
    const msg = entryMessage(entry, this.tokens);
    if (msg) this.write(msg);
    else this.stop('badToken'); // nothing left to rejoin with
  }

  private write(msg: Client.ClientMessage): boolean {
    const data = JSON.stringify(msg);
    if (new TextEncoder().encode(data).length > maxMessageBytes) return false;
    this.ws?.send(data);
    return true;
  }

  private receive(data: unknown): void {
    let msg: Server.ServerMessage;
    try {
      msg = JSON.parse(String(data));
    } catch {
      return;
    }
    if (typeof msg !== 'object' || msg === null) return;
    switch (msg.type) {
      case 'welcome':
        return this.welcome(msg);
      case 'error':
        return this.error(msg);
      case 'view':
      case 'events':
        if (this.joined) this.applied(this.store.apply(msg));
        return;
      default:
        return; // a message type from a newer server
    }
  }

  private welcome(msg: Server.Welcome): void {
    if (msg.protocolVersion > protocolVersion) return this.stop('outdated');
    if (msg.token && this.room) this.tokens.save(this.room, msg.token);
    this.pending = undefined;
    this.joined = true;
    this.attempt = 0;
    this.store.apply(msg);
    this.store.setConnection({ state: 'joined' });
  }

  private error(msg: Server.ErrorMessage): void {
    const answering = this.pending;
    this.store.apply(msg, situationOf(answering));
    const room = this.room;
    switch (msg.code) {
      case 'replaced_by_other_device':
        if (room) this.tokens.forget(room); // the other phone holds the new token
        return this.stop('replaced');
      case 'kicked_by_host':
        return this.stop('kicked'); // the token still reads the games played before
    }
    if (answering && answering.kind !== 'token') {
      // A refused join or PIN rejoin: the socket stays open for the next try.
      this.store.setConnection({ state: 'idle' });
      return;
    }
    if (answering?.kind !== 'token') return;
    if (msg.code === 'bad_credentials') {
      if (room) this.tokens.forget(room);
      return this.stop('badToken');
    }
    if (msg.code === 'room_not_found') return this.stop('roomGone'); // token kept for history
    // Any other refusal of a token rejoin (rate_limited …): try again later.
    this.ws?.close();
  }

  private applied(result: Applied): void {
    // Events were missed: rejoin, which sends the whole list again.
    if (result === 'gap') this.ws?.close();
  }

  private dropped(): void {
    const rejoining = this.joined || this.pending?.kind === 'token';
    this.ws = undefined;
    this.isOpen = false;
    this.joined = false;
    if (this.stopped) return;
    const room = this.room;
    if (!rejoining || room === undefined || this.tokens.get(room) === undefined) {
      // Never got in (a refused join, or the server closed after 8 refusals):
      // nothing to rejoin with, the join screen tries again.
      this.pending = undefined;
      this.store.setConnection({ state: 'idle' });
      return;
    }
    this.pending = { kind: 'token', room };
    const delay = retryDelay(this.attempt++, this.env.random);
    this.store.setConnection({ state: 'waiting', retryInMs: delay });
    this.timer = this.env.setTimeout(() => {
      this.timer = undefined;
      this.connect();
    }, delay);
  }

  private stop(reason: StopReason): void {
    this.stopped = true;
    this.pending = undefined;
    this.close();
    this.store.setConnection({ state: 'stopped', reason });
    this.store.leaveRoom();
    this.room = undefined;
  }

  private close(): void {
    if (this.timer !== undefined) this.env.clearTimeout(this.timer);
    this.timer = undefined;
    const ws = this.ws;
    this.ws = undefined;
    this.isOpen = false;
    this.joined = false;
    ws?.close();
  }
}
