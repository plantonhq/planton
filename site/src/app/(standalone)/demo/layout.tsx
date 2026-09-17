import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './demo.css';
import { HeaderLogo } from '../_components/HeaderLogo';

export const metadata: Metadata = {
  title: 'Platform Demo - Planton',
  description:
    'Interactive demo showcasing Planton DevOps platform features and capabilities',
};

const inter = Inter({
  weight: ['400', '500', '600', '700'],
  subsets: ['latin'],
  display: 'swap',
});

export default function DemoLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative">
      <HeaderLogo className="absolute top-[23px] left-8 z-[9999]" />
      <div className={`isolate ${inter.className}`}>{children}</div>
    </div>
  );
}
