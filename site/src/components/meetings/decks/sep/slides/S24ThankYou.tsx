'use client';

import { Slide } from '@/components/deck';

export default function S24ThankYou() {
  return (
    <Slide variant="cover">
      <div className="text-center max-w-2xl mx-auto">
        {/* Main title */}
        <h1 className="text-4xl sm:text-5xl md:text-6xl font-extrabold mb-4 text-white">
          Thank You
        </h1>

        {/* Subtitle */}
        <p className="text-xl sm:text-2xl text-white/80 mb-12">
          Looking Forward to Working with SEP
        </p>

        {/* Links */}
        <div className="space-y-2 text-sm text-white/50">
          <p>✉️ swarup@planton.ai</p>
          <p>🌐 planton.ai</p>
          <p>📚 planton.ai/docs</p>
          <p>💻 github.com/plantonhq/planton</p>
        </div>
      </div>
    </Slide>
  );
}
