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
          This is the desktop app — free forever, runs on any developer&apos;s
          laptop, deploys into their own cloud account.
        </p>
      </Callout>

      <div className="max-w-2xl mx-auto mb-6">
        <p className="text-xs uppercase tracking-wider text-white/50 mb-3 text-center">
          Three things to watch for
        </p>
        <Checklist
          items={[
            'The architecture drawing itself from one sentence',
            'The cloud bill and IAM policy, before anything exists',
            'Publishing it as a template any developer can deploy',
          ]}
        />
      </div>
    </Slide>
  );
}
