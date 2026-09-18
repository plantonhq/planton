import type { ReactNode } from 'react';

/**
 * The website's navigation as data: the header's menus are the source, and
 * the footer is a projection of them. A link appears here once, and every
 * surface that shows it reads it, so the header and the footer can never
 * disagree about where a page lives or what it is called.
 *
 * Every href is a path the site's route registry knows (or a console path
 * the link gate allows, or an absolute URL); the site's build proves each
 * one resolves, because the header and footer render on every exported
 * page. This package cannot import the registry (the console consumes the
 * shell too), so the hrefs are literal and the gate is the proof. Labels
 * are Title Case (they are chrome); a sub-label is one line of menu copy,
 * distinct from the paragraph the registry holds as the page's description.
 * The call to action for /signup is "Start Free" here as it is on every
 * door on the site.
 */

export interface MenuItem {
  label: string;
  subLabel?: string;
  href: string;
  icon?: ReactNode;
}

export interface MenuSection {
  title?: string;
  items: MenuItem[];
}

export interface FooterGroup {
  id: string;
  title: string;
  items: MenuItem[];
}

// ---------------------------------------------------------------------------
// Header: the Product mega-menu
// ---------------------------------------------------------------------------

// The order is the order a platform engineer meets the product: the two hubs,
// the coding agent as a first-class user, the terminal, the catalog, what you
// already have, and the open source underneath.
export const menuProduct: MenuItem[] = [
  { label: 'Infra Hub', subLabel: 'Cost and permissions verified before anything is created', href: '/product/infra-hub' },
  { label: 'Service Hub', subLabel: 'Every push built, deployed, and written back to GitHub', href: '/product/service-hub' },
  { label: 'Coding Agents', subLabel: 'Your agent deploys through the same door as everyone', href: '/product/coding-agents' },
  { label: 'CLI', subLabel: 'Everything Planton does, from your terminal', href: '/product/cli' },
  { label: 'Catalog', subLabel: '700+ component kinds, each with its own fact sheet', href: '/product/catalog' },
  { label: 'Import', subLabel: 'Bring what already exists under the record', href: '/product/import' },
  { label: 'Open Source', subLabel: 'Every module Apache 2.0; leave with your manifests', href: '/product/open-source' },
];

// Where the platform runs (one model, three shapes).
export const menuDistributions: MenuItem[] = [
  { label: 'Hosted', subLabel: 'Nothing to run; your account, your keys', href: '/distributions/hosted' },
  { label: 'Self-Hosted', subLabel: 'Two Helm installs on your own cluster', href: '/distributions/self-hosted' },
  { label: 'Desktop', subLabel: 'Free for individuals, commercial use included', href: '/desktop' },
];

// The site's map in one column: the section indexes and the page that
// compares Planton with what a reader already runs. The header's Explore
// column and the footer's Explore group both read this list.
export const menuExplore: MenuItem[] = [
  { label: 'All Product', href: '/product' },
  { label: 'Distributions', href: '/distributions' },
  { label: 'Solutions', href: '/solutions' },
  { label: 'How Planton Compares', href: '/compare' },
];

// ---------------------------------------------------------------------------
// Header: the Solutions mega-menu
// ---------------------------------------------------------------------------

// The five people the story is told to, one page each, in the story's own
// order (the user first, then the one who signs). The labels are the persona
// names exactly as the site's persona records spell them; a person's name
// needs no sub-label and no icon.
export const menuSolutions: MenuItem[] = [
  { label: 'Platform Engineer', href: '/solutions/platform-engineer' },
  { label: 'Engineering Leader', href: '/solutions/engineering-leader' },
  { label: 'IT Consultancy', href: '/solutions/it-consultancy' },
  { label: 'Startup Founder', href: '/solutions/startup-founder' },
  { label: 'Security and Governance Leader', href: '/solutions/security-and-governance-leader' },
];

// ---------------------------------------------------------------------------
// Header: the Resources mega-menu
// ---------------------------------------------------------------------------

export const menuResources: MenuItem[] = [
  { label: 'Docs', subLabel: 'Guides, references, and platform overview', href: '/docs' },
  { label: 'Tutorials', subLabel: 'Step-by-step deployment walkthroughs', href: '/tutorials' },
  { label: 'Blog', subLabel: 'Product updates and engineering insights', href: '/blog' },
  { label: 'Changelog', subLabel: 'What shipped in every release', href: '/changelog' },
];

// ---------------------------------------------------------------------------
// The doors the header and footer share
// ---------------------------------------------------------------------------

export const START_FREE: MenuItem = { label: 'Start Free', href: '/signup' };
export const SIGN_IN: MenuItem = { label: 'Sign In', href: '/login' };
export const DOWNLOAD_DESKTOP: MenuItem = { label: 'Download Planton Desktop', href: '/desktop/download' };

// ---------------------------------------------------------------------------
// Footer: a projection of the menus, plus the two groups no menu holds
// ---------------------------------------------------------------------------

/** A menu as a footer group: labels and hrefs only, the sub-labels are the header's. */
const asGroup = (id: string, title: string, items: MenuItem[]): FooterGroup => ({
  id,
  title,
  items: items.map(({ label, href }) => ({ label, href })),
});

export const footerGroups: FooterGroup[] = [
  asGroup('product', 'Product', menuProduct),
  {
    // What is on GitHub, for the reader who wants the source rather than the page about it.
    id: 'open_source',
    title: 'Open Source',
    items: [
      { label: 'Planton on GitHub', href: 'https://github.com/plantonhq/planton' },
      { label: 'The Catalog', href: 'https://github.com/plantonhq/planton/tree/main/catalog' },
      { label: 'Infra Charts', href: 'https://github.com/plantonhq/planton/tree/main/charts' },
    ],
  },
  {
    id: 'get_started',
    title: 'Get Started',
    items: [DOWNLOAD_DESKTOP, START_FREE, { label: 'Pricing', href: '/pricing' }, { label: 'Book a Demo', href: '/book-demo' }],
  },
  asGroup('resources', 'Resources', menuResources),
  asGroup('explore', 'Explore', menuExplore),
];

// The legal pages live under /legal/. A status page joins this line the day
// one exists; a link to nothing is not a promise.
export const footerTermsLinks: MenuItem[] = [
  { label: 'Privacy', href: '/legal/privacy' },
  { label: 'Terms', href: '/legal/terms' },
  { label: 'Refunds', href: '/legal/refund-policy' },
];

export const DISCORD_URL = 'https://discord.gg/pwcSapdQAp';
