import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import InvestHeader from '@/components/invest/InvestHeader';
import './invest.css';
import { HeaderLogo } from '../_components/HeaderLogo';

export const metadata: Metadata = {
  title: 'Invest in Planton - The Self-Service Cloud Platform',
  description:
    'Join us in building the platform that makes cloud infrastructure accessible to every company. Seed stage investment opportunity.',
  // The investor pages carry the round's terms and the cap table walkthrough.
  // They are shared by URL with the people they are for, never search results.
  // Crawling stays allowed so crawlers can read this directive; a robots.txt
  // Disallow would leave shared links indexable.
  robots: { index: false, follow: false, nocache: true },
};

const inter = Inter({
  weight: ['400', '500', '600', '700', '800'],
  subsets: ['latin'],
  display: 'swap',
});

/**
 * Layout for all investor-related pages.
 * 
 * Paints the standalone logo overlay in its own layout.
 * Adds InvestHeader with Home button for navigation back to /invest.
 * Provides Inter font and isolate context for all /invest/* routes.
 */
export default function InvestLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative">
      <HeaderLogo className="absolute top-[23px] left-8 z-[9999]" />
      <div className={`isolate ${inter.className}`}>
        <InvestHeader />
        {children}
      </div>
    </div>
  );
}
