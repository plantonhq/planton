'use client';
import { HeaderLogo } from './_components/HeaderLogo';

export default function DemoLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative">
      <HeaderLogo className="absolute top-[23px] left-8 z-[9999]" />
      {children}
    </div>
  );
}
