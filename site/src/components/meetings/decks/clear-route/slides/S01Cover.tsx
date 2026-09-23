'use client';

import { Slide } from '@/components/deck';

export default function S01Cover() {
  return (
    <Slide variant="cover">
      <div className="text-center max-w-3xl mx-auto">
        <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-extrabold mb-4">
          <span className="text-white">
            Planton
          </span>
          <span className="text-white"> for ClearRoute</span>
        </h1>

        <p className="text-xl sm:text-2xl md:text-3xl text-white/80 font-medium mb-3">
          AI-Native Self-Service DevOps
        </p>

        <p className="text-base sm:text-lg md:text-xl text-white/60 mb-12">
          A platform you deploy for your clients — not another tool they buy
        </p>

        <div className="space-y-2 text-sm sm:text-base text-white/50">
          <p>
            <span className="text-white/70 font-medium">Date:</span> August 12,
            2026
          </p>
          <p>
            <span className="text-white/70 font-medium">Location:</span> Virtual
          </p>
        </div>
      </div>
    </Slide>
  );
}
