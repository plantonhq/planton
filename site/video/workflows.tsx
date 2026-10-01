import { HERO } from '../src/data/homepage-experience';
import { HeroScene } from '../src/components/marketing/HeroScene';
import { AbsoluteFill, Composition, registerRoot, useCurrentFrame } from 'remotion';
import { WORKFLOWS, WORKFLOW_COPY, type WorkflowId } from '../src/data/workflow-explainers';
import { WorkflowScene } from '../src/components/marketing/workflows/WorkflowScene';
import { duration, FPS, sample } from '../src/components/marketing/workflows/timeline';
import { workflowDarkTokens } from '../src/theme/workflows';
import { HomepageOverview } from './Overview';
import { OVERVIEW_VIDEO } from '../src/data/homepage-video';

/** Export-only adapter. Frame time enters the same scene used on the website;
 * Remotion and its renderer are never imported by the Next application. */
function WorkflowVideo({ storyId }: { storyId: WorkflowId }) {
  const story = WORKFLOWS[storyId];
  const palette = workflowDarkTokens;
  const seconds = useCurrentFrame() / FPS;
  const { phase } = sample(story, seconds);
  return (
    <AbsoluteFill
      style={{
        background: palette.surface.canvas,
        color: palette.text.primary,
        fontFamily: 'Arial, sans-serif',
        padding: 60,
      }}
    >
      <div style={{ fontSize: 25, marginBottom: 38, color: palette.text.secondary }}>
        {story.context}
      </div>
      <div
        style={{ fontSize: 54, lineHeight: 1.1, letterSpacing: -2, maxWidth: 900, minHeight: 130 }}
      >
        {story.title}
      </div>
      <div
        style={{
          background: palette.surface.panel,
          border: `1px solid ${palette.edge.default}`,
          borderRadius: 16,
          marginTop: 24,
          padding: '24px 4px',
        }}
      >
        <WorkflowScene story={story} seconds={seconds} idPrefix={`video-${storyId}`} />
      </div>
      <div style={{ marginTop: 38, display: 'flex', gap: 20 }}>
        <div style={{ fontSize: 22, color: palette.text.secondary }}>
          {String(phase + 1).padStart(2, '0')}
        </div>
        <div>
          <div style={{ fontSize: 30, marginBottom: 14 }}>{story.phases[phase].title}</div>
          <div style={{ fontSize: 25, lineHeight: 1.5, color: palette.text.secondary }}>
            {story.phases[phase].text}
          </div>
        </div>
      </div>
      <div
        style={{
          position: 'absolute',
          bottom: 36,
          left: 60,
          right: 60,
          display: 'flex',
          justifyContent: 'space-between',
          fontSize: 18,
          color: palette.text.secondary,
        }}
      >
        <span>{WORKFLOW_COPY.illustration}</span>
        <span>planton.ai</span>
      </div>
    </AbsoluteFill>
  );
}

function HeroVideo() {
  const seconds = useCurrentFrame() / FPS,
    phase = Math.min(3, Math.floor(seconds / 4)),
    p = workflowDarkTokens;
  return (
    <AbsoluteFill
      style={{
        background: p.surface.canvas,
        color: p.text.primary,
        fontFamily: 'Arial,sans-serif',
        padding: 60,
      }}
    >
      <div style={{ fontSize: 48, lineHeight: 1.2, marginBottom: 40 }}>{HERO.title}</div>
      <HeroScene seconds={seconds} id="hero-video" />
      <div style={{ marginTop: 28, fontSize: 29 }}>
        {seconds >= 16 ? HERO.complete : HERO.phases[phase].title}
      </div>
      <div style={{ marginTop: 14, fontSize: 24, color: p.text.secondary, lineHeight: 1.5 }}>
        {seconds >= 16 ? HERO.illustration : HERO.phases[phase].text}
      </div>
      <div
        style={{
          position: 'absolute',
          bottom: 36,
          right: 60,
          color: p.text.secondary,
          fontSize: 20,
        }}
      >
        planton.ai
      </div>
    </AbsoluteFill>
  );
}

function Root() {
  return (
    <>
      <Composition id={OVERVIEW_VIDEO.id} component={HomepageOverview} durationInFrames={OVERVIEW_VIDEO.duration * OVERVIEW_VIDEO.fps} fps={OVERVIEW_VIDEO.fps} width={1920} height={1080} />
      <Composition
        id="hero"
        component={HeroVideo}
        durationInFrames={20 * FPS}
        fps={FPS}
        width={1080}
        height={1080}
      />
      {Object.values(WORKFLOWS).map((story) => (
        <Composition
          key={story.id}
          id={story.id}
          component={WorkflowVideo}
          defaultProps={{ storyId: story.id }}
          durationInFrames={duration(story) * FPS}
          fps={FPS}
          width={1080}
          height={1080}
        />
      ))}
    </>
  );
}
registerRoot(Root);
