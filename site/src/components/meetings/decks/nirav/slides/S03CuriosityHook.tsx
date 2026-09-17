'use client';

import { Slide, SlideHeader, Card, CardTitle, Callout } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

export default function S03CuriosityHook(_props: SlideComponentProps) {
  return (
    <Slide>
      <SlideHeader
        sectionTag="The Question"
        title="AI Writes Great Code. Why Won't Anyone Let It Operate Infrastructure?"
      />

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-6 mb-6 sm:mb-8 max-w-3xl mx-auto">
        <Card variant="success" className="text-left">
          <CardTitle className="mb-3 !text-base sm:!text-lg">AI Writing Code</CardTitle>
          <ul className="space-y-2">
            {[
              'Cursor, Copilot, Codex — massive adoption',
              'AI writes application code AND Terraform',
              'Code review provides a safety net before merge',
              'Entire companies built with AI coding agents',
            ].map((item) => (
              <li key={item} className="flex items-start gap-2 text-xs sm:text-sm text-white/70">
                <span className="text-[#10b981] shrink-0 mt-0.5">✓</span>
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </Card>

        <Card variant="danger" className="text-left">
          <CardTitle className="mb-3 !text-base sm:!text-lg">AI Operating Infrastructure</CardTitle>
          <ul className="space-y-2">
            {[
              'No central visibility or auditability',
              'No AI orchestrating deployments end-to-end',
              'No trust to let AI manage state and secrets',
              'Ops teams won\'t let AI touch production',
            ].map((item) => (
              <li key={item} className="flex items-start gap-2 text-xs sm:text-sm text-white/70">
                <span className="text-red-400 shrink-0 mt-0.5">✗</span>
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </Card>
      </div>

      <Callout variant="warning" className="max-w-2xl mx-auto">
        <p className="text-base sm:text-lg md:text-xl text-white font-semibold text-center mb-1">
          Why?
        </p>
        <p className="text-sm sm:text-base text-white/60 text-center">
          We spent 2 years and $500K discovering the answer.
        </p>
      </Callout>
    </Slide>
  );
}
