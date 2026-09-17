'use client';

import { Slide, SlideTitle, SlideSubtitle, DemoBadge } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

export default function S07Demo(_props: SlideComponentProps) {
  return (
    <Slide variant="gradient">
      <div className="text-center">
        <DemoBadge className="mb-6 sm:mb-8" />

        <SlideTitle className="mb-3 sm:mb-4">
          Let Me Show You How It Works
        </SlideTitle>

        <SlideSubtitle className="mb-8 sm:mb-12 max-w-xl mx-auto">
          A walkthrough of Planton — from connecting a cloud provider
          to deploying infrastructure with AI teammates.
        </SlideSubtitle>

        <div className="inline-flex items-center gap-3 px-6 py-3 bg-white/5 border border-white/10 rounded-2xl">
          <div className="w-3 h-3 rounded-full bg-[#10b981] animate-pulse" />
          <span className="text-sm sm:text-base text-white/70">Switching to live product...</span>
        </div>
      </div>
    </Slide>
  );
}
