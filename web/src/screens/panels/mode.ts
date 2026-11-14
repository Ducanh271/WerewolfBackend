// "Đổi chế độ" (rules §1b, lobby.md configureMode): a real change puts the
// sound back to the new mode's default and drops the seats; picking the
// current mode sends nothing.
import type { Ctx } from '../../app/context.ts';
import { vi } from '../../i18n/vi.ts';
import type { State } from '../../state/store.ts';
import { button, h, type Mounted } from '../../ui/dom.ts';
import { panelShell, type Send } from './shell.ts';

export function modePanel(ctx: Ctx, send: Send, close: () => void): Mounted {
  const pick = (mode: 'inPerson' | 'online') => () => {
    if (ctx.store.state.view?.mode !== mode) send({ type: 'command', name: 'configureMode', args: { mode } });
    close();
  };
  const inPerson = button(vi.modeInPerson, pick('inPerson'));
  const online = button(vi.modeOnline, pick('online'));
  function draw(s: State = ctx.store.state): void {
    inPerson.setAttribute('aria-pressed', String(s.view?.mode === 'inPerson'));
    online.setAttribute('aria-pressed', String(s.view?.mode === 'online'));
  }
  const body = h(
    'div',
    {},
    h('div', { class: 'grid2' }, inPerson, online, h('span', { class: 'muted' }, vi.modeInPersonHint), h('span', { class: 'muted' }, vi.modeOnlineHint)),
    h('p', { class: 'muted' }, vi.modeChangeNote),
  );
  draw();
  return { el: panelShell(vi.modeButton, body, [], close), update: draw };
}
