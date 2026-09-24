import { workflowDarkTokens as palette } from '../../../theme/workflows';
import type { createFlowRoute } from './flowGeometry';

/** Arrowheads have a fixed size in scene coordinates, independent of stroke
 * weight. A filled tip joins the line instead of reading as a detached glyph. */
export function FlowArrowMarkers({ id }: { id: string }) {
  return <defs>{[false, true].map(active => <marker key={String(active)} id={`${id}-${active ? 'active' : 'idle'}`} viewBox="0 0 8 8" refX="7" refY="4" markerWidth="8" markerHeight="8" markerUnits="userSpaceOnUse" orient="auto">
    <path d="M1 1 7 4 1 7Z" fill={active ? palette.semantic.flow : palette.text.secondary} />
  </marker>)}</defs>;
}

export function FlowConnection({ route, active, seconds, marker, radius = 3.5 }: {
  route: ReturnType<typeof createFlowRoute>; active: boolean; seconds: number; marker: string; radius?: number;
}) {
  return <>
    <path d={route.path} fill="none" stroke={active ? palette.semantic.flow : palette.edge.default} strokeWidth="1.5" opacity={active ? 0.7 : 1} markerEnd={`url(#${marker}-${active ? 'active' : 'idle'})`} />
    {active && [0, 1, 2].map(packet => {
      const [cx, cy] = route.at((seconds * 90 / route.length + packet / 3) % 1);
      return <circle key={packet} cx={cx} cy={cy} r={radius} fill={palette.semantic.flow} />;
    })}
  </>;
}
