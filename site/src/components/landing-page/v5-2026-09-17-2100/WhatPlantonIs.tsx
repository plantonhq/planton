import { Box } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { BodyText, Card, FeatureTitle, Grid } from '@/components/marketing';
import { POSITIONING } from '@/data/positioning';
import { chapter } from '@/data/story';
import { ChapterSection } from './ChapterSection';

/**
 * Chapter 2. The two halves, each with its one analogy, and how the visitor's
 * agent reaches them. This is the only place on the page the Level 2
 * analogies appear; every other section speaks of Planton without one.
 * "Infra Chart" is the product noun and "template" the plain word beside it.
 */

const ch = chapter('what-planton-is');

const HUBS = [
  { ...POSITIONING.infraHub, href: '/features/infra-hub' },
  { ...POSITIONING.serviceHub, href: '/features/service-hub' },
];

export const WhatPlantonIs: FC = () => (
  <ChapterSection chapter={ch}>
    <Grid cols={2} className="max-w-6xl mx-auto">
      {HUBS.map((hub) => (
        <Card key={hub.name}>
          <Box className="flex items-baseline justify-between gap-3 mb-3 flex-wrap">
            <FeatureTitle>{hub.name}</FeatureTitle>
            <BodyText className="text-xs text-fg-muted">{hub.analogy}</BodyText>
          </Box>
          <BodyText className="mb-4">{hub.line}</BodyText>
          <Link href={hub.href} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4">
            {`${hub.name} \u2192`}
          </Link>
        </Card>
      ))}
    </Grid>
    <BodyText className="text-center max-w-2xl mx-auto mt-8 text-fg-secondary">
      {ch.proof[2]}{' '}
      <Link href="/docs/coding-agents" className="text-fg-secondary hover:text-white underline underline-offset-4">
        {'Connect Your Agent \u2192'}
      </Link>
    </BodyText>
  </ChapterSection>
);
