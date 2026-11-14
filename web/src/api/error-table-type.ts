// The shape of the generated src/api/errors.ts (docs/api/errors.md).

/** The sentences of one error code. */
export interface ErrorTexts {
  /** The sentence when no situation matches. */
  text: string;
  /** A sentence per situation, when the code means two things at the table
   * (e.g. `bad_credentials` answering a token rejoin: `token`). */
  situations?: Record<string, string>;
}

export interface ErrorTable {
  /** Every code that has its own sentence. */
  codes: Record<string, ErrorTexts>;
  /** The shared sentence of the codes of a buggy client or server, and of
   * any code the app does not know. */
  fallback: string;
  /** The codes that show the fallback on purpose. */
  fallbackCodes: string[];
}
