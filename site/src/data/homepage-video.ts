import story from './homepage-video-story.json' with { type: 'json' };

/** Shared editorial source for the film, native player, transcript, and analytics.
 * Bump the version when publishing different bytes; CDN objects are immutable. */
export const OVERVIEW_VIDEO = {
  id: 'homepage-overview',
  version: '2026-09-25-v2',
  title: 'From repository to running service.',
  label: 'See Planton in 60 seconds.',
  description: 'Your platform team sets the foundation. Developers ship within its controls. See where Planton fits.',
  poster: '/_site/images/product/homepage-overview.webp',
  base: 'https://assets.planton.ai/videos/homepage/2026-09-25-v2',
  duration: 60,
  fps: 30,
} as const;

export const OVERVIEW_CHAPTERS = story;
