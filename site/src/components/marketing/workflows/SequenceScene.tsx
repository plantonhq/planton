import { WORKFLOW_COPY as copy, type WorkflowStory } from '../../../data/workflow-explainers';
import { workflowDarkTokens as palette } from '../../../theme/workflows';
import { createFlowRoute, type Point } from './flowGeometry';
import { nodeState, PHASE_SECONDS, sample } from './timeline';
import { FlowArrowMarkers, FlowConnection } from './FlowConnection';

// Three stages per row preserve generous connector lanes. The return travels
// outside the cards; numbering and arrow direction make the folded order clear.
const desktop: readonly Point[] = [[20, 12], [400, 12], [780, 12], [780, 236], [400, 236], [20, 236]];
const mobile: readonly Point[] = Array.from({ length: 6 }, (_, i) => [60, 8 + i * 172] as const);
const routes = [desktop, mobile].map((points, layout) => points.slice(0, -1).map(([sx, sy], i) => {
  const [tx, ty] = points[i + 1];
  if (layout) return createFlowRoute([sx + 120, sy + 124], [sx + 120, sy + 148], [tx + 120, ty - 29], [tx + 120, ty - 5]);
  if (i === 2) return createFlowRoute([sx + 240, sy + 62], [sx + 345, sy + 62], [tx + 345, ty + 62], [tx + 245, ty + 62]);
  const right = tx > sx;
  const a: Point = [sx + (right ? 240 : 0), sy + 62];
  const b: Point = [tx + (right ? -6 : 246), ty + 62];
  const middle = (a[0] + b[0]) / 2;
  return createFlowRoute(a, [middle, a[1]], [middle, b[1]], b);
}));

export function SequenceScene({ story, seconds, compact, idPrefix }: {
  story: WorkflowStory; seconds: number; compact: boolean; idPrefix: string;
}) {
  const { phase, time, finished } = sample(story, seconds);
  const marker = `${idPrefix}-delivery-arrow`;
  const flow = palette.semantic.flow;
  return <svg viewBox={compact ? '0 0 360 1032' : '0 0 1120 400'} width="100%" style={{ display: 'block' }} role="img" aria-label={story.title}>
    <desc>{story.scope} {story.phases.map(p => `${p.title}. ${p.text}`).join(' ')}</desc>
    <FlowArrowMarkers id={marker} />
    {story.edges.map((edge, index) => {
      const route = routes[compact ? 1 : 0][index];
      const target = story.nodes.find(n => n.id === edge.to)!;
      // A protected action waits until the preceding human-decision phase ends.
      const active = !finished && phase === target.phase;
      return <g key={`${edge.from}-${edge.to}`} data-edge={`${edge.from}-${edge.to}`} data-transferring={active}>
        <FlowConnection route={route} active={active} seconds={time} marker={marker} />
      </g>;
    })}
    {story.nodes.map((node, i) => {
      const [x, y] = (compact ? mobile : desktop)[i];
      const state = nodeState(node.phase, phase, finished);
      const gate = node.gate;
      const accent = gate && state === 'active' ? palette.semantic.warn : flow;
      const progress = gate ? (state === 'complete' ? 1 : 0) : Math.min(1, Math.max(0, (time - node.phase * PHASE_SECONDS) / PHASE_SECONDS));
      const status = gate && state !== 'waiting'
        ? state === 'active' ? copy.approvalWaiting : copy.approvalComplete
        : state === 'complete' ? `✓ ${copy.complete}` : copy[state];
      return <g key={node.id} transform={`translate(${x} ${y})`} data-node={node.id} data-state={state}>
        <rect width="240" height="124" rx="10" fill={palette.surface.canvas} stroke={gate && state === 'active' ? accent : palette.edge.default} strokeWidth="1.2" />
        {state === 'active' && <rect width="240" height="124" rx="10" fill={accent} opacity="0.05" />}
        <text x="14" y="23" fontSize="15" fill={palette.text.secondary}>{String(i + 1).padStart(2, '0')}</text>
        <text x="14" y="47" fontSize="21" fontWeight="500" fill={palette.text.primary}>{node.label}</text>
        <text x="14" y="68" fontSize="15" fill={palette.text.secondary}>{node.detail}</text>
        <text x="14" y="96" fontSize={gate ? 14 : 16} fill={gate && state === 'active' ? accent : palette.text.secondary}>{status}</text>
        <rect x="14" y="110" width="212" height="3" rx="1.5" fill={palette.edge.default} />
        <rect x="14" y="110" width={212 * progress} height="3" rx="1.5" fill={accent} />
      </g>;
    })}
  </svg>;
}
