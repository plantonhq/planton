'use client';

import Image from 'next/image';
import { useEffect, useId, useRef, useState, useSyncExternalStore } from 'react';
import { PROVIDER_ORDER, PROVIDER_NAMES } from '../../../data/architecture-stories';
import { WORKFLOW_ICONS } from '../../../data/workflow-icons';
import { WORKFLOW_COPY as copy, type ProviderId } from '../../../data/workflow-explainers';
import { WorkflowExplainer } from './WorkflowExplainer';
import styles from './workflows.module.css';

const subscribe = () => () => {};
const subscribeMotion = (notify: () => void) => {
  const media = matchMedia('(prefers-reduced-motion: reduce)');
  media.addEventListener('change', notify);
  return () => media.removeEventListener('change', notify);
};

/** Selection and playback are distinct: manual exploration disables cycling,
 * while Pause survives tab changes. Only the selected story owns a clock. */
export function ArchitectureTabs() {
  const ready = useSyncExternalStore(subscribe, () => true, () => false);
  const reduced = useSyncExternalStore(subscribeMotion, () => matchMedia('(prefers-reduced-motion: reduce)').matches, () => true);
  const [selected, setSelected] = useState<ProviderId>('aws');
  const [previous, setPrevious] = useState<ProviderId | null>(null);
  const [autoCycle, setAutoCycle] = useState(true);
  const [playing, setPlaying] = useState(true);
  const [restart, setRestart] = useState(0);
  const tabs = useRef<(HTMLButtonElement | null)[]>([]);
  const id = `architectures-${useId().replaceAll(':', '')}`;
  // Exit cleanup is independent of CSS animation events (including interrupted
  // transitions and motion-preference changes). This timer never advances a story.
  useEffect(() => {
    if (!previous) return;
    const cleanup = setTimeout(() => setPrevious(null), 200);
    return () => clearTimeout(cleanup);
  }, [previous, selected, restart, reduced]);
  const inspect = () => setAutoCycle(false);
  const select = (provider: ProviderId) => { inspect(); if (provider !== selected) setPrevious(selected); setSelected(provider); setRestart(n => n + 1); };
  const complete = () => {
    if (autoCycle && !reduced) { setPrevious(selected); setSelected(PROVIDER_ORDER[(PROVIDER_ORDER.indexOf(selected) + 1) % PROVIDER_ORDER.length]); }
  };
  const cycling = autoCycle && !reduced;
  const cycleControl = <button data-auto-cycle type="button" disabled={reduced} aria-pressed={cycling} onClick={() => {
    if (cycling) inspect();
    else { setAutoCycle(true); setPlaying(true); setRestart(n => n + 1); }
  }}>{cycling ? copy.autoOn : copy.autoOff}</button>;
  return <div className={styles.architectures} data-architectures data-selected={selected} data-cycling={ready && cycling} onFocusCapture={event => {
    // Pointer clicks on controls must retain their current state through click.
    // Keyboard focus stops rotation; automatic transitions never move focus.
    if ((event.target as HTMLElement).matches(':focus-visible')) inspect();
  }}>
    {ready && <div className={styles.providerTabs} role="tablist" aria-label={copy.providers}>
      {PROVIDER_ORDER.map((provider, index) => <button key={provider} ref={element => { tabs.current[index] = element; }} role="tab" type="button" id={`${id}-${provider}-tab`} aria-controls={`${id}-panel`} aria-selected={provider === selected} tabIndex={provider === selected ? 0 : -1} onClick={() => select(provider)} onKeyDown={event => {
        const next = event.key === 'ArrowRight' ? (index + 1) % PROVIDER_ORDER.length : event.key === 'ArrowLeft' ? (index + PROVIDER_ORDER.length - 1) % PROVIDER_ORDER.length : event.key === 'Home' ? 0 : event.key === 'End' ? PROVIDER_ORDER.length - 1 : -1;
        if (next >= 0) { event.preventDefault(); select(PROVIDER_ORDER[next]); tabs.current[next]?.focus({ preventScroll: true }); tabs.current[next]?.scrollIntoView({ block: 'nearest', inline: 'nearest' }); }
      }}><Image unoptimized src={WORKFLOW_ICONS[provider === 'aws' ? 'aws-light' : provider]} width="24" height="24" alt="" />{PROVIDER_NAMES[provider]}</button>)}
    </div>}
    {ready ? <div className={styles.panelStack} id={`${id}-panel`} role="tabpanel" aria-labelledby={`${id}-${selected}-tab`}>
      {previous && !reduced && <div key={`${previous}-${selected}-${restart}`} className={styles.panelExit} aria-hidden="true" inert><WorkflowExplainer storyId={previous} staticOnly /></div>}
      <div key={`${selected}-${restart}`} className={styles.panelEnter}>
        <WorkflowExplainer storyId={selected} playback={{ playing, onPlayingChange: setPlaying, onComplete: complete, loop: false }} onInspect={inspect} extraControls={cycleControl} />
      </div>
    </div> : <div className={styles.staticArchitectures}>{PROVIDER_ORDER.map(provider => <WorkflowExplainer key={provider} storyId={provider} staticOnly />)}</div>}
  </div>;
}
