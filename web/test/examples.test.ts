// The contract's real messages (../docs/api/examples/) all read back, each
// with the `type` its folder allows. store.test.ts replays them through the store.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';

const examples = join(import.meta.dirname, '..', '..', 'docs', 'api', 'examples');

function read(folder: string): [string, Record<string, unknown>][] {
  return readdirSync(join(examples, folder))
    .filter((f) => f.endsWith('.json'))
    .map((f) => [f, JSON.parse(readFileSync(join(examples, folder, f), 'utf8'))]);
}

test('every server example is one of the four server messages', () => {
  const files = read('server');
  assert.ok(files.length > 50, `only ${files.length} server examples`);
  for (const [f, msg] of files) {
    assert.ok(['welcome', 'view', 'events', 'error'].includes(msg.type as string), `${f}: type ${msg.type}`);
    assert.ok(f.startsWith(`${msg.type}.`), `${f} holds a ${msg.type} message`);
  }
});

test('every client example names its message, commands their name', () => {
  for (const [f, msg] of read('client')) {
    assert.equal(typeof msg.type, 'string', f);
    const expected = msg.type === 'command' ? `command.${msg.name}` : (msg.type as string);
    assert.ok(f.startsWith(expected), `${f} holds ${expected}`);
  }
});

test('every event example is filed under its type', () => {
  for (const [f, ev] of read('events')) {
    assert.equal(f, `${ev.type}.json`);
    assert.ok('data' in ev || Object.keys(ev).length === 1, `${f}: no data`);
  }
});

test('every HTTP example reads', () => {
  assert.ok(read('http').length > 0);
});
