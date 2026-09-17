import type { PropsWithChildren } from 'react';

/**
 * Surfaces that live outside the website shell: decks, the interactive demo,
 * the tour, the investor pages, and single-task frames like booking a demo.
 * They share the site's one theme (mounted in the root layout) and nothing
 * else; each surface declares its own chrome in its own layout.
 */
export default function StandaloneLayout({ children }: PropsWithChildren) {
  return <>{children}</>;
}
