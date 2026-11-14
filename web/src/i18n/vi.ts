// The web game's own Vietnamese texts (CLAUDE.md "Conventions"): whole
// sentences with placeholders, never pieces glued together in code, so another
// language is one more file. Texts the server sends (`say`, `notice`,
// `summary`, the guide, error sentences in src/api/errors.ts) are shown as
// given and never copied here.
export const vi = {
  /** Portrait phone overlay (docs/frontend/giao-dien-3d.md §1). */
  rotatePhone: 'Xoay ngang điện thoại',
} as const;

export type TextKey = keyof typeof vi;
