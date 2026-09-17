import { Metadata } from 'next';
import { Box } from '@mui/material';
import { InternalDeveloperPlatform } from '@/components/product/solutions/internal-developer-platform';

export const metadata: Metadata = {
  title: 'Internal Developer Platform | Planton',
  description:
    'Build an IDP without building an IDP. Self-service infrastructure, managed CI/CD, access control, and AI assistance — out of the box.',
};

export default function InternalDeveloperPlatformPage() {
  return (
    <Box>
      <InternalDeveloperPlatform />
    </Box>
  );
}
