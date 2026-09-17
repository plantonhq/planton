'use client';

import { Box, Typography } from '@mui/material';
import {
  Shield as ShieldIcon,
  Code as CodeIcon,
  CloudQueue as CloudIcon,
  Speed as SpeedIcon,
} from '@mui/icons-material';
import { Section } from '@/components/marketing';
import { ReactNode } from 'react';

const TrustItem = ({ icon, title, description }: { icon: ReactNode; title: string; description: string }) => (
  <Box className="flex flex-col items-center text-center p-4">
    <Box className="w-12 h-12 rounded-full bg-white/5 border border-[#2a2a2a] flex items-center justify-center mb-3 text-[#a0a0a0]">
      {icon}
    </Box>
    <Typography className="text-sm font-semibold text-white mb-1">{title}</Typography>
    <Typography className="text-xs text-[#666] leading-relaxed max-w-[200px]">{description}</Typography>
  </Box>
);

export const ProductTrustBar = () => {
  return (
    <Section>
      <Box className="grid grid-cols-2 md:grid-cols-4 gap-4 md:gap-8">
        <TrustItem
          icon={<SpeedIcon fontSize="small" />}
          title="Minutes, Not Weeks"
          description="From zero to deployed infrastructure in minutes. No ops team required."
        />
        <TrustItem
          icon={<CloudIcon fontSize="small" />}
          title="Your Cloud, Your Control"
          description="SaaS orchestration with execution in your cloud account. You own your resources."
        />
        <TrustItem
          icon={<CodeIcon fontSize="small" />}
          title="Open Source Foundation"
          description="Planton open source powers every deployment. Export manifests, run anywhere. Zero lock-in."
        />
        <TrustItem
          icon={<ShieldIcon fontSize="small" />}
          title="Enterprise Security"
          description="Encrypted tunnels, secrets resolved at execution time, full audit trails for every change."
        />
      </Box>
    </Section>
  );
};
