import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import InvestHeader from '@/components/invest/InvestHeader';
import { HeaderLogo } from '../_components/HeaderLogo';

export const metadata: Metadata = {
  title: 'Legal - Planton',
  description: 'Legal information and investor updates for Planton',
};

const inter = Inter({
  weight: ['400', '500', '600', '700', '800'],
  subsets: ['latin'],
  display: 'swap',
});

/**
 * Layout for legal pages including investor updates.
 * Paints the standalone logo overlay in its own layout.
 * Adds InvestHeader with Home button for navigation back to /invest.
 */
export default function LegalLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative">
      <HeaderLogo className="absolute top-[23px] left-8 z-[9999]" />
      <div className={`isolate ${inter.className}`}>
        <InvestHeader alwaysShow />
        {children}
      </div>
    </div>
  );
}
