// Entry of the web game: the store, the saved tokens, the socket and the
// HTTP API, and the app that draws the screens from them.
import './ui/styles.css';
import { App } from './app/app.ts';
import { vi } from './i18n/vi.ts';
import { CurrentRoom } from './net/current.ts';
import { Api } from './net/http.ts';
import { GameSocket } from './net/socket.ts';
import { TokenBook } from './net/tokens.ts';
import { Store } from './state/store.ts';
import { installAssetVars } from './ui/assets.ts';

const root = document.getElementById('app');
if (root) {
  installAssetVars(document.documentElement);
  const rotate = document.createElement('div');
  rotate.id = 'rotate';
  rotate.textContent = vi.rotatePhone;
  document.body.append(rotate);

  const store = new Store();
  const tokens = new TokenBook();
  const socket = new GameSocket(store, tokens);
  const app = new App(root, { store, socket, api: new Api(), tokens, current: new CurrentRoom() });
  app.start(location.pathname);

  // Back online or back on the page: rejoin now instead of waiting out the delay.
  window.addEventListener('online', () => socket.retryNow());
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') socket.retryNow();
  });
}
