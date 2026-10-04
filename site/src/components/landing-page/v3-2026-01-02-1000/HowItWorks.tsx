'use client';

import { Box, Stack, Typography } from '@mui/material';
import Image from 'next/image';
import { FC } from 'react';
import { Layers } from 'lucide-react';
import { Section, SectionTitle, SectionSubtitle, Badge } from './shared';
import { PLATFORM_STATS } from '@/data/platform-stats';

const _GitHubIcon = () => (
  <svg className="w-8 h-8 text-white" viewBox="0 0 24 24" fill="currentColor">
    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
  </svg>
);

const MultiCloudIcon = () => (
  <Box className="flex items-center justify-center gap-1">
    <Image src="/_site/images/providers/aws.svg" alt="AWS" width={28} height={28} className="w-7 h-7 md:w-8 md:h-8" />
    <Image src="/_site/images/providers/gcp.svg" alt="GCP" width={28} height={28} className="w-7 h-7 md:w-8 md:h-8" />
    <Image src="/_site/images/providers/azure.svg" alt="Azure" width={28} height={28} className="w-7 h-7 md:w-8 md:h-8" />
  </Box>
);

const IaCIcon = () => (
  <Box className="flex items-center justify-center gap-2">
    <Image src="/_site/images/providers/terraform.svg" alt="Terraform" width={48} height={48} className="w-9 h-9 md:w-10 md:h-10" />
    <Image src="/_site/images/providers/pulumi.svg" alt="Pulumi" width={48} height={48} className="w-10 h-10 md:w-12 md:h-12" />
  </Box>
);

const GitProvidersIcon = () => (
  <Box className="flex items-center justify-center gap-3">
    <Image src="/_site/images/providers/github-dark.svg" alt="GitHub" width={48} height={48} className="w-9 h-9 md:w-11 md:h-11" />
    <Image src="/_site/images/providers/gitlab.svg" alt="GitLab" width={48} height={48} className="w-10 h-10 md:w-12 md:h-12" />
  </Box>
);

const infraSteps = [
  {
    title: 'Connect Cloud',
    description: 'AWS, GCP, or Azure via OAuth',
    iconType: 'multiCloud' as const,
  },
  {
    title: 'Choose Infra',
    description: `${PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} pre-built kinds`,
    iconType: 'lucide' as const,
    icon: Layers,
  },
  {
    title: 'Deploy',
    description: 'Live Terraform / Pulumi visualization',
    iconType: 'iac' as const,
  },
];

const serviceSteps = [
  {
    title: 'Connect Git',
    description: 'GitHub or GitLab',
    iconType: 'gitProviders' as const,
  },
  {
    title: 'Deploy Services',
    description: 'Git push → production with Tekton',
    iconType: 'image' as const,
    iconSrc: '/_site/images/tekton.svg',
  },
];

export const HowItWorks: FC = () => {
  return (
    <Section id="how-it-works">
      <Stack className="items-center text-center mb-12">
        <Badge className="mb-6">HOW IT WORKS</Badge>
        <SectionTitle>
          From Zero to Production in Minutes
        </SectionTitle>
        <SectionSubtitle className="mx-auto">
          Self-service by design. No weeks of configuration. Deploy infrastructure AND services.
        </SectionSubtitle>
      </Stack>

      <Box className="mb-12">
        <Typography className="text-sm font-medium text-[#a0a0a0] text-center mb-6">
          Infrastructure
        </Typography>
        <Box className="flex flex-row items-start justify-center gap-2 md:gap-4 lg:gap-6">
          {infraSteps.map((step, index) => (
            <Box key={index} className="contents">
              <Box className="flex flex-col items-center text-center flex-1 max-w-[200px]">
                <Box className="w-14 h-14 md:w-16 md:h-16 flex items-center justify-center mb-4">
                  {step.iconType === 'multiCloud' ? (
                    <MultiCloudIcon />
                  ) : step.iconType === 'iac' ? (
                    <IaCIcon />
                  ) : (
                    <step.icon className="w-10 h-10 md:w-12 md:h-12 text-white" />
                  )}
                </Box>
                
                <Typography className="text-lg font-semibold text-white mb-2">
                  {step.title}
                </Typography>
                
                <Typography className="text-sm text-[#a0a0a0] max-w-xs leading-relaxed">
                  {step.description}
                </Typography>
              </Box>
              
              {index < infraSteps.length - 1 && (
                <Box className="hidden sm:flex items-center justify-center pt-6 text-white/30">
                  <span className="text-xl">→</span>
                </Box>
              )}
            </Box>
          ))}
        </Box>
      </Box>

      <Box className="mb-12">
        <Typography className="text-sm font-medium text-[#a0a0a0] text-center mb-6">
          Services
        </Typography>
        <Box className="flex flex-row items-start justify-center gap-4 md:gap-8 lg:gap-12 max-w-2xl mx-auto">
          {serviceSteps.map((step, index) => (
            <Box key={index} className="contents">
              <Box className="flex flex-col items-center text-center flex-1 max-w-[240px]">
                <Box className="w-14 h-14 md:w-16 md:h-16 flex items-center justify-center mb-4">
                  {step.iconType === 'gitProviders' ? (
                    <GitProvidersIcon />
                  ) : (
                    <Image 
                      src={step.iconSrc!} 
                      alt={step.title} 
                      width={40} 
                      height={40} 
                      className="w-10 h-10 md:w-12 md:h-12"
                    />
                  )}
                </Box>
                
                <Typography className="text-lg font-semibold text-white mb-2">
                  {step.title}
                </Typography>
                
                <Typography className="text-sm text-[#a0a0a0] max-w-xs leading-relaxed">
                  {step.description}
                </Typography>
              </Box>
              
              {index < serviceSteps.length - 1 && (
                <Box className="hidden sm:flex items-center justify-center pt-6 text-white/30">
                  <span className="text-xl">→</span>
                </Box>
              )}
            </Box>
          ))}
        </Box>
      </Box>

      <Box className="bg-[#151515] border border-[#2a2a2a] rounded-xl p-6 max-w-xl mx-auto text-center">
        <Typography className="text-lg font-medium text-white">
          ⚡ Infrastructure + Services in &lt;1 Hour
        </Typography>
      </Box>
    </Section>
  );
};
