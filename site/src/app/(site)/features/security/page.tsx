import { Metadata } from 'next';
import { Box } from '@mui/material';
import { SecurityHero, SecurityCapabilities, SecurityCTA } from '@/components/product/security';

export const metadata: Metadata = {
  title: 'Security | Planton',
  description:
    'Built-in secrets management, identity and access control, full audit trails, and zero-trust architecture. Security is native to every layer of Planton.',
};

export default function SecurityPage() {
  return (
    <Box>
      <SecurityHero />
      <SecurityCapabilities />
      <SecurityCTA />
    </Box>
  );
}
