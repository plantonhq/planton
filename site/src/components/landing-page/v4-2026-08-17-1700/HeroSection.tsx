'use client';

import { Box, Stack, Typography } from '@mui/material';
import Link from 'next/link';
import Image from 'next/image';
import { FC, useEffect, useState } from 'react';
import { PrimaryButton, ArrowRightIcon } from './shared';
import { FREE_TIER_SEATS } from '@/data/pricing';
import { POSITIONING } from '@/data/positioning';
import { SITE } from '@/data/site-pages';

/**
 * The hero states the umbrella positioning (never a hub analogy — that is
 * the three-level rule in src/data/positioning.ts) and shows the product
 * doing its defining move: one sentence in, an architecture with a cost
 * and a policy out. The mockup is CSS-built on purpose — when real desktop
 * recordings exist they replace AnimatedCompose in this same slot.
 *
 * The desktop app is described in copy but carries NO download CTA
 * (decided 2026-08-17): there is no download page yet and the desktop
 * release is on hold, and a button may never point at either.
 */

const cloudProviders = [
  { src: '/_site/images/providers/aws.svg', alt: 'AWS' },
  { src: '/_site/images/providers/gcp.svg', alt: 'GCP' },
  { src: '/_site/images/providers/azure.svg', alt: 'Azure' },
  { src: '/_site/images/providers/digital-ocean.svg', alt: 'DigitalOcean' },
  { src: '/_site/images/providers/kubernetes.svg', alt: 'Kubernetes' },
  { src: '/_site/images/providers/cloudflare.svg', alt: 'Cloudflare' },
];

const PROMPT_TEXT =
  'A production environment on AWS: private VPC, ECS services behind a load balancer, RDS PostgreSQL encrypted with our own KMS key.';

interface ComposedComponent {
  name: string;
  cost: string;
}

/** Example figures, labeled "est." — real numbers come from the verified catalog. */
const composedComponents: ComposedComponent[] = [
  { name: 'AWS VPC', cost: 'No charge' },
  { name: 'NAT Gateway', cost: '~$33/mo est.' },
  { name: 'Application Load Balancer', cost: '~$16/mo est.' },
  { name: 'ECS Service', cost: 'Usage-based' },
  { name: 'RDS PostgreSQL (Multi-AZ)', cost: '~$122/mo est.' },
  { name: 'KMS Key', cost: '$1/mo est.' },
];

const AnimatedCompose: FC = () => {
  const [typedChars, setTypedChars] = useState(0);
  const [visibleComponents, setVisibleComponents] = useState(0);
  const [showFooter, setShowFooter] = useState(false);
  // Incremented once per completed run; the effect keys on it so each
  // cycle schedules exactly one timeline and cleans it up on unmount.
  const [cycle, setCycle] = useState(0);

  useEffect(() => {
    const timeouts: NodeJS.Timeout[] = [];

    // Type the prompt, then reveal components one by one, then the verdict row.
    for (let i = 1; i <= PROMPT_TEXT.length; i++) {
      timeouts.push(setTimeout(() => setTypedChars(i), 300 + i * 18));
    }
    const typingDone = 300 + PROMPT_TEXT.length * 18;
    composedComponents.forEach((_, i) => {
      timeouts.push(
        setTimeout(() => setVisibleComponents(i + 1), typingDone + 500 + i * 420)
      );
    });
    timeouts.push(
      setTimeout(
        () => setShowFooter(true),
        typingDone + 500 + composedComponents.length * 420 + 400
      )
    );
    // Hold the finished frame, then reset and start the next cycle.
    timeouts.push(
      setTimeout(() => {
        setTypedChars(0);
        setVisibleComponents(0);
        setShowFooter(false);
        setCycle((c) => c + 1);
      }, typingDone + 500 + composedComponents.length * 420 + 7000)
    );

    return () => timeouts.forEach(clearTimeout);
  }, [cycle]);

  return (
    <Box className="rounded-lg overflow-hidden border border-[#2a2a2a] bg-black/60 backdrop-blur-sm text-left w-full">
      <Box className="flex items-center gap-3 px-4 py-2.5 bg-[#111]/80 border-b border-[#2a2a2a]">
        <Box className="flex gap-1.5">
          <Box className="w-3 h-3 rounded-full bg-[#ef4444]/80" />
          <Box className="w-3 h-3 rounded-full bg-[#f59e0b]/80" />
          <Box className="w-3 h-3 rounded-full bg-[#10b981]/80" />
        </Box>
        <Typography className="text-sm text-[#666]">
          Planton Desktop — Planton Studio
        </Typography>
      </Box>

      <Box className="p-4 min-h-[300px]">
        <Box className="rounded-lg border border-[#2a2a2a] bg-[#111] px-4 py-3 mb-4">
          <Typography className="text-xs text-[#555] mb-1">
            What do you want to build?
          </Typography>
          <Typography className="text-sm text-gray-100 font-mono">
            {PROMPT_TEXT.slice(0, typedChars)}
            <span className="animate-pulse text-white">▊</span>
          </Typography>
        </Box>

        <Box className="grid grid-cols-2 sm:grid-cols-3 gap-2 mb-4">
          {composedComponents.slice(0, visibleComponents).map((component) => (
            <Box
              key={component.name}
              className="rounded-lg border border-[#2a2a2a] bg-[#151515] px-3 py-2"
            >
              <Typography className="text-xs text-white font-medium leading-tight">
                {component.name}
              </Typography>
              <Typography className="text-[11px] text-[#666] mt-0.5">
                {component.cost}
              </Typography>
            </Box>
          ))}
        </Box>

        {showFooter && (
          <Box className="flex flex-wrap items-center gap-2">
            <Box className="rounded-full border border-[#2a2a2a] bg-[#151515] px-3 py-1.5">
              <Typography className="text-xs text-white">
                Estimated Monthly Bill: <strong>~$172 est.</strong>
              </Typography>
            </Box>
            <Box className="rounded-full border border-[#2a2a2a] bg-[#151515] px-3 py-1.5">
              <Typography className="text-xs text-white">
                Least-Privilege IAM Policy: Ready
              </Typography>
            </Box>
            <Box className="rounded-full border border-[#10b981]/30 bg-[#10b981]/10 px-3 py-1.5">
              <Typography className="text-xs text-[#10b981] font-medium">
                Deploy · Publish as Template
              </Typography>
            </Box>
          </Box>
        )}
      </Box>
    </Box>
  );
};

