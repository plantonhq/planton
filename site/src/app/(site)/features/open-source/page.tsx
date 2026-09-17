import { Metadata } from 'next';
import { Box } from '@mui/material';
import { OpenSourceHero, OpenSourceCapabilities, OpenSourceCTA } from '@/components/product/open-source';

export const metadata: Metadata = {
  title: 'Open Source | Planton',
  description:
    'Planton open source is the open-source foundation powering Planton. Protobuf-defined APIs, Pulumi and Terraform modules, portable KRM YAML manifests — no vendor lock-in.',
};

export default function OpenSourcePage() {
  return (
    <Box>
      <OpenSourceHero />
      <OpenSourceCapabilities />
      <OpenSourceCTA />
    </Box>
  );
}
