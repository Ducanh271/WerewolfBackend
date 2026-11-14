import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { errorTable } from '../src/api/errors.ts';
import { errorText } from '../src/api/error-text.ts';
import { parseErrorTable } from '../scripts/error-table.ts';

const api = join(import.meta.dirname, '..', '..', 'docs', 'api');

test('src/api/errors.ts is what docs/api/errors.md says (run npm run gen:types)', () => {
  assert.deepEqual(errorTable, parseErrorTable(readFileSync(join(api, 'errors.md'), 'utf8')));
});

test('every error code of the examples has its sentence', () => {
  const dir = join(api, 'examples', 'server');
  for (const f of readdirSync(dir).filter((f) => f.startsWith('error.'))) {
    const code = JSON.parse(readFileSync(join(dir, f), 'utf8')).code as string;
    const known = code in errorTable.codes || errorTable.fallbackCodes.includes(code);
    assert.ok(known, `${f}: code ${code} is not in errors.md`);
  }
});

test('errorText follows "How to show an error"', () => {
  const rows: [string, string | undefined, number | undefined, string][] = [
    ['room_not_found', undefined, undefined, errorTable.codes.room_not_found!.text],
    ['wrong_phase', 'join', undefined, errorTable.codes.wrong_phase!.situations!.join!],
    ['wrong_phase', 'nominate', undefined, errorTable.codes.wrong_phase!.text],
    ['bad_credentials', 'token', undefined, errorTable.codes.bad_credentials!.situations!.token!],
    ['no_such_code_yet', undefined, undefined, errorTable.fallback],
    ['internal', undefined, undefined, errorTable.fallback],
  ];
  for (const [code, situation, retry, want] of rows) {
    assert.equal(errorText(code, situation, retry), want, `${code}/${situation}`);
  }
  const locked = errorText('pin_locked', undefined, 41_200);
  assert.ok(locked.includes('42') && !locked.includes('%s'), locked);
});
