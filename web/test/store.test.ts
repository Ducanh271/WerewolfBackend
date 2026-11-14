// Replays the contract's real server messages (../docs/api/examples/server/)
// into the Store: each view replaces the last one whole, events follow
// `reset` and `from` (docs/api/README.md "Server → client").
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import type { Server } from '../src/api/types.ts';
import { Store } from '../src/state/store.ts';

const folder = join(import.meta.dirname, '..', '..', 'docs', 'api', 'examples', 'server');

function example<T = Server.ServerMessage>(name: string): T {
  return JSON.parse(readFileSync(join(folder, name), 'utf8')) as T;
}

const all = readdirSync(folder).filter((f) => f.endsWith('.json'));
const views = all.filter((f) => f.startsWith('view.'));
const filler = (n: number): Server.Event[] =>
  Array.from({ length: n }, (_, i) => ({ type: 'PlayerJoined', data: { player: `p${i + 1}`, nickname: `N${i + 1}` } }));

test('every server example applies without error', () => {
  const store = new Store();
  for (const f of all) store.apply(example(f));
});

test('each view replaces the last one with its deadline, speaker and people', () => {
  const store = new Store();
  for (const prev of views) {
    for (const next of views) {
      store.apply(example(prev));
      const msg = example<Server.ViewMessage>(next);
      store.apply(msg);
      const s = store.state;
      assert.deepEqual(s.view, msg.view, `${prev} → ${next}`);
      assert.equal(s.deadline, msg.deadline, `${prev} → ${next}: deadline`);
      assert.equal(s.speaker, msg.speaker === true, `${prev} → ${next}: speaker`);
      assert.deepEqual(s.people, msg.people, `${prev} → ${next}: people`);
    }
  }
});

test('the view keeps the room mode of the message', () => {
  const store = new Store();
  store.apply(example('view.lobby.online.json'));
  assert.equal(store.state.view?.mode, 'online');
  store.apply(example('view.lobby.json'));
  assert.equal(store.state.view?.mode, 'inPerson');
});

test('a view without deadline or people drops the old ones', () => {
  const store = new Store();
  store.apply(example('view.night.asleep.json'));
  assert.ok(store.state.deadline);
  store.apply(example('view.lobby.json'));
  assert.equal(store.state.deadline, undefined);
  assert.ok(store.state.people?.length);
  store.apply(example('view.night.asleep.json'));
  assert.equal(store.state.people, undefined);
});

test('a reset replaces the whole list, never appends', () => {
  const store = new Store();
  store.apply({ type: 'events', events: filler(5), from: 0, reset: true });
  const reset = example<Server.EventsMessage>('events.reset.json');
  store.apply(reset);
  assert.deepEqual(store.state.events, reset.events);
});

test('events from the end of the list continue it', () => {
  const store = new Store();
  store.apply({ type: 'events', events: filler(13), from: 0, reset: true });
  const batch = example<Server.EventsMessage>('events.batch.json');
  assert.equal(store.apply(batch), 'ok');
  assert.equal(store.state.events.length, 13 + batch.events.length);
  assert.deepEqual(store.state.events.slice(13), batch.events);
});

test('events the phone already holds are not doubled', () => {
  const store = new Store();
  store.apply({ type: 'events', events: filler(14), from: 0, reset: true });
  const batch = example<Server.EventsMessage>('events.batch.json');
  store.apply(batch);
  assert.equal(store.state.events.length, 14);
  assert.deepEqual(store.state.events[13], batch.events[0]);
});

test('events past the end of the list are a gap and change nothing', () => {
  const store = new Store();
  const reset = example<Server.EventsMessage>('events.reset.json');
  store.apply(reset);
  assert.equal(store.apply(example('events.batch.json')), 'gap');
  assert.deepEqual(store.state.events, reset.events);
});

test('an event type the app does not know still counts in the list', () => {
  const store = new Store();
  const unknown = { type: 'SomethingNew', data: {} } as unknown as Server.Event;
  store.apply({ type: 'events', events: [unknown], from: 0, reset: true });
  assert.equal(store.apply(example('events.batch.json')), 'gap');
  store.apply({ type: 'events', events: filler(1), from: 1, reset: false });
  assert.equal(store.state.events.length, 2);
});

test('welcome sets the player; the PIN only comes with a join', () => {
  const store = new Store();
  store.apply(example('welcome.join.json'));
  assert.equal(store.state.player, 'p1');
  assert.equal(store.state.pin, '6663');
  store.clearPin();
  store.apply(example('welcome.rejoin-token.json'));
  assert.equal(store.state.player, 'p3');
  assert.equal(store.state.pin, undefined);
});

test('errors keep their code, situation and retryAfter, never the message', () => {
  const store = new Store();
  for (const f of all.filter((n) => n.startsWith('error.'))) {
    const msg = example<Server.ErrorMessage>(f);
    const before = store.state.error?.seq ?? 0;
    store.apply(msg, 'join');
    const e = store.state.error;
    assert.equal(e?.code, msg.code, f);
    assert.equal(e?.situation, 'join', f);
    assert.equal(e?.retryAfter, msg.retryAfter, f);
    assert.ok(e && e.seq > before, `${f}: seq`);
    assert.ok(!('message' in (e ?? {})), f);
  }
});

test('subscribers hear every change until they stop', () => {
  const store = new Store();
  let calls = 0;
  const stop = store.subscribe(() => calls++);
  store.apply(example('view.lobby.json'));
  store.apply(example('events.reset.json'));
  stop();
  store.apply(example('view.day.json'));
  assert.equal(calls, 2);
});

test('leaving the room drops the game but keeps the error', () => {
  const store = new Store();
  store.apply(example('view.night.seer.json'));
  store.apply(example('events.reset.json'));
  store.apply(example('error.kicked_by_host.json'));
  store.leaveRoom();
  assert.equal(store.state.view, undefined);
  assert.equal(store.state.events.length, 0);
  assert.equal(store.state.error?.code, 'kicked_by_host');
});
