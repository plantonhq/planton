import { notFound } from 'next/navigation';
import { MeetingDeck, getMeetingDeck, listMeetingDecks } from '@/components/meetings';

interface MeetsPageProps {
  params: Promise<{ guest: string; date: string }>;
}

/** One page per registered "guest/date"; the static export needs the list up front. */
export function generateStaticParams() {
  return listMeetingDecks().map((key) => {
    const [guest, date] = key.split('/');
    return { guest, date };
  });
}

/**
 * /meets/[guest]/[date] is a meeting's permanent address (yyyy-mm-dd-hhmm,
 * 24-hour clock), so an earlier deck can still be opened after a newer one
 * takes over /meets/[guest].
 */
export default async function MeetsPage({ params }: MeetsPageProps) {
  const { guest, date } = await params;
  const deck = getMeetingDeck(guest, date);
  if (!deck) return notFound();
  return <MeetingDeck {...deck} />;
}
