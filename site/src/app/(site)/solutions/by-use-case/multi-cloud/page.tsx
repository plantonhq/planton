import { Metadata } from 'next';
import { Box } from '@mui/material';
import { MultiCloud } from '@/components/product/solutions/multi-cloud';

export const metadata: Metadata = {
  title: 'Multi-Cloud | Planton',
  description:
    'Same workflow, every cloud. One YAML manifest format, one CLI, one console — for AWS, GCP, Azure, and beyond.',
};

export default function MultiCloudPage() {
  return (
    <Box>
      <MultiCloud />
    </Box>
  );
}
