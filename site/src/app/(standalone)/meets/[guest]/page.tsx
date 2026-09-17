import { notFound } from 'next/navigation';
import { MeetingDeck, getLatestMeetingDeck, listMeetingDecks } from '@/components/meetings';

interface GuestPageProps {
  params: Promise<{ guest: string }>;
}

/** One page per guest; the static export needs the list up front. */
export function generateStaticParams() {
  const guests = [...new Set(listMeetingDecks().map((key) => key.split('/')[0]))];
  return guests.map((guest) => ({ guest }));
}

/**
 * /meets/[guest] renders the guest's newest deck. It is the URL a guest is
 * sent, so it stays stable while the deck behind it moves forward.
 */
export default async function GuestPage({ params }: GuestPageProps) {
  const { guest } = await params;
  const deck = getLatestMeetingDeck(guest);
  if (!deck) return notFound();
  return <MeetingDeck {...deck} />;
}
