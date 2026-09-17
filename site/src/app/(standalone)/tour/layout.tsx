import type { Metadata } from 'next';
import { HeaderLogo } from '../_components/HeaderLogo';

export const metadata: Metadata = {
  title: 'Planton Tour',
  description: 'Interactive tour of the Planton platform',
};

export default function TourLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative">
      <HeaderLogo className="absolute top-[23px] left-8 z-[9999]" />
      <div className="tour-layout">{children}</div>
    </div>
  );
}
