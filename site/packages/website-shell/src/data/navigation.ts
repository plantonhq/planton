/**
 * The website's navigation as data: the header's links, its two doors, and
 * the footer's groups. A link appears here once, and every surface that
 * shows it reads it, so the header and the footer can never disagree about
 * where a page lives or what it is called.
 *
 * The header is deliberately small: four plain links a stranger needs, a
 * quiet Sign In, and Start Free as the one primary door. Hosted is the
 * front door, so the desktop download and the demo booking live in the
 * footer and on their own pages, never beside Start Free.
 *
 * Every href is a path the site's route registry knows (or a console path
 * the link gate allows, or an absolute URL); the site's build proves each
 * one resolves, because the header and footer render on every exported
 * page. This package cannot import the registry (the console consumes the
 * shell too), so the hrefs are literal and the gate is the proof. Labels
 * are Title Case (they are chrome).
 */

export interface MenuItem {
  label: string;
  href: string;
}

export interface FooterGroup {
  id: string;
  title: string;
  items: MenuItem[];
}

// ---------------------------------------------------------------------------
// Header
// ---------------------------------------------------------------------------

/** The header's links, in the order a stranger needs them: what it is, what it costs, how to use it, what shipped. */
export const headerLinks: MenuItem[] = [
  { label: 'Product', href: '/product' },
  { label: 'Pricing', href: '/pricing' },
  { label: 'Docs', href: '/docs' },
  { label: 'Changelog', href: '/changelog' },
];

/** The one primary door: a new organization, whose first step is the Setup Prompt. */
export const START_FREE: MenuItem = { label: 'Start Free', href: 'https://planton.ai/signup' };
export const SIGN_IN: MenuItem = { label: 'Sign In', href: 'https://planton.ai/login' };

// ---------------------------------------------------------------------------
// Footer
// ---------------------------------------------------------------------------

export const DISCORD_URL = 'https://discord.gg/pwcSapdQAp';
export const STATUS_URL = 'https://status.planton.ai';

export const footerGroups: FooterGroup[] = [
  {
    // What Planton is and what it costs, plus the one page a cautious reader
    // looks for before anything else: what Planton can and cannot do in their cloud.
    id: 'product',
    title: 'Product',
    items: [
      { label: 'Overview', href: '/product' },
      { label: 'Pricing', href: '/pricing' },
      { label: 'Trust', href: '/trust' },
      { label: 'Changelog', href: '/changelog' },
    ],
  },
  {
    id: 'get_started',
    title: 'Get Started',
    items: [START_FREE, { label: 'Planton Desktop', href: '/desktop' }],
  },
  {
    id: 'learn',
    title: 'Learn',
    items: [
      { label: 'Docs', href: '/docs' },
      { label: 'Blog', href: '/blog' },
      { label: 'Tutorials', href: '/tutorials' },
    ],
  },
  {
    // What is on GitHub, for the reader who wants the source rather than the page about it.
    id: 'open_source',
    title: 'Open Source',
    items: [
      { label: 'Planton on GitHub', href: 'https://github.com/plantonhq/planton' },
      { label: 'The Catalog', href: 'https://github.com/plantonhq/planton/tree/main/catalog' },
    ],
  },
  {
    // Whether Planton is up, and where a person answers.
    id: 'help',
    title: 'Help',
    items: [
      { label: 'Status', href: STATUS_URL },
      { label: 'Discord Community', href: DISCORD_URL },
    ],
  },
];

/** The legal pages, on the footer's bottom line. */
export const footerTermsLinks: MenuItem[] = [
  { label: 'Privacy', href: '/legal/privacy' },
  { label: 'Terms', href: '/legal/terms' },
  { label: 'Refunds', href: '/legal/refund-policy' },
];
