'use client';

import { Deck, type SlideConfig } from '@/components/deck';

/**
 * A deck prepared for one meeting with one guest. The engine renders the
 * slides; the meeting facts (who, when, where, who presents) are the record
 * the registry keeps so a deck can be found again by guest and date. They
 * are deliberately not painted on the slides: a guest's name belongs in the
 * room, not in a screenshot.
 */
export interface MeetingDeckProps {
  slides: SlideConfig[];
  guest: string;
  meetingDate: string;
  presenter?: string;
  company?: string;
  location?: string;
}

export function MeetingDeck({ slides }: MeetingDeckProps) {
  return <Deck slides={slides} />;
}
