import type { PropsWithChildren } from 'react';
import { LightMarketingSurface } from '@/components/marketing/LightMarketingSurface';
import { FocusFrame } from '../_components/FocusFrame';

export default function Layout({ children }: PropsWithChildren) {
  return <LightMarketingSurface><FocusFrame>{children}</FocusFrame></LightMarketingSurface>;
}
