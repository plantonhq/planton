import type { MeetingDeckProps } from '../MeetingDeck';

/** The record a meeting deck is filed under: its slides plus the meeting facts. */
export type MeetingDeckConfig = MeetingDeckProps;

import { sepConfig } from './sep/config';
import { niravConfig } from './nirav/config';
import { clearRouteConfig } from './clear-route/config';
import { rahulGulatiConfig } from './rahul-gulati/config';

/**
 * Every meeting deck, keyed "guest/yyyy-mm-dd-hhmm" (24-hour clock), e.g.
 * "sep/2026-01-23-1400". The guest segment is also the public URL under
 * /meets/, which is why a key is never renamed once it has been shared.
 */
const meetingDeckRegistry: Record<string, MeetingDeckConfig> = {
  'sep/2026-01-23-1400': sepConfig,
  'nirav/2026-05-08-2030': niravConfig,
  'clear-route/2026-08-12-1100': clearRouteConfig,
  'rahul-gulati/2026-08-17-1700': rahulGulatiConfig,
};

/** The deck for one guest on one date, or null. */
export function getMeetingDeck(guest: string, date: string): MeetingDeckConfig | null {
  return meetingDeckRegistry[`${guest}/${date}`] || null;
}

/**
 * The most recent deck for a guest. /meets/[guest] renders this one, so the
 * URL a guest was sent keeps pointing at the newest meeting; the date format
 * sorts lexically, which is what makes the plain sort correct.
 */
export function getLatestMeetingDeck(guest: string): MeetingDeckConfig | null {
  const dates = Object.keys(meetingDeckRegistry)
    .filter((key) => key.startsWith(`${guest}/`))
    .map((key) => key.split('/')[1])
    .sort()
    .reverse();
  if (dates.length === 0) return null;
  return meetingDeckRegistry[`${guest}/${dates[0]}`] || null;
}

/** Every "guest/date" key; the static export pre-renders one page per key. */
export function listMeetingDecks(): string[] {
  return Object.keys(meetingDeckRegistry);
}
