/**
 * Decks prepared for meetings with prospects and customers, one folder per
 * guest under decks/, found by guest and date through the registry. The
 * public URL is /meets/<guest>; the newest deck answers it.
 */
export { MeetingDeck } from './MeetingDeck';
export type { MeetingDeckProps } from './MeetingDeck';
export { getMeetingDeck, getLatestMeetingDeck, listMeetingDecks } from './decks';
export type { MeetingDeckConfig } from './decks';
