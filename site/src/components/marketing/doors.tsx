/**
 * A page's two doors, rendered from the door vocabulary in src/data/doors.ts.
 * The pair is named by key, so the same door reads the same way everywhere
 * it appears and a destination changes in one place; a section that wants a
 * call to action renders this and never writes its own button. The user's
 * pair (start free, download the desktop app) is the default; Trust pages
 * pass the buyer's pair.
 */
import { Stack } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { ArrowRightIcon } from './icons';
import { PrimaryButton, SecondaryButton } from './buttons';
import { DOORS, START_DOORS, type DoorId } from '@/data/doors';

export interface DoorsProps {
  primary?: DoorId;
  secondary?: DoorId;
  className?: string;
}

export const Doors: FC<DoorsProps> = ({ primary = START_DOORS.primary, secondary = START_DOORS.secondary, className = '' }) => {
  const lead = DOORS[primary];
  const beside = DOORS[secondary];
  return (
    <Stack direction={{ xs: 'column', sm: 'row' }} className={`gap-3 items-center ${className}`}>
      <Link href={lead.href}>
        <PrimaryButton className="text-sm px-8 py-3">
          {lead.label}
          <ArrowRightIcon />
        </PrimaryButton>
      </Link>
      <Link href={beside.href}>
        <SecondaryButton className="text-sm px-8 py-3">{beside.label}</SecondaryButton>
      </Link>
    </Stack>
  );
};
