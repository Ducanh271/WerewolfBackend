// The host's settings panels turn into the contract's commands
// (docs/api/screens/lobby.md "What the phone may send", rules §2).
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import type { Server } from '../src/api/types.ts';
import { initialChoice, rolesCommand, stepRole, totalRoles } from '../src/settings/roles.ts';
import { initialOrder, seatsCommand, swapSeats } from '../src/settings/seats.ts';
import { formatDuration, stepTimer, timerFields, timersCommand, visibleTimerFields } from '../src/settings/timers.ts';

const examples = join(import.meta.dirname, '..', '..', 'docs', 'api', 'examples');
const read = <T,>(p: string): T => JSON.parse(readFileSync(join(examples, p), 'utf8')) as T;
const lobby = read<Server.ViewMessage>('server/view.lobby.host.json');

test('the timers panel covers every timer the server sends', () => {
  assert.deepEqual(timerFields.map((f) => f.key).sort(), Object.keys(lobby.view.timers!).sort());
});

test('configureTimers carries only what changed, in ms', () => {
  // The example lobby was captured after that command; start from the rules §2 defaults of the two.
  const current = { ...lobby.view.timers!, day: 420000, defense: 60000 };
  assert.equal(timersCommand(current, { ...current }), undefined);
  const want = read<{ args: { timers: Record<string, number> } }>('client/command.configureTimers.json');
  const edited = { ...current, ...want.args.timers };
  assert.deepEqual(timersCommand(current, edited), want);
});

test('timer steps stay inside the ranges of rules §2', () => {
  const field = (k: string) => timerFields.find((f) => f.key === k)!;
  assert.equal(stepTimer(field('nightStep'), 5000, -1), 5000);
  assert.equal(stepTimer(field('nightStep'), 600000, 1), 600000);
  assert.equal(stepTimer(field('day'), 60000, -1), 60000);
  assert.equal(stepTimer(field('day'), 1800000, 1), 1800000);
  assert.equal(stepTimer(field('announce'), 60000, 1), 60000);
  assert.equal(stepTimer(field('nominationLock'), 10000, 1), 15000);
});

test('an online room shows no phone check time', () => {
  assert.ok(visibleTimerFields('inPerson').some((f) => f.key === 'phoneCheck'));
  assert.ok(!visibleTimerFields('online').some((f) => f.key === 'phoneCheck'));
});

test('a countdown shows mm:ss and never ends early', () => {
  assert.equal(formatDuration(420000), '07:00');
  assert.equal(formatDuration(29001), '00:30');
  assert.equal(formatDuration(1), '00:01');
  assert.equal(formatDuration(0), '00:00');
});

test('configureRoles is the role list, counts for Villagers and Wolves, one of the rest', () => {
  let choice = initialChoice({});
  for (let i = 0; i < 8; i++) choice = stepRole(choice, 'Villager', 1);
  for (let i = 0; i < 4; i++) choice = stepRole(choice, 'Werewolf', 1);
  assert.deepEqual(rolesCommand(choice), read('client/command.configureRoles.json'));
  choice = stepRole(stepRole(choice, 'Seer', 1), 'Seer', 1);
  assert.equal(choice.Seer, 1);
  choice = stepRole(choice, 'Seer', -1);
  assert.equal('Seer' in choice, false);
  assert.equal(totalRoles(choice), 12);
});

test('the roles panel opens on the room list', () => {
  assert.deepEqual(initialChoice({ roles: { Villager: 2, Seer: 1, Guard: 0 } }), { Villager: 2, Seer: 1 });
});

test('setSeats lists everybody in the room once', () => {
  const people = lobby.people!;
  const order = initialOrder(people, []);
  assert.deepEqual(order, people.map((p) => p.id));
  const swapped = swapSeats(order, 0, 2);
  assert.equal(swapped[0], order[2]);
  assert.equal(swapped[2], order[0]);
  assert.deepEqual(seatsCommand(swapped, people)?.args.order, swapped);
  assert.equal(seatsCommand(order.slice(1), people), undefined, 'someone missing');
  assert.equal(seatsCommand([...order.slice(1), order[1]!], people), undefined, 'someone twice');
});

test('setSeats as the contract writes it, from the seats already set', () => {
  const want = read<{ args: { order: string[] } }>('client/command.setSeats.json');
  const people = want.args.order.map((id) => ({ id, nickname: id, connected: true, isHost: false }));
  const players = [...want.args.order].reverse().map((id, i) => ({ id, nickname: id, seat: want.args.order.length - i, alive: true }));
  assert.deepEqual(initialOrder(people, players), want.args.order);
  assert.deepEqual(seatsCommand(want.args.order, people), want);
  const someoneNew = [...people, { id: 'p99', nickname: 'x', connected: true, isHost: false }];
  assert.equal(initialOrder(someoneNew, players).at(-1), 'p99', 'seats no longer fit: join order');
});
