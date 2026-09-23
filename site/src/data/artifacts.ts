/**
 * The proof a page shows beside its claims, found by the page's path. Trust,
 * Product, and Distributions records each carry one artifact (the shape of
 * the record the product stamps, or the commands a person types); a page
 * that points at one of them as its proof (a persona page's beat) shows the
 * same artifact rather than drawing its own, so an illustrated record is
 * defined once and a real capture replaces it in one place.
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators without a bundler.
 */
import { DISTRIBUTION_PAGES } from './distributions.ts';
import type { Artifact } from './page-shapes.ts';
import { PRODUCT_PAGES } from './product.ts';
import { TRUST_PAGES } from './trust.ts';

/** The artifact of the page at `path`, or undefined for a page that carries none (an index, a persona page). */
export function artifactOf(path: string): Artifact | undefined {
  return (
    TRUST_PAGES.find((p) => p.path === path)?.artifact ??
    PRODUCT_PAGES.find((p) => p.path === path)?.artifact ??
    DISTRIBUTION_PAGES.find((p) => p.path === path)?.artifact
  );
}
