import { Box, Typography } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { BodyText, FeatureTitle, Section, SectionTitle } from '@/components/marketing';
import { chapter } from '@/data/story';
import { Doors } from './Doors';

/**
 * Chapters 11 and 12 folded into one close: how Planton sits beside the
 * tools a reader already has (categories described, never a vendor named),
 * then the same two doors the hero opened with, and the one place a demo is
 * offered. Prices are not typed here; the pricing page carries them.
 */

const compare = chapter('how-it-compares');
const start = chapter('start');

/** Chapter 11's three contrasts, each a category described and never a vendor named. */
const CONTRASTS = [
  { label: 'Beside Governance Tools', text: compare.proof[0] },
  { label: 'Beside Terraform', text: compare.proof[1] },
  { label: 'Beside a Portal', text: compare.proof[2] },
];

export const Close: FC = () => (
  <Section id="start">
    <Box className="max-w-3xl mx-auto text-center flex flex-col items-center gap-6">
      <SectionTitle>Where It Sits, and Where to Start</SectionTitle>
      <Box className="grid grid-cols-1 md:grid-cols-3 gap-6 text-left w-full">
        {CONTRASTS.map((c) => (
          <Box key={c.label} className="border-l-2 border-edge-hover pl-4">
            <FeatureTitle className="text-sm md:text-base mb-1">{c.label}</FeatureTitle>
            <BodyText>{c.text}</BodyText>
          </Box>
        ))}
      </Box>
      <BodyText className="text-fg">{start.claim}</BodyText>
      <Doors className="mt-2" />
      <Typography className="text-xs text-fg-faint">
        <Link href="/pricing" className="text-fg-secondary hover:text-white underline underline-offset-4">Pricing</Link>
        {' \u00b7 '}
        <Link href="/book-demo" className="text-fg-secondary hover:text-white underline underline-offset-4">Book a Demo With the Founder</Link>
      </Typography>
    </Box>
  </Section>
);
