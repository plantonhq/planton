import { pageMetadata } from '@/lib/page-metadata';
import { Box } from '@mui/material';
import { InternalDeveloperPlatform } from '@/components/product/solutions/internal-developer-platform';

export const metadata = pageMetadata('/solutions/by-use-case/internal-developer-platform');

export default function InternalDeveloperPlatformPage() {
  return (
    <Box>
      <InternalDeveloperPlatform />
    </Box>
  );
}
