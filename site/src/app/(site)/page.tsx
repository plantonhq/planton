import { pageMetadata } from '@/lib/page-metadata';
import {
  Hero,
  TheWall,
  WhatPlantonIs,
  VerifiedBeforeItExists,
  YourRulesHold,
  TheRecord,
  ServicesShipFromGit,
  BringWhatYouHave,
  RunsWhereYouDecide,
  WhoItIsFor,
  Proof,
  Close,
} from '@/components/landing-page';

export const metadata = pageMetadata('/');

/** The home page is the story in order; the sections are the chapters. */
export default function Home() {
  return (
    <main className="overflow-x-hidden">
      <Hero />
      <TheWall />
      <WhatPlantonIs />
      <VerifiedBeforeItExists />
      <YourRulesHold />
      <TheRecord />
      <ServicesShipFromGit />
      <BringWhatYouHave />
      <RunsWhereYouDecide />
      <WhoItIsFor />
      <Proof />
      <Close />
    </main>
  );
}
