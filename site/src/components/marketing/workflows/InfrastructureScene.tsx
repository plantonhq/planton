import { useMemo } from 'react';
import { WORKFLOW_COPY as copy, type WorkflowStory, type ProviderId } from '../../../data/workflow-explainers';
import { ResourceIcon } from './ResourceIcon';
import { workflowDarkTokens as palette } from '../../../theme/workflows';
import { nodeState, PHASE_SECONDS, sample } from './timeline';
import type { Point } from './flowGeometry';
import { ARCHITECTURE_LAYOUTS, ARCHITECTURE_SIZE, architectureRoute } from './architectureLayout';
import { FlowArrowMarkers, FlowConnection } from './FlowConnection';

export function InfrastructureScene({ story, seconds, compact, idPrefix }: {
  story: WorkflowStory; seconds: number; compact: boolean; idPrefix: string;
}) {
  const { phase, time, finished } = sample(story, seconds);
  const marker = `${idPrefix}-prerequisite-arrow`;
  const points: readonly Point[] = useMemo(() => compact ? story.nodes.map(node => [74, 12 + [...story.nodes].sort((a, b) => a.phase - b.phase).findIndex(n => n.id === node.id) * 142] as const) : ARCHITECTURE_LAYOUTS[story.id as ProviderId], [story, compact]);
  const routes = useMemo(() => story.edges.map(edge => architectureRoute(story.id as ProviderId, `${edge.from}-${edge.to}`, points, story.nodes.findIndex(n => n.id === edge.from), story.nodes.findIndex(n => n.id === edge.to), compact)), [points, story, compact]);
  const { width, height, viewWidth, viewHeight } = ARCHITECTURE_SIZE;
  return <svg viewBox={compact ? `0 0 360 ${story.nodes.length * 142}` : `0 0 ${viewWidth} ${viewHeight}`} width="100%" style={{ display: 'block' }} role="img" aria-label={story.title}>
    <desc>{story.scope} {story.phases.map(p => `${p.title} ${p.text}`).join(' ')}</desc>
    <FlowArrowMarkers id={marker} />
    {!compact && story.architecture?.boundary && <g aria-hidden="true">
      <rect x="306" y="120" width="900" height="232" rx="12" fill="none" stroke={palette.edge.default} strokeDasharray="4 6" />
      <text x="326" y="154" fill={palette.text.secondary} fontSize="12" letterSpacing="1">{story.architecture.boundary}</text>
    </g>}
    {story.edges.map((edge, index) => {
      const target = story.nodes.findIndex(n => n.id === edge.to);
      const active = !finished && phase === (edge.transferPhase ?? story.nodes[target].phase);
      return <g key={`${edge.from}-${edge.to}`} data-edge={`${edge.from}-${edge.to}`} data-transferring={active} aria-hidden="true">
        <FlowConnection route={routes[index]} active={active} seconds={time} marker={marker} />
      </g>;
    })}
    {story.nodes.map((node, i) => {
      const [x, y] = points[i], status = nodeState(node.phase, phase, finished);
      const progress = Math.min(1, Math.max(0, (time - node.phase * PHASE_SECONDS) / PHASE_SECONDS));
      return <g key={node.id} transform={`translate(${x} ${y})`} data-node={node.id} data-state={status}>
        <rect width={width} height={height} rx="10" fill={palette.surface.canvas} stroke={palette.edge.default} strokeWidth="1.2" />
        {status === 'active' && <rect width={width} height={height} rx="10" fill={palette.semantic.flow} opacity="0.07" />}
        <ResourceIcon kind={node.icon!} x={12} y={12} />
        <text x="58" y="29" fontSize="17" fontWeight="500" fill={palette.text.primary}>{node.label}</text>
        <text x="58" y="49" fontSize="12" fill={palette.text.secondary}>{node.detail}</text>
        <text x="14" y="74" fontSize="13" fill={palette.text.secondary}>{status === 'complete' ? `✓ ${copy.ready}` : status === 'active' ? copy.deploying : copy.waiting}</text>
        <rect x="14" y="86" width={width - 28} height="2.5" rx="1" fill={palette.edge.default} />
        <rect x="14" y="86" width={(width - 28) * progress} height="2.5" rx="1" fill={palette.semantic.flow} />
      </g>;
    })}
  </svg>;
}
