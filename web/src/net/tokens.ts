// TokenBook keeps the rejoin token of each room this phone joined
// (docs/api/README.md "Joining and rejoining"), newest room first. A token
// stays after the player leaves the room: it still reads that room's
// finished games (docs/api/screens/history.md "Asking"), which takes at most
// 20 tokens, so the book keeps the 20 newest rooms. It is held in memory and
// mirrored to localStorage; storage that is missing or throws (a private
// window, blocked site data) only loses the copy, never the game.

/** The part of localStorage the book uses. */
export interface KeyValueStore {
  getItem(key: string): string | null;
  setItem(key: string, value: string): void;
}

interface Entry {
  room: string;
  token: string;
}

const storageKey = 'masoi.tokens';
export const maxRooms = 20;

/** Room codes are case-insensitive (docs/api/README.md "Connecting"). */
export function roomKey(room: string): string {
  return room.trim().toUpperCase();
}

function browserStorage(): KeyValueStore | undefined {
  try {
    return globalThis.localStorage;
  } catch {
    return undefined;
  }
}

export class TokenBook {
  private entries: Entry[];
  private readonly storage: () => KeyValueStore | undefined;

  constructor(storage: () => KeyValueStore | undefined = browserStorage) {
    this.storage = storage;
    this.entries = this.load();
  }

  /** The token saved for a room, if any. */
  get(room: string): string | undefined {
    const key = roomKey(room);
    return this.entries.find((e) => e.room === key)?.token;
  }

  /** Saves the newest token of a room; an older one stops working on the server. */
  save(room: string, token: string): void {
    const key = roomKey(room);
    this.entries = [{ room: key, token }, ...this.entries.filter((e) => e.room !== key)].slice(0, maxRooms);
    this.store();
  }

  /** Drops a room's token once the server no longer takes it. */
  forget(room: string): void {
    const key = roomKey(room);
    this.entries = this.entries.filter((e) => e.room !== key);
    this.store();
  }

  /** Every saved token, newest room first (the body of POST /history). */
  all(): string[] {
    return this.entries.map((e) => e.token);
  }

  private load(): Entry[] {
    try {
      const raw = this.storage()?.getItem(storageKey);
      if (!raw) return [];
      const parsed: unknown = JSON.parse(raw);
      if (!Array.isArray(parsed)) return [];
      return parsed
        .filter(
          (e): e is Entry =>
            typeof e === 'object' && e !== null && typeof e.room === 'string' && typeof e.token === 'string',
        )
        .slice(0, maxRooms);
    } catch {
      return [];
    }
  }

  private store(): void {
    try {
      this.storage()?.setItem(storageKey, JSON.stringify(this.entries));
    } catch {
      // Only the copy is lost; the book in memory still serves this visit.
    }
  }
}
