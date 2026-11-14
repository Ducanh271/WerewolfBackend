// Trang đầu (giao-dien-3d.md §10 "Trước sảnh"): create a room or join one.
import type { Ctx } from '../app/context.ts';
import { stateErrorText } from '../app/messages.ts';
import { vi } from '../i18n/vi.ts';
import { button, frame, h, type Mounted } from '../ui/dom.ts';

export function homeScreen(ctx: Ctx): Mounted {
  const error = h('p', { class: 'error', role: 'alert' });
  const status = h('p', { class: 'muted' });
  const el = h(
    'main',
    { class: 'screen' },
    frame(
      h('h1', {}, vi.appTitle),
      h(
        'div',
        { class: 'row' },
        button(vi.createRoom, () => ctx.go({ name: 'create' }), 'primary'),
        button(vi.joinRoom, () => ctx.go({ name: 'join' })),
      ),
      status,
      error,
    ),
  );
  return {
    el,
    update(s) {
      status.textContent = s.connection.state === 'connecting' ? vi.connecting : '';
      error.textContent = stateErrorText(s);
    },
  };
}
