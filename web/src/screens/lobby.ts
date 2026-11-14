// A temporary lobby to try the connection end to end; the real lobby is 3D
// (Bước 3b, giao-dien-3d.md §10). Drawn from the latest view only
// (docs/api/screens/lobby.md "What drives the lobby"): `people` for the list,
// view.me.id + people[].isHost for the host's controls, view.mode,
// view.roleCards/roles for the role list. The settings panels are 2D.
import type { Ctx } from '../app/context.ts';
import { stateErrorText } from '../app/messages.ts';
import { isHost } from '../app/screen.ts';
import { fmt, vi } from '../i18n/vi.ts';
import type { State } from '../state/store.ts';
import { button, fill, frame, h, type Mounted } from '../ui/dom.ts';
import { modePanel } from './panels/mode.ts';
import { rolesPanel } from './panels/roles.ts';
import { seatsPanel } from './panels/seats.ts';
import type { Send } from './panels/shell.ts';
import { timersPanel } from './panels/timers.ts';

type Panel = (ctx: Ctx, send: Send, close: () => void) => Mounted;

export function lobbyScreen(ctx: Ctx): Mounted {
  let panel: Mounted | undefined;
  let panelInPersonOnly = false; // the seats panel closes if the room turns online
  const send: Send = (msg) => {
    ctx.store.clearError();
    ctx.socket.send(msg);
  };

  const band = h('div', { class: 'band', hidden: true }, vi.reconnecting);
  const room = h('span', { class: 'muted' });
  const count = h('span', { class: 'muted' });
  const mode = h('span', { class: 'muted' });
  const list = h('ul', { class: 'people' });
  const chips = h('div', { class: 'chips' });
  const controls = h('div', { class: 'row' });
  const error = h('p', { class: 'error', role: 'alert' });

  function open(make: Panel, inPersonOnly = false): void {
    closePanel();
    panelInPersonOnly = inPersonOnly;
    panel = make(ctx, send, closePanel);
    el.append(panel.el);
  }
  function closePanel(): void {
    panel?.destroy?.();
    panel?.el.remove();
    panel = undefined;
  }

  function drawPeople(s: State, host: boolean): void {
    const me = s.view?.me?.id;
    fill(
      list,
      ...(s.people ?? []).map((p) =>
        h(
          'li',
          {},
          h('span', { class: 'name' }, p.nickname),
          p.isHost && h('span', { class: 'tag host' }, vi.hostTag),
          p.id === me && h('span', { class: 'tag' }, vi.youTag),
          !p.connected && h('span', { class: 'tag' }, vi.offlineTag),
          host && p.id !== me && button(vi.makeHost, () => ctx.socket.send({ type: 'transferHost', target: p.id }), 'normal', { class: 'btn-small' }),
          host && p.id !== me && button(vi.kick, () => ctx.socket.send({ type: 'kick', target: p.id }), 'danger', { class: 'btn-small' }),
        ),
      ),
    );
  }

  function drawRoles(s: State): void {
    const cards = s.view?.roleCards ?? [];
    if (cards.length === 0) {
      fill(chips, h('span', { class: 'muted' }, vi.noRolesYet));
      return;
    }
    fill(
      chips,
      ...cards.map((c) => h('span', { class: 'chip' }, fmt(vi.roleCount, { name: c.name, n: s.view?.roles?.[c.role] ?? 1 }))),
    );
  }

  function drawControls(s: State, host: boolean): void {
    if (!host) {
      fill(controls, h('span', { class: 'muted' }, vi.waitForHost));
      return;
    }
    const inPerson = s.view?.mode === 'inPerson';
    fill(
      controls,
      inPerson && button(vi.seatsButton, () => open(seatsPanel, true)),
      button(vi.rolesButton, () => open(rolesPanel)),
      button(vi.timersButton, () => open(timersPanel)),
      button(vi.modeButton, () => open(modePanel)),
      button(vi.deal, () => send({ type: 'command', name: 'dealRoles' }), 'primary'),
    );
  }

  function update(s: State): void {
    const host = isHost(s);
    band.hidden = s.connection.state === 'joined';
    room.textContent = fmt(vi.roomIs, { code: s.room ?? '' });
    count.textContent = fmt(vi.peopleCount, { n: (s.people ?? []).length });
    mode.textContent = fmt(vi.modeIs, { mode: s.view?.mode === 'online' ? vi.modeOnline : vi.modeInPerson });
    drawPeople(s, host);
    drawRoles(s);
    drawControls(s, host);
    error.textContent = stateErrorText(s);
    if (panel && (!host || (panelInPersonOnly && s.view?.mode !== 'inPerson'))) closePanel();
    panel?.update(s);
  }

  const el = h(
    'main',
    { class: 'screen' },
    band,
    h(
      'div',
      { class: 'lobby' },
      frame(h('div', { class: 'row' }, h('h2', {}, vi.lobbyTitle), room, count, mode), list),
      frame(
        h('span', { class: 'title' }, vi.rolesInGame),
        chips,
        controls,
        button(vi.leave, () => ctx.socket.leave(), 'danger'),
        error,
      ),
    ),
  );
  return { el, update, destroy: closePanel };
}
