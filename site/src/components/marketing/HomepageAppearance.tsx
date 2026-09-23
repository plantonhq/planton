import type { CSSProperties, ReactNode } from 'react';
import { homepageLightTokens as p } from '@/theme/homepage';
import styles from './homepage.module.css';

/** The approved, light-only homepage. No browser preference or hydration switch. */
export function HomepageAppearance({ children }: { children: ReactNode }) {
  const variables = {
    '--hp-canvas':p.surface.canvas,'--hp-panel':p.surface.panel,'--hp-raised':p.surface.raised,
    '--hp-ink':p.text.primary,'--hp-secondary':p.text.secondary,'--hp-line':p.edge.default,
    '--hp-button':p.cta.background,'--hp-button-text':p.cta.text,
  } as CSSProperties;
  return <div className={styles.home} style={variables} data-homepage-appearance="light">{children}</div>;
}
