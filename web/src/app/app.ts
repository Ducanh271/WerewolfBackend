// The app: mounts the screen pickScreen chooses and updates it from each new
// state (a screen is rebuilt only when it changes, so a half-typed form
// survives a status change). It also remembers the room being played
// (CurrentRoom) and moves to the right form when the phone stops.
import type { Api } from '../net/http.ts';
import type { CurrentRoom } from '../net/current.ts';
import type { GameSocket } from '../net/socket.ts';
import type { TokenBook } from '../net/tokens.ts';
import type { State, Store } from '../state/store.ts';
import type { Mounted } from '../ui/dom.ts';
import { createScreen } from '../screens/create.ts';
import { gameScreen } from '../screens/game.ts';
import { homeScreen } from '../screens/home.ts';
import { joinScreen } from '../screens/join.ts';
import { lobbyScreen } from '../screens/lobby.ts';
import { pinScreen } from '../screens/pin.ts';
import type { Ctx } from './context.ts';
import { afterStop, bootPlan } from './flow.ts';
import { httpErrorText } from './messages.ts';
import { pickScreen, type Route, type ScreenName } from './screen.ts';

export interface AppDeps {
  store: Store;
  socket: GameSocket;
  api: Api;
  tokens: TokenBook;
  current: CurrentRoom;
}

export class App {
  private readonly root: HTMLElement;
  private readonly deps: AppDeps;
  private readonly ctx: Ctx;
  private route: Route = { name: 'home' };
  private notice = ''; // a sentence for the join form (the QR code's room was not found)
  private mounted: { key: string; m: Mounted } | undefined;
  private lastRoom: string | undefined;
  private automatic = false;
  private prev: State['connection']['state'] = 'idle';

  constructor(root: HTMLElement, deps: AppDeps) {
    this.root = root;
    this.deps = deps;
    this.ctx = { store: deps.store, socket: deps.socket, api: deps.api, go: (r) => this.go(r) };
    deps.store.subscribe((s) => this.changed(s));
  }

  start(path: string): void {
    const plan = bootPlan(path, (r) => this.deps.tokens.get(r) !== undefined, this.deps.current.get());
    this.route = plan.route;
    if (plan.rejoin) {
      this.automatic = true;
      this.deps.socket.enter({ kind: 'token', room: plan.rejoin });
    }
    if (plan.check) {
      void this.deps.api.checkRoom(plan.check).then((r) => {
        if (r.ok || this.route.name !== 'join') return;
        this.notice = httpErrorText(r.code, r.retryAfter);
        this.render(true);
      });
    }
    this.render();
  }

  private go(route: Route): void {
    this.route = route;
    this.notice = '';
    this.automatic = false;
    this.deps.store.clearError();
    if (route.name === 'home' && location.pathname !== '/') history.replaceState(null, '', '/');
    this.render();
  }

  private changed(s: State): void {
    const now = s.connection.state;
    if (now === 'joined' && this.prev !== 'joined' && s.room) {
      this.deps.current.set(s.room);
      this.lastRoom = s.room;
      this.automatic = false;
    }
    if (s.connection.state === 'stopped' && this.prev !== 'stopped') {
      this.deps.current.clear();
      // The socket reports the stop before it drops the game, so s.room is still the room.
      const next = afterStop(s.connection.reason, s.room ?? this.lastRoom, this.automatic);
      this.automatic = false;
      this.prev = now;
      this.route = next.route;
      if (next.clearError) this.deps.store.clearError();
    }
    this.prev = now;
    this.render();
  }

  private render(force = false): void {
    const s = this.deps.store.state;
    const name = pickScreen(s, this.route);
    const key = name === 'join' && this.route.name === 'join' ? `join:${this.route.code ?? ''}` : name;
    if (force || this.mounted?.key !== key) {
      this.mounted?.m.destroy?.();
      const m = this.build(name);
      this.root.replaceChildren(m.el);
      this.mounted = { key, m };
    }
    this.mounted.m.update(s);
  }

  private build(name: ScreenName): Mounted {
    switch (name) {
      case 'home':
        return homeScreen(this.ctx);
      case 'create':
        return createScreen(this.ctx);
      case 'join':
        return joinScreen(this.ctx, this.route.name === 'join' ? (this.route.code ?? '') : '', this.notice);
      case 'pin':
        return pinScreen(this.ctx);
      case 'lobby':
        return lobbyScreen(this.ctx);
      case 'game':
        return gameScreen(this.ctx);
    }
  }
}
