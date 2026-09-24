'use client';
import dynamic from 'next/dynamic';
const Architecture = dynamic(() =>
  import('./workflows/ArchitectureTabs').then((m) => m.ArchitectureTabs)
);
const Explainer = dynamic(() =>
  import('./workflows/WorkflowExplainer').then((m) => m.WorkflowExplainer)
);
export function ArchitectureStories() {
  return <Architecture />;
}
export function DeliveryStory() {
  return <Explainer storyId="delivery" />;
}
export function AgentStory() {
  return <Explainer storyId="agents" />;
}
