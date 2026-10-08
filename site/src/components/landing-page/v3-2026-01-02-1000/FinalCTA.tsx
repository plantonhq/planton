'use client';

import { Box, Stack, Typography } from '@mui/material';
import Link from 'next/link';
import { FC } from 'react';
import { Section, SectionTitle, PrimaryButton, SecondaryButton, ArrowRightIcon } from './shared';
import { SITE } from '@/data/site-pages';

const metrics = [
  { value: '450+', label: 'Infrastructure Deployments' },
  { value: '<1 hr', label: 'Average Setup Time' },
  { value: '100%', label: 'Customer Retention' },
];

export const FinalCTA: FC = () => {
  return (
    <Section id="cta">
      <Box>
        <Box className="grid grid-cols-3 gap-4 max-w-3xl mx-auto mb-12">
          {metrics.map((metric, index) => (
            <Box key={index} className="text-center">
              <Typography className="text-2xl md:text-3xl font-bold text-white">
                {metric.value}
              </Typography>
              <Typography className="text-xs text-[#a0a0a0] mt-1">{metric.label}</Typography>
            </Box>
          ))}
        </Box>
        
        <Stack className="items-center text-center mb-12">
          <SectionTitle>
            Start Deploying Today
          </SectionTitle>
          <Typography className="text-lg text-[#a0a0a0] max-w-2xl mt-4 mb-8">
            Free tier. No credit card required. Deploy your first infrastructure in under an hour.
          </Typography>
          
          <Stack direction={{ xs: 'column', sm: 'row' }} className="gap-4">
            <Link href={`${SITE.url}/login`} target="_blank">
              <PrimaryButton className="px-8 py-4 text-lg">
                Start Free
                <ArrowRightIcon />
              </PrimaryButton>
            </Link>
            <Link href="/book-demo">
              <SecondaryButton className="px-8 py-4">
                Book a Demo
              </SecondaryButton>
            </Link>
          </Stack>
        </Stack>
      </Box>
    </Section>
  );
};
