import type { PropsWithChildren } from 'react';
import { FocusFrame } from '../_components/FocusFrame';

export default function Layout({ children }: PropsWithChildren) {
  return <FocusFrame>{children}</FocusFrame>;
}
