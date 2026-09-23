'use client';

import { Slide, SlideHeader, Callout } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

const milestones = [
  {
    year: '2023',
    title: 'Platform Foundation',
    description: 'Built Planton as a DevOps automation platform — infrastructure + CI/CD, multi-cloud, open-source IaC.',
    status: 'done' as const,
  },
  {
    year: 'Early 2025',
    title: 'The Co-Pilot Attempt',
    description: 'Built an AI co-pilot for infrastructure. No matter what we tried, it was never deterministic enough. Teams didn\'t trust it.',
    status: 'failed' as const,
  },
  {
    year: 'Mid 2025',
    title: 'The Breakthrough',
    description: 'Pivoted to "AI as orchestration layer." AI doesn\'t write IaC on the fly — it selects and configures the right deterministic tooling.',
    status: 'breakthrough' as const,
  },
  {
    year: '2026',
    title: 'AI Teammates Ship',
    description: 'Launched autonomous AI DevOps teammates powered by 370+ protobuf-modeled cloud resource kinds and deterministic execution.',
    status: 'current' as const,
  },
];

const statusStyles = {
  done: 'border-white/20 bg-white/[0.03]',
  failed: 'border-red-500/30 bg-red-500/5',
  breakthrough: 'border-[#10b981]/30 bg-[#10b981]/5',
  current: 'border-white/30 bg-white/10',
};

const dotStyles = {
  done: 'bg-white/30',
  failed: 'bg-red-500/70',
  breakthrough: 'bg-[#10b981]',
  current: 'bg-white',
};

export default function S05Journey(_props: SlideComponentProps) {
  return (
    <Slide>
      <SlideHeader
        sectionTag="Our Journey"
        title="From Co-Pilot to Orchestration Layer"
        subtitle="Two years of learning what actually works"
      />

      <div className="max-w-3xl mx-auto mb-6 sm:mb-8">
        <div className="space-y-3 sm:space-y-4">
          {milestones.map((m) => (
            <div
              key={m.year}
              className={`flex items-start gap-3 sm:gap-4 p-3 sm:p-4 rounded-xl border ${statusStyles[m.status]}`}
            >
              <div className="flex flex-col items-center shrink-0 pt-1">
                <div className={`w-3 h-3 rounded-full ${dotStyles[m.status]}`} />
              </div>
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xs sm:text-sm font-semibold text-white/50">{m.year}</span>
                  <span className="text-sm sm:text-base font-semibold text-white">{m.title}</span>
                </div>
                <p className="text-xs sm:text-sm text-white/60">{m.description}</p>
              </div>
            </div>
          ))}
        </div>
      </div>

      <Callout variant="success" className="max-w-2xl mx-auto">
        <p className="text-sm sm:text-base text-white font-medium text-center">
          Key insight: AI should not invent infrastructure code on the fly.
          <br />
          AI should select and configure the <span className="text-[#10b981]">right deterministic tooling</span>.
        </p>
      </Callout>
    </Slide>
  );
}
