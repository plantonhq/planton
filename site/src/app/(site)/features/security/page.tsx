import { pageMetadata } from '@/lib/page-metadata';
import { Box } from '@mui/material';
import { SecurityHero, SecurityCapabilities, SecurityCTA } from '@/components/product/security';

export const metadata = pageMetadata('/features/security');

export default function SecurityPage() {
  return (
    <Box>
      <SecurityHero />
      <SecurityCapabilities />
      <SecurityCTA />
    </Box>
  );
}
