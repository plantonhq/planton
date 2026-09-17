'use client';

import {
  Slide,
  SlideHeader,
  DemoBadge,
  Checklist,
  Callout,
} from '@/components/deck';

export default function S08DemoHandoff() {
  return (
    <Slide>
      <SlideHeader title="Let's Open the Desktop App" />

      <div className="text-center mb-6">
        <DemoBadge />
      </div>

      <Callout variant="highlight" className="max-w-3xl mx-auto mb-6">
        <p className="text-sm sm:text-base text-white text-center">
          Planton ships on four surfaces — desktop, web, CLI and mobile. The
          desktop app is where the AI teammate lives today, and it is the one we
          put in front of engineers.
        </p>
      </Callout>

      <div className="max-w-2xl mx-auto mb-6">
        <p className="text-xs uppercase tracking-wider text-white/50 mb-3 text-center">
          What to watch for
        </p>
        <Checklist
          items={[
            'Infrastructure described in plain language, then actually created',
            'A typed catalog underneath — every field has a home, nothing is invented',
            'Real Pulumi and OpenTofu execution, with live logs',
          ]}
        />
      </div>

      <Callout variant="warning" className="max-w-3xl mx-auto">
        <p className="text-xs sm:text-sm text-white/70 text-center">
          This demo is Infra Hub. Service Hub runs on the hosted and self-hosted
          consoles — it reaches the desktop app in a coming release.
        </p>
      </Callout>
    </Slide>
  );
}
