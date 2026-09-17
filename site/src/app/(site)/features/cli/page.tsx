import { Metadata } from 'next';
import { Box } from '@mui/material';
import { CliHero, CliCapabilities, CliCTA } from '@/components/product/cli';

export const metadata: Metadata = {
  title: 'CLI | Planton',
  description:
    'Everything Planton does, from your terminal. Manifest-driven deployments, real-time stack job streaming, Kubernetes access, and environment config — one CLI.',
};

export default function CliPage() {
  return (
    <Box>
      <CliHero />
      <CliCapabilities />
      <CliCTA />
    </Box>
  );
}
