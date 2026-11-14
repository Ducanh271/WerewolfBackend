// The countdown runs on the server's clock (docs/api/README.md "Server →
// client", secrecy.md "Time comes from the server"): two phones whose clocks
// disagree show the same time left.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import type { Server } from '../src/api/types.ts';
import { remaining, serverNow } from '../src/state/clock.ts';
import { Store } from '../src/state/store.ts';

const folder = join(import.meta.dirname, '..', '..', 'docs', 'api', 'examples', 'server');
const example = <T,>(name: string): T => JSON.parse(readFileSync(join(folder, name), 'utf8')) as T;

const night = example<Server.ViewMessage>('view.night.asleep.json');
const welcome = example<Server.Welcome>('welcome.rejoin-token.json');

/** A phone whose clock is `skew` ms off the server's, receiving `msg` after `latency` ms. */
function phone(skew: number) {
  let local = 0;
  const store = new Store(() => local);
  return {
    store,
    receive(msg: Server.ServerMessage, latency = 0) {
      if (msg.type === 'view') local = msg.serverTime + skew + latency;
      store.apply(msg);
    },
    at: (serverTime: number) => serverTime + skew,
  };
}

test('the night example has a server time before its deadline', () => {
  assert.equal(typeof night.serverTime, 'number');
  assert.ok(night.deadline! > night.serverTime);
});

test('phones with wrong clocks show the same time left', () => {
  const fast = phone(+7000);
  const slow = phone(-4000);
  fast.receive(night);
  slow.receive(night);
  const t = night.serverTime + 3000;
  const left = night.deadline! - t;
  assert.equal(remaining(fast.store.state, fast.at(t)), left);
  assert.equal(remaining(slow.store.state, slow.at(t)), left);
  assert.equal(serverNow(fast.store.state, fast.at(t)), t);
});

test('the offset keeps the view that came the fastest', () => {
  const p = phone(2000);
  p.receive(night, 400);
  assert.equal(p.store.state.clockOffset, -2400);
  p.receive({ ...night, serverTime: night.serverTime + 5000 }, 50);
  assert.equal(p.store.state.clockOffset, -2050);
  p.receive({ ...night, serverTime: night.serverTime + 9000 }, 900); // a slow one does not undo it
  assert.equal(p.store.state.clockOffset, -2050);
});

test('a welcome starts the offset again', () => {
  const p = phone(0);
  p.receive(night, 0);
  assert.equal(p.store.state.clockOffset, 0);
  p.store.apply(welcome);
  assert.equal(p.store.state.clockOffset, undefined);
  p.receive(night, 300);
  assert.equal(p.store.state.clockOffset, -300);
});

test('no deadline, nothing to count; past it, zero', () => {
  const p = phone(0);
  p.receive(example('view.lobby.json'));
  assert.equal(remaining(p.store.state, 0), undefined);
  p.receive(night);
  assert.equal(remaining(p.store.state, night.deadline! + 5000), 0);
});

test('every example view reads a server time', () => {
  const store = new Store(() => 0);
  for (const name of ['view.lobby.json', 'view.day.json', 'view.ended.json', 'view.hunterShot.json']) {
    const msg = example<Server.ViewMessage>(name);
    store.apply(welcome);
    store.apply(msg);
    assert.equal(store.state.clockOffset, msg.serverTime, name);
  }
});
