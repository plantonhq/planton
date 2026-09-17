import { Metadata } from 'next';
import { Box } from '@mui/material';
import { ServiceHubHero, ServiceHubCapabilities, ServiceHubCTA } from '@/components/product/service-hub';

export const metadata: Metadata = {
  title: 'Service Hub | Planton',
  description:
    'Ship code from Git to production. Managed CI/CD with Tekton pipelines, multi-environment promotion, deploy to Kubernetes, ECS, or Cloud Run, and Kustomize-native config — all from one workflow.',
};

export default function ServiceHubPage() {
  return (
    <Box>
      <ServiceHubHero />
      <ServiceHubCapabilities />
      <ServiceHubCTA />
    </Box>
  );
}
