// "Xếp chỗ" (inPerson only; lobby.md "What the phone may send"): tap one
// seat, then another, and the two players swap. Opens on the seats already
// set, else the join order (settings/seats.ts).
import type { Ctx } from '../../app/context.ts';
import { fmt, vi } from '../../i18n/vi.ts';
import { initialOrder, seatsCommand, swapSeats } from '../../settings/seats.ts';
import type { State } from '../../state/store.ts';
import { button, fill, h, type Mounted } from '../../ui/dom.ts';
import { panelShell, type Send } from './shell.ts';

export function seatsPanel(ctx: Ctx, send: Send, close: () => void): Mounted {
  const start = ctx.store.state;
  let order = initialOrder(start.people ?? [], start.view?.players);
  let picked: number | undefined;
  let stale = false;
  const list = h('div', { class: 'grid2' });
  const note = h('p', { class: 'muted' }, vi.seatsHint);
  const save = button(vi.save, () => {
    const cmd = seatsCommand(order, ctx.store.state.people ?? []);
    if (!cmd) return;
    send(cmd);
    close();
  }, 'primary');

  function draw(s: State = ctx.store.state): void {
    const people = s.people ?? [];
    if (!seatsCommand(order, people)) {
      order = initialOrder(people, s.view?.players);
      picked = undefined;
      stale = true;
    }
    const names = new Map(people.map((p) => [p.id, p.nickname]));
    fill(
      list,
      ...order.map((id, i) =>
        button(
          `${fmt(vi.seatNumber, { n: i + 1 })} · ${names.get(id) ?? id}`,
          () => tap(i),
          'normal',
          { class: `seat${picked === i ? ' selected' : ''}`, 'aria-pressed': String(picked === i) },
        ),
      ),
    );
    note.textContent = stale ? vi.seatsStale : vi.seatsHint;
  }

  function tap(i: number): void {
    if (picked === undefined) picked = i;
    else {
      order = swapSeats(order, picked, i);
      picked = undefined;
    }
    draw();
  }

  draw();
  return { el: panelShell(vi.seatsButton, h('div', {}, note, list), [save], close), update: draw };
}
