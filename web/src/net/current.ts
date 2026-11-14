// CurrentRoom remembers the room this phone is playing in, so reopening the
// app at "/" rejoins it with the saved token (TokenBook). It is forgotten
// when the phone stops (left, kicked, replaced, refused token, room gone).
// Storage that is missing or throws only loses the convenience.
import { roomKey, type KeyValueStore } from './tokens.ts';

const storageKey = 'masoi.current';

function browserStorage(): KeyValueStore | undefined {
  try {
    return globalThis.localStorage;
  } catch {
    return undefined;
  }
}

export class CurrentRoom {
  private readonly storage: () => (KeyValueStore & { removeItem?(key: string): void }) | undefined;

  constructor(storage: () => (KeyValueStore & { removeItem?(key: string): void }) | undefined = browserStorage) {
    this.storage = storage;
  }

  get(): string | undefined {
    try {
      return this.storage()?.getItem(storageKey) || undefined;
    } catch {
      return undefined;
    }
  }

  set(room: string): void {
    try {
      this.storage()?.setItem(storageKey, roomKey(room));
    } catch {
      // only the convenience is lost
    }
  }

  clear(): void {
    try {
      const s = this.storage();
      if (s?.removeItem) s.removeItem(storageKey);
      else s?.setItem(storageKey, '');
    } catch {
      // only the convenience is lost
    }
  }
}

/** The room code in a /j/{code} address (docs/api/README.md "Connecting"), if any. */
export function codeFromPath(path: string): string | undefined {
  const m = /^\/j\/([^/?#]+)\/?$/.exec(path);
  if (!m?.[1]) return undefined;
  try {
    return roomKey(decodeURIComponent(m[1]));
  } catch {
    return undefined; // a malformed address
  }
}
