// The frame every lobby settings panel sits in (2D over the lobby,
// giao-dien-3d.md §10), and the − value + stepper the panels share.
import { vi } from '../../i18n/vi.ts';
import type { Outgoing } from '../../net/socket.ts';
import { button, h, type Child } from '../../ui/dom.ts';

/** Sends a command after clearing the last error, so a new refusal shows on its own. */
export type Send = (msg: Outgoing) => void;

export function panelShell(title: string, body: HTMLElement, footer: Child[], close: () => void): HTMLElement {
  return h(
    'div',
    {
      class: 'panel-back',
      role: 'dialog',
      'aria-modal': 'true',
      'aria-label': title,
      onkeydown: (e: KeyboardEvent) => e.key === 'Escape' && close(),
    },
    h(
      'div',
      { class: 'khung' },
      h('h2', {}, title),
      h('div', { class: 'panel-body' }, body),
      h('div', { class: 'row' }, button(vi.close, close), ...footer),
    ),
  );
}

export interface Stepper {
  el: HTMLElement;
  set(value: string, canLess: boolean, canMore: boolean): void;
}

export function stepper(label: string, onLess: () => void, onMore: () => void): Stepper {
  const value = h('span', { class: 'value' });
  const less = button('−', onLess, 'normal', { class: 'btn-small', 'aria-label': `${vi.fewer}: ${label}` });
  const more = button('+', onMore, 'normal', { class: 'btn-small', 'aria-label': `${vi.more}: ${label}` });
  const el = h('div', { class: 'stepper' }, h('span', {}, label), h('span', { class: 'row' }, less, value, more));
  return {
    el,
    set(v, canLess, canMore) {
      value.textContent = v;
      less.disabled = !canLess;
      more.disabled = !canMore;
    },
  };
}
