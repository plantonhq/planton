import { pageMetadata } from '@/lib/page-metadata';
import { Box } from '@mui/material';
import { MultiCloud } from '@/components/product/solutions/multi-cloud';

export const metadata = pageMetadata('/solutions/by-use-case/multi-cloud');

export default function MultiCloudPage() {
  return (
    <Box>
      <MultiCloud />
    </Box>
  );
}
