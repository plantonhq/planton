import { generatedStats } from "@/site/stats.generated";

/**
 * Display-ready statistics derived from the code-generated counts. Copy uses
 * these conservative forms (per the messaging spec) so numbers never rot:
 * "400+ components", "17 providers", "dozens of stacks" (exact where it helps).
 */

/** Round down to the nearest hundred, e.g. 409 -> "400+". */
function roundedPlus(n: number): string {
  return `${Math.floor(n / 100) * 100}+`;
}

export const stats = {
  components: generatedStats.components,
  providers: generatedStats.providers,
  charts: generatedStats.charts,

  /** "400+" */
  componentsLabel: roundedPlus(generatedStats.components),
  /** "17" */
  providersLabel: String(generatedStats.providers),
  /** exact chart count, e.g. "49" */
  chartsLabel: String(generatedStats.charts),
  /** copy-friendly, rot-proof phrasing */
  chartsPhrase: "dozens of ready-made stacks",
} as const;

export type Stats = typeof stats;
