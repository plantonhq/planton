/**
 * The route registry: every page the website serves that is not generated
 * from a content folder, with the words search engines and agents read about
 * it and the chapters of the story it renders.
 *
 * One list feeds four things, so they can never disagree:
 *   - src/app/sitemap.ts and robots.ts (what is advertised to crawlers);
 *   - the page metadata helper in src/lib/page-metadata.ts (title,
 *     description, canonical, Open Graph) that every registered page uses;
 *   - scripts/generate-llms.mjs (llms.txt and the per-page markdown), whose
 *     coverage check fails the build when an exported page is missing here;
 *   - scripts/check-apex-routing.mjs, which proves every top-level path is
 *     passed through at the edge and reserved as a platform handle.
 *
 * Content pages (docs, blog, changelog, tutorials) are walked from their
 * markdown folders by those same scripts; only their index pages appear here.
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators without a bundler.
 */
import { HOMEPAGE } from './homepage.ts';
import { COMMUNITY_SEAT_LIMIT, FREE_TIER_SEATS } from './pricing.ts';
import { PLATFORM_STATS } from './platform-stats.ts';
import type { ChapterId } from './story.ts';

export type PageGroup =
  | 'home'
  | 'product'
  | 'distributions'
  | 'trust'
  | 'pricing'
  | 'content'
  | 'company'
  | 'legal'
  | 'standalone';

/**
 * The groups the `llms.txt` index lists pages under, in reading order, with
 * the heading each is given. Held here beside the type so a new group cannot
 * be added to the registry and forgotten by the index: the generator walks
 * this record and refuses a marketing page whose group is not in it. `home`
 * is the index's own opening line, `content` is walked from markdown, and
 * `standalone` is never listed, so none of the three has a heading.
 */
export const PAGE_GROUP_HEADINGS: Readonly<Partial<Record<PageGroup, string>>> = {
  trust: 'Trust',
  product: 'Product',
  distributions: 'Distributions',
  pricing: 'Pricing',
  company: 'Company',
  legal: 'Legal',
};

export interface SitePage {
  /** The route, with a leading slash and no trailing slash ("/" for the home page). */
  path: string;
  /** Title Case; rendered as `<title>` with the site suffix added by the metadata helper. */
  title: string;
  /** Sentence case, one or two sentences, under 160 characters. The description crawlers and agents read. */
  description: string;
  group: PageGroup;
  /** The chapters this page renders from, in order; empty for pages with no story content. */
  chapters?: readonly ChapterId[];
  /**
   * False keeps the page out of the sitemap and marks it noindex. Shared-by-URL
   * surfaces (meeting decks, investor pages) and single-task frames are not
   * search results.
   */
  index?: boolean;
}

const SITE_URL = 'https://planton.ai';

export const SITE = {
  url: SITE_URL,
  name: 'Planton',
  /** Appended to every registered page's title. */
  titleSuffix: ' | Planton',
  defaultOgImage: `${SITE_URL}/_site/images/og/desktop.png`,
} as const;

/**
 * The playbook a coding agent reads when a person hands it their Setup Prompt
 * (every Planton's Getting Started writes one): a Markdown file at the root,
 * served beside llms.txt, and listed in llms.txt's Start section so an agent
 * that begins from the index finds it.
 */
export const AGENT_PLAYBOOK = {
  path: '/agents.md',
  title: 'Set Up Planton (for coding agents)',
  description:
    'A coding agent given a Planton Setup Prompt follows this: install the CLI and the planton skill, sign in with the prompt\'s code, and take the repository to its first deployment.',
} as const;

