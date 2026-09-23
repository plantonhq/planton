import type { ProviderId } from '../../../data/workflow-types';
import { createRoundedRoute, type Point } from './flowGeometry.ts';

/** Editorial geometry is intentionally independent of resource relationships.
 * Each edge names explicit ports and reserved lanes; no distance heuristic is
 * allowed to invent a route through another card. Coordinates use one canvas. */
export const ARCHITECTURE_LAYOUTS: Record<ProviderId, readonly Point[]> = {
  aws: [[8, 12], [8, 244], [326, 12], [326, 128], [326, 244], [644, 128], [962, 128]],
  gcp: [[8, 12], [326, 12], [644, 12], [644, 128], [326, 244], [962, 128], [962, 244]],
  azure: [[8, 12], [8, 244], [644, 244], [326, 128], [644, 12], [962, 244]],
  cloudflare: [[8, 12], [8, 244], [644, 244], [326, 128], [644, 128], [962, 128]],
  digitalocean: [[8, 12], [326, 12], [644, 12], [962, 12], [326, 244], [644, 244]],
};
export const ARCHITECTURE_SIZE = { width: 238, height: 96, viewWidth: 1208, viewHeight: 354 };
export const ARCHITECTURE_ROUTES: Record<ProviderId, Record<string, readonly Point[]>> = {
  aws: {
    'documents-knowledge': [[246,44],[320,44]],
    'vectors-knowledge': [[246,292],[286,292],[286,84],[320,84]],
    'knowledge-api': [[564,60],[604,60],[604,152],[638,152]],
    'model-api': [[564,176],[638,176]],
    'guardrail-api': [[564,292],[604,292],[604,200],[638,200]],
    'api-endpoint': [[882,176],[956,176]],
  },
  gcp: {
    'network-cluster': [[246,60],[320,60]],
    'cluster-gpu': [[564,60],[638,60]],
    'cluster-vectors': [[504,108],[504,238]],
    'gpu-model': [[763,108],[763,122]],
    'model-app': [[882,160],[956,160]],
    'vectors-app': [[564,292],[920,292],[920,200],[956,200]],
    'app-endpoint': [[1081,224],[1081,238]],
  },
  azure: {
    'environment-api': [[246,44],[638,44]],
    'database-api': [[246,292],[286,292],[286,110],[604,110],[604,84],[638,84]],
    'access-api': [[564,160],[800,160],[800,114]],
    'queue-jobs': [[882,316],[956,316]],
    'access-jobs': [[564,200],[604,200],[604,220],[920,220],[920,276],[956,276]],
  },
  cloudflare: {
    'files-worker': [[246,60],[286,60],[286,152],[320,152]],
    'data-worker': [[246,292],[286,292],[286,200],[320,200]],
    'worker-queue': [[564,176],[638,176]],
    'queue-api': [[882,152],[956,152]],
    'config-api': [[882,292],[920,292],[920,200],[956,200]],
  },
  digitalocean: {
    'network-database': [[246,60],[320,60]],
    'database-servers': [[564,44],[638,44]],
    'assets-servers': [[564,276],[604,276],[604,84],[638,84]],
    'servers-endpoint': [[882,60],[956,60]],
    'assets-cdn': [[564,316],[638,316]],
  },
};
export function architectureRoute(provider: ProviderId, edge: string, points: readonly Point[], source: number, target: number, compact: boolean) {
  if (!compact) return createRoundedRoute(ARCHITECTURE_ROUTES[provider][edge]);
  const [sx, sy] = points[source], [tx, ty] = points[target];
  const { width, height } = ARCHITECTURE_SIZE;
  // Distinct mobile gutters separate long prerequisites. Only adjacent cards
  // use the central vertical lane; cards remain full size on small screens.
  const gutter = 12 + source * 7;
  return createRoundedRoute(ty - sy === 142
    ? [[sx + width/2, sy + height], [tx + width/2, ty - 6]]
    : [[sx, sy + 64], [gutter, sy + 64], [gutter, ty + 32], [tx - 6, ty + 32]]);
}
