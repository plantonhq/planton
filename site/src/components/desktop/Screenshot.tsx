import { Box } from '@mui/material';
import Image from 'next/image';
import type { FC } from 'react';
import type { DesktopScreenshot } from '@/data/desktop-screenshots';

/**
 * A real capture of the desktop app in its own window chrome, framed only by
 * the library's border. A scene with no capture yet renders nothing: a
 * placeholder would be a picture of nothing pretending to be the product.
 */
export const Screenshot: FC<{ shot: DesktopScreenshot | null; className?: string; priority?: boolean }> = ({ shot, className = '', priority = false }) =>
  shot ? (
    <Box className={`rounded-xl border border-edge overflow-hidden bg-canvas ${className}`}>
      <Image src={shot.src} alt={shot.alt} width={shot.width} height={shot.height} priority={priority} className="w-full h-auto block" />
    </Box>
  ) : null;
