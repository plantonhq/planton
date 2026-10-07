import story from './homepage-video-story.json' with { type: 'json' };

/** Shared editorial source for the film, native player, transcript, and analytics.
 * Bump the version when publishing different bytes; CDN objects are immutable. */
export const OVERVIEW_VIDEO = {
  id: 'homepage-overview',
  version: '2026-10-08-v1',
  title: 'Governance built in for coding agents.',
  label: 'See Planton in 45 seconds.',
  description: 'Proven infrastructure charts. Agents that work within your policies. Production approvals stay with your team.',
  poster: '/_site/images/product/homepage-overview-2026-10-08-v1.webp',
  captions: '/_site/videos/homepage-overview-2026-10-08-v1-en.vtt',
  base: 'https://assets.planton.ai/videos/homepage/2026-10-08-v1',
  duration: 44.5,
  fps: 30,
} as const;

export const OVERVIEW_CHAPTERS = story;
