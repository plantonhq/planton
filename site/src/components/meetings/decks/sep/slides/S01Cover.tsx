'use client';

import { Slide } from '@/components/deck';

export default function S01Cover() {
  return (
    <Slide variant="cover">
      <div className="text-center max-w-3xl mx-auto">
        {/* Main title */}
        <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-extrabold mb-4">
          <span className="text-white">
            Planton
          </span>
        </h1>

        {/* Subtitle */}
        <p className="text-xl sm:text-2xl md:text-3xl text-white/80 font-medium mb-3">
          From Days to Minutes, Every Time
        </p>

        {/* Tagline */}
        <p className="text-base sm:text-lg md:text-xl text-white/60 mb-12">
          How to Make Your DevOps Success Repeatable
        </p>

        {/* Meeting details */}
        <div className="space-y-2 text-sm sm:text-base text-white/50">
          <p>
            <span className="text-white/70 font-medium">Date:</span> January 23,
            2026
          </p>
          <p>
            <span className="text-white/70 font-medium">Location:</span>{' '}
            Westfield, Indiana
          </p>
        </div>
      </div>
    </Slide>
  );
}
