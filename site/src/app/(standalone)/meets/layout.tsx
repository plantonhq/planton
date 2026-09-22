import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './meets.css';
import { HeaderLogo } from '../_components/HeaderLogo';

export const metadata: Metadata = {
  title: 'Planton Meets',
  description: 'Prospect meeting presentations for Planton',
  // Decks name prospects, quote their internal positioning, and in the investor
  // case carry fundraising terms. They are shared by URL, never public pages.
  // Crawling stays allowed on purpose: a robots.txt Disallow would prevent
  // crawlers from ever reading this directive, leaving shared links indexable.
  robots: { index: false, follow: false, nocache: true },
};

const inter = Inter({
  weight: ['400', '500', '600', '700', '800'],
  subsets: ['latin'],
  display: 'swap',
});

export default function MeetsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative">
      <HeaderLogo className="absolute top-[23px] left-8 z-[9999]" />
      <div className={`isolate ${inter.className}`}>{children}</div>
    </div>
  );
}
