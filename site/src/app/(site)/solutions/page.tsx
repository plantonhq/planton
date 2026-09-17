import { Metadata } from 'next';
import { Box } from '@mui/material';
import { SolutionsHub } from '@/components/product/solutions/hub';

export const metadata: Metadata = {
  title: 'Solutions | Planton',
  description:
    'Explore how Planton solves infrastructure challenges for teams of every size and role.',
};

export default function SolutionsPage() {
  return (
    <Box>
      <SolutionsHub />
    </Box>
  );
}
