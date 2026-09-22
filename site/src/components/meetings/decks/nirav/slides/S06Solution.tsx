'use client';

import { Slide, SlideHeader, FlowDiagram, Card, CardTitle, CardText, Callout } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

export default function S06Solution(_props: SlideComponentProps) {
  return (
    <Slide variant="solution">
      <SlideHeader
        sectionTag="The Solution"
        title="AI Orchestrates. Tools Execute."
        subtitle="How Planton makes AI safe for infrastructure"
      />

      <FlowDiagram
        steps={[
          { icon: '🤖', label: 'AI Teammate' },
          { icon: '🧠', label: 'Understands Intent' },
          { icon: '📋', label: 'Selects from Catalog' },
          { icon: '⚙️', label: 'Runs Deterministic Tools' },
          { icon: '✅', label: 'Predictable Result' },
        ]}
        className="mb-6 sm:mb-8"
      />

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 sm:gap-4 mb-6 sm:mb-8 max-w-4xl mx-auto">
        <Card className="text-center">
          <div className="text-3xl mb-2">📐</div>
          <CardTitle className="!text-sm sm:!text-base mb-1">Protobuf-Modeled</CardTitle>
          <CardText>370+ cloud resource kinds structured with Protocol Buffers. AI excels at working with structured data.</CardText>
        </Card>
        <Card className="text-center">
          <div className="text-3xl mb-2">🔧</div>
          <CardTitle className="!text-sm sm:!text-base mb-1">Deterministic Tooling</CardTitle>
          <CardText>Terraform, Helm, Kustomize — what you write is what you get. The tools are fully predictable.</CardText>
        </Card>
        <Card className="text-center">
          <div className="text-3xl mb-2">🌐</div>
          <CardTitle className="!text-sm sm:!text-base mb-1">Multi-Cloud Catalog</CardTitle>
          <CardText>AWS, GCP, Azure, Kubernetes, Cloudflare, and 9 more providers — all in one unified catalog.</CardText>
        </Card>
      </div>

      <Callout variant="success" className="max-w-2xl mx-auto">
        <p className="text-sm sm:text-base md:text-lg text-white font-semibold text-center">
          &ldquo;AI orchestrates. Tools execute. Results are deterministic.&rdquo;
        </p>
      </Callout>
    </Slide>
  );
}
