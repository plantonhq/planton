'use client';

import { Slide, SlideTitle, SlideSubtitle, Badge, Metric } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

export default function S01Cover(_props: SlideComponentProps) {
  return (
    <Slide variant="cover">
      <div className="text-center">
        <div className="mb-4 sm:mb-6">
          <span className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold text-white tracking-tight">
            Planton
          </span>
        </div>

        <SlideTitle className="mb-2 sm:mb-3">
          Autonomous AI Teammates
        </SlideTitle>
        <SlideSubtitle className="!mt-0 mb-8 sm:mb-10">
          for Cloud Infrastructure
        </SlideSubtitle>

        <div className="flex flex-wrap justify-center gap-3 sm:gap-6 mb-8 sm:mb-10">
          <Metric value="3+" label="Years Building" />
          <Metric value="$500K+" label="Self-Funded" />
          <Metric value="370+" label="Cloud Resources" />
          <Metric value="0%" label="Churn" />
        </div>

        <Badge variant="default" className="mb-4">
          Seed Stage &middot; $7M Valuation Cap &middot; SAFE
        </Badge>

        <p className="text-xs sm:text-sm text-white/40 mt-4">
          May 2026
        </p>
      </div>
    </Slide>
  );
}
