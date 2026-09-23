'use client';

import { Slide, SlideHeader, TwoCol, Card, CardTitle, Callout } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

export default function S04DeterminismInsight(_props: SlideComponentProps) {
  return (
    <Slide>
      <SlideHeader
        sectionTag="The Insight"
        title="The Determinism Gap"
        subtitle="Why AI works in coding but not in infrastructure"
      />

      <TwoCol className="mb-6 sm:mb-8 max-w-4xl mx-auto">
        <Card className="text-left !border-[#10b981]/30">
          <CardTitle className="mb-3">
            <span className="text-[#10b981]">Coding</span> — Tolerance for Error
          </CardTitle>
          <ul className="space-y-3">
            <li className="flex items-start gap-2 text-xs sm:text-sm text-white/70">
              <span className="text-[#10b981] shrink-0 mt-0.5">→</span>
              <span>AI output is <strong className="text-white">non-deterministic</strong> — and that&apos;s OK</span>
            </li>
            <li className="flex items-start gap-2 text-xs sm:text-sm text-white/70">
              <span className="text-[#10b981] shrink-0 mt-0.5">→</span>
              <span>Code review catches mistakes before merge</span>
            </li>
            <li className="flex items-start gap-2 text-xs sm:text-sm text-white/70">
              <span className="text-[#10b981] shrink-0 mt-0.5">→</span>
              <span>Cost of bad code is <strong className="text-white">manageable</strong></span>
            </li>
          </ul>
        </Card>

        <Card className="text-left !border-red-500/30">
          <CardTitle className="mb-3">
            <span className="text-red-400">Infrastructure</span> — Zero Margin for Error
          </CardTitle>
          <ul className="space-y-3">
            <li className="flex items-start gap-2 text-xs sm:text-sm text-white/70">
              <span className="text-red-400 shrink-0 mt-0.5">→</span>
              <span>Non-deterministic AI is <strong className="text-white">dangerous</strong></span>
            </li>
            <li className="flex items-start gap-2 text-xs sm:text-sm text-white/70">
              <span className="text-red-400 shrink-0 mt-0.5">→</span>
              <span>Ops teams are anxious about AI touching prod infra</span>
            </li>
            <li className="flex items-start gap-2 text-xs sm:text-sm text-white/70">
              <span className="text-red-400 shrink-0 mt-0.5">→</span>
              <span>Cost of misconfiguration can be <strong className="text-white">catastrophic</strong></span>
            </li>
          </ul>
        </Card>
      </TwoCol>

      <Callout variant="highlight" className="max-w-3xl mx-auto">
        <p className="text-sm sm:text-base md:text-lg text-white font-semibold text-center">
          AI must be paired with deterministic tools to earn trust in infrastructure.
        </p>
        <p className="text-xs sm:text-sm text-white/50 text-center mt-1">
          AI should orchestrate. Tools should execute. The outcome must be predictable.
        </p>
      </Callout>
    </Slide>
  );
}
