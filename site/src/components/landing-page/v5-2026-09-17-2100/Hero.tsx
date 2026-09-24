import { Box, Stack, Typography } from '@mui/material';
import Image from 'next/image';
import type { FC } from 'react';
import { POSITIONING } from '@/data/positioning';
import { HERO_HEADLINE, HERO_SUBHEAD } from '@/data/story';
import { FREE_TIER_SEATS } from '@/data/pricing';
import { PLATFORM_STATS } from '@/data/platform-stats';
import { Doors } from '@/components/marketing';
import { asset } from '@/lib/assets';
import { PersonaRouter } from './PersonaRouter';
import { ProofMoment } from './ProofMoment';

/**
 * The first screen answers four things without scrolling: what this is (the
 * umbrella name as the eyebrow, never a hub analogy), why a visitor with an
 * agent and a CLI would add it (chapter 1's line, the story's first sentence,
 * as the headline), where it sits beside what they have (one sentence), and
 * what to do next (two doors, the same two the whole page uses). The proof
 * moment below is a still frame of the product doing its defining move.
 *
 * Never here: a hub's analogy, a price as a literal, a savings figure,
 * motion that depends on a clock, a third door.
 */

const PROVIDERS = [
  { src: asset('images/providers/aws.svg'), alt: 'AWS' },
  { src: asset('images/providers/gcp.svg'), alt: 'Google Cloud' },
  { src: asset('images/providers/azure.svg'), alt: 'Azure' },
  { src: asset('images/providers/digital-ocean.svg'), alt: 'DigitalOcean' },
  { src: asset('images/providers/kubernetes.svg'), alt: 'Kubernetes' },
  { src: asset('images/providers/cloudflare.svg'), alt: 'Cloudflare' },
];

export const Hero: FC = () => (
  <Box component="section" className="relative overflow-hidden bg-canvas">
    <Box className="relative w-full max-w-7xl mx-auto px-4 md:px-8 pt-16 md:pt-24 pb-12 md:pb-16">
      <Stack className="items-center text-center gap-6">
        <Typography className="text-xs md:text-sm font-medium tracking-wide text-fg-muted">
          Planton &middot; {POSITIONING.umbrella.tagline}
        </Typography>

        <Typography
          variant="h1"
          className="text-3xl sm:text-4xl md:text-5xl font-semibold text-white leading-[1.15] tracking-tight max-w-4xl"
        >
          {HERO_HEADLINE}
        </Typography>

        <Typography className="text-base md:text-lg text-fg-secondary max-w-2xl leading-relaxed text-balance">{HERO_SUBHEAD}</Typography>

        <Stack className="items-center gap-3 mt-2">
          <Doors />
          <Typography className="text-sm text-fg-secondary max-w-xl text-balance">
            Hosted is free for up to {FREE_TIER_SEATS} seats, no card. Desktop is free for individuals, commercial use included.
          </Typography>
          <PersonaRouter />
        </Stack>

        <Box className="relative w-full max-w-3xl lg:max-w-4xl mt-6">
          <ProofMoment />
        </Box>

        <Box className="w-full max-w-3xl">
          <Typography className="text-xs text-fg-muted mb-4">Deploys to {PLATFORM_STATS.CLOUD_PROVIDER_COUNT} providers, among them</Typography>
          <Box className="flex flex-wrap items-center justify-center gap-6 md:gap-8">
            {PROVIDERS.map((provider) => (
              <Box key={provider.alt} className="opacity-50 hover:opacity-100 transition-opacity" sx={{ height: { xs: 24, sm: 32 }, width: 'auto' }}>
                <Image src={provider.src} alt={provider.alt} width={80} height={32} style={{ height: '100%', width: 'auto', objectFit: 'contain' }} />
              </Box>
            ))}
          </Box>
        </Box>
      </Stack>
    </Box>
  </Box>
);
