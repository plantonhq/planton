'use client';
import { useId, type CSSProperties } from 'react';
import { HERO as H } from '@/data/homepage-experience';
import { WORKFLOW_COPY as C } from '@/data/workflow-explainers';
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
/** Play once, then hold. Explicit controls own playback; pointer motion never does. */
export function HeroExperience() {
  const id = useId().replaceAll(':', '');
  const { rootRef, seconds, ready, reduced, playing, toggle, replay } = useWorkflowPlayback(H, {
    loop: false,
  });
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
        {ready && !reduced && (
          <div className={styles.playback}>
            <button
              type="button"
              onClick={toggle}
              disabled={finished}
              aria-label={`${playing ? C.pause : C.play}: ${H.title}`}
            >
              {playing ? C.pause : C.play}
            </button>
            <button type="button" onClick={replay} aria-label={`${C.replay}: ${H.title}`}>
              {C.replay}
            </button>
          </div>
        )}
      </figcaption>
    </figure>
  );
}
