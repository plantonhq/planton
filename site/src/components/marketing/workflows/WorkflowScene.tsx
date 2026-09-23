import type { WorkflowStory } from '../../../data/workflow-explainers.ts';
import { InfrastructureScene } from './InfrastructureScene';
import { SequenceScene } from './SequenceScene';

/** Pure authored SVGs share the browser and video timeline. Keep layout choices
 * here, and leave timers, visibility and interaction in the playback adapter. */
export function WorkflowScene(props: {
  story: WorkflowStory; seconds: number; compact?: boolean; idPrefix: string;
}) {
  return Boolean(props.story.architecture)
    ? <InfrastructureScene {...props} compact={props.compact ?? false} />
    : <SequenceScene {...props} compact={props.compact ?? false} />;
}
