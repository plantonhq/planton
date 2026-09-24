import type { WorkflowStory } from '../../../data/workflow-explainers.ts';

export const PHASE_SECONDS = 4;
export const HOLD_SECONDS = 4;
export const FPS = 30;
export const duration = (story: { phases: readonly unknown[] }) =>
  story.phases.length * PHASE_SECONDS + HOLD_SECONDS;

/** Pure sampling is the seam between browser playback, paused inspection and video.
 * Never read wall time or randomness in a scene: every frame must be reproducible. */
export function sample(story: WorkflowStory, seconds: number) {
  const time = Math.max(0, Math.min(seconds, duration(story)));
  const finished = time >= story.phases.length * PHASE_SECONDS;
  const phase = Math.min(Math.floor(time / PHASE_SECONDS), story.phases.length - 1);
  return { time, phase, finished, progress: time / duration(story) };
}

export function nodeState(nodePhase: number, phase: number, finished: boolean) {
  return finished || nodePhase < phase ? 'complete' : nodePhase === phase ? 'active' : 'waiting';
}
