import { Box } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { Badge, BodyText, Card, ChapterSection, FeatureTitle, Grid } from '@/components/marketing';
import { chapter } from '@/data/story';
import { POSITIONING } from '@/data/positioning';
import { COMMUNITY_SEAT_LIMIT, FREE_TIER_SEATS } from '@/data/pricing';
import { DESKTOP_LANDING_PATH } from '@/data/desktop-download';
/**
 * Chapter 8. Three shapes, one model; then keyless connections, open source,
 * and the exit path said concretely. Prices read from the pricing data. No
 * "zero lock-in" badge; the exit path is the sentence. No mobile store link.
 */

const ch = chapter('runs-where-you-decide');

const SHAPES = [
  {
    name: 'Hosted',
    badge: `Free for Up to ${FREE_TIER_SEATS} Seats`,
    text: 'Sign up at planton.ai and connect your cloud. Your account, your keys; Planton holds the record.',
    href: '/signup',
    cta: 'Start Free',
  },
  {
    name: 'Self-Hosted',
    badge: `Free for Up to ${COMMUNITY_SEAT_LIMIT} Seats`,
    text: 'The whole platform on your own Kubernetes cluster, community edition free; a license key that verifies offline for larger teams.',
    href: '/pricing',
    cta: 'Licenses',
  },
  {
    name: POSITIONING.desktop.name,
    badge: 'Free for Individuals',
    text: POSITIONING.desktop.whatItAdds[6],
    href: DESKTOP_LANDING_PATH,
    cta: 'Planton Desktop',
  },
];

const POINTS = [
  { label: 'Keyless Connections', text: ch.proof[0] },
  { label: 'Open Source, and the Way Out', text: ch.proof[1] },
  { label: 'Secrets on the Desktop', text: ch.proof[2] },
];

export const RunsWhereYouDecide: FC = () => (
  <ChapterSection chapter={ch} readMore={{ href: '/trust/your-cloud-your-keys', label: 'Your Cloud, Your Keys' }}>
    <Grid cols={3} className="max-w-6xl mx-auto mb-8">
      {SHAPES.map((shape) => (
        <Card key={shape.name}>
          <Box className="flex flex-col gap-3 h-full">
            <FeatureTitle>{shape.name}</FeatureTitle>
            <Badge className="self-start">{shape.badge}</Badge>
            <BodyText className="flex-1">{shape.text}</BodyText>
            <Link href={shape.href} className="text-sm text-fg-secondary hover:text-white underline underline-offset-4">
              {`${shape.cta} \u2192`}
            </Link>
          </Box>
        </Card>
      ))}
    </Grid>
    <Box className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-6xl mx-auto">
      {POINTS.map((point) => (
        <Box key={point.label} className="border-l-2 border-edge-hover pl-4">
          <FeatureTitle className="text-sm md:text-base mb-1">{point.label}</FeatureTitle>
          <BodyText>{point.text}</BodyText>
        </Box>
      ))}
    </Box>
  </ChapterSection>
);
