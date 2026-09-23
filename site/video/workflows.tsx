import { AbsoluteFill, Composition, registerRoot, useCurrentFrame } from 'remotion';
import { WORKFLOWS, WORKFLOW_COPY, type WorkflowId } from '../src/data/workflow-explainers';
import { WorkflowScene } from '../src/components/marketing/workflows/WorkflowScene';
import { duration, FPS, sample } from '../src/components/marketing/workflows/timeline';
import { workflowDarkTokens } from '../src/theme/workflows';

/** Export-only adapter. Frame time enters the same scene used on the website;
 * Remotion and its renderer are never imported by the Next application. */
function WorkflowVideo({ storyId }: { storyId: WorkflowId }) {
  const story = WORKFLOWS[storyId];
  const palette = workflowDarkTokens;
  const seconds = useCurrentFrame() / FPS;
  const { phase } = sample(story, seconds);
  return <AbsoluteFill style={{ background: palette.surface.canvas, color: palette.text.primary, fontFamily: 'Arial, sans-serif', padding: 60 }}>
    <div style={{ fontSize: 25, marginBottom: 38, color: palette.text.secondary }}>{story.context}</div>
    <div style={{ fontSize: 54, lineHeight: 1.1, letterSpacing: -2, maxWidth: 900, minHeight: 130 }}>{story.title}</div>
    <div style={{ background: palette.surface.panel, border: `1px solid ${palette.edge.default}`, borderRadius: 16, marginTop: 24, padding: '24px 4px' }}>
      <WorkflowScene story={story} seconds={seconds} idPrefix={`video-${storyId}`} />
    </div>
    <div style={{ marginTop: 38, display: 'flex', gap: 20 }}>
      <div style={{ fontSize: 22, color: palette.text.secondary }}>{String(phase + 1).padStart(2, '0')}</div>
      <div><div style={{ fontSize: 30, marginBottom: 14 }}>{story.phases[phase].title}</div><div style={{ fontSize: 25, lineHeight: 1.5, color: palette.text.secondary }}>{story.phases[phase].text}</div></div>
    </div>
    <div style={{ position: 'absolute', bottom: 36, left: 60, right: 60, display: 'flex', justifyContent: 'space-between', fontSize: 18, color: palette.text.secondary }}><span>{WORKFLOW_COPY.illustration}</span><span>planton.ai</span></div>
  </AbsoluteFill>;
}

function Root() {
  return <>{Object.values(WORKFLOWS).map(story => <Composition key={story.id} id={story.id} component={WorkflowVideo} defaultProps={{ storyId: story.id }} durationInFrames={duration(story) * FPS} fps={FPS} width={1080} height={1080} />)}</>;
}
registerRoot(Root);
