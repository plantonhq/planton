/**
 * GENERATED FILE — do not edit by hand.
 *
 * Written by scripts/generate-stats.ts (runs as part of `yarn build`) from the
 * canonical sources in the repo: the cloud-resource-kind enum, the provider
 * tree, and charts/. This keeps the site's numbers single-sourced from code so
 * they never drift into a stale hardcoded count.
 */
export const generatedStats = {
  /** Annotated cloud-resource kinds (catalog breadth). */
  components: 409,
  /** Cloud providers in the catalog. */
  providers: 17,
  /** Ready-made infra charts (stacks). */
  charts: 49,
} as const;

export type GeneratedStats = typeof generatedStats;
