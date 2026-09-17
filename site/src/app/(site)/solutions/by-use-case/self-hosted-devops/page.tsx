import { pageMetadata } from '@/lib/page-metadata';
import { Box } from '@mui/material';
import { SelfHostedDevOps } from '@/components/product/solutions/self-hosted-devops';

export const metadata = pageMetadata('/solutions/by-use-case/self-hosted-devops');

export default function SelfHostedDevOpsPage() {
  return (
    <Box>
      <SelfHostedDevOps />
    </Box>
  );
}
