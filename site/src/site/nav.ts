/**
 * External destinations and site navigation — the single source for every link
 * so no CTA can dead-end. Presentation components import from here; they never
 * hardcode a URL.
 */

export const GITHUB_REPO = "plantonhq/planton";
export const GITHUB_URL = `https://github.com/${GITHUB_REPO}`;
export const RELEASES_URL = `${GITHUB_URL}/releases`;

// Reused from planton.ai (shared community server).
export const DISCORD_URL = "https://discord.gg/pwcSapdQAp";

export const PLANTON_AI_URL = "https://planton.ai";

/**
 * Charts link. Until the on-site charts catalog generator ships, this points at
 * the real charts/ directory on GitHub (a working target, never a 404). Flip to
 * "/docs/charts" once that catalog exists.
 */
export const CHARTS_URL = `${GITHUB_URL}/tree/main/charts`;

export interface NavLink {
  label: string;
  href: string;
  /** External links open in a new tab and get rel="noreferrer". */
  external?: boolean;
}

/** Primary header navigation (brand + Download live outside this list). */
export const HEADER_NAV: NavLink[] = [
  { label: "Docs", href: "/docs" },
  { label: "Charts", href: CHARTS_URL, external: true },
];

/** Footer link row, mirroring the design: Docs · Charts · GitHub · Discord · planton.ai */
export const FOOTER_LINKS: NavLink[] = [
  { label: "Docs", href: "/docs" },
  { label: "Charts", href: CHARTS_URL, external: true },
  { label: "GitHub", href: GITHUB_URL, external: true },
  { label: "Discord", href: DISCORD_URL, external: true },
  { label: "planton.ai", href: PLANTON_AI_URL, external: true },
];
