// Tạo phòng (giao-dien-3d.md §10): pick the room's mode (rules §1b) and a
// nickname. POST /rooms gives the code and the hostKey, kept in memory only
// and sent with this phone's first join (docs/api/README.md "Connecting").
// A refused join (a bad nickname) retries in the same room, never a new one.
import type { Ctx } from '../app/context.ts';
import { httpErrorText, stateErrorText } from '../app/messages.ts';
import { vi } from '../i18n/vi.ts';
import { button, frame, h, type Mounted } from '../ui/dom.ts';

type Mode = 'inPerson' | 'online';

export function createScreen(ctx: Ctx): Mounted {
  let mode: Mode = 'inPerson';
  let created: { room: string; hostKey: string } | undefined;
  let httpError = '';
  let sending = false;

  const nickname = h('input', { type: 'text', autocomplete: 'nickname', enterKeyHint: 'go' });
  const inPerson = button(vi.modeInPerson, () => pick('inPerson'));
  const online = button(vi.modeOnline, () => pick('online'));
  const hint = h('p', { class: 'muted' });
  const error = h('p', { class: 'error', role: 'alert' });
  const submit = button(vi.createRoom, () => void send(), 'primary');

  function pick(m: Mode): void {
    if (created) return; // the room exists: its mode is changed in the lobby
    mode = m;
    draw();
  }

  async function send(): Promise<void> {
    const name = nickname.value;
    httpError = '';
    if (!created) {
      sending = true;
      draw();
      const r = await ctx.api.createRoom(mode);
      sending = false;
      if (!r.ok) {
        httpError = httpErrorText(r.code, r.retryAfter);
        draw();
        return;
      }
      created = r.value;
    }
    ctx.socket.enter({ kind: 'join', room: created.room, nickname: name, hostKey: created.hostKey });
  }

  function draw(): void {
    inPerson.setAttribute('aria-pressed', String(mode === 'inPerson'));
    online.setAttribute('aria-pressed', String(mode === 'online'));
    inPerson.disabled = online.disabled = created !== undefined;
    hint.textContent = mode === 'online' ? vi.modeOnlineHint : vi.modeInPersonHint;
    const busy = sending || ctx.store.state.connection.state === 'connecting';
    submit.disabled = busy;
    submit.textContent = busy ? vi.connecting : vi.createRoom;
    error.textContent = httpError || stateErrorText(ctx.store.state);
  }

  nickname.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !submit.disabled) void send();
  });

  const el = h(
    'main',
    { class: 'screen' },
    frame(
      h('h2', {}, vi.createRoom),
      h('span', { class: 'muted' }, vi.modeLabel),
      h('div', { class: 'row' }, inPerson, online),
      hint,
      h('label', {}, vi.nickname, nickname),
      h('div', { class: 'row' }, button(vi.back, () => ctx.go({ name: 'home' })), submit),
      error,
    ),
  );
  draw();
  queueMicrotask(() => nickname.focus());
  return { el, update: draw };
}
