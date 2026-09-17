'use client';

import { Box, Typography } from '@mui/material';
import { Badge } from '@/components/marketing';

export const ProductOverviewHero = () => {
  return (
    <Box className="w-full pt-20 md:pt-24 pb-8 md:pb-10 px-4 md:px-8 bg-[#0a0a0a]">
      <Box className="max-w-7xl mx-auto text-center">
        <Badge className="mb-4">The Planton Platform</Badge>
        <Typography
          variant="h1"
          className="text-2xl md:text-3xl font-semibold text-white tracking-tight mb-2"
        >
          Everything you need to deploy and operate infrastructure
        </Typography>
        <Typography className="text-sm md:text-base text-[#666] max-w-2xl mx-auto">
          Seven modules that work together to replace your patchwork of DevOps tools.
        </Typography>
      </Box>
    </Box>
  );
};
