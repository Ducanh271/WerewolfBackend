// parseErrorTable reads the generated tables of docs/api/errors.md: groups of
// `| code | situation | when | text |` rows, then a last group whose codes
// all show one shared sentence (the `> ...` line before its table).
import type { ErrorTable } from '../src/api/error-table-type.ts';

export function parseErrorTable(md: string): ErrorTable {
  const start = md.indexOf('<!-- generated: errors -->');
  const end = md.indexOf('<!-- /generated -->');
  if (start < 0 || end < start) throw new Error('errors.md: no <!-- generated: errors --> region');
  const table: ErrorTable = { codes: {}, fallback: '', fallbackCodes: [] };
  let columns: string[] = [];
  for (const raw of md.slice(start, end).split('\n')) {
    const line = raw.trim();
    if (line.startsWith('> ')) {
      table.fallback = line.slice(2).trim();
      continue;
    }
    if (!line.startsWith('|')) continue;
    const cells = line.slice(1, -1).split('|').map((c) => c.trim());
    if (cells[0] === 'code') {
      columns = cells;
      continue;
    }
    if (cells.every((c) => /^-+$/.test(c))) continue;
    const row: Record<string, string> = Object.fromEntries(columns.map((c, i) => [c, cells[i] ?? '']));
    const code = (row.code ?? '').replace(/`/g, '');
    if (!code) continue;
    if (!row.text) {
      table.fallbackCodes.push(code);
      continue;
    }
    const entry = (table.codes[code] ??= { text: '' });
    if (row.situation) {
      (entry.situations ??= {})[row.situation.replace(/`/g, '')] = row.text;
    } else {
      entry.text = row.text;
    }
  }
  if (!table.fallback) throw new Error('errors.md: no shared sentence for the last group');
  for (const [code, e] of Object.entries(table.codes)) {
    if (!e.text) throw new Error(`errors.md: ${code} has only situation rows`);
  }
  return table;
}
