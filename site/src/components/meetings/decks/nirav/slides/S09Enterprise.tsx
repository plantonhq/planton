'use client';

import { Slide, SlideHeader, Grid, Card, CardTitle, CardText, Callout } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

const features = [
  {
    icon: '🏗️',
    title: 'Self-Hosted Runner',
    description: 'Kubernetes operator that deploys completely inside your private network. Your infrastructure, your control.',
  },
  {
    icon: '🔀',
    title: 'Hybrid Model',
    description: 'SaaS control plane + private data plane. Or go fully self-hosted. Flexible deployment for any security posture.',
  },
  {
    icon: '📡',
    title: 'Outbound Only',
    description: 'Runners need only outbound connections. Zero inbound ports required. Works in the most locked-down environments.',
  },
  {
    icon: '📱',
    title: 'Multi-Surface',
    description: 'Web, Desktop (Tauri/Rust), Mobile (Flutter), CLI. DevOps teammates go where you go — like Slack for infrastructure.',
  },
];

export default function S09Enterprise(_props: SlideComponentProps) {
  return (
    <Slide>
      <SlideHeader
        sectionTag="Enterprise Ready"
        title="Built for 50–500+ Engineer Organizations"
        subtitle="The same platform works as SaaS, hybrid, or fully self-hosted"
      />

      <Grid cols={2} gap="sm" className="mb-6 sm:mb-8 max-w-4xl mx-auto">
        {features.map((f) => (
          <Card key={f.title} className="text-left">
            <div className="text-2xl mb-2">{f.icon}</div>
            <CardTitle className="mb-1 !text-sm sm:!text-base">{f.title}</CardTitle>
            <CardText>{f.description}</CardText>
          </Card>
        ))}
      </Grid>

      <Callout variant="highlight" className="max-w-3xl mx-auto">
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 sm:gap-8 text-center">
          <div>
            <div className="text-xs text-white/50 mb-1">Quick Start</div>
            <div className="text-sm sm:text-base text-white font-medium">SaaS — minutes to deploy</div>
          </div>
          <div className="hidden sm:block text-white/20">|</div>
          <div>
            <div className="text-xs text-white/50 mb-1">Maximum Control</div>
            <div className="text-sm sm:text-base text-white font-medium">Self-hosted — &lt;1 day setup</div>
          </div>
          <div className="hidden sm:block text-white/20">|</div>
          <div>
            <div className="text-xs text-white/50 mb-1">Best of Both</div>
            <div className="text-sm sm:text-base text-white font-medium">Hybrid — SaaS brain, private execution</div>
          </div>
        </div>
      </Callout>
    </Slide>
  );
}
