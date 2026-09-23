'use client';

import {
  Slide,
  SlideHeader,
  Comparison,
  Callout,
} from '@/components/deck';

export default function S05NotAnotherCodingAgent() {
  return (
    <Slide>
      <SlideHeader
        sectionTag="The Positioning"
        title="AI Makes Experts Faster. Planton Makes Teams Self-Serve."
        subtitle="A coding agent speeds up the person using it. Planton turns that person's work into something the whole team can reuse."
      />

      <Comparison
        className="mb-8"
        before={{
          label: 'Coding Agent',
          value: 'Code',
          subtext: 'Terraform one expert understands — each deployment starts from scratch',
        }}
        after={{
          label: 'Planton',
          value: 'Capability',
          subtext: 'A governed template — repeat deployments are a form fill',
        }}
      />

      <div className="max-w-3xl mx-auto mb-6 space-y-2 text-sm sm:text-base text-white/60 text-center">
        <p>
          Cost, compliance posture, and least-privilege permissions are{' '}
          <strong className="text-white">verified data, not model memory</strong>{' '}
          — the bill and the IAM policy exist before the infrastructure does.
        </p>
        <p>
          The AI never writes infrastructure code. It composes from a typed
          catalog; deterministic modules execute.
        </p>
      </div>

      <Callout variant="success" className="max-w-3xl mx-auto">
        <p className="text-center text-sm sm:text-base text-white">
          The test of Self-Service: the second developer needs only the
          template, not the AI.
        </p>
      </Callout>
    </Slide>
  );
}
