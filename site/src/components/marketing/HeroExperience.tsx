'use client';
import { useId, type CSSProperties } from 'react';
import { HERO as H } from '@/data/homepage-experience';
import { workflowDarkTokens as p } from '@/theme/workflows';
import { HeroScene } from './HeroScene';
import { useWorkflowPlayback } from './workflows/useWorkflowPlayback';
import styles from './experience.module.css';

export const experienceTheme = {
  '--ex-bg': p.surface.canvas,
  '--ex-panel': p.surface.raised,
  '--ex-text': p.text.primary,
  '--ex-muted': p.text.secondary,
  '--ex-line': p.edge.default,
  '--ex-flow': p.semantic.flow,
} as CSSProperties;
/** The hero workflow loops continuously while it is visible. */
export function HeroExperience() {
  const id = useId().replaceAll(':', '');
  const { rootRef, seconds } = useWorkflowPlayback(H);
  const phase = Math.min(3, Math.floor(seconds / 4)),
    finished = seconds >= 16;
  return (
    <figure
      ref={rootRef}
      className={styles.hero}
      style={experienceTheme}
      data-hero-phase={phase}
      data-hero-time={seconds.toFixed(2)}
    >
      <div className={styles.heroLabel}>{H.title}</div>
      <div className={styles.heroWide}>
        <HeroScene seconds={seconds} id={`${id}-wide`} />
      </div>
      <div className={styles.heroCompact}>
        <HeroScene seconds={seconds} id={`${id}-compact`} compact />
      </div>
      <figcaption className={styles.heroCaption}>
        <div>
          <strong>{finished ? H.complete : H.phases[phase].title}</strong>
          <span>{finished ? H.illustration : H.phases[phase].text}</span>
        </div>
      </figcaption>
    </figure>
  );
}
