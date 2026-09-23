 'use client';
import { usePathname } from 'next/navigation';
import { WebsiteShell } from '@planton/website-shell';
import { lightVariables } from '@/theme/homepage';
import { lightWebsiteTheme } from '@/theme/light-website';
import type { ReactNode } from 'react';

/** Only the homepage opts into light chrome during the incremental rollout. */
export function MarketingShell({children}:{children:ReactNode}) {
  const light = usePathname() === '/';
  return <div data-marketing-appearance={light ? 'light' : 'dark'} style={light ? {...lightVariables,color:lightWebsiteTheme.palette.text.primary} : undefined}><WebsiteShell theme={light ? lightWebsiteTheme : undefined}>{children}</WebsiteShell></div>;
}
