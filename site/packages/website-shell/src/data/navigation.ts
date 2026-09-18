import type { ReactNode } from 'react';

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
  items: { title: string; url: string }[];
}

// ---------------------------------------------------------------------------
// Header — Product mega-menu
// ---------------------------------------------------------------------------

// The order is the order a platform engineer meets the product: the two hubs,
// the coding agent as a first-class user, the terminal, the catalog, what you
// already have, and the open source underneath. Sub-labels are one line each
// and claim nothing the page does not; the page's registry entry carries the
// full description.
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

export const menuExplorer: MenuItem[] = [
  { label: 'All Product', href: '/product' },
  { label: 'How Planton Compares', href: '/compare' },
  { label: 'Documentation', href: '/docs' },
  { label: 'Tutorials', href: '/tutorials' },
  { label: 'Blog', href: '/blog' },
  { label: 'Changelog', href: '/changelog' },
];

// ---------------------------------------------------------------------------
// Header — Solutions mega-menu
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
// Header — Resources mega-menu
// ---------------------------------------------------------------------------

export const menuResources: MenuItem[] = [
  { label: 'Docs', subLabel: 'Guides, references, and platform overview', href: '/docs' },
  { label: 'Tutorials', subLabel: 'Step-by-step deployment walkthroughs', href: '/tutorials' },
  { label: 'Blog', subLabel: 'Product updates and engineering insights', href: '/blog' },
  { label: 'Changelog', subLabel: 'What shipped in every release', href: '/changelog' },
  { label: 'Tour', subLabel: 'Interactive walkthrough of the console', href: '/tour' },
  { label: 'Demo', subLabel: 'See Planton in action', href: '/demo' },
];

// ---------------------------------------------------------------------------
// Footer link groups — canonical URLs (aligned with header)
// ---------------------------------------------------------------------------

export const footerGroups: FooterGroup[] = [
  {
    title: 'Product',
    id: 'product',
    items: [
      { title: 'Infra Hub', url: '/product/infra-hub' },
      { title: 'Service Hub', url: '/product/service-hub' },
      { title: 'Coding Agents', url: '/product/coding-agents' },
      { title: 'CLI', url: '/product/cli' },
      { title: 'Catalog', url: '/product/catalog' },
      { title: 'Import', url: '/product/import' },
    ],
  },
  {
    title: 'Open Source',
    id: 'open_source',
    items: [
      { title: 'Planton Open Source', url: '/product/open-source' },
      { title: 'Infra Charts', url: 'https://github.com/plantonhq/planton/tree/main/charts' },
    ],
  },
  {
    title: 'Get Started',
    id: 'get_started',
    items: [
      { title: 'Download Planton Desktop', url: '/desktop/download' },
      { title: 'Sign Up', url: '/signup' },
      { title: 'Pricing', url: '/pricing' },
      { title: 'Book a Demo', url: '/book-demo' },
    ],
  },
  {
    title: 'Resources',
    id: 'resources',
    items: [
      { title: 'Documentation', url: '/docs' },
      { title: 'Tutorials', url: '/tutorials' },
      { title: 'Blog', url: '/blog' },
      { title: 'Changelog', url: '/changelog' },
    ],
  },
  {
    title: 'Explore',
    id: 'explore',
    items: [
      { title: 'All Product', url: '/product' },
      { title: 'Distributions', url: '/distributions' },
      { title: 'Solutions', url: '/solutions' },
      { title: 'Tour', url: '/tour' },
    ],
  },
];

export const footerTermsLinks = [
  { title: 'Status', url: '/' },
  { title: 'Privacy', url: '/legal/privacy' },
  { title: 'Terms', url: '/legal/terms' },
  { title: 'Refunds', url: '/legal/refund-policy' },
];

export const DISCORD_URL = 'https://discord.gg/pwcSapdQAp';
