'use client';

import { Slide, SlideHeader, TwoCol, Card, CardTitle, Checklist, Callout, Badge } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

export default function S11Stigmer(_props: SlideComponentProps) {
  return (
    <Slide>
      <SlideHeader
        sectionTag="The Spinoff"
        title="Stigmer — Open-Source AI Agent Platform"
        subtitle="Born from Planton's AI research, now its own company"
      />

      <TwoCol className="mb-6 sm:mb-8 max-w-4xl mx-auto">
        <Card className="text-left">
          <CardTitle className="mb-3 !text-base sm:!text-lg">What Is Stigmer</CardTitle>
          <Checklist
            items={[
              'Open-source AI agent platform (Apache 2.0)',
              'YAML-defined agents with skills, tools, and approval flows',
              'Self-deployable Kubernetes operator',
              'Can run completely inside a private network',
              'Mirrors Temporal\'s model: OSS + managed cloud',
            ]}
          />
        </Card>

        <Card className="text-left">
          <CardTitle className="mb-3 !text-base sm:!text-lg">Traction</CardTitle>
          <Checklist
            items={[
              'Planton is Stigmer\'s first customer',
              'Separate Delaware C-Corp, independently operated',
              'Spun off January 2026 — 4 months ago',
              '2 pilot users in Hyderabad building agentic experiences',
              'Growing community resonance in the developer ecosystem',
            ]}
          />
        </Card>
      </TwoCol>

      <Callout variant="highlight" className="max-w-3xl mx-auto">
        <div className="flex flex-col sm:flex-row items-center justify-center gap-3 sm:gap-6">
          <div className="text-center">
            <p className="text-sm sm:text-base text-white font-medium">
              All of Planton&apos;s AI agent research led to Stigmer.
            </p>
            <p className="text-xs sm:text-sm text-white/50 mt-1">
              The same deterministic execution layer that powers Planton&apos;s AI teammates
              is now available as a standalone platform for anyone building agentic products.
            </p>
          </div>
          <Badge variant="default" className="shrink-0">
            Separate $7M Cap
          </Badge>
        </div>
      </Callout>
    </Slide>
  );
}
