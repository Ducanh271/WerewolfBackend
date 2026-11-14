// The host's "Xếp chỗ" panel (inPerson only: lobby.md "What the phone may
// send"): the circle clockwise, seat 1 first. Tap one player, then another,
// and they swap. setSeats must list everybody in the room exactly once.
import type { Client, Server } from '../api/types.ts';

/**
 * The order the panel opens with: the seats already set when they still
 * hold exactly the people in the room, else the join order.
 */
export function initialOrder(people: readonly Server.Person[], players: readonly Server.PlayerView[] = []): string[] {
  const ids = people.map((p) => p.id);
  const seated = [...players].sort((a, b) => a.seat - b.seat).map((p) => p.id);
  const same = seated.length === ids.length && seated.every((id) => ids.includes(id));
  return same ? seated : ids;
}

export function swapSeats(order: readonly string[], a: number, b: number): string[] {
  const next = [...order];
  const x = next[a];
  const y = next[b];
  if (x === undefined || y === undefined) return next;
  next[a] = y;
  next[b] = x;
  return next;
}

/** setSeats for the order, or undefined when it no longer lists the room exactly (someone came or left). */
export function seatsCommand(
  order: readonly string[],
  people: readonly Server.Person[],
): Client.SetSeatsCommand | undefined {
  const ids = new Set(people.map((p) => p.id));
  const unique = new Set(order);
  if (unique.size !== order.length || order.length !== ids.size || !order.every((id) => ids.has(id))) return undefined;
  return { type: 'command', name: 'setSeats', args: { order: [...order] } };
}
