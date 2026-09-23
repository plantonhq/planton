export type ProviderId = 'aws' | 'gcp' | 'azure' | 'cloudflare' | 'digitalocean';
export type WorkflowId = ProviderId | 'delivery' | 'agents';
export interface WorkflowNode {
  id: string;
  label: string;
  detail: string;
  gate?: boolean;
  phase: number;
  icon?: string;
}
export interface WorkflowEdge {
  from: string;
  to: string;
  /** Output labels and transfer phases apply to resource-dependency diagrams. */
  label?: string;
  transferPhase?: number;
}
export interface WorkflowStory {
  id: WorkflowId;
  version: number;
  title: string;
  context: string;
  takeaway: string;
  scope: string;
  setup?: string;
  architecture?: { provider: ProviderId; resources: readonly ArchitectureResource[]; boundary?: string };
  phases: readonly { title: string; text: string }[];
  nodes: readonly WorkflowNode[];
  edges: readonly WorkflowEdge[];
}


export interface ArchitectureResource { id: string; label: string; kind: string; group: string; requires: readonly string[]; note?: string }
