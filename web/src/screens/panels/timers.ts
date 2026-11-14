// "Giờ & âm thanh": every phase time and the sound mode, drawn from
// view.timers / view.sound alone (lobby.md "What drives the lobby"). An
// online room hides the phone check time (giao-dien-3d.md §10). Saving sends
// configureTimers with what changed and configureSound if it changed.
import type { Server } from '../../api/types.ts';
import type { Ctx } from '../../app/context.ts';
import { vi } from '../../i18n/vi.ts';
import { formatDuration, stepTimer, timersCommand, visibleTimerFields } from '../../settings/timers.ts';
import { button, h, type Mounted } from '../../ui/dom.ts';
import { panelShell, stepper, type Send } from './shell.ts';

type Sound = NonNullable<Server.View['sound']>;

export function timersPanel(ctx: Ctx, send: Send, close: () => void): Mounted {
  const view = ctx.store.state.view;
  const current = view?.timers;
  const currentSound = view?.sound;
  const draft: Server.Timers | undefined = current && { ...current };
  let sound: Sound | undefined = currentSound;

  const fields = visibleTimerFields(view?.mode ?? 'inPerson');
  const rows = fields.map((f) => {
    const st = stepper(
      vi.timerLabel[f.key],
      () => change(f.key, -1),
      () => change(f.key, 1),
    );
    return { f, st };
  });
  const one = button(vi.soundOneSpeaker, () => pickSound('oneSpeaker'));
  const all = button(vi.soundAllPhones, () => pickSound('allPhones'));

  function change(key: keyof Server.Timers, dir: 1 | -1): void {
    const row = rows.find((r) => r.f.key === key);
    if (!draft || !row) return;
    draft[key] = stepTimer(row.f, draft[key], dir);
    draw();
  }
  function pickSound(m: Sound): void {
    sound = m;
    draw();
  }
  function draw(): void {
    for (const { f, st } of rows) {
      const v = draft?.[f.key];
      st.set(v === undefined ? '—' : formatDuration(v), v !== undefined && v > f.min, v !== undefined && v < f.max);
    }
    one.setAttribute('aria-pressed', String(sound === 'oneSpeaker'));
    all.setAttribute('aria-pressed', String(sound === 'allPhones'));
  }

  const save = button(vi.save, () => {
    const cmd = current && draft && timersCommand(current, draft);
    if (cmd) send(cmd);
    if (sound && sound !== currentSound) send({ type: 'command', name: 'configureSound', args: { mode: sound } });
    close();
  }, 'primary');

  const body = h(
    'div',
    {},
    h('span', { class: 'muted' }, vi.soundTitle),
    h('div', { class: 'row' }, one, all),
    h('span', { class: 'muted' }, vi.timersTitle),
    h('div', { class: 'grid2' }, ...rows.map((r) => r.st.el)),
  );
  draw();
  // The panel edits a draft; a new view does not overwrite what the host is changing.
  return { el: panelShell(vi.timersButton, body, [save], close), update: () => {} };
}
