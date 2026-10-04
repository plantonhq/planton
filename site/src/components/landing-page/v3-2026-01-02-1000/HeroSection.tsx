'use client';

import { Box, Stack, Typography } from '@mui/material';
import Link from 'next/link';
import Image from 'next/image';
import { FC, useState, useEffect, useCallback } from 'react';
import {
  Shield as ShieldIcon,
  Code as CodeIcon,
  CloudQueue as CloudIcon,
  Speed as SpeedIcon,
} from '@mui/icons-material';
import { PrimaryButton, ArrowRightIcon } from './shared';
import { ReactNode } from 'react';
import { FREE_TIER_SEATS } from '@/data/pricing';

const cloudProviders = [
  { src: '/_site/images/providers/aws.svg', alt: 'AWS' },
  { src: '/_site/images/providers/gcp.svg', alt: 'GCP' },
  { src: '/_site/images/providers/azure.svg', alt: 'Azure' },
  { src: '/_site/images/providers/digital-ocean.svg', alt: 'DigitalOcean' },
  { src: '/_site/images/providers/kubernetes.svg', alt: 'Kubernetes' },
  { src: '/_site/images/providers/cloudflare.svg', alt: 'Cloudflare' },
];

interface TerminalLine {
  type: 'command' | 'info' | 'success' | 'final' | 'endpoint' | 'output';
  text: string;
  delay: number;
}

interface Scenario {
  label: string;
  lines: TerminalLine[];
}

const scenarios: Scenario[] = [
  {
    label: 'Infra',
    lines: [
      { type: 'command', text: 'planton chart install api aws-ecs-environment -f values.yaml', delay: 600 },
      { type: 'output', text: '', delay: 400 },
      { type: 'info', text: 'Resolved from platform chart catalog', delay: 300 },
      { type: 'output', text: '', delay: 300 },
      { type: 'success', text: '✓ VPC and networking provisioned (2m 15s)', delay: 500 },
      { type: 'success', text: '✓ Load balancer configured (1m 30s)', delay: 600 },
      { type: 'success', text: '✓ ECS cluster created (3m 45s)', delay: 700 },
      { type: 'success', text: '✓ Container service deployed (1m 18s)', delay: 800 },
      { type: 'success', text: '✓ DNS and TLS certificates ready (42s)', delay: 900 },
      { type: 'output', text: '', delay: 300 },
      { type: 'final', text: '⚡ Environment ready in 9 minutes', delay: 400 },
      { type: 'output', text: '', delay: 300 },
      { type: 'endpoint', text: '→ https://api.acmecorp.io', delay: 400 },
    ],
  },
  {
    label: 'Services',
    lines: [
      { type: 'command', text: 'git push origin main', delay: 600 },
      { type: 'output', text: '', delay: 400 },
      { type: 'info', text: 'Pipeline #247 triggered...', delay: 500 },
      { type: 'output', text: '', delay: 300 },
      { type: 'success', text: '✓ Source cloned (8s)', delay: 500 },
      { type: 'success', text: '✓ Docker image built (1m 42s)', delay: 600 },
      { type: 'success', text: '✓ Pushed to ECR (22s)', delay: 700 },
      { type: 'success', text: '✓ Deployed to dev (45s)', delay: 800 },
      { type: 'success', text: '✓ Health check passed (12s)', delay: 900 },
      { type: 'output', text: '', delay: 300 },
      { type: 'final', text: '⚡ Live in 3 minutes', delay: 400 },
      { type: 'output', text: '', delay: 300 },
      { type: 'endpoint', text: '→ https://api.dev.acmecorp.io', delay: 400 },
    ],
  },
  {
    label: 'CLI',
    lines: [
      { type: 'command', text: 'planton apply -f service.yaml', delay: 600 },
      { type: 'output', text: '', delay: 400 },
      { type: 'info', text: 'Applying Service/acme-api to environment dev...', delay: 500 },
      { type: 'output', text: '', delay: 300 },
      { type: 'success', text: '✓ Manifest validated', delay: 500 },
      { type: 'success', text: '✓ Infra Job created (job-4a7f)', delay: 600 },
      { type: 'success', text: '✓ Container built (1m 18s)', delay: 700 },
      { type: 'success', text: '✓ Deployed to ECS (2m 05s)', delay: 800 },
      { type: 'success', text: '✓ Ingress configured (15s)', delay: 900 },
      { type: 'output', text: '', delay: 300 },
      { type: 'final', text: '⚡ Service live', delay: 400 },
      { type: 'output', text: '', delay: 300 },
      { type: 'endpoint', text: '→ https://api.acmecorp.io', delay: 400 },
    ],
  },
];

