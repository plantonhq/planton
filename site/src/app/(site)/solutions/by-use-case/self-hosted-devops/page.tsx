import { Metadata } from 'next';
import { Box } from '@mui/material';
import { SelfHostedDevOps } from '@/components/product/solutions/self-hosted-devops';

export const metadata: Metadata = {
  title: 'Self-Hosted DevOps | Planton',
  description:
    'Enterprise security with SaaS convenience. Your credentials never leave your cloud boundary.',
};

export default function SelfHostedDevOpsPage() {
  return (
    <Box>
      <SelfHostedDevOps />
    </Box>
  );
}
