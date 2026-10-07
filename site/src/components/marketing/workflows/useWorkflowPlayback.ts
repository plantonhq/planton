'use client';

import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from 'react';
import { sendGAEvent } from '@next/third-parties/google';
import { duration, FPS } from './timeline';

const subscribeMotion = (notify: () => void) => {
  const media = matchMedia('(prefers-reduced-motion: reduce)');
  media.addEventListener('change', notify);
  return () => media.removeEventListener('change', notify);
};
const motionSnapshot = () => matchMedia('(prefers-reduced-motion: reduce)').matches;
const subscribeHydration = () => () => {};

export interface PlaybackOptions {
  onComplete?: () => void;
  loop?: boolean;
  suspended?: boolean;
}

/** The sole clock owns completion, including the final hold. Visibility and explicit suspension
 * stop the clock; they never seek forward. Provider selection remounts this hook
 * so clocks and completion bookkeeping cannot leak into another story. */
export function useWorkflowPlayback(
  story: { id: string; version: number; phases: readonly unknown[] },
  options: PlaybackOptions = {}
) {
  const rootRef = useRef<HTMLElement>(null);
  const clock = useRef(0);
  const [seconds, setSeconds] = useState(0);
  const ready = useSyncExternalStore(
    subscribeHydration,
    () => true,
    () => false
  );
  const reduced = useSyncExternalStore(subscribeMotion, motionSnapshot, () => true);
  const [visible, setVisible] = useState(false);
  const callback = useRef(options.onComplete);
  const onComplete = options.onComplete;
  useEffect(() => {
    callback.current = onComplete;
  }, [onComplete]);
  const completed = useRef(false),
    notified = useRef(false),
    eligibleCycle = useRef(true),
    viewed = useRef(false);

  useEffect(() => {
    const element = rootRef.current;
    if (!element) return;
    let intersecting = false;
    let exposureTimer: ReturnType<typeof setTimeout> | undefined;
    const update = () => {
      const inView = intersecting && !document.hidden;
      setVisible(inView);
      clearTimeout(exposureTimer);
      if (inView && !viewed.current)
        exposureTimer = setTimeout(() => {
          viewed.current = true;
          sendGAEvent('event', 'workflow_explainer_view', {
            story_id: story.id,
            story_version: story.version,
            motion_mode: matchMedia('(prefers-reduced-motion: reduce)').matches
              ? 'static'
              : 'animated',
          });
        }, 1000);
    };
    const observer = new IntersectionObserver(
      ([entry]) => {
        intersecting = entry.isIntersecting && entry.intersectionRatio >= 0.2;
        update();
      },
      { threshold: [0, 0.2] }
    );
    observer.observe(element);
    document.addEventListener('visibilitychange', update);
    return () => {
      observer.disconnect();
      clearTimeout(exposureTimer);
      document.removeEventListener('visibilitychange', update);
    };
  }, [story]);

  useEffect(() => {
    if (!ready || reduced || !visible || options.suspended) return;
    let frame: number;
    let previous = performance.now(),
      lastPaint = previous;
    const tick = (now: number) => {
      clock.current += (now - previous) / 1000;
      previous = now;
      if (clock.current >= duration(story)) {
        if (!completed.current && eligibleCycle.current) {
          completed.current = true;
          sendGAEvent('event', 'workflow_explainer_cycle_complete', {
            story_id: story.id,
            story_version: story.version,
            motion_mode: 'animated',
          });
        }
        if (options.loop === false) {
          clock.current = duration(story);
          setSeconds(clock.current);
          if (!notified.current) {
            notified.current = true;
            callback.current?.();
          }
          return;
        }
        clock.current %= duration(story);
        eligibleCycle.current = true;
      }
      if (now - lastPaint >= 1000 / FPS) {
        setSeconds(clock.current);
        lastPaint = now;
      }
      frame = requestAnimationFrame(tick);
    };
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [ready, reduced, story, visible, options.suspended, options.loop]);

  const seek = useCallback(
    (time: number) => {
      eligibleCycle.current = false;
      clock.current = Math.max(0, Math.min(time, duration(story)));
      setSeconds(clock.current);
    },
    [story]
  );
  return {
    rootRef,
    seconds: reduced ? duration(story) : seconds,
    ready,
    reduced,
    visible,
    seek,
  };
}