const AnimatedTerminal: FC = () => {
  const [visibleLines, setVisibleLines] = useState<number>(0);
  const [scenarioIndex, setScenarioIndex] = useState(0);

  const currentScenario = scenarios[scenarioIndex];

  useEffect(() => {
    const lines = scenarios[scenarioIndex].lines;

    let currentLine = 0;
    let totalDelay = 0;
    const timeoutIds: NodeJS.Timeout[] = [];

    const showNextLine = () => {
      if (currentLine < lines.length) {
        const line = lines[currentLine];
        totalDelay += line.delay;

        const timeoutId = setTimeout(() => {
          setVisibleLines(prev => prev + 1);
        }, totalDelay);
        timeoutIds.push(timeoutId);

        currentLine++;
        showNextLine();
      } else {
        const resetTimeout = setTimeout(() => {
          setVisibleLines(0);
          setScenarioIndex(prev => (prev + 1) % scenarios.length);
        }, 4000 + totalDelay);
        timeoutIds.push(resetTimeout);
      }
    };

    showNextLine();

    return () => {
      timeoutIds.forEach(id => clearTimeout(id));
    };
  }, [scenarioIndex]);

  const goToScenario = useCallback((index: number) => {
    if (index !== scenarioIndex) {
      setVisibleLines(0);
      setScenarioIndex(index);
    }
  }, [scenarioIndex]);

  return (
    <Box className="rounded-lg overflow-hidden border border-[#2a2a2a] bg-black/60 backdrop-blur-sm text-left w-full">
      <Box className="flex items-center justify-between px-4 py-2.5 bg-[#111]/80 border-b border-[#2a2a2a]">
        <Box className="flex items-center gap-3">
          <Box className="flex gap-1.5">
            <Box className="w-3 h-3 rounded-full bg-[#ef4444]/80" />
            <Box className="w-3 h-3 rounded-full bg-[#f59e0b]/80" />
            <Box className="w-3 h-3 rounded-full bg-[#10b981]/80" />
          </Box>
          <Box className="flex items-center gap-1.5 text-sm text-[#666]">
            <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <span>Terminal</span>
          </Box>
        </Box>
        <Box className="flex gap-1">
          {scenarios.map((s, i) => (
            <button
              key={s.label}
              onClick={() => goToScenario(i)}
              className={`px-2.5 py-1 rounded text-xs font-medium transition-all ${
                i === scenarioIndex
                  ? 'bg-white/10 text-white'
                  : 'text-[#555] hover:text-[#999] hover:bg-white/5'
              }`}
            >
              {s.label}
            </button>
          ))}
        </Box>
      </Box>

      <Box className="p-4 font-mono text-xs sm:text-sm h-[280px] overflow-hidden">
        {currentScenario.lines.slice(0, visibleLines).map((line, index) => (
          <Box key={`${scenarioIndex}-${index}`} className="mb-1.5">
            {line.type === 'command' && (
              <Box className="flex items-start gap-2 overflow-hidden">
                <span className="text-[#a0a0a0] select-none flex-shrink-0">$</span>
                <code className="text-gray-100 break-words overflow-hidden">{line.text}</code>
              </Box>
            )}
            {line.type === 'info' && (
              <Typography className="text-gray-400 text-sm">{line.text}</Typography>
            )}
            {line.type === 'success' && (
              <Typography className="text-[#a0a0a0] text-sm">{line.text}</Typography>
            )}
            {line.type === 'final' && (
              <Typography className="text-amber-400 text-sm font-semibold">{line.text}</Typography>
            )}
            {line.type === 'endpoint' && (
              <Typography className="text-white text-sm font-medium">{line.text}</Typography>
            )}
            {line.type === 'output' && (
              <Box className="h-2" />
            )}
          </Box>
        ))}
        {(visibleLines === 0 || visibleLines < currentScenario.lines.length) && (
          <Box className="inline-block animate-pulse">
            <span className="text-white">▊</span>
          </Box>
        )}
      </Box>
    </Box>
  );
};

const TrustItem = ({ icon, title, description }: { icon: ReactNode; title: string; description: string }) => (
  <Box className="flex flex-col items-center text-center">
    <Box className="w-10 h-10 rounded-full bg-white/5 border border-[#2a2a2a] flex items-center justify-center mb-2 text-[#a0a0a0]">
      {icon}
    </Box>
    <Typography className="text-xs font-semibold text-white mb-0.5">{title}</Typography>
    <Typography className="text-[11px] text-[#555] leading-snug max-w-[180px]">{description}</Typography>
  </Box>
);

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
            One platform. Infra and apps.
            <br />
            Any cloud.
          </Typography>

          <Typography className="text-base md:text-lg text-[#a0a0a0] max-w-2xl leading-relaxed">
            Infrastructure deployment, application CI/CD, and AI operations &mdash; all in your cloud.
            Open source foundation. Zero vendor lock-in.
          </Typography>

          <Stack className="items-center gap-3 mt-2">
            <Stack direction={{ xs: 'column', sm: 'row' }} className="gap-3 items-center">
              <Link href="/signup">
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
              Free for up to {FREE_TIER_SEATS} people &middot; No credit card required
            </Typography>
          </Stack>

          <Box className="relative w-full max-w-3xl mt-6">
            <AnimatedTerminal />
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
                    style={{
                      height: '100%',
                      width: 'auto',
                      objectFit: 'contain',
                    }}
                  />
                </Box>
              ))}
            </Box>
          </Box>

          <Box className="w-full max-w-4xl mt-4 pt-6 border-t border-[#2a2a2a]/50">
            <Box className="grid grid-cols-2 md:grid-cols-4 gap-6 md:gap-8">
              <TrustItem
                icon={<SpeedIcon sx={{ fontSize: 18 }} />}
                title="Minutes, Not Weeks"
                description="From zero to deployed infrastructure in minutes."
              />
              <TrustItem
                icon={<CloudIcon sx={{ fontSize: 18 }} />}
                title="Your Cloud, Your Control"
                description="SaaS orchestration with execution in your cloud account."
              />
              <TrustItem
                icon={<CodeIcon sx={{ fontSize: 18 }} />}
                title="Open Source Foundation"
                description="Planton open source powers every deployment. Export and run anywhere."
              />
              <TrustItem
                icon={<ShieldIcon sx={{ fontSize: 18 }} />}
                title="Enterprise Security"
                description="Encrypted tunnels, secrets at execution time, full audit trails."
              />
            </Box>
          </Box>
        </Stack>
      </Box>
    </Box>
  );
};
