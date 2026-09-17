'use client';

import { FC } from 'react';
import Link from 'next/link';
import { Box, Typography } from '@mui/material';
import { CheckIcon } from '@/components/marketing';

/**
 * One quiet line directly under the plan grid: self-hosted runners are
 * included on every plan, and a runner inside a private network reaches
 * the platform outbound-only. This fact lives HERE, visibly, because the
 * value matrix renders row descriptions as hover tooltips — a scanning
 * buyer would never see it there. Deliberately a single line, not a
 * second band: the enterprise band directly below keeps its weight.
 */
export const RunnersNote: FC = () => {
  return (
    <Box className="w-full px-4 md:px-8 pb-2 bg-[#0a0a0a]">
      <Box className="max-w-7xl mx-auto flex items-start justify-center gap-2 text-center">
        <Box className="mt-0.5 flex-shrink-0">
          <CheckIcon />
        </Box>
        <Typography className="text-xs md:text-sm text-[#8a8a8a] max-w-[820px]">
          Self-hosted runners are included in every plan — they run inside
          your private network and connect outbound-only, so you never open
          an inbound firewall port.{' '}
          <Link
            href="/features/runner"
            className="text-[#c0c0c0] underline underline-offset-4 hover:text-white transition-colors"
          >
            Learn About Runner
          </Link>
        </Typography>
      </Box>
    </Box>
  );
};
