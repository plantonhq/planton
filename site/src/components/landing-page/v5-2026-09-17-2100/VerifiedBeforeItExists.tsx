import type { FC } from 'react';
import { RecordWindow } from '@/components/marketing';
import { chapter } from '@/data/story';
import { ChapterSection } from './ChapterSection';
import { ProofList } from './ProofList';

/**
 * Chapter 3, the first proof moment: the three verdicts stamped before
 * anything exists, beside the artifact that carries them. Never a savings
 * figure, never "compliant", never an estimate dressed as a bill.
 */

const ch = chapter('verified-before-it-exists');

const ITEMS = [
  { label: 'The Monthly Cost, With Its Coverage', text: ch.proof[0] },
  { label: 'Least-Privilege Permissions', text: ch.proof[2] },
  { label: 'The Controls Each Component Enforces', text: ch.proof[3] },
];

const ROWS = [
  { label: 'verified cost', value: '~$172/mo est. · 5 of 6 components priced, 1 usage-based' },
  { label: 'catalog release', value: '2026.09.2 · prices verified against provider documents' },
  { label: 'against today', value: '+$16/mo est. · the load balancer is new' },
  { label: 'permissions', value: 'least-privilege policy · 14 actions · ready to download' },
  { label: 'controls', value: 'encryption at rest · encryption in transit · no public exposure' },
  { label: 'uncovered', value: 'none of the 6 components is without a control profile' },
];

export const VerifiedBeforeItExists: FC = () => (
  <ChapterSection
    chapter={ch}
    layout="split"
    aside={<ProofList items={ITEMS} />}
    readMore={{ href: '/trust/verified-before-deploy', label: 'Verified Before Deploy' }}
  >
    <RecordWindow title="Before deploy" rows={ROWS} footer="An illustration of the shape. Figures marked est. are examples; a real deploy carries its own." />
  </ChapterSection>
);
