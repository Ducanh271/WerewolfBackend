// A tiny DOM builder for the 2D screens (no framework: web/CLAUDE.md "Stack").
// Text always goes in as text nodes, never as HTML (secrecy.md "Chat text is
// untrusted"; nicknames are typed by players too).
import type { State } from '../state/store.ts';

export type Child = Node | string | number | null | undefined | false;
type Props = Record<string, unknown>;

export function h<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  props: Props = {},
  ...children: Child[]
): HTMLElementTagNameMap[K] {
  const el = document.createElement(tag);
  for (const [key, value] of Object.entries(props)) {
    if (value === undefined || value === false) continue;
    if (key === 'class') el.className = String(value);
    else if (key.startsWith('on') && typeof value === 'function') {
      el.addEventListener(key.slice(2).toLowerCase(), value as EventListener);
    } else if (key in el && key !== 'list') (el as unknown as Props)[key] = value;
    else el.setAttribute(key, value === true ? '' : String(value));
  }
  append(el, children);
  return el;
}

export function append(el: Element, children: Child[]): void {
  for (const c of children) {
    if (c === null || c === undefined || c === false) continue;
    el.append(typeof c === 'string' || typeof c === 'number' ? document.createTextNode(String(c)) : c);
  }
}

/** Replaces an element's children. */
export function fill(el: Element, ...children: Child[]): void {
  el.replaceChildren();
  append(el, children);
}

export type ButtonKind = 'normal' | 'primary' | 'danger';

/** A button of ban-giao-frontend.md §4 "Nút". */
export function button(label: string, onClick: () => void, kind: ButtonKind = 'normal', props: Props = {}): HTMLButtonElement {
  const cls = ['btn', kind === 'primary' && 'btn-primary', kind === 'danger' && 'btn-danger', props.class]
    .filter(Boolean)
    .join(' ');
  return h('button', { type: 'button', ...props, class: cls, onclick: onClick }, label);
}

/** A framed panel (ban-giao-frontend.md §4 "Khung"). */
export function frame(...children: Child[]): HTMLElement {
  return h('div', { class: 'khung' }, ...children);
}

/** One mounted screen: built once, then updated from each new state. */
export interface Mounted {
  el: HTMLElement;
  update(s: State): void;
  destroy?(): void;
}
