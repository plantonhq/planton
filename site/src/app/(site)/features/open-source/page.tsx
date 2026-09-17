import { pageMetadata } from '@/lib/page-metadata';
import { Box } from '@mui/material';
import { OpenSourceHero, OpenSourceCapabilities, OpenSourceCTA } from '@/components/product/open-source';

export const metadata = pageMetadata('/features/open-source');

export default function OpenSourcePage() {
  return (
    <Box>
      <OpenSourceHero />
      <OpenSourceCapabilities />
      <OpenSourceCTA />
    </Box>
  );
}
