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
 * The clouds shown in the "Deploys to" row. Every entry maps to a real logo in
 * public/images/providers.
 *
 * Missing real assets (add here once available — never invent a mark):
 *   - Oracle Cloud (oci), Alibaba Cloud (alicloud), Hetzner Cloud (hetznercloud)
 */
export const DEPLOY_CLOUDS: Brand[] = [
  { name: "AWS", logo: logo("aws") },
  { name: "Google Cloud", logo: logo("gcp") },
  { name: "Azure", logo: logo("azure") },
  { name: "Kubernetes", logo: logo("kubernetes") },
  { name: "Cloudflare", logo: logo("cloudflare") },
  { name: "DigitalOcean", logo: logo("digital-ocean") },
  { name: "Civo", logo: logo("civo") },
  { name: "Scaleway", logo: logo("scaleway") },
  { name: "OpenStack", logo: logo("openstack") },
];

