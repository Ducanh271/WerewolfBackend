// GameSocket against a stand-in server: what it sends (checked against
// ../docs/api/examples/client/), how it saves tokens, when it reconnects and
// when it stops (docs/api/README.md "Joining and rejoining").
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { GameSocket, retryDelay, type Outgoing, type SocketEnv, type WebSocketLike } from '../src/net/socket.ts';
import { TokenBook, maxRooms, type KeyValueStore } from '../src/net/tokens.ts';
import { Store } from '../src/state/store.ts';

const examples = join(import.meta.dirname, '..', '..', 'docs', 'api', 'examples');
const read = (path: string): Record<string, unknown> => JSON.parse(readFileSync(join(examples, path), 'utf8'));

class FakeSocket implements WebSocketLike {
  sent: Record<string, unknown>[] = [];
  closed = false;
  onopen: (() => void) | null = null;
  onmessage: ((ev: { data: unknown }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  send(data: string): void {
    this.sent.push(JSON.parse(data));
  }
  close(): void {
    this.closed = true;
  }
  open(): void {
    this.onopen?.();
  }
  receive(msg: unknown): void {
    this.onmessage?.({ data: JSON.stringify(msg) });
  }
  drop(): void {
    this.onclose?.();
  }
}

class MemoryStorage implements KeyValueStore {
  data = new Map<string, string>();
  getItem(k: string): string | null {
    return this.data.get(k) ?? null;
  }
  setItem(k: string, v: string): void {
    this.data.set(k, v);
  }
}

function setup(storage: () => KeyValueStore | undefined = () => new MemoryStorage()) {
  const sockets: FakeSocket[] = [];
  const timers: { fn: () => void; ms: number; done: boolean }[] = [];
  const env: SocketEnv = {
    url: 'ws://test/ws',
    open: (url) => {
      assert.equal(url, 'ws://test/ws');
      const s = new FakeSocket();
      sockets.push(s);
      return s;
    },
    setTimeout: (fn, ms) => {
      const t = { fn, ms, done: false };
      timers.push(t);
      return t;
    },
    clearTimeout: (h) => {
      (h as { done: boolean }).done = true;
    },
    random: () => 0.5,
  };
  const store = new Store();
  const tokens = new TokenBook(storage);
  const socket = new GameSocket(store, tokens, env);
  const last = (): FakeSocket => sockets.at(-1)!;
  const tick = (): number => {
    const t = timers.find((x) => !x.done);
    assert.ok(t, 'no timer waiting');
    t.done = true;
    t.fn();
    return t.ms;
  };
  return { socket, store, tokens, sockets, timers, last, tick };
}

const welcomeJoin = read('server/welcome.join.json');
const lobby = read('server/view.lobby.json');
const reset = read('server/events.reset.json');

/** A joined phone in room K7QP, with its token saved. */
function joined() {
  const t = setup();
  t.socket.enter({ kind: 'join', room: 'K7QP', nickname: 'Bình' });
  t.last().open();
  t.last().receive(welcomeJoin);
  t.last().receive(lobby);
  t.last().receive(reset);
  return t;
}

test('join, host join and both rejoins send the contract messages', () => {
  const cases = [
    [{ kind: 'join', room: 'K7QP', nickname: 'Bình' }, 'client/join.json'],
    [{ kind: 'join', room: 'K7QP', nickname: 'An', hostKey: '5f1c9e0b7a2d4e6f8a9b0c1d2e3f4a5b' }, 'client/join.host.json'],
    [{ kind: 'pin', room: 'K7QP', nickname: 'Dũng', pin: '3967' }, 'client/rejoin.pin.json'],
  ] as const;
  for (const [entry, file] of cases) {
    const t = setup();
    t.socket.enter(entry);
    assert.equal(t.last().sent.length, 0, 'nothing before the socket opens');
    t.last().open();
    assert.deepEqual(t.last().sent, [read(file)], file);
  }
  const t = setup();
  t.tokens.save('K7QP', '52300c648d7bf3be9967637971f2df72');
  t.socket.enter({ kind: 'token', room: 'k7qp' });
  t.last().open();
  assert.deepEqual(t.last().sent, [read('client/rejoin.token.json')]);
});

test('welcome saves the token by room; the PIN is shown, never stored', () => {
  const storage = new MemoryStorage();
  const t = setup(() => storage);
  t.socket.enter({ kind: 'join', room: 'k7qp', nickname: 'An' });
  t.last().open();
  t.last().receive(welcomeJoin);
  assert.equal(t.tokens.get('K7QP'), welcomeJoin.token);
  assert.equal(t.store.state.pin, '6663');
  assert.equal(t.store.state.connection.state, 'joined');
  assert.ok(![...storage.data.values()].some((v) => v.includes('6663')), 'PIN in storage');
  assert.equal(new TokenBook(() => storage).get('K7QP'), welcomeJoin.token, 'token survives a reload');
});

test('a new token from a PIN rejoin replaces the old one', () => {
  const t = setup();
  t.tokens.save('K7QP', 'old');
  t.socket.enter({ kind: 'pin', room: 'K7QP', nickname: 'Dũng', pin: '3967' });
  t.last().open();
  t.last().receive(read('server/welcome.rejoin-pin.json'));
  assert.equal(t.tokens.get('K7QP'), '119c8cc7d15760b9cd2b12898717c24b');
  assert.deepEqual(t.tokens.all(), ['119c8cc7d15760b9cd2b12898717c24b']);
});

test('views and events reach the store once joined', () => {
  const t = joined();
  assert.equal(t.store.state.view?.phase, 'Lobby');
  assert.equal(t.store.state.events.length, 2);
});

test('a refused join keeps the socket for another try, with the situation', () => {
  const t = setup();
  t.socket.enter({ kind: 'join', room: 'K7QP', nickname: 'An' });
  t.last().open();
  t.last().receive(read('server/error.nickname_taken.json'));
  t.last().receive({ type: 'error', code: 'wrong_phase', message: 'x' });
  assert.equal(t.store.state.error?.situation, 'join');
  assert.equal(t.socket.send({ type: 'command', name: 'dealRoles' }), false, 'no commands before welcome');
  t.socket.enter({ kind: 'join', room: 'K7QP', nickname: 'An Nhiên' });
  assert.equal(t.sockets.length, 1);
  assert.equal(t.last().sent.length, 2);
  assert.equal(t.store.state.error, undefined);
});

test('a dropped line rejoins with the token after a growing delay', () => {
  const t = joined();
  const delays: number[] = [];
  for (let i = 0; i < 4; i++) {
    t.last().drop();
    assert.equal(t.store.state.connection.state, 'waiting');
    delays.push(t.tick());
    t.last().open();
    assert.deepEqual(t.last().sent, [{ type: 'rejoin', room: 'K7QP', token: welcomeJoin.token }]);
    if (i < 3) continue;
    t.last().receive(read('server/welcome.rejoin-token.json'));
  }
  assert.deepEqual(delays, [0, 1000, 2000, 4000]);
  assert.equal(t.store.state.connection.state, 'joined');
  t.last().drop();
  assert.equal(t.tick(), 0, 'a welcome starts the delays again');
});

test('the delay stops growing at 30 seconds', () => {
  assert.equal(retryDelay(10, () => 0.5), 30_000);
  assert.equal(retryDelay(20, () => 1), 37_500);
  assert.equal(retryDelay(2, () => 0), 1500);
});

test('a socket that never got in does not reconnect', () => {
  const t = setup();
  t.socket.enter({ kind: 'join', room: 'K7QP', nickname: 'An' });
  t.last().open();
  t.last().receive(read('server/error.nickname_taken.json'));
  t.last().drop();
  assert.equal(t.timers.length, 0);
  assert.equal(t.store.state.connection.state, 'idle');
});

const stops = [
  ['replaced_by_other_device', 'replaced', false],
  ['kicked_by_host', 'kicked', true],
] as const;

for (const [code, reason, keepsToken] of stops) {
  test(`${code} stops reconnecting and drops the game`, () => {
    const t = joined();
    t.last().receive(read(`server/error.${code}.json`));
    t.last().drop();
    assert.equal(t.timers.filter((x) => !x.done).length, 0);
    assert.deepEqual(t.store.state.connection, { state: 'stopped', reason });
    assert.equal(t.store.state.view, undefined);
    assert.equal(t.store.state.error?.code, code);
    assert.equal(t.tokens.get('K7QP') !== undefined, keepsToken, 'token kept');
  });
}

test('a token rejoin refused with bad_credentials stops and forgets the token', () => {
  const t = joined();
  t.last().drop();
  t.tick();
  t.last().open();
  t.last().receive(read('server/error.bad_credentials.json'));
  assert.deepEqual(t.store.state.connection, { state: 'stopped', reason: 'badToken' });
  assert.equal(t.store.state.error?.situation, 'token');
  assert.equal(t.tokens.get('K7QP'), undefined);
});

test('a token rejoin to an expired room stops and keeps the token for history', () => {
  const t = joined();
  t.last().drop();
  t.tick();
  t.last().open();
  t.last().receive({ type: 'error', code: 'room_not_found', message: 'x' });
  assert.deepEqual(t.store.state.connection, { state: 'stopped', reason: 'roomGone' });
  assert.equal(t.tokens.get('K7QP'), welcomeJoin.token);
});

test('a PIN rejoin with a wrong PIN has no situation and does not stop', () => {
  const t = setup();
  t.socket.enter({ kind: 'pin', room: 'K7QP', nickname: 'Dũng', pin: '0000' });
  t.last().open();
  t.last().receive(read('server/error.pin_locked.json'));
  assert.equal(t.store.state.error?.situation, undefined);
  assert.ok(t.store.state.error?.retryAfter);
  assert.equal(t.last().closed, false);
});

test('leave sends leave, stops and keeps the token', () => {
  const t = joined();
  t.socket.leave();
  assert.deepEqual(t.sockets[0]!.sent.at(-1), read('client/leave.json'));
  assert.deepEqual(t.store.state.connection, { state: 'stopped', reason: 'left' });
  t.last().drop();
  assert.equal(t.timers.length, 0);
  assert.equal(t.tokens.get('K7QP'), welcomeJoin.token);
});

test('a newer protocol stops before playing', () => {
  const t = setup();
  t.socket.enter({ kind: 'join', room: 'K7QP', nickname: 'An' });
  t.last().open();
  t.last().receive({ ...welcomeJoin, protocolVersion: 2 });
  assert.deepEqual(t.store.state.connection, { state: 'stopped', reason: 'outdated' });
  assert.equal(t.tokens.get('K7QP'), undefined);
});

test('missed events close the line and rejoin for the whole list', () => {
  const t = joined();
  t.last().receive(read('server/events.batch.json')); // from 13, the phone holds 2
  assert.equal(t.last().closed, true);
  assert.equal(t.store.state.events.length, 2);
  t.last().drop();
  t.tick();
  t.last().open();
  assert.equal(t.last().sent[0]?.type, 'rejoin');
});

test('commands go out once joined, as the contract writes them', () => {
  const t = joined();
  for (const name of ['nominate', 'dealRoles', 'sendChat'] as const) {
    const msg = read(`client/command.${name}.json`);
    assert.equal(t.socket.send(msg as unknown as Outgoing), true);
    assert.deepEqual(t.last().sent.at(-1), msg);
  }
  assert.equal(t.socket.send(read('client/kick.json') as unknown as Outgoing), true);
});

test('a message over 4 KB is not sent', () => {
  const t = joined();
  const before = t.last().sent.length;
  assert.equal(t.socket.send({ type: 'command', name: 'sendChat', args: { text: 'x'.repeat(5000) } }), false);
  assert.equal(t.last().sent.length, before);
});

test('messages from an old socket are ignored', () => {
  const t = joined();
  const old = t.last();
  old.drop();
  t.tick();
  old.receive(read('server/view.day.json'));
  assert.equal(t.store.state.view?.phase, 'Lobby');
});

test('storage that throws loses only the copy', () => {
  const broken = (): KeyValueStore => {
    throw new Error('blocked');
  };
  const t = setup(broken);
  t.socket.enter({ kind: 'join', room: 'K7QP', nickname: 'An' });
  t.last().open();
  t.last().receive(welcomeJoin);
  assert.equal(t.tokens.get('K7QP'), welcomeJoin.token);
  t.last().drop();
  t.tick();
  t.last().open();
  assert.equal(t.last().sent[0]?.token, welcomeJoin.token);
  const throwing = new TokenBook(() => ({
    getItem: () => {
      throw new Error('x');
    },
    setItem: () => {
      throw new Error('x');
    },
  }));
  throwing.save('K7QP', 't');
  assert.equal(throwing.get('k7qp'), 't');
});

test('the token book keeps the newest token of the 20 newest rooms', () => {
  const book = new TokenBook(() => new MemoryStorage());
  for (let i = 0; i < maxRooms + 5; i++) book.save(`R${i}`, `t${i}`);
  book.save('R10', 't10b');
  const tokens = book.all();
  assert.equal(tokens.length, maxRooms);
  assert.equal(tokens[0], 't10b');
  assert.ok(!tokens.includes('t10'));
  assert.equal(book.get('R0'), undefined);
});
