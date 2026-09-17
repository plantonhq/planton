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

  // Retired product concepts and old entry points.
  { from: '/agents', to: '/features/agent-fleet', reason: 'the agents page became the Agent Fleet product page' },
  { from: '/cli', to: '/features/cli', reason: 'the CLI page moved under the product group' },
  { from: '/features/planton-copilot', to: '/features/agent-fleet', reason: 'Planton Copilot was retired in favor of Agent Fleet' },
  { from: '/features/iac-workflows', to: '/features/infra-hub', reason: 'IaC workflows are Infra Hub' },
  { from: '/features/self-service-devops', to: '/features', reason: 'a retired concept page' },
  { from: '/features/auditable-intelligence', to: '/features', reason: 'a retired concept page' },
  { from: '/features/kubernetes-dashboard', to: '/features', reason: 'a retired concept page' },
  { from: '/docs/infrastructure/openmcf', to: '/docs/infrastructure/open-source', reason: 'the docs page was renamed' },
  { from: '/solutions/by-role/devops', to: '/solutions/by-role/platform-engineers', reason: 'the DevOps persona became the platform engineer' },
  { from: '/solutions/by-use-case/chat-ops', to: '/solutions', reason: 'a retired concept page' },
  { from: '/enterprise', to: '/pricing/enterprise', reason: 'enterprise pricing lives under pricing' },
] as const;

export function retiredRoute(from: string): RetiredRoute {
  const found = RETIRED_ROUTES.find((r) => r.from === from);
  if (!found) throw new Error(`"${from}" is not a retired route (src/data/retired-routes.ts)`);
  return found;
}
