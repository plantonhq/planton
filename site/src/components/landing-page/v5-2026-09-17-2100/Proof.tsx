import Link from 'next/link';
import type { FC } from 'react';
import { BodyText, ChapterSection, Grid, PlatformCounts, TestimonialCard } from '@/components/marketing';
import { chapter } from '@/data/story';
import { TESTIMONIALS } from '@/data/testimonials';

/**
 * Chapter 10. Numbers from the platform statistics and people in their own
 * words. Every quote is verbatim from src/data/testimonials.ts and named;
 * no quote carries a dollar figure and no number is typed here.
 */

const ch = chapter('proof-it-works');

export const Proof: FC = () => (
  <ChapterSection chapter={ch}>
    <PlatformCounts className="mb-12" />
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
