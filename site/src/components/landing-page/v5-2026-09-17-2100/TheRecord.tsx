import type { FC } from 'react';
import { ChapterSection, ProofList, RecordWindow } from '@/components/marketing';
import { chapter } from '@/data/story';
/**
 * Chapter 5. One deploy record beside what it holds. The record leads with
 * the human facts (what, where, who) and keeps the identifier last. Never a
 * framework verdict, never "audit-ready", never a dashboard.
 */

const ch = chapter('every-deployment-leaves-a-record');

const ITEMS = [
  { label: 'The Configuration, Frozen', text: ch.proof[0] },
  { label: 'Kept and Queryable', text: ch.proof[1] },
  { label: 'One Stream, Every Surface', text: ch.proof[2] },
  { label: 'Tagged in Your Cloud', text: ch.proof[3] },
];

const ROWS = [
  { label: 'deploy', value: 'production environment · prod · succeeded' },
  { label: 'requested by', value: 'coding agent, on behalf of s.rao' },
  { label: 'approved by', value: 'a.patel · 2026-09-17 09:14 · “the replica is intentional”' },
  { label: 'configuration', value: 'embedded at creation; never changes' },
  { label: 'verified cost', value: '~$312/mo est. · catalog 2026.09.2' },
  { label: 'phases', value: 'init · refresh · preview · apply · capture' },
  { label: 'snapshot', value: '7 resources · tagged planton.ai/environment=prod' },
];

export const TheRecord: FC = () => (
  <ChapterSection
    chapter={ch}
    layout="split"
    aside={<ProofList items={ITEMS} />}
    readMore={{ href: '/trust/the-record', label: 'The Record' }}
  >
    <RecordWindow title="Deploy record" rows={ROWS} footer="An illustration of the shape. Every deploy leaves one of these." />
  </ChapterSection>
);