export const SITE_PAGES: readonly SitePage[] = [
  // Home
  {
    path: '/',
    title: HOMEPAGE.title,
    description: HOMEPAGE.description,
    group: 'home',
    chapters: ['the-wall', 'verified-before-it-exists', 'services-ship-from-git', 'start'],
  },

  // Trust: the buyer's half of the story
  {
    path: '/trust',
    title: 'Trust',
    description:
      'What Planton proves about every deployment before it exists and keeps afterward: the cost, the rules, the record, the security posture, and whose keys it runs on.',
    group: 'trust',
    chapters: ['verified-before-it-exists', 'your-rules-hold', 'every-deployment-leaves-a-record', 'runs-where-you-decide'],
  },
  {
    path: '/trust/verified-before-deploy',
    title: 'Verified Before Deploy',
    description:
      'Before anything is created, Planton states the monthly cost with its coverage, the least-privilege permissions, and the controls each kind enforces.',
    group: 'trust',
    chapters: ['verified-before-it-exists'],
  },
  {
    path: '/trust/rules-and-approvals',
    title: 'Rules and Approvals',
    description:
      'Deployment budgets that pause a deploy, protected environments that refuse self-approval, a curated catalog, and sensitive fields that only take managed secrets.',
    group: 'trust',
    chapters: ['your-rules-hold'],
  },
  {
    path: '/trust/the-record',
    title: 'The Record',
    description:
      'Every infrastructure change is one Infra Job, kept and queryable with its configuration, cost fact, verdicts, approvals, and the snapshot of what exists afterward.',
    group: 'trust',
    chapters: ['every-deployment-leaves-a-record'],
  },
  {
    path: '/trust/security-posture',
    title: 'Security Posture',
    description:
      'Seventeen technical controls with evidence per kind, stated honestly: a kind enforces controls; it is never called compliant.',
    group: 'trust',
    chapters: ['verified-before-it-exists', 'runs-where-you-decide'],
  },
  {
    path: '/trust/your-cloud-your-keys',
    title: 'Your Cloud, Your Keys',
    description:
      'Planton deploys into your account with your keys, your state, and your bill. Keyless connections store nothing; every module is open source; you can leave with your manifests.',
    group: 'trust',
    chapters: ['runs-where-you-decide'],
  },

  // Product: one page per thing a person uses
  {
    path: '/product',
    title: 'Product',
    description:
      'Infra Hub, Service Hub, your coding agent, the CLI, the catalog, import, and the open source underneath: what Planton does in your cloud.',
    group: 'product',
    chapters: ['what-planton-is'],
  },
  {
    path: '/product/infra-hub',
    title: 'Infra Hub',
    description:
      'Every infrastructure change as a plan with its monthly cost and the permissions it needs, applied only after you review it, and saved as an Infra Chart your team redeploys.',
    group: 'product',
    chapters: ['what-planton-is', 'verified-before-it-exists', 'your-rules-hold', 'every-deployment-leaves-a-record'],
  },
  {
    path: '/product/service-hub',
    title: 'Service Hub',
    description:
      'Connect a repository and every push is built, containerized, deployed, and promoted in your environments\u2019 order, with the result written back into GitHub.',
    group: 'product',
    chapters: ['services-ship-from-git', 'your-rules-hold', 'every-deployment-leaves-a-record'],
  },
  {
    path: '/product/coding-agents',
    title: 'Coding Agents',
    description:
      'Your coding agent reaches Planton through the Planton skills and the Planton MCP server, or through the CLI, and deploys through the same door as everyone.',
    group: 'product',
    chapters: ['what-planton-is', 'the-wall', 'your-rules-hold'],
  },
  {
    path: '/product/cli',
    title: 'CLI',
    description:
      'Everything Planton does, from your terminal: validate a manifest, deploy a component or a directory in dependency order, stream the job, install an Infra Chart.',
    group: 'product',
    chapters: ['runs-where-you-decide', 'your-rules-hold', 'every-deployment-leaves-a-record'],
  },
  {
    path: '/product/catalog',
    title: 'Catalog',
    description: `${PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} catalog kinds across ${PLATFORM_STATS.CLOUD_PROVIDER_COUNT} providers, each with a cost fact sheet, a control posture with evidence, and least-privilege permissions.`,
    group: 'product',
    chapters: ['proof-it-works', 'verified-before-it-exists', 'your-rules-hold'],
  },
  {
    path: '/product/import',
    title: 'Import',
    description:
      'Bring infrastructure that already exists under Planton\u2019s record without redeploying it: adopt it, import its state in one verified step, keep the record.',
    group: 'product',
    chapters: ['bring-what-you-have', 'every-deployment-leaves-a-record'],
  },
  {
    path: '/product/open-source',
    title: 'Open Source',
    description:
      'Every infrastructure module is Apache 2.0: the catalog, the CLI and its engine. If you leave, you take your manifests and keep deploying them.',
    group: 'product',
    chapters: ['runs-where-you-decide', 'proof-it-works'],
  },

  // Distributions: where the platform runs (chapter 8); the desktop pages carry the desktop landing's own data at Planton Desktop's permanent address
  {
    path: '/distributions',
    title: 'Distributions',
    description:
      'Hosted at planton.ai, self-hosted on your Kubernetes cluster with a license that verifies offline, or free on your laptop as Planton Desktop. One model.',
    group: 'distributions',
    chapters: ['runs-where-you-decide'],
  },
  {
    path: '/distributions/hosted',
    title: 'Hosted',
    description: `The control plane at planton.ai; every deploy in your own cloud account, under your own keys, with nothing to run. Free for up to ${FREE_TIER_SEATS} seats with no card.`,
    group: 'distributions',
    chapters: ['runs-where-you-decide', 'start'],
  },
  {
    path: '/distributions/self-hosted',
    title: 'Self-Hosted',
    description: `The whole platform on your Kubernetes cluster: two Helm commands, an operator, a license that verifies offline and never bricks. Free for up to ${COMMUNITY_SEAT_LIMIT} seats.`,
    group: 'distributions',
    chapters: ['runs-where-you-decide', 'start'],
  },
  { path: '/desktop', title: 'Planton Desktop', description: 'The whole platform on your laptop, deploying to your own cloud with the sign-ins already on your machine. No account. Free for individuals, commercial use too.', group: 'distributions', chapters: ['runs-where-you-decide', 'the-wall'] },
  { path: '/desktop/download', title: 'Download Planton Desktop', description: 'Install Planton Desktop for macOS, Windows, or Linux. Free for individuals, including commercial use.', group: 'distributions', chapters: ['start'] },

  // Pricing
  { path: '/pricing', title: 'Pricing', description: 'Plans for every stage, on planton.ai or your own infrastructure. A free tier that never bills, one team plan, and self-hosted licenses that verify offline.', group: 'pricing', chapters: ['start'] },

  // Content indexes (their children are walked from markdown)
  { path: '/docs', title: 'Documentation', description: 'Guides and reference for Planton: getting started, Infra Hub, Service Hub, distributions, the CLI, and the catalog.', group: 'content' },
  { path: '/blog', title: 'Blog', description: 'Writing from the Planton team.', group: 'content' },
  { path: '/changelog', title: 'Changelog', description: 'What shipped in Planton, release by release.', group: 'content' },
  { path: '/tutorials', title: 'Tutorials', description: 'Step-by-step tutorials for deploying infrastructure and services with Planton.', group: 'content' },

  // Company and legal
  { path: '/branding/design-system', title: 'Design System', description: 'The visual laws every Planton surface follows: monochrome chrome, color only where it carries meaning, one typeface, one palette.', group: 'company' },
  { path: '/legal/privacy', title: 'Privacy Policy', description: 'How Planton collects, uses, and protects your personal data when you use the platform.', group: 'legal' },
  { path: '/legal/terms', title: 'Terms of Service', description: 'The terms governing your use of the Planton platform.', group: 'legal' },
  { path: '/legal/refund-policy', title: 'Refund Policy', description: 'How refunds work for Planton team subscriptions, self-hosted licenses, and prepaid AI credits.', group: 'legal' },

  // Standalone surfaces: reachable by link, not search results
  { path: '/book-demo', title: 'Book a Demo', description: 'See how Planton connects cloud infrastructure and application delivery. Share your details, then choose a time for a walkthrough.', group: 'standalone', index: false },
  { path: '/desktop/open', title: 'Open Planton Desktop', description: 'Hands a link off to the desktop app installed on this machine.', group: 'standalone', index: false },

] as const;

export function sitePage(path: string): SitePage {
  const found = SITE_PAGES.find((p) => p.path === path);
  if (!found) throw new Error(`"${path}" is not in the route registry (src/data/site-pages.ts)`);
  return found;
}

/** Pages advertised to crawlers. */
export const INDEXED_PAGES: readonly SitePage[] = SITE_PAGES.filter((p) => p.index !== false);

/**
 * Route prefixes that are exported but deliberately never registered: shared
 * by URL and marked noindex by their own layouts. The coverage check skips
 * them; anything else exported and unregistered fails the build.
 */
export const UNREGISTERED_PREFIXES: readonly string[] = ['/meets', '/invest', '/legal/investor-updates'];
