import { pageMetadata } from '@/lib/page-metadata';
import { Box } from '@mui/material';
import { CliHero, CliCapabilities, CliCTA } from '@/components/product/cli';

export const metadata = pageMetadata('/features/cli');

export default function CliPage() {
  return (
    <Box>
      <CliHero />
      <CliCapabilities />
      <CliCTA />
    </Box>
  );
}
