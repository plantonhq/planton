/**
 * `src/site` — the domain/config layer for planton.dev.
 *
 * All site-specific data (copy inputs, links, install methods, stats, provider
 * logos) lives here, separate from presentation. Components import values from
 * this layer and stay dumb and reusable. See README.md for the boundary.
 */
export * from "./nav";
export * from "./download";
export * from "./providers";
export { stats } from "./stats";
