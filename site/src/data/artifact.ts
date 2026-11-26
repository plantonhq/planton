/**
 * The shapes of the real things the story pages show beside their words:
 * a file, a policy or a command's output (code), a small table of facts, and
 * the check that lets a reader prove a sentence without taking the page's
 * word for it. Home's sections and the Security page fill these records;
 * `components/marketing/Artifact.tsx` renders them, flat and labelled, so
 * every page shows an artifact the same way (site/design-system.md,
 * "Components").
 */

/** A real file, policy or command output, shown flat and readable. */
export interface CodeArtifact {
  kind: 'code';
  /** What the reader is looking at and where it comes from. */
  label: string;
  language: 'yaml' | 'json' | 'shell';
  lines: readonly string[];
  /**
   * Long lines wrap under their own text (a comment after its `# `, a
   * command after its `$ `), the way a person reads a file; without it a
   * long line scrolls inside the panel, which suits a policy or JSON.
   */
  wrap?: boolean;
}

/** A small table of facts (a permission and its use, a role and whether it may act). */
export interface TableArtifact {
  kind: 'table';
  label: string;
  columns: readonly string[];
  rows: readonly (readonly string[])[];
}

export type Artifact = CodeArtifact | TableArtifact;

/** How a reader proves a sentence without taking the page's word for it. */
export interface Check {
  text: string;
  /** A command they can run, shown in monospace. */
  command?: string;
  /** A page that shows the same thing. */
  href?: string;
}
