'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Stack, Typography } from '@mui/material';

export const EnterpriseRedirect = () => {
  const router = useRouter();
  useEffect(() => {
    router.replace('/pricing/enterprise');
  }, [router]);
  return (
    <Stack className="items-center py-24 px-4 text-center gap-2 bg-[#0a0a0a]">
      <Typography className="text-sm text-[#a0a0a0]">
        Enterprise pricing has moved.
      </Typography>
      <Link href="/pricing/enterprise" className="text-sm text-white underline">
        Continue to Enterprise Pricing
      </Link>
    </Stack>
  );
};
