import type { FC } from 'react';
import { ProofList, RecordWindow } from '@/components/marketing';
import { chapter } from '@/data/story';
import { ChapterSection } from './ChapterSection';

/**
 * Chapter 7, short by design. Existing infrastructure joins the record
 * through adopt and verified import. The AI-assisted account scan is not
 * released and does not appear.
 */

const ch = chapter('bring-what-you-have');

const ITEMS = [
  { label: 'Adopt Without Redeploying', text: ch.proof[0] },
  { label: 'Import, Verified Atomically', text: ch.proof[1] },
  { label: 'Recipes for the Common Kinds', text: ch.proof[2] },
];

const ROWS = [
  { label: 'resource', value: 'S3 bucket · billing-exports · existing' },
  { label: 'state', value: 'adopted, not yet deployed' },
  { label: 'import', value: 'verified · live state matches the manifest · 0 changes planned' },
  { label: 'record', value: 'from here on, every change is a stack job' },
];

export const BringWhatYouHave: FC = () => (
  <ChapterSection chapter={ch} layout="split" aside={<ProofList items={ITEMS} />}>
    <RecordWindow title="Import" rows={ROWS} footer="An illustration of the shape." />
  </ChapterSection>
);
