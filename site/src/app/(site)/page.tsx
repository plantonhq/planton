'use client';

import { Box } from '@mui/material';
import {
  HeroSection,
  HowItWorks,
  TwoHubs,
  VerifiedBeforeDeploy,
  YourCloudYourControl,
  OpenSource,
  ThreeWaysToRun,
  Proof,
  FinalCTA,
} from '@/components/landing-page';

export default function Home() {
  return (
    <Box className="overflow-x-hidden">
      <HeroSection />
      <HowItWorks />
      <TwoHubs />
      <VerifiedBeforeDeploy />
      <YourCloudYourControl />
      <OpenSource />
      <ThreeWaysToRun />
      <Proof />
      <FinalCTA />
    </Box>
  );
}
