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
import type { ChapterId } from './story.ts';

export type PageGroup =
  | 'home'
  | 'product'
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

  // Product (today's routes; the Product rebuild renames the group and retires the stale pages)
  { path: '/features', title: 'Product', description: 'The Self-Service Cloud Platform: AI-designed infrastructure and Git-to-production deployments, in your own cloud account, on an open-source foundation.', group: 'product', chapters: ['what-planton-is'] },
  { path: '/features/infra-hub', title: 'Infra Hub', description: 'Deploy any cloud resource, from databases to Kubernetes clusters, in minutes. 700+ resource types, Infra Charts, dependency-aware pipelines, and auditable stack jobs.', group: 'product', chapters: ['what-planton-is', 'verified-before-it-exists'] },
  { path: '/features/service-hub', title: 'Service Hub', description: 'Ship code from Git to production. Managed CI/CD with Tekton pipelines, multi-environment promotion, deploy to Kubernetes, ECS, or Cloud Run, and Kustomize-native config.', group: 'product', chapters: ['services-ship-from-git'] },
  { path: '/features/cloud-catalog', title: 'Cloud Catalog', description: 'Browse 700+ pre-built deployment modules across 8 cloud providers. Filter by provider, preview YAML configurations, and deploy to your cloud in minutes.', group: 'product', chapters: ['what-planton-is'] },
  { path: '/features/runner', title: 'Runner', description: 'Self-hosted execution agent that runs in your cloud. Planton orchestrates, Runner executes; your credentials never leave your account.', group: 'product', chapters: ['runs-where-you-decide'] },
  { path: '/features/security', title: 'Security', description: 'Built-in secrets management, identity and access control, full audit trails, and zero-trust architecture. Security is native to every layer of Planton.', group: 'product', chapters: ['runs-where-you-decide', 'every-deployment-leaves-a-record'] },
  { path: '/features/agent-fleet', title: 'Agent Fleet', description: 'Purpose-built AI agents for DevOps. Browse the marketplace, encode your runbooks as skills, orchestrate sub-agents, and stream every action in real time.', group: 'product' },
  { path: '/features/cli', title: 'CLI', description: 'Everything Planton does, from your terminal. Manifest-driven deployments, real-time stack job streaming, Kubernetes access, and environment config in one CLI.', group: 'product', chapters: ['runs-where-you-decide'] },
  { path: '/features/open-source', title: 'Open Source', description: 'The open-source foundation powering Planton. Protobuf-defined APIs, Pulumi and Terraform modules, portable KRM YAML manifests, no vendor lock-in.', group: 'product', chapters: ['runs-where-you-decide'] },
  { path: '/features/desktop', title: 'Planton Desktop', description: 'Your coding agent can already create cloud infrastructure. Planton Desktop makes it verifiable, recorded, and reusable, in your account, on your laptop, free.', group: 'product', chapters: ['runs-where-you-decide', 'the-wall'] },
  { path: '/features/desktop/download', title: 'Download Planton Desktop', description: 'Install Planton Desktop for macOS, Windows, or Linux. Free for individuals, including commercial use.', group: 'product', chapters: ['start'] },

  // Solutions (today's routes; the persona rebuild replaces them with five pages from personas.ts)
  { path: '/solutions', title: 'Solutions', description: 'How Planton serves platform engineers, engineering leaders, consultancies, founders, and security leaders.', group: 'solutions', chapters: ['who-it-is-for'] },
  { path: '/solutions/by-role/developers', title: 'For Developers', description: 'Deploy your code, not your weekend. Git-to-deploy, self-service infrastructure, without deep Kubernetes expertise.', group: 'solutions', chapters: ['who-it-is-for'] },
  { path: '/solutions/by-role/engineering-leader', title: 'For Engineering Leaders', description: 'Visibility without micromanagement. Audit trails, team autonomy with guardrails, and the proof of what was deployed.', group: 'solutions', chapters: ['who-it-is-for'] },
  { path: '/solutions/by-role/platform-engineers', title: 'For Platform Engineers', description: 'Build golden paths, not bottleneck queues. Define standards, govern credentials, and let developers self-serve.', group: 'solutions', chapters: ['who-it-is-for'] },
  { path: '/solutions/by-role/startup-founders', title: 'For Startup Founders', description: 'Ship production infrastructure without an ops hire, and redo nothing when you become a team.', group: 'solutions', chapters: ['who-it-is-for'] },
  { path: '/solutions/by-size/enterprises', title: 'Enterprises', description: 'Enterprise controls without enterprise friction. Runner security, honest control posture, multi-cloud governance.', group: 'solutions', chapters: ['who-it-is-for'] },
  { path: '/solutions/by-size/growing-teams', title: 'Growing Teams', description: 'Scale your infrastructure without scaling your ops team. Self-service, standards enforcement, and team visibility.', group: 'solutions', chapters: ['who-it-is-for'] },
  { path: '/solutions/by-size/startups', title: 'Startups', description: 'Ship production infrastructure without growing your ops team. Free tier, open-source foundation, no lock-in.', group: 'solutions', chapters: ['who-it-is-for'] },
  { path: '/solutions/by-use-case/internal-developer-platform', title: 'Internal Developer Platform', description: 'Build an IDP without building an IDP. Self-service infrastructure, managed CI/CD, access control, and AI assistance, out of the box.', group: 'solutions', chapters: ['who-it-is-for'] },
  { path: '/solutions/by-use-case/multi-cloud', title: 'Multi-Cloud', description: 'Same workflow, every cloud. One YAML manifest format, one CLI, one console, for AWS, GCP, Azure, and beyond.', group: 'solutions', chapters: ['who-it-is-for'] },
  { path: '/solutions/by-use-case/self-hosted-devops', title: 'Self-Hosted DevOps', description: 'Enterprise security with SaaS convenience. Your credentials never leave your cloud boundary.', group: 'solutions', chapters: ['who-it-is-for'] },

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
