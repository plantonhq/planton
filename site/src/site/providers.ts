/**
 * Provider / tool logos used across the marketing surface. Every entry maps to a
 * REAL asset in public/images/providers — we never invent brand marks. Rows are
 * rendered by mapping over these lists (add a provider = edit data, not JSX).
 */

export interface Brand {
  name: string;
  /** Path under public/. */
  logo: string;
}

const logo = (slug: string) => `/images/providers/${slug}.svg`;

/**
 * The clouds shown in the "Deploys to" row. This is the DEPLOY-PARITY set only —
 * clouds you can actually deploy to today — NOT the full catalog breadth (that's
 * "17 providers"). Claiming deploy for a cloud not at parity is an overclaim, so
 * keep this list honest and short. Every entry maps to a real logo.
 */
export const DEPLOY_CLOUDS: Brand[] = [
  { name: "AWS", logo: logo("aws") },
  { name: "Google Cloud", logo: logo("gcp") },
  { name: "Azure", logo: logo("azure") },
  { name: "Kubernetes", logo: logo("kubernetes") },
  { name: "DigitalOcean", logo: logo("digital-ocean") },
];

