import type { FC } from 'react';
import { RecordWindow } from '@/components/marketing';
import { chapter } from '@/data/story';
import { ChapterSection } from './ChapterSection';
import { ProofList } from './ProofList';

/**
 * Chapter 4, the second proof moment: a deploy that broke a rule, paused,
 * beside the rules a platform team writes once. Only shipped behavior: the
 * budget pause, protected environments, catalog curation, managed secrets.
 * Rules over the content of a configuration are not shipped and do not appear.
 */

const ch = chapter('your-rules-hold');

const ITEMS = [
  { label: 'Deployment Budgets', text: ch.proof[0] },
  { label: 'Protected Environments', text: ch.proof[1] },
  { label: 'A Curated Catalog', text: ch.proof[2] },
  { label: 'Secrets Stay Secret', text: ch.proof[3] },
];

const ROWS = [
  { label: 'requested by', value: 'coding agent, on behalf of s.rao' },
  { label: 'environment', value: 'prod · protected · budget $250/mo' },
  { label: 'verified cost', value: '~$312/mo est.' },
  { label: 'verdict', value: 'over budget by ~$62/mo · paused for a decision' },
  { label: 'who may approve', value: 'anyone with approve access on prod · never the requester' },
  { label: 'resolution', value: 'approved by a.patel · 2026-09-17 09:14 · “the replica is intentional”' },
];

export const YourRulesHold: FC = () => (
  <ChapterSection
    chapter={ch}
    layout="split-reverse"
    aside={<ProofList items={ITEMS} />}
    readMore={{ href: '/trust/rules-and-approvals', label: 'Rules and Approvals' }}
  >
    <RecordWindow title="Deploy paused" rows={ROWS} footer="An illustration of the shape. Figures marked est. are examples; a real deploy carries its own." />
  </ChapterSection>
);
