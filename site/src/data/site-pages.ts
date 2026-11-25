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
 * Content pages (docs, changelog) are walked from their
 * markdown folders by those same scripts; only their index pages appear here.
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators without a bundler.
 */
import { HOMEPAGE } from './homepage.ts';
import type { ChapterId } from './story.ts';

export type PageGroup =
  | 'home'
  | 'pricing'
  | 'security'
  | 'desktop'
  | 'content'
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
  pricing: 'Pricing',
  security: 'Security',
  desktop: 'Planton Desktop',
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
  /** Home's card: every page without a card of its own is shared with the story's first screen. */
  defaultOgImage: `${SITE_URL}/_site/images/og/homepage.png`,
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

  // Pricing
  { path: '/pricing', title: 'Pricing', description: 'Plans for every stage, on planton.ai or your own infrastructure. A free tier that never bills, one team plan, and self-hosted licenses that verify offline.', group: 'pricing', chapters: ['start'] },

  // Security: what Planton can touch in a customer's cloud and GitHub, and how to cut it off
  {
    path: '/security',
    title: 'Security',
    description:
      'What Planton can touch in your cloud and GitHub, why it stores no cloud keys, who approves what, the record it keeps, and how to cut it off.',
    group: 'security',
    chapters: ['verified-before-it-exists', 'runs-where-you-decide', 'your-rules-hold', 'every-deployment-leaves-a-record'],
  },

  // Planton Desktop: the free edition on a laptop; the CLI's install hint and the README link here
  { path: '/desktop', title: 'Planton Desktop', description: 'The whole platform on your laptop, deploying to your own cloud with the sign-ins already on your machine. No account. Free for individuals, commercial use too.', group: 'desktop', chapters: ['runs-where-you-decide', 'the-wall'] },
  { path: '/desktop/download', title: 'Download Planton Desktop', description: 'Install Planton Desktop for macOS, Windows, or Linux. Free for individuals, including commercial use.', group: 'desktop', chapters: ['start'] },

  // Content indexes (their children are walked from markdown)
  { path: '/docs', title: 'Documentation', description: 'Guides and reference for Planton: getting started, Infra Hub, Service Hub, the CLI, and the catalog.', group: 'content' },
  { path: '/changelog', title: 'Changelog', description: 'What shipped in Planton, release by release.', group: 'content' },

  // Legal
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
