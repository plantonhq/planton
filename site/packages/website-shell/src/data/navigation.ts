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

/**
 * The header's links, in the order a stranger needs them once Home has told
 * the story: how to use it, what it costs, and what it can touch in their
 * cloud. Sign In and Start Free follow as the header's doors.
 */
export const headerLinks: MenuItem[] = [
  { label: 'Docs', href: '/docs' },
  { label: 'Pricing', href: '/pricing' },
  { label: 'Security', href: '/security' },
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
    // The site's own pages beyond Home: what it costs, what it can touch in
    // your cloud, what shipped, and the free laptop edition.
    id: 'planton',
    title: 'Planton',
    items: [
      { label: 'Pricing', href: '/pricing' },
      { label: 'Security', href: '/security' },
      { label: 'Changelog', href: '/changelog' },
      { label: 'Planton Desktop', href: '/desktop' },
    ],
  },
  {
    // The two ways in: a new organization, or the docs that teach it.
    id: 'get_started',
    title: 'Get Started',
    items: [START_FREE, { label: 'Docs', href: '/docs' }],
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
