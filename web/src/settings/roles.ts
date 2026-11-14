// The host's "Chọn vai" panel. The roles to pick from and their names come
// from GET /guide `roles[]` (docs/api/screens/guide.md), the current choice
// from view.roles. Rules §2 "Số lượng mỗi vai": Villagers and plain Wolves
// take a count, every other role is unique (on or off). The wolf limit and
// the player-count conditions are the server's to check: its error says what
// to fix (lobby.md "Errors").
import type { Client, Server } from '../api/types.ts';

/** Roles that take a count; the others are at most one. */
export const countedRoles: ReadonlySet<string> = new Set(['Villager', 'Werewolf']);

export type RoleChoice = Record<string, number>;

/** The host's choice as the panel opens: the room's current list. */
export function initialChoice(view: Pick<Server.View, 'roles'>): RoleChoice {
  const out: RoleChoice = {};
  for (const [role, n] of Object.entries(view.roles ?? {})) if (n) out[role] = n;
  return out;
}

/** One more (+1) or one fewer (−1) of a role: counted roles go 0…30, the others 0…1. */
export function stepRole(choice: RoleChoice, role: string, dir: 1 | -1): RoleChoice {
  const max = countedRoles.has(role) ? 30 : 1;
  const n = Math.min(max, Math.max(0, (choice[role] ?? 0) + dir));
  const next = { ...choice };
  if (n === 0) delete next[role];
  else next[role] = n;
  return next;
}

export function totalRoles(choice: RoleChoice): number {
  return Object.values(choice).reduce((a, b) => a + b, 0);
}

export function rolesCommand(choice: RoleChoice): Client.ConfigureRolesCommand {
  const roles: RoleChoice = {};
  for (const [role, n] of Object.entries(choice)) if (n > 0) roles[role] = n;
  return { type: 'command', name: 'configureRoles', args: { roles } };
}
