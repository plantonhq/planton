import { Stack } from '@mui/material';
import Link from 'next/link';
import type { FC } from 'react';
import { ArrowRightIcon, PrimaryButton, SecondaryButton } from '@/components/marketing';
import { DESKTOP_DOWNLOAD_PATH } from '@/data/desktop-download';

/**
 * The page's two doors, in the same words and the same order everywhere
 * they appear. The hosted free tier leads because it works on every device
 * a visitor might be holding, including the phone a thread link is often
 * opened on; the free desktop app is the second door and the one the fine
 * print explains. A section that wants a door renders this and never writes
 * its own button.
 */
export const DOOR = {
  hosted: { label: 'Start Free', href: '/signup' },
  desktop: { label: 'Download Planton Desktop', href: DESKTOP_DOWNLOAD_PATH },
} as const;

export const Doors: FC<{ className?: string }> = ({ className = '' }) => (
  <Stack direction={{ xs: 'column', sm: 'row' }} className={`gap-3 items-center ${className}`}>
    <Link href={DOOR.hosted.href}>
      <PrimaryButton className="text-sm px-8 py-3">
        {DOOR.hosted.label}
        <ArrowRightIcon />
      </PrimaryButton>
    </Link>
    <Link href={DOOR.desktop.href}>
      <SecondaryButton className="text-sm px-8 py-3">{DOOR.desktop.label}</SecondaryButton>
    </Link>
  </Stack>
);
