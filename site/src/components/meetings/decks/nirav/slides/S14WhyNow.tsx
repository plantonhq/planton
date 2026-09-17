'use client';

import { Slide, SlideHeader, IconList, Callout } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

export default function S14WhyNow(_props: SlideComponentProps) {
  return (
    <Slide>
      <SlideHeader
        sectionTag="Why Now"
        title="The Timing Is Right"
        subtitle="Five converging forces"
      />

      <div className="max-w-2xl mx-auto mb-6 sm:mb-8">
        <IconList
          items={[
            {
              icon: '🧠',
              title: 'AI Models Are Finally Capable Enough',
              description: 'LLMs can now reliably orchestrate structured tooling — the missing piece for infrastructure automation.',
            },
            {
              icon: '🛡️',
              title: 'Deterministic Tooling Is Our Moat',
              description: 'Competitors are still trying raw AI on infrastructure. Our insight — pair AI with deterministic tools — took 2 years and $500K to discover.',
            },
            {
              icon: '🏢',
              title: 'Enterprise Demand Is Immediate',
              description: 'Every enterprise struggles to hire DevOps talent. The pain is acute and growing — not a future problem.',
            },
            {
              icon: '🏁',
              title: 'First Mover in AI Teammates for Infrastructure',
              description: 'We are the first to ship named, skilled AI teammates backed by a validated 370+ component cloud catalog.',
            },
            {
              icon: '🎯',
              title: 'Two Companies, Two Bets',
              description: 'Planton (the product) and Stigmer (the agent platform) — both independently valuable, together a powerful ecosystem.',
            },
          ]}
        />
      </div>

      <Callout variant="success" className="max-w-2xl mx-auto">
        <p className="text-sm sm:text-base text-white font-medium text-center">
          The window is now. AI has crossed the capability threshold for infrastructure.
          <br />
          <span className="text-white/60">The company that earns enterprise trust first wins the market.</span>
        </p>
      </Callout>
    </Slide>
  );
}
