'use client';

import { Slide, SlideTitle, SlideSubtitle } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

export default function S15Close(_props: SlideComponentProps) {
  return (
    <Slide variant="cover">
      <div className="text-center">
        <div className="mb-6 sm:mb-8">
          <span className="text-4xl sm:text-5xl md:text-6xl font-bold text-white tracking-tight">
            Planton
          </span>
        </div>

        <SlideTitle className="mb-3 sm:mb-4 max-w-3xl mx-auto">
          The future of DevOps is not more engineers.
          <br />
          <span className="text-[#10b981]">It is better teammates.</span>
        </SlideTitle>

        <SlideSubtitle className="mb-8 sm:mb-12 max-w-xl mx-auto">
          Thank you, Nirav. We are excited about the possibilities.
        </SlideSubtitle>

        <div className="space-y-2 mb-8">
          <p className="text-sm sm:text-base text-white/70">Swarup Donepudi</p>
          <p className="text-xs sm:text-sm text-white/40">swarup@planton.ai</p>
          <p className="text-xs sm:text-sm text-white/40">planton.ai</p>
        </div>

        <div className="inline-flex items-center gap-2 px-4 py-2 bg-white/5 border border-white/10 rounded-full">
          <div className="w-2 h-2 rounded-full bg-[#10b981]" />
          <span className="text-xs sm:text-sm text-white/50">
            Press <kbd className="px-1.5 py-0.5 bg-white/10 rounded text-white/70 text-xs">N</kbd> for presenter notes
            &middot; <kbd className="px-1.5 py-0.5 bg-white/10 rounded text-white/70 text-xs">F</kbd> for fullscreen
          </span>
        </div>
      </div>
    </Slide>
  );
}
