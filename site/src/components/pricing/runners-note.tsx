'use client';

import { FC } from 'react';
import Link from 'next/link';
import { Box, Typography } from '@mui/material';
import { CheckIcon } from '@/components/marketing';

/**
 * One quiet line directly under the plan grid: where deploys run on each
 * kind of plan, and that runner minutes are never billed. This fact lives
 * HERE, visibly, because the value matrix renders row descriptions as hover
 * tooltips — a scanning buyer would never see it there. Deliberately a
 * single line, not a second band.
 */
export const RunnersNote: FC = () => {
  return (
    <Box className="w-full px-4 md:px-8 pb-2 bg-[#f6f6f3]">
      <Box className="max-w-7xl mx-auto flex items-start justify-center gap-2 text-center">
        <Box className="mt-0.5 flex-shrink-0">
          <CheckIcon />
        </Box>
        <Typography className="text-xs md:text-sm text-[#616161] max-w-[820px]">
          Deploys run on Planton’s runners on every hosted plan, and runner
          minutes are never billed. A runner of your own, inside your
          network, is included on self-hosted and coming soon on hosted.{' '}
          <Link
            href="/docs/runner/planton-hosted-runners"
            className="text-[#454545] underline underline-offset-4 hover:text-[#171717] transition-colors"
          >
            Learn About Runner
          </Link>
        </Typography>
      </Box>
    </Box>
  );
};
