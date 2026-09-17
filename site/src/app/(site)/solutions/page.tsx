import { pageMetadata } from '@/lib/page-metadata';
import { Box } from '@mui/material';
import { SolutionsHub } from '@/components/product/solutions/hub';

export const metadata = pageMetadata('/solutions');

export default function SolutionsPage() {
  return (
    <Box>
      <SolutionsHub />
    </Box>
  );
}
