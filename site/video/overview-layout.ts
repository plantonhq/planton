import { createRoundedRoute, type Point } from '../src/components/marketing/workflows/flowGeometry';

export type VideoCard = { x: number; y: number; width: number };
export const CARD_HEIGHT = 228;
const row = (xs: number[], y: number, width: number): VideoCard[] => xs.map(x => ({ x, y, width }));
export const OVERVIEW_LAYOUT = {
  intro: row([120, 750, 1380], 440, 420),
  foundation: [{ x: 120, y: 485, width: 420 }, { x: 750, y: 345, width: 420 }, { x: 750, y: 640, width: 420 }],
  delivery: row([90, 550, 1010, 1470], 455, 360),
  production: row([120, 750, 1380], 455, 420),
} as const;

// Reserve explicit lanes and perpendicular ports. Geometry stays independent of
// chapter copy and handoff timing, and uses the browser's sampled motion paths.
const route = (points: Point[]) => createRoundedRoute(points, 34);
const connections = (cards: readonly VideoCard[]) => cards.slice(1).map((b, i) => {
  const a = cards[i];
  return route([[a.x + a.width, a.y + CARD_HEIGHT / 2], [b.x - 14, b.y + CARD_HEIGHT / 2]]);
});
export const OVERVIEW_ROUTES = {
  intro: connections(OVERVIEW_LAYOUT.intro),
  foundation: [
    route([[540, 565], [630, 565], [630, 459], [736, 459]]),
    route([[540, 633], [665, 633], [665, 754], [736, 754]]),
  ],
  delivery: connections(OVERVIEW_LAYOUT.delivery),
  production: connections(OVERVIEW_LAYOUT.production),
} as const;
