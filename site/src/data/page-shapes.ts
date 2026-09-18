/**
 * The shapes a story page's record is built from, shared by the Trust,
 * Product, and Distributions data files so the templates that render them
 * compose the same primitives. A page states nothing the record does not
 * carry; the record states nothing the story does not back.
 */

/** A proof point: a short Title Case label the page chooses, over a sentence the story states. */
export interface ProofPoint {
  label: string;
  text: string;
}

/**
 * The proof beside the claims, as the shape of a real record: the product's
 * field names as labels, illustrated values, and a footer that says so
 * wherever a value is a number.
 */
export interface RecordArtifact {
  kind: 'record';
  title: string;
  rows: readonly { label: string; value: string }[];
  footer: string;
}

/**
 * The proof beside the claims, as the commands a person actually types,
 * copied from the documentation they are quoted from and captioned with it.
 */
export interface CommandsArtifact {
  kind: 'commands';
  title: string;
  /** A file the commands act on, shown above them so the input is in view too. */
  file?: { name: string; body: string };
  commands: readonly string[];
  /** Screen-reader name for the copy button. */
  label: string;
  /** Sentence case; what these commands do, in one line. */
  caption: string;
  /** Where they are documented; rendered as the caption's link. */
  source: { label: string; href: string };
}

export type Artifact = RecordArtifact | CommandsArtifact;

/**
 * The footer vocabulary every illustrated record shares, so the same fact is
 * said the same way under every record on the site. A record's footer is the
 * illustration sentence, then the figures sentence when any row carries an
 * example figure marked `est.`, then one clause of its own only where the
 * record needs one (a control profile says why it has no verdict). A record
 * whose rows are real counts names its source and date instead.
 */
export const ILLUSTRATED_RECORD = 'An illustration of the record the product shows.';
export const EXAMPLE_FIGURES = 'Figures marked est. are examples; a real record carries its own.';

/** The footer of an illustrated record: the shared sentences, then the record's own clause when it has one. */
export function illustratedFooter(options: { figures?: boolean; note?: string } = {}): string {
  return [ILLUSTRATED_RECORD, options.figures ? EXAMPLE_FIGURES : undefined, options.note].filter(Boolean).join(' ');
}

/** Which strip of facts, if any, a page shows under its hero. */
export type HeroStrip = 'providers' | 'counts';

/** One step of a how-it-works sequence, in the order a person experiences it. */
export interface Step {
  /** Title Case. */
  title: string;
  /** Sentence case; one or two sentences. */
  text: string;
}
