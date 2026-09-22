import type { PropsWithChildren } from 'react';
import { HeaderLogo } from '../_components/HeaderLogo';

/**
 * The persona decks' chrome: the logo as a door home, and nothing else. A
 * deck fills the screen; its title, canonical, and noindex come from the
 * route registry through the page's metadata, and its palette is the site's.
 */
export default function DecksLayout({ children }: PropsWithChildren) {
  return (
    <div className="relative">
      <HeaderLogo className="absolute top-[23px] left-8 z-[9999]" />
      {children}
    </div>
  );
}
