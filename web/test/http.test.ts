// The HTTP calls read the contract's bodies (docs/api/examples/http/) and
// turn errors into codes; CurrentRoom survives storage that throws.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { CurrentRoom } from '../src/net/current.ts';
import { Api, networkError, type FetchLike } from '../src/net/http.ts';

const http = join(import.meta.dirname, '..', '..', 'docs', 'api', 'examples', 'http');
const read = (f: string): unknown => JSON.parse(readFileSync(join(http, f), 'utf8'));

function fake(ok: boolean, body: unknown, seen: { url: string; init?: Parameters<FetchLike>[1] }[] = []): FetchLike {
  return async (url, init) => {
    seen.push({ url, init });
    return { ok, json: async () => body };
  };
}

test('createRoom posts the mode and reads the code and hostKey', async () => {
  const seen: { url: string; init?: Parameters<FetchLike>[1] }[] = [];
  const api = new Api(fake(true, read('create-room.reply.json'), seen));
  const r = await api.createRoom('online');
  assert.deepEqual(r, { ok: true, value: read('create-room.reply.json') });
  assert.equal(seen[0]?.url, '/rooms');
  assert.equal(seen[0]?.init?.method, 'POST');
  assert.deepEqual(JSON.parse(seen[0]!.init!.body!), read('create-room.online.request.json'));
});

test('checkRoom: the room, or room_not_found', async () => {
  const seen: { url: string }[] = [];
  assert.deepEqual(await new Api(fake(true, read('join-room.reply.json'), seen)).checkRoom('K7QP'), {
    ok: true,
    value: read('join-room.reply.json'),
  });
  assert.equal(seen[0]?.url, '/join/K7QP');
  const missing = await new Api(fake(false, read('join-room.error.room_not_found.json'))).checkRoom('ZZZZ');
  assert.deepEqual(missing, { ok: false, code: 'room_not_found' });
});

test('no answer at all is the network code', async () => {
  const offline: FetchLike = async () => {
    throw new TypeError('failed to fetch');
  };
  assert.deepEqual(await new Api(offline).checkRoom('K7QP'), { ok: false, code: networkError });
  const notJson = await new Api(async () => ({ ok: false, json: async () => Promise.reject(new SyntaxError('x')) })).createRoom('inPerson');
  assert.deepEqual(notJson, { ok: false, code: networkError });
});

test('the guide is fetched once; a failure is tried again', async () => {
  let calls = 0;
  let fail = true;
  const api = new Api(async () => {
    calls++;
    return fail ? Promise.reject(new TypeError('offline')) : { ok: true, json: async () => read('guide.reply.json') };
  });
  assert.equal((await api.guide()).ok, false);
  fail = false;
  const r = await api.guide();
  assert.ok(r.ok && r.value.roles.length >= 12);
  await api.guide();
  assert.equal(calls, 2);
});

test('the current room is kept, cleared, and lost quietly when storage throws', () => {
  const data = new Map<string, string>();
  const cur = new CurrentRoom(() => ({ getItem: (k) => data.get(k) ?? null, setItem: (k, v) => void data.set(k, v), removeItem: (k) => void data.delete(k) }));
  cur.set('k7qp');
  assert.equal(cur.get(), 'K7QP');
  cur.clear();
  assert.equal(cur.get(), undefined);
  const broken = new CurrentRoom(() => {
    throw new Error('blocked');
  });
  broken.set('K7QP');
  assert.equal(broken.get(), undefined);
  broken.clear();
});
