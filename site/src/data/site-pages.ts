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
 * Retired paths live in ./retired-routes.ts.
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators without a bundler.
 */
import { COMMUNITY_SEAT_LIMIT, FREE_TIER_SEATS } from './pricing.ts';
import { PLATFORM_STATS } from './platform-stats.ts';
import type { ChapterId } from './story.ts';

export type PageGroup =
  | 'home'
  | 'product'
  | 'distributions'
  | 'trust'
  | 'solutions'
  | 'pricing'
  | 'content'
  | 'company'
  | 'legal'
  | 'standalone';

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
   * surfaces (decks, investor pages) and single-task frames are not search
   * results.
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

export const SITE_PAGES: readonly SitePage[] = [
  // Home
  {
    path: '/',
    title: 'The Self-Service Cloud Platform',
    description:
      'Planton turns your own cloud account into a self-service platform. Cost, permissions, and controls are verified before anything is created, and every deployment leaves an immutable record.',
    group: 'home',
    chapters: [
      'the-wall',
      'what-planton-is',
      'verified-before-it-exists',
      'your-rules-hold',
      'every-deployment-leaves-a-record',
      'services-ship-from-git',
      'bring-what-you-have',
      'runs-where-you-decide',
      'who-it-is-for',
      'proof-it-works',
      'how-it-compares',
      'start',
    ],
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
      'Before anything is created, Planton states the monthly cost with its coverage, the least-privilege permissions, and the controls each component enforces.',
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
      'Every infrastructure change is one stack job, kept and queryable with its configuration, cost fact, verdicts, approvals, and the snapshot of what exists afterward.',
    group: 'trust',
    chapters: ['every-deployment-leaves-a-record'],
  },
  {
    path: '/trust/security-posture',
    title: 'Security Posture',
    description:
      'Seventeen technical controls with evidence per component, crosswalked to four frameworks, stated honestly: a component enforces controls; it is never called compliant.',
    group: 'trust',
    chapters: ['verified-before-it-exists', 'runs-where-you-decide'],
  },
  {
    path: '/trust/your-cloud-your-keys',
    title: 'Your Cloud, Your Keys',
    description:
      'Planton runs in your account with your keys, your state, and your bill. Keyless connections store nothing; every module is open source; you can leave with your manifests.',
    group: 'trust',
    chapters: ['runs-where-you-decide'],
  },

  // Product: the platform engineer's half of the story, one page per thing a person uses
  {
    path: '/product',
    title: 'Product',
    description:
      'Infra Hub, Service Hub, your coding agent, the CLI, the catalog, import, and the open source underneath: what Planton is, as a platform engineer meets it.',
    group: 'product',
    chapters: ['what-planton-is'],
  },
  {
    path: '/product/infra-hub',
    title: 'Infra Hub',
    description:
      'Describe what you need, see the monthly cost and the IAM policy before anything is created, deploy, and publish it as an Infra Chart your team reuses.',
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
    description: `${PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} component kinds across ${PLATFORM_STATS.CLOUD_PROVIDER_COUNT} providers, each with a cost fact sheet, a control posture with evidence, and least-privilege permissions; ${PLATFORM_STATS.INFRA_CHART_COUNT} Infra Charts.`,
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
      'Every infrastructure module is Apache 2.0: the catalog, the Infra Charts, the CLI and its engine. If you leave, you take your manifests and keep deploying them.',
    group: 'product',
    chapters: ['runs-where-you-decide', 'proof-it-works'],
  },

  // Distributions: where the platform runs (chapter 8); the desktop pages carry the desktop landing's own data
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
  { path: '/features/desktop', title: 'Planton Desktop', description: 'The whole platform on your laptop, deploying to your own cloud with the sign-ins already on your machine. No account. Free for individuals, commercial use too.', group: 'distributions', chapters: ['runs-where-you-decide', 'the-wall'] },
  { path: '/features/desktop/download', title: 'Download Planton Desktop', description: 'Install Planton Desktop for macOS, Windows, or Linux. Free for individuals, including commercial use.', group: 'distributions', chapters: ['start'] },

  // Solutions: chapter 9 as the index, then one page per persona. The five
  // slugs are the personas' own (src/data/personas.ts); a page's chapters are
  // its persona's beats, in that person's order.
  { path: '/solutions', title: 'Solutions', description: 'Planton for the platform engineer who runs the platform, the leader who signs for it, the consultancy that delivers it, the founder who ships on it, and the security leader who governs it.', group: 'solutions', chapters: ['who-it-is-for'] },
  { path: '/solutions/platform-engineer', title: 'Planton for Platform Engineers', description: 'Self-service your developers and their coding agents cannot break: rules written once, cost and permissions verified before anything exists, a record of every deploy.', group: 'solutions', chapters: ['your-rules-hold', 'verified-before-it-exists', 'every-deployment-leaves-a-record', 'what-planton-is', 'runs-where-you-decide', 'services-ship-from-git', 'bring-what-you-have'] },
  { path: '/solutions/engineering-leader', title: 'Planton for Engineering Leaders', description: 'What your team deploys, proven before it exists: the cost, the rule that held, and the record, without a new team and without opening a console.', group: 'solutions', chapters: ['verified-before-it-exists', 'your-rules-hold', 'every-deployment-leaves-a-record', 'who-it-is-for', 'what-planton-is', 'runs-where-you-decide'] },
  { path: '/solutions/it-consultancy', title: 'Planton for IT Consultancies', description: 'One organization per client, a client environment from a published template, and everything handed back as manifests when the engagement ends.', group: 'solutions', chapters: ['what-planton-is', 'runs-where-you-decide', 'verified-before-it-exists', 'every-deployment-leaves-a-record', 'services-ship-from-git'] },
  { path: '/solutions/startup-founder', title: 'Planton for Startup Founders', description: 'Ship without an ops hire: push to deploy, the monthly cost before it exists, free to start, and nothing redone when you become a team.', group: 'solutions', chapters: ['what-planton-is', 'services-ship-from-git', 'verified-before-it-exists', 'runs-where-you-decide', 'every-deployment-leaves-a-record'] },
  { path: '/solutions/security-and-governance-leader', title: 'Planton for Security and Governance Leaders', description: 'Rules that hold at the moment of creation, controls stated with evidence and never called compliant, and a record of every change; a complement to your posture tools.', group: 'solutions', chapters: ['your-rules-hold', 'verified-before-it-exists', 'every-deployment-leaves-a-record', 'how-it-compares', 'runs-where-you-decide', 'what-planton-is'] },

  // Pricing
  { path: '/pricing', title: 'Pricing', description: 'Plans for every stage, on planton.ai or your own infrastructure. A free tier that never bills, one team plan, and self-hosted licenses that verify offline.', group: 'pricing', chapters: ['start'] },
  { path: '/pricing/enterprise', title: 'Enterprise', description: 'Enterprise at Planton: a published rate card in your market, enterprise identity, air-gap, compliance reporting, and real SLAs. Self-serve below 25 seats.', group: 'pricing', chapters: ['start'] },

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
  { path: '/book-demo', title: 'Book a Demo', description: 'Pick a time to see Planton with the founder.', group: 'standalone', index: false },
  { path: '/desktop/open', title: 'Open Planton Desktop', description: 'Hands a link off to the desktop app installed on this machine.', group: 'standalone', index: false },
  { path: '/tour', title: 'Tour', description: 'An interactive tour of Planton.', group: 'standalone', index: false },
  { path: '/demo', title: 'Interactive Demo', description: 'A guided walk through Planton.', group: 'standalone', index: false },
  { path: '/hackathon/mobile-vibe-2025', title: 'MobileVibe Hackathon 2025', description: 'A 2025 hackathon page.', group: 'standalone', index: false },

  // The persona decks: the story told for one person, with presenter notes,
  // at a stable address the founder pastes into a meeting invite. Registered
  // so one helper gives each its title, canonical, and noindex; unindexed
  // because a deck is opened by the person it was sent to, not found.
  { path: '/decks/platform-engineer', title: 'The Planton Story for Platform Engineers', description: 'The Planton story told for platform engineers, with presenter notes under every slide.', group: 'standalone', index: false },
  { path: '/decks/engineering-leader', title: 'The Planton Story for Engineering Leaders', description: 'The Planton story told for engineering leaders, with presenter notes under every slide.', group: 'standalone', index: false },
  { path: '/decks/it-consultancy', title: 'The Planton Story for IT Consultancies', description: 'The Planton story told for IT consultancies, with presenter notes under every slide.', group: 'standalone', index: false },
  { path: '/decks/startup-founder', title: 'The Planton Story for Startup Founders', description: 'The Planton story told for startup founders, with presenter notes under every slide.', group: 'standalone', index: false },
  { path: '/decks/security-and-governance-leader', title: 'The Planton Story for Security and Governance Leaders', description: 'The Planton story told for security and governance leaders, with presenter notes under every slide.', group: 'standalone', index: false },
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
export const UNREGISTERED_PREFIXES: readonly string[] = ['/meets', '/invest', '/legal/investor-updates', '/enterprise'];
