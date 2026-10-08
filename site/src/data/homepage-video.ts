import story from './homepage-video-story.json' with { type: 'json' };

/** Shared editorial source for the homepage film, transcript exports, and analytics.
 * Bump the version when publishing different bytes; CDN objects are immutable. */
export const OVERVIEW_VIDEO = {
  id: 'homepage-overview',
  version: '2026-10-08-v3',
  title: 'Governance built in for coding agents.',
  label: 'See Planton in action.',
  description: 'Proven infrastructure charts. Agents that work within your policies. Production approvals stay with your team.',
  poster: '/_site/images/product/homepage-overview-2026-10-08-v3.webp',
  base: 'https://assets.planton.ai/videos/homepage/2026-10-08-v3',
  duration: 47,
  fps: 30,
} as const;

export const OVERVIEW_CHAPTERS = story;
