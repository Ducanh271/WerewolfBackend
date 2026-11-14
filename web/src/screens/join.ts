// Vào phòng (docs/api/screens/lobby.md "Join screen"): room code (from the
// QR link /j/{code} or typed), nickname, and "I was already in this room"
// for a rejoin by nickname + PIN from another phone.
import type { Ctx } from '../app/context.ts';
import { stateErrorText } from '../app/messages.ts';
import { vi } from '../i18n/vi.ts';
import { roomKey } from '../net/tokens.ts';
import { button, frame, h, type Mounted } from '../ui/dom.ts';

export function joinScreen(ctx: Ctx, code = '', notice = ''): Mounted {
  let shownNotice = notice;
  const room = h('input', {
    type: 'text',
    value: code,
    autocapitalize: 'characters',
    autocomplete: 'off',
    spellcheck: false,
    maxLength: 4,
  });
  const nickname = h('input', { type: 'text', autocomplete: 'nickname' });
  const already = h('input', { type: 'checkbox' });
  const pin = h('input', { type: 'tel', inputMode: 'numeric', autocomplete: 'off', maxLength: 4 });
  const pinField = h('label', {}, vi.pin, pin);
  const error = h('p', { class: 'error', role: 'alert' });
  const submit = button(vi.joinRoom, send, 'primary');

  function send(): void {
    shownNotice = '';
    const r = roomKey(room.value);
    if (already.checked) ctx.socket.enter({ kind: 'pin', room: r, nickname: nickname.value, pin: pin.value.trim() });
    else ctx.socket.enter({ kind: 'join', room: r, nickname: nickname.value });
  }

  function draw(): void {
    pinField.hidden = !already.checked;
    const busy = ctx.store.state.connection.state === 'connecting';
    submit.disabled = busy;
    submit.textContent = busy ? vi.connecting : already.checked ? vi.rejoin : vi.joinRoom;
    error.textContent = shownNotice || stateErrorText(ctx.store.state);
  }

  already.addEventListener('change', draw);
  for (const input of [room, nickname, pin]) {
    input.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' && !submit.disabled) send();
    });
  }

  const el = h(
    'main',
    { class: 'screen' },
    frame(
      h('h2', {}, vi.joinRoom),
      h('div', { class: 'form' }, h('label', {}, vi.roomCode, room), h('label', {}, vi.nickname, nickname)),
      h('div', { class: 'row' }, h('label', { class: 'check' }, already, vi.alreadyInRoom), pinField),
      h('div', { class: 'row' }, button(vi.back, () => ctx.go({ name: 'home' })), submit),
      error,
    ),
  );
  draw();
  queueMicrotask(() => (code ? nickname : room).focus());
  return { el, update: draw };
}
