/**
 * The proof strip's numbers: four counts from the platform statistics and
 * the one line that says where and when they were counted. The landing's
 * proof chapter and every persona page show the same strip, so a number
 * appears on the site exactly as `@/data/platform-stats` states it and its
 * provenance is never left off. This component carries no figure of its own.
 */
import { Box } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { PLATFORM_COUNTS, PLATFORM_STATS } from '@/data/platform-stats';
import { Metric } from './metric';
import { BodyText } from './typography';

const COUNTS = [
  { value: PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT, label: 'Component Kinds' },
  { value: PLATFORM_STATS.CLOUD_PROVIDER_COUNT, label: 'Providers' },
  { value: PLATFORM_STATS.CONTROL_COUNT, label: 'Controls with Evidence' },
  { value: `Since ${PLATFORM_STATS.IN_PRODUCTION_SINCE}`, label: 'In Production' },
];

export const PlatformCounts: FC<{ className?: string }> = ({ className = '' }) => (
  <Box className={className}>
    <Box className="grid grid-cols-2 md:grid-cols-4 gap-6 max-w-3xl mx-auto mb-3">
      {COUNTS.map((count) => (
        <Metric key={count.label} value={count.value} label={count.label} />
      ))}
    </Box>
    <BodyText className="text-center text-xs text-fg-muted">
      Catalog figures counted from the open-source repository on {PLATFORM_COUNTS.countedOn} (
      <Link href="https://github.com/plantonhq/planton" className="underline underline-offset-4 hover:text-white">github.com/plantonhq/planton</Link>
      ); {PLATFORM_STATS.IN_PRODUCTION_SINCE} is the year the first customer went to production.
    </BodyText>
  </Box>
);
