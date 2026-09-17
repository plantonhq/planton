import { Box } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { BodyText, Grid, Metric, TestimonialCard } from '@/components/marketing';
import { chapter } from '@/data/story';
import { PLATFORM_COUNTS, PLATFORM_STATS } from '@/data/platform-stats';
import { TESTIMONIALS } from '@/data/testimonials';
import { ChapterSection } from './ChapterSection';

/**
 * Chapter 10. Numbers from the platform statistics and people in their own
 * words. Every quote is verbatim from src/data/testimonials.ts and named;
 * no quote carries a dollar figure and no number is typed here.
 */

const ch = chapter('proof-it-works');

const STATS = [
  { value: PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT, label: 'Component Kinds' },
  { value: PLATFORM_STATS.CLOUD_PROVIDER_COUNT, label: 'Providers' },
  { value: PLATFORM_STATS.CONTROL_COUNT, label: 'Controls With Evidence' },
  { value: `Since ${PLATFORM_STATS.IN_PRODUCTION_SINCE}`, label: 'In Production' },
];

export const Proof: FC = () => (
  <ChapterSection chapter={ch}>
    <Box className="grid grid-cols-2 md:grid-cols-4 gap-6 max-w-3xl mx-auto mb-3">
      {STATS.map((stat) => (
        <Metric key={stat.label} value={stat.value} label={stat.label} />
      ))}
    </Box>
    <BodyText className="text-center text-xs text-fg-muted mb-12">
      Catalog figures counted from the open-source repository on {PLATFORM_COUNTS.countedOn} (
      <Link href="https://github.com/plantonhq/planton" className="underline underline-offset-4 hover:text-white">github.com/plantonhq/planton</Link>
      ); {PLATFORM_STATS.IN_PRODUCTION_SINCE} is the year the first customer went to production.
    </BodyText>
    <Grid cols={2} className="max-w-6xl mx-auto mb-8">
      {TESTIMONIALS.map((t) => (
        <TestimonialCard key={t.name} name={t.name} role={t.role} company={t.company} location={t.location} quote={t.quote} />
      ))}
    </Grid>
    <BodyText className="text-center text-fg-secondary">
      {ch.proof[2]}{' '}
      <Link href="/product/open-source" className="text-fg-secondary hover:text-white underline underline-offset-4">
        {'Open-Source Modules \u2192'}
      </Link>
    </BodyText>
  </ChapterSection>
);
