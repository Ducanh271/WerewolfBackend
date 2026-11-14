// "Chọn vai": the roles and their names from GET /guide roles[] (screens/
// guide.md), counts for Villagers and Wolves, on/off for the rest (rules §2
// "Số lượng mỗi vai"; settings/roles.ts). The server checks the wolf limit
// and the player counts and says what to fix.
import type { Http } from '../../api/types.ts';
import type { Ctx } from '../../app/context.ts';
import { httpErrorText } from '../../app/messages.ts';
import { fmt, vi } from '../../i18n/vi.ts';
import { countedRoles, initialChoice, rolesCommand, stepRole, totalRoles } from '../../settings/roles.ts';
import type { State } from '../../state/store.ts';
import { button, fill, h, type Mounted } from '../../ui/dom.ts';
import { panelShell, stepper, type Send, type Stepper } from './shell.ts';

export function rolesPanel(ctx: Ctx, send: Send, close: () => void): Mounted {
  let choice = initialChoice(ctx.store.state.view ?? {});
  let rows: { role: string; step: Stepper }[] = [];
  const list = h('div', { class: 'grid2' });
  const total = h('p', { class: 'muted' });
  const status = h('p', { class: 'muted' }, vi.rolesLoading);
  const save = button(vi.save, () => {
    send(rolesCommand(choice));
    close();
  }, 'primary', { disabled: true });

  function draw(s: State = ctx.store.state): void {
    for (const { role, step } of rows) {
      const n = choice[role] ?? 0;
      step.set(String(n), n > 0, n < (countedRoles.has(role) ? 30 : 1));
    }
    total.textContent = fmt(vi.rolesTotal, { chosen: totalRoles(choice), n: (s.people ?? []).length });
  }

  function show(roles: Http.GuideEntry[]): void {
    rows = roles.map((r) => ({
      role: r.role,
      step: stepper(
        r.title,
        () => ((choice = stepRole(choice, r.role, -1)), draw()),
        () => ((choice = stepRole(choice, r.role, 1)), draw()),
      ),
    }));
    fill(list, ...rows.map((r) => r.step.el));
    status.hidden = true;
    save.disabled = false;
    draw();
  }

  void ctx.api.guide().then((r) => {
    if (r.ok) show(r.value.roles);
    else status.textContent = httpErrorText(r.code, r.retryAfter);
  });
  draw();
  return { el: panelShell(vi.rolesButton, h('div', {}, status, list, total), [save], close), update: draw };
}
