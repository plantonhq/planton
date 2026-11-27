/**
 * The shapes of the real things the story pages show beside their words:
 * a file, a policy or a command's output (code), a small table of facts, a
 * screen of the product itself, and the check that lets a reader prove a
 * sentence without taking the page's word for it. Home's sections and the Security page fill these records;
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

/**
 * A screen of the product, as a person saw it: a crop of one demo universe
 * scene (`demo-universe/<company>/company.yaml`), never a drawn or edited
 * picture. The scene is the source, so a recapture refreshes the picture and
 * `scripts/check-scenes.mjs` refuses a screen whose scene doesn't exist. The
 * image is the site's own (content/assets/images/, served from the assets
 * bucket), at twice the width it is shown.
 */
export interface ScreenArtifact {
  kind: 'screen';
  label: string;
  /** The universe scene it is cropped from, `<company>.<area>.<screen>.<state>`. */
  scene: string;
  src: string;
  /** What the picture shows, for a reader who can't see it; it carries the meaning, not the pixels. */
  alt: string;
  width: number;
  height: number;
  /** A narrower crop of the same scene for a phone, where the whole screen's text would be too small to read. */
  phone?: { src: string; width: number; height: number };
}

export type Artifact = CodeArtifact | TableArtifact | ScreenArtifact;

/** How a reader proves a sentence without taking the page's word for it. */
export interface Check {
  text: string;
  /** A command they can run, shown in monospace. */
  command?: string;
  /** A page that shows the same thing. */
  href?: string;
}
