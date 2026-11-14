// The HTTP side of the contract (docs/api/README.md "Connecting"): create a
// room, check a room code, read the rules guide. An HTTP error has the body
// of a WebSocket error (`HTTPError`); a request that never got an answer is
// the app's own `network` code (shown with vi.network, not errors.md).
import type { Http } from '../api/types.ts';

export type HttpResult<T> = { ok: true; value: T } | { ok: false; code: string; retryAfter?: number };

/** No answer from the server at all (offline, server down). */
export const networkError = 'network';

export type FetchLike = (url: string, init?: { method?: string; headers?: Record<string, string>; body?: string }) => Promise<{
  ok: boolean;
  json(): Promise<unknown>;
}>;

async function call<T>(fetchFn: FetchLike, url: string, init?: Parameters<FetchLike>[1]): Promise<HttpResult<T>> {
  let res: Awaited<ReturnType<FetchLike>>;
  try {
    res = await fetchFn(url, init);
  } catch {
    return { ok: false, code: networkError };
  }
  let body: unknown;
  try {
    body = await res.json();
  } catch {
    body = undefined;
  }
  if (res.ok && body !== undefined) return { ok: true, value: body as T };
  const err = body as Partial<Http.HTTPError> | undefined;
  if (err && typeof err.code === 'string') {
    return { ok: false, code: err.code, ...(typeof err.retryAfter === 'number' && { retryAfter: err.retryAfter }) };
  }
  return { ok: false, code: networkError };
}

export class Api {
  private readonly fetchFn: FetchLike;
  private guideCache: Promise<HttpResult<Http.GuideReply>> | undefined;

  constructor(fetchFn: FetchLike = (url, init) => fetch(url, init)) {
    this.fetchFn = fetchFn;
  }

  /** POST /rooms: the code and the hostKey (keep it in memory only, README "Connecting"). */
  createRoom(mode: 'inPerson' | 'online'): Promise<HttpResult<Http.CreateRoomReply>> {
    const body: Http.CreateRoomRequest = { mode };
    return call(this.fetchFn, '/rooms', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  }

  /** GET /join/{code}: the room's code when it exists, else room_not_found. */
  checkRoom(code: string): Promise<HttpResult<Http.JoinRoomReply>> {
    return call(this.fetchFn, `/join/${encodeURIComponent(code)}`);
  }

  /** GET /guide, fetched once and kept (docs/api/screens/guide.md); a failure is not kept. */
  guide(): Promise<HttpResult<Http.GuideReply>> {
    if (!this.guideCache) {
      this.guideCache = call<Http.GuideReply>(this.fetchFn, '/guide').then((r) => {
        if (!r.ok) this.guideCache = undefined;
        return r;
      });
    }
    return this.guideCache;
  }
}
