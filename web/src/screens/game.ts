// Placeholder for every phase after the lobby: the game screens come in
// Bước 3b onwards. It shows the countdown on the server's clock
// (state/clock.ts remaining()) so the connection can be tried end to end.
import { vi, fmt } from '../i18n/vi.ts';
import { remaining } from '../state/clock.ts';
import type { State } from '../state/store.ts';
import { formatDuration } from '../settings/timers.ts';
import { frame, h, type Mounted } from '../ui/dom.ts';
import type { Ctx } from '../app/context.ts';

export function gameScreen(ctx: Ctx): Mounted {
  const clock = h('div', { class: 'clock' });
  const band = h('div', { class: 'band', hidden: true }, vi.reconnecting);
  const el = h('main', { class: 'screen' }, band, frame(h('h2', {}, vi.gameTitle), h('p', {}, vi.gameLater), clock));

  function tick(s: State = ctx.store.state): void {
    const left = remaining(s);
    clock.textContent = left === undefined ? '' : fmt(vi.timeLeft, { time: formatDuration(left) });
  }
  const timer = setInterval(() => tick(), 250);
  return {
    el,
    update(s) {
      band.hidden = s.connection.state === 'joined';
      tick(s);
    },
    destroy: () => clearInterval(timer),
  };
}
