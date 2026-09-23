import type { FC } from 'react';
import { ChapterSection, ProofList, RecordWindow } from '@/components/marketing';
import { chapter } from '@/data/story';
/**
 * Chapter 6. Service Hub gets its full beat: push, build, deploy, written
 * back to GitHub, through the same protected environments. The analogy for
 * this hub lives in chapter 2 and is not repeated here. Image scanning and
 * health checks are not shipped and do not appear.
 */

const ch = chapter('services-ship-from-git');

const ITEMS = [
  { label: 'Build', text: ch.proof[0] },
  { label: 'Promote', text: ch.proof[1] },
  { label: 'Protect', text: ch.proof[2] },
  { label: 'Cost', text: ch.proof[3] },
];

const ROWS = [
  { label: 'push', value: 'main · a41f2c9 · “add the invoices endpoint”' },
  { label: 'build', value: 'buildpacks · 2m 04s · image pushed' },
  { label: 'deploy', value: 'dev · succeeded · 1m 12s' },
  { label: 'promote', value: 'staging succeeded · prod paused (protected)' },
  { label: 'github', value: 'check planton/deploy passed · deployments dev, staging' },
];

export const ServicesShipFromGit: FC = () => (
  <ChapterSection
    chapter={ch}
    layout="split-reverse"
    aside={<ProofList items={ITEMS} />}
    readMore={{ href: '/product/service-hub', label: 'Service Hub' }}
  >
    <RecordWindow title="Service run" rows={ROWS} footer="An illustration of the shape; times are examples." />
  </ChapterSection>
);
