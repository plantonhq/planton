'use client';

import { useId, type CSSProperties, type ReactNode } from 'react';
import { workflowDarkTokens } from '../../../theme/workflows';
import { WORKFLOW_COPY as copy, WORKFLOWS, type WorkflowId } from '../../../data/workflow-explainers.ts';
import { WorkflowScene } from './WorkflowScene';
import { PHASE_SECONDS, sample } from './timeline';
import { useWorkflowPlayback } from './useWorkflowPlayback';
import styles from './workflows.module.css';
import { ResourceGraph } from './ResourceGraph';
import type { PlaybackOptions } from './useWorkflowPlayback';

/** A full-width explainer for a product story. Native text and a static SVG ship
 * with the page; JS progressively adds playback and deliberate step inspection. */
export function WorkflowExplainer({ storyId, playback, onInspect, extraControls, staticOnly = false }: {
  storyId: WorkflowId; playback?: PlaybackOptions; onInspect?: () => void; extraControls?: ReactNode; staticOnly?: boolean;
}) {
  const story = WORKFLOWS[storyId];
  const controller = useWorkflowPlayback(story, { ...playback, suspended: staticOnly || playback?.suspended });
  const { rootRef, ready, reduced, playing, visible, seek, replay, toggle } = controller;
  const seconds = staticOnly ? story.phases.length * PHASE_SECONDS + 4 : controller.seconds;
  const state = sample(story, seconds);
  const id = useId().replaceAll(':', '');
  const infrastructure = Boolean(story.architecture);
  const dark = workflowDarkTokens;
  const theme = {
    '--hp-canvas': dark.surface.canvas, '--hp-panel': dark.surface.panel,
    '--hp-ink': dark.text.primary, '--hp-secondary': dark.text.secondary,
    '--hp-line': dark.edge.default, '--workflow-flow': dark.semantic.flow,
  } as CSSProperties;
  const controls = <div className={styles.controls}>{!staticOnly && extraControls}
    {ready && !reduced && !staticOnly ? <>
      <button type="button" onClick={() => { onInspect?.(); toggle(); }} aria-label={`${playing ? copy.pause : copy.play}: ${story.title}`}><span aria-hidden="true">{playing ? 'Ⅱ' : '▶'}</span>{playing ? copy.pause : copy.play}</button>
      <button type="button" onClick={() => { onInspect?.(); replay(); }} aria-label={`${copy.replay}: ${story.title}`}><span aria-hidden="true">↺</span>{copy.replay}</button>
    </> : <span className={styles.static}>{copy.static}</span>}
  </div>;
  return <figure ref={rootRef} className={`${styles.figure} ${styles.dark}`} style={theme} data-kind={infrastructure ? 'architecture' : 'sequence'} data-workflow={story.id} data-phase={state.phase} data-playing={playing && visible && !reduced && !staticOnly && !state.finished} aria-labelledby={`${id}-title`}>
    <div className={styles.top}>
      <div><span className={styles.context}>{story.context}</span><h3 id={`${id}-title`}>{story.title}</h3></div>
    </div>
    {story.setup && <p className={styles.setup}>{story.setup}</p>}
    <div className={styles.scene} aria-hidden="true">
      <div className={styles.wide}><WorkflowScene story={story} seconds={seconds} idPrefix={`${id}-wide`} /></div>
      <div className={styles.compact}><WorkflowScene story={story} seconds={seconds} compact idPrefix={`${id}-compact`} /></div>
    </div>
    <div className={styles.readout}>
      <span className={styles.phaseNumber} aria-hidden="true">{String(state.phase + 1).padStart(2, '0')}{` / ${String(story.phases.length).padStart(2, '0')}`}</span>
      <div className={styles.notes}>{story.phases.map((phase, i) => <div key={phase.title} className={styles.phaseNote} data-active={state.phase === i} aria-hidden={state.phase !== i}><strong>{phase.title}</strong><p>{phase.text}</p></div>)}</div>
    </div>
    <figcaption className={styles.caption}>
    <ol style={{ gridTemplateColumns: `repeat(${story.phases.length}, minmax(0, 1fr))` }} className={styles.phases} aria-label={copy.explanation}>
      {story.phases.map((step, i) => <li key={step.title}>
        <button type="button" disabled={!ready || reduced || staticOnly} onClick={() => { onInspect?.(); seek(i * PHASE_SECONDS + 0.5); }} aria-label={`${copy.selectPhase} ${i + 1}: ${step.title}`} aria-pressed={state.phase === i}>
          <span className={styles.stepIndex}>{String(i + 1).padStart(2, '0')}</span><span className={styles.stepName}>{step.title}</span>
          <span className={styles.stepProgress} aria-hidden="true" style={{ transform: `scaleX(${state.finished || state.phase > i ? 1 : state.phase === i ? (state.time % PHASE_SECONDS) / PHASE_SECONDS : 0})` }} />
        </button>
      </li>)}
    </ol>
      <div className={styles.legend}><span><i aria-hidden="true" />{infrastructure ? copy.resourceOutput : story.id === 'agents' ? copy.agentFlow : copy.deliveryFlow}</span><span>{copy.illustration}</span></div>
      {controls}
    </figcaption>
    <details className={styles.transcript} onToggle={event => { if (event.currentTarget.open) { onInspect?.(); seek(seconds); } }}><summary>{infrastructure ? copy.explore : copy.explanation}</summary>
      {infrastructure && <ResourceGraph story={story} />}
      <ol>{story.phases.map(step => <li key={step.title}><strong>{step.title}</strong><p>{step.text}</p></li>)}</ol>
      <p>{story.scope}</p>
    </details>
  </figure>;
}
