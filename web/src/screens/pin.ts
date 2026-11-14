// The PIN, shown once after a join (docs/api/screens/lobby.md "Join
// screen"; rules §2 "Vào lại phòng"). It lives only in memory and is gone
// once the player taps that they remember it.
import type { Ctx } from '../app/context.ts';
import { vi } from '../i18n/vi.ts';
import { button, frame, h, type Mounted } from '../ui/dom.ts';

export function pinScreen(ctx: Ctx): Mounted {
  const code = h('div', { class: 'big-code' });
  const el = h(
    'main',
    { class: 'screen' },
    frame(
      h('h2', {}, vi.pinTitle),
      code,
      h('p', {}, vi.pinNote),
      h('div', { class: 'row' }, button(vi.pinDone, () => ctx.store.clearPin(), 'primary')),
    ),
  );
  return {
    el,
    update(s) {
      code.textContent = s.pin ?? '';
    },
  };
}
