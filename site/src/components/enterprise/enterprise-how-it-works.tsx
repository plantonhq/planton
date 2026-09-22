'use client';

import { FC } from 'react';
import { Box, Grid2, Stack } from '@mui/material';
import {
  Badge,
  BodyText,
  Card,
  FeatureTitle,
  SectionTitle,
} from '@/components/marketing';
import { useMarket } from '@/components/market';
import { EVALUATION_DAYS } from '@/data/pricing';

/**
 * How an enterprise purchase actually works. The land move is deliberate:
 * a team proves value self-serve before procurement ever hears the word
 * "vendor" -- the buying steps word the real issuance machinery. The
 * self-serve price range reads from the pricing-truth module in the
 * visitor's market -- prose never re-authors a number.
 */

export const EnterpriseHowItWorks: FC = () => {
  const { market } = useMarket();
  const sizeLow = market.licenses[0].perYearCompact;
  const sizeHigh = market.licenses[market.licenses.length - 1].perYearCompact;

  const steps = [
    {
      title: 'Prove It First',
      body: `Run the community edition free, or unlock everything for ${EVALUATION_DAYS} days with a self-serve evaluation key on your own cluster. Many teams start with a ${sizeLow}–${sizeHigh} self-serve license — no procurement required.`,
    },
    {
      title: 'Quote and Invoice',
      body: 'When you outgrow self-serve, we issue a quote and invoice with net-30/60 bank-transfer terms — or contract on your paper with Enterprise Plus.',
    },
    {
      title: 'Your Key Arrives',
      body: 'The license key is emailed the moment payment lands — same machinery as the self-serve licenses, bigger bounds. It verifies offline, works air-gapped, and never breaks a running deployment.',
    },
  ];

  return (
    <Stack className="items-center px-4 md:px-8 py-10 gap-6 bg-[#0a0a0a]">
      <SectionTitle>How Buying Works</SectionTitle>
      <Grid2 container spacing={4} className="w-full max-w-5xl items-stretch">
        {steps.map((step, index) => (
          <Grid2 size={{ xs: 12, md: 4 }} key={step.title} className="flex">
            <Card className="w-full">
              <Stack className="gap-3">
                <Box>
                  <Badge className="!px-2 !py-0.5 !text-[11px]">Step {index + 1}</Badge>
                </Box>
                <FeatureTitle>{step.title}</FeatureTitle>
                <BodyText>{step.body}</BodyText>
              </Stack>
            </Card>
          </Grid2>
        ))}
      </Grid2>
    </Stack>
  );
};
