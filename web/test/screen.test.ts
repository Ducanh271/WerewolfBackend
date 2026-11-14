// Which screen shows, from the store and the form the player is on, and
// where the app starts and goes after a stop (src/app/screen.ts, flow.ts).
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import type { Server } from '../src/api/types.ts';
import { afterStop, bootPlan } from '../src/app/flow.ts';
import { stateErrorText } from '../src/app/messages.ts';
import { isHost, pickScreen, screenOfPhase } from '../src/app/screen.ts';
import { codeFromPath } from '../src/net/current.ts';
import { Store } from '../src/state/store.ts';

const folder = join(import.meta.dirname, '..', '..', 'docs', 'api', 'examples', 'server');
const example = <T,>(name: string): T => JSON.parse(readFileSync(join(folder, name), 'utf8')) as T;
const views = readdirSync(folder).filter((f) => f.startsWith('view.'));

function joinedWith(view: string): Store {
  const store = new Store(() => 0);
  store.setConnection({ state: 'joined' });
  store.apply(example('welcome.rejoin-token.json'));
  store.apply(example(view));
  return store;
}

test('every example view picks the lobby or the game', () => {
  for (const f of views) {
    const s = joinedWith(f).state;
    const want = s.view!.phase === 'Lobby' ? 'lobby' : 'game';
    assert.equal(pickScreen(s, { name: 'home' }), want, f);
  }
});

test('only the Lobby phase is the lobby', () => {
  assert.equal(screenOfPhase('Lobby'), 'lobby');
  for (const p of ['RoleView', 'Night', 'Day', 'Ended', 'NominationLock'] as const) assert.equal(screenOfPhase(p), 'game');
});

test('the PIN comes before the lobby, until the player dismisses it', () => {
  const store = joinedWith('view.lobby.json');
  store.apply(example('welcome.join.json'));
  assert.equal(pickScreen(store.state, { name: 'join' }), 'pin');
  store.clearPin();
  assert.equal(pickScreen(store.state, { name: 'join' }), 'lobby');
});

test('a PIN still on screen never hides the game once the roles are dealt', () => {
  const store = joinedWith('view.roleView.json');
  store.apply(example('welcome.join.json'));
  assert.equal(pickScreen(store.state, { name: 'home' }), 'game');
});

test('without a room the form the player is on shows', () => {
  const store = new Store();
  assert.equal(pickScreen(store.state, { name: 'create' }), 'create');
  store.setConnection({ state: 'connecting' });
  assert.equal(pickScreen(store.state, { name: 'join', code: 'K7QP' }), 'join');
});

test('a dropped phone keeps its screen while it rejoins', () => {
  const store = joinedWith('view.night.asleep.json');
  store.setConnection({ state: 'waiting', retryInMs: 1000 });
  assert.equal(pickScreen(store.state, { name: 'home' }), 'game');
  store.setConnection({ state: 'stopped', reason: 'kicked' });
  assert.equal(pickScreen(store.state, { name: 'join' }), 'join');
});

test('the host controls show for the host only', () => {
  const host = joinedWith('view.lobby.host.json').state;
  assert.equal(isHost(host), true);
  const other = joinedWith('view.lobby.json').state;
  assert.equal(isHost(other), other.people!.some((p) => p.id === other.view!.me!.id && p.isHost));
});

test('/j/{code} reads the code, case ignored', () => {
  assert.equal(codeFromPath('/j/k7qp'), 'K7QP');
  assert.equal(codeFromPath('/j/K7QP/'), 'K7QP');
  assert.equal(codeFromPath('/'), undefined);
  assert.equal(codeFromPath('/j/%E0%A4%A'), undefined);
});

test('start: a QR link rejoins with its token, else checks the code', () => {
  const has = (r: string) => r === 'K7QP';
  assert.deepEqual(bootPlan('/j/k7qp', has, undefined), { route: { name: 'join', code: 'K7QP' }, rejoin: 'K7QP' });
  assert.deepEqual(bootPlan('/j/ABCD', has, 'K7QP'), { route: { name: 'join', code: 'ABCD' }, check: 'ABCD' });
});

test('start: "/" rejoins the room being played, if its token is kept', () => {
  const has = (r: string) => r === 'K7QP';
  assert.deepEqual(bootPlan('/', has, 'K7QP'), { route: { name: 'home' }, rejoin: 'K7QP' });
  assert.deepEqual(bootPlan('/', has, 'ABCD'), { route: { name: 'home' } });
  assert.deepEqual(bootPlan('/', has, undefined), { route: { name: 'home' } });
});

test('after a stop: leave goes home, a silent expired room too, the rest to the join form', () => {
  assert.deepEqual(afterStop('left', 'K7QP', false), { route: { name: 'home' }, clearError: false });
  assert.deepEqual(afterStop('roomGone', 'K7QP', true), { route: { name: 'home' }, clearError: true });
  assert.deepEqual(afterStop('roomGone', 'K7QP', false), { route: { name: 'join', code: 'K7QP' }, clearError: false });
  for (const r of ['kicked', 'replaced', 'badToken'] as const) {
    assert.deepEqual(afterStop(r, 'K7QP', true), { route: { name: 'join', code: 'K7QP' }, clearError: false }, r);
  }
});

test('the sentence shown: the error code, or the app being too old', () => {
  const store = new Store();
  store.apply(example<Server.ErrorMessage>('error.kicked_by_host.json'));
  assert.equal(stateErrorText(store.state), 'Chủ phòng đã mời bạn ra khỏi phòng.');
  store.apply(example<Server.ErrorMessage>('error.wrong_phase.json'), 'join');
  assert.match(stateErrorText(store.state), /Ván đã bắt đầu/);
  store.clearError();
  store.setConnection({ state: 'stopped', reason: 'outdated' });
  assert.match(stateErrorText(store.state), /cập nhật/);
});
