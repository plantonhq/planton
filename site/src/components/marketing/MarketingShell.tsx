'use client';
import { trackExperience } from '@/lib/demo-analytics';
import { usePathname } from 'next/navigation';
import { WebsiteShell } from '@planton/website-shell';
import { lightVariables } from '@/theme/homepage';
import { lightWebsiteTheme } from '@/theme/light-website';
import type { ReactNode } from 'react';

/** All public website pages share the homepage's light appearance. */
export function MarketingShell({ children }: { children: ReactNode }) {
  const homepage = usePathname() === '/';
  return (
    <div
      data-marketing-appearance="light"
      style={{
        ...lightVariables,
        color: lightWebsiteTheme.palette.text.primary,
        background: lightWebsiteTheme.palette.background.default,
        minHeight: '100vh',
      }}
    >
      <WebsiteShell
        theme={lightWebsiteTheme}
        onSelfServiceAction={
          homepage
            ? () =>
                trackExperience('self_service_click', { location: 'navigation', door: 'hosted' })
            : undefined
        }
      >
        {children}
      </WebsiteShell>
    </div>
  );
}
