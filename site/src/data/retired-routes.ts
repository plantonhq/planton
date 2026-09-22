/**
 * Paths the site used to serve and where each one goes now. One list feeds
 * three things: the RetiredRoute component (the client-side forward that
 * catches an old link on any origin that serves the export directly), the
 * edge redirect declared in the estate (a true 301 on the apex), and the
 * sitemap, which never lists a retired path.
 *
 * A path is added here the day its page is retired and never removed: old
 * links live in search results, bookmarks, and other people's documents for
 * years, and a retired path that 404s is a broken promise.
 */
export interface RetiredRoute {
  /** The path that no longer has a page. */
  from: string;
  /** The live page that answers for it. */
  to: string;
  /** One line on why, for the reader of this file. */
  reason: string;
}

export const RETIRED_ROUTES: readonly RetiredRoute[] = [
  // Legal pages moved under /legal/, the prefix the apex passes through to the site.
  { from: '/privacy', to: '/legal/privacy', reason: 'legal pages live under /legal/' },
  { from: '/terms', to: '/legal/terms', reason: 'legal pages live under /legal/' },
  { from: '/refund-policy', to: '/legal/refund-policy', reason: 'legal pages live under /legal/' },

  // The Product group moved from /features to /product, and the pages that
  // were not products went where their subject lives: Security to the Trust
  // section, Runner to the hosted distribution, Agent Fleet (retired with
  // Copilot before it) to the coding-agents page that describes what shipped.
  { from: '/features', to: '/product', reason: 'the Product group lives at /product' },
  { from: '/features/infra-hub', to: '/product/infra-hub', reason: 'the Product group lives at /product' },
  { from: '/features/service-hub', to: '/product/service-hub', reason: 'the Product group lives at /product' },
  { from: '/features/cli', to: '/product/cli', reason: 'the Product group lives at /product' },
  { from: '/features/open-source', to: '/product/open-source', reason: 'the Product group lives at /product' },
  { from: '/features/cloud-catalog', to: '/product/catalog', reason: 'the catalog page is /product/catalog' },
  { from: '/features/agent-fleet', to: '/product/coding-agents', reason: 'Agent Fleet was retired; the coding-agents page describes what shipped' },
  { from: '/features/planton-copilot', to: '/product/coding-agents', reason: 'Planton Copilot was retired; the coding-agents page describes what shipped' },
  { from: '/features/security', to: '/trust/security-posture', reason: 'security is a Trust page, not a product' },
  { from: '/features/runner', to: '/distributions/hosted', reason: 'the runner is how the hosted distribution reaches your cloud' },
  { from: '/features/iac-workflows', to: '/product/infra-hub', reason: 'IaC workflows are Infra Hub' },
  { from: '/features/self-service-devops', to: '/product', reason: 'a retired concept page' },
  { from: '/features/auditable-intelligence', to: '/product', reason: 'a retired concept page' },
  { from: '/features/kubernetes-dashboard', to: '/product', reason: 'a retired concept page' },
  { from: '/agents', to: '/product/coding-agents', reason: 'the agents page became the coding-agents product page' },
  // Planton Desktop's permanent address is /desktop. The release pipeline printed
  // the old address into every installed cask and installer, so these two
  // forwards are kept for as long as any of those exist.
  { from: '/features/desktop', to: '/desktop', reason: 'Planton Desktop lives at /desktop' },
  { from: '/features/desktop/download', to: '/desktop/download', reason: 'Planton Desktop lives at /desktop' },
  { from: '/cli', to: '/product/cli', reason: 'the CLI page lives under the product group' },
  { from: '/docs/infrastructure/openmcf', to: '/docs/infrastructure/open-source', reason: 'the docs page was renamed' },
  // The Solutions section became five persona pages. Each old page forwards
  // to the person who would have read it; the ones that were about a shape
  // or a product go where that subject lives. Developers go to the coding
  // agents page: the developer's door in 2026 is the agent, and the story
  // never headlines "built for developers".
  { from: '/solutions/by-role/devops', to: '/solutions/platform-engineer', reason: 'the DevOps persona became the platform engineer' },
  { from: '/solutions/by-role/platform-engineers', to: '/solutions/platform-engineer', reason: 'the persona page replaced the role page' },
  { from: '/solutions/by-use-case/internal-developer-platform', to: '/solutions/platform-engineer', reason: 'the platform engineer builds the platform' },
  { from: '/solutions/by-role/engineering-leader', to: '/solutions/engineering-leader', reason: 'the persona page replaced the role page' },
  { from: '/solutions/by-size/growing-teams', to: '/solutions/engineering-leader', reason: 'the page argued no new team; that is the leader\u2019s chapter' },
  { from: '/solutions/by-size/enterprises', to: '/solutions/security-and-governance-leader', reason: 'the page was about security posture' },
  { from: '/solutions/by-role/startup-founders', to: '/solutions/startup-founder', reason: 'the persona page replaced the role page' },
  { from: '/solutions/by-size/startups', to: '/solutions/startup-founder', reason: 'a startup is its founder\u2019s page' },
  { from: '/solutions/by-use-case/multi-cloud', to: '/solutions/it-consultancy', reason: 'many clouds is the consultancy\u2019s wall' },
  { from: '/solutions/by-use-case/self-hosted-devops', to: '/distributions/self-hosted', reason: 'a distribution, not a persona' },
  { from: '/solutions/by-role/developers', to: '/product/coding-agents', reason: 'the developer\u2019s door is the agent' },
  { from: '/solutions/by-use-case/chat-ops', to: '/solutions', reason: 'a retired concept page' },
  { from: '/enterprise', to: '/pricing/enterprise', reason: 'enterprise pricing lives under pricing' },
  // The interactive demo and the tour told a 2025 story and no longer worked;
  // both will be redone from first principles, never patched. The hackathon
  // was a 2025 event. A past event and a broken walkthrough answer to the
  // pages that tell the story now.
  { from: '/demo', to: '/', reason: 'the interactive demo was retired; the landing carries the proof record and offers the live demo' },
  { from: '/tour', to: '/product', reason: 'the console tour was retired; the product index is its map' },
  { from: '/hackathon/mobile-vibe-2025', to: '/', reason: 'a 2025 event' },
] as const;

export function retiredRoute(from: string): RetiredRoute {
  const found = RETIRED_ROUTES.find((r) => r.from === from);
  if (!found) throw new Error(`"${from}" is not a retired route (src/data/retired-routes.ts)`);
  return found;
}
