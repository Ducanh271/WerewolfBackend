// What every screen may use: the store to read, the socket and the HTTP API
// to act, and `go` to move between the forms before the room.
import type { Api } from '../net/http.ts';
import type { GameSocket } from '../net/socket.ts';
import type { Store } from '../state/store.ts';
import type { Route } from './screen.ts';

export interface Ctx {
  store: Store;
  socket: GameSocket;
  api: Api;
  go(route: Route): void;
}