export const HeroSection: FC = () => {
  return (
    <Box className="relative min-h-[80vh] flex items-center overflow-hidden bg-[#0a0a0a]">
      <Box className="absolute inset-0 overflow-hidden">
        <Box
          className="absolute inset-0 opacity-[0.03]"
          sx={{
            backgroundImage: `
              linear-gradient(rgba(255,255,255,0.1) 1px, transparent 1px),
              linear-gradient(90deg, rgba(255,255,255,0.1) 1px, transparent 1px)
            `,
            backgroundSize: '60px 60px',
          }}
        />
      </Box>

      <Box className="relative z-10 w-full max-w-7xl mx-auto px-4 md:px-8 py-16 md:py-24">
        <Stack className="items-center text-center gap-6">
          <Typography
            variant="h1"
            className="text-3xl sm:text-4xl md:text-5xl font-semibold text-white leading-[1.15] tracking-tight max-w-4xl"
          >
            {POSITIONING.umbrella.tagline}
          </Typography>

          <Typography className="text-base md:text-lg text-[#a0a0a0] max-w-2xl leading-relaxed">
            {POSITIONING.umbrella.sentence}
          </Typography>

          <Stack className="items-center gap-3 mt-2">
            <Stack direction={{ xs: 'column', sm: 'row' }} className="gap-3 items-center">
              <Link href={`${SITE.url}/login`}>
                <PrimaryButton className="text-sm px-8 py-3">
                  Start Free
                  <ArrowRightIcon />
                </PrimaryButton>
              </Link>
              <Link
                href="/book-demo"
                className="text-[#a0a0a0] hover:text-white transition-colors text-sm font-medium"
              >
                Book a Demo &rarr;
              </Link>
            </Stack>

            <Typography className="text-xs text-[#555]">
              Free for up to {FREE_TIER_SEATS} people &middot; No credit card
              required &middot; Also a free desktop app that deploys with your
              own cloud logins
            </Typography>
          </Stack>

          <Box className="relative w-full max-w-3xl mt-6">
            <AnimatedCompose />
          </Box>

          <Box className="w-full max-w-3xl">
            <Box className="flex flex-wrap items-center justify-center gap-6 md:gap-8">
              {cloudProviders.map((provider, index) => (
                <Box
                  key={index}
                  className="transition-opacity hover:opacity-100 opacity-50"
                  sx={{ height: { xs: 24, sm: 32 }, width: 'auto' }}
                >
                  <Image
                    src={provider.src}
                    alt={provider.alt}
                    width={80}
                    height={32}
                    style={{ height: '100%', width: 'auto', objectFit: 'contain' }}
                  />
                </Box>
              ))}
            </Box>
          </Box>
        </Stack>
      </Box>
    </Box>
  );
};
