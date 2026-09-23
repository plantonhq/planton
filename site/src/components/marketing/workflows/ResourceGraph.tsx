'use client';

import { useEffect, useId, useMemo, useRef, useState } from 'react';
import { WORKFLOW_COPY as copy, type WorkflowStory } from '../../../data/workflow-explainers';
import { ResourceIcon } from './ResourceIcon';
import { workflowDarkTokens as palette } from '../../../theme/workflows';
import { createFlowRoute, type Point } from './flowGeometry';
import { FlowArrowMarkers } from './FlowConnection';
import styles from './workflows.module.css';

/** A separate, deliberately explorable detail view. The animation never needs
 * to cram the complete graph into its viewport. Native selection and the
 * resource list keep the same information available to keyboard/no-JS readers. */
export function ResourceGraph({ story }: { story: WorkflowStory }) {
  const resources = story.architecture!.resources;
  const { byId, layers, height, positions, edges } = useMemo(() => {
    const byId = new Map(resources.map(resource => [resource.id, resource]));
    const rank = (id: string): number => {
      const dependencies = byId.get(id)!.requires;
      return dependencies.length ? 1 + Math.max(...dependencies.map(rank)) : 0;
    };
    const ranks = resources.map(resource => rank(resource.id));
    const layers = Array.from({ length: Math.max(...ranks) + 1 }, (_, i) => resources.filter((_, index) => ranks[index] === i));
    const height = Math.max(...layers.map(layer => layer.length)) * 80 + 32;
    const positions = new Map(layers.flatMap((layer, column) => layer.map((resource, row) => [resource.id, [16 + column * 260, (height - layer.length * 80) / 2 + row * 80] as const] as const)));
    const hasAncestor = (id: string, ancestor: string): boolean => byId.get(id)!.requires.some(parent => parent === ancestor || hasAncestor(parent, ancestor));
    const edges = resources.flatMap(resource => resource.requires.map(from => {
      const [sx, sy] = positions.get(from)!, [tx, ty] = positions.get(resource.id)!;
      const redundant = resource.requires.some(other => other !== from && hasAncestor(other, from));
      const midpoint = (sx + 210 + tx - 7) / 2;
      return { from, to: resource.id, redundant, path: createFlowRoute([sx + 210, sy + 28], [midpoint, sy + 28], [midpoint, ty + 28], [tx - 7, ty + 28]).path };
    }));
    return { byId, layers, height, positions, edges };
  }, [resources]);
  const [selected, setSelected] = useState('');
  const canvas = useRef<HTMLDivElement>(null);
  const marker = `resource-${useId().replaceAll(':', '')}`;
  const resource = byId.get(selected);
  const focusedHeight = Math.max(resource?.requires.length ?? 0, 1) * 80 + 32;
  const targetX = resource?.requires.length ? 600 : 300;
  const visibleNodes = resource ? [...resource.requires.map(id => byId.get(id)!), resource] : resources;
  const visiblePositions = resource ? new Map<string, Point>([
    ...resource.requires.map((id, index) => [id, [16, 16 + index * 80] as const] as const),
    [resource.id, [targetX, focusedHeight / 2 - 26] as const] as const,
  ]) : positions;
  const visibleEdges = resource ? resource.requires.map(from => {
    const [sx, sy] = visiblePositions.get(from)!, [tx, ty] = visiblePositions.get(resource.id)!;
    const midpoint = (sx + 210 + tx - 7) / 2;
    return { from, to: resource.id, path: createFlowRoute([sx + 210, sy + 26], [midpoint, sy + 26], [midpoint, ty + 26], [tx - 7, ty + 26]).path };
  }) : edges.filter(edge => !edge.redundant);
  useEffect(() => {
    const element = canvas.current;
    if (!element) return;
    const focused = byId.get(selected);
    const position = focused ? [focused.requires.length ? 600 : 300] : undefined;
    // Focus the requested resource within the horizontal graph only. Do not
    // move the page away from the selector or introduce automatic animation.
    element.scrollLeft = position ? Math.max(0, position[0] - element.clientWidth / 2 + 105) : 0;
  }, [selected, byId]);
  return <div className={styles.resourceExplorer} data-resource-explorer>
    <p>{copy.resourceIntro}</p>
    <label className={styles.resourceSelect}>{copy.resourceFocus}
      <select value={selected} onChange={event => setSelected(event.target.value)}>
        <option value="">{copy.resourceAll}</option>
        {resources.map(node => <option key={node.id} value={node.id}>{node.label}</option>)}
      </select>
    </label>
    <p className={styles.resourceSelection} aria-live="polite">{resource ? <><strong>{resource.label}</strong> · {resource.kind}<br />{copy.resourceRequires} {resource.requires.map(id => byId.get(id)!.label).join(', ') || copy.resourceNone}</> : copy.resourceHelp}</p>
    <div ref={canvas} className={styles.resourceCanvas} tabIndex={0} role="region" aria-label={copy.resourceRegion}>
      <svg viewBox={`0 0 ${resource ? 840 : layers.length * 260} ${resource ? focusedHeight : height}`} width={resource ? 840 : layers.length * 260} height={resource ? focusedHeight : height} aria-hidden="true">
        <FlowArrowMarkers id={marker} />
        {visibleEdges.map(edge => <path key={`${edge.from}-${edge.to}`} d={edge.path} fill="none" stroke={selected ? palette.semantic.flow : palette.edge.default} strokeWidth="1.5" markerEnd={`url(#${marker}-${selected ? 'active' : 'idle'})`} />)}
        {visibleNodes.map(node => {
          const [x, y] = visiblePositions.get(node.id)!;
          return <g key={node.id} transform={`translate(${x} ${y})`} data-resource={node.id}>
            <rect width="210" height="56" rx="7" fill={palette.surface.canvas} stroke={node.id === selected ? palette.semantic.flow : palette.edge.default} />
            <ResourceIcon kind={node.kind} x={9} y={12} size={32} /><text x="49" y="33" fontSize="12" fill={palette.text.primary}>{node.label}</text>
          </g>;
        })}
      </svg>
    </div>
    <details><summary>{copy.resourceInventory}</summary><div className={styles.resourceInventory}>{story.nodes.map(group => <section key={group.id}>
      <h4>{group.label}</h4>
      <ul>{resources.filter(node => node.group === group.id).map(node => <li key={node.id}><strong>{node.label}</strong><span>{node.kind}</span><span>{copy.resourceRequires} {node.requires.map(id => byId.get(id)!.label).join(', ') || copy.resourceNone}</span>{node.note && <span>{node.note}</span>}</li>)}</ul>
    </section>)}</div></details>
  </div>;
}
