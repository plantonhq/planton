'use client';

import {
  Slide,
  SlideHeader,
  Callout,
  FlowDiagram,
  Comparison,
} from '@/components/deck';

export default function S07AINativeDeterministic() {
  return (
    <Slide>
      <SlideHeader
        sectionTag="AI-Native"
        title="The AI Never Writes Infrastructure Code"
        subtitle="Code review catches a bad function. Nothing catches a misconfigured production VPC."
      />

      <Comparison
        className="mb-8"
        before={{
          label: 'Generated IaC',
          value: 'Plausible',
          subtext: 'Reviewed by hope',
        }}
        after={{
          label: 'Planton',
          value: 'Typed',
          subtext: 'Validated before it runs',
        }}
      />

      <FlowDiagram
        className="mb-6"
        steps={[
          { label: 'Natural language' },
          { label: 'Select from typed catalog' },
          { label: 'Deterministic module runs' },
        ]}
      />

      <Callout variant="success" className="max-w-3xl mx-auto">
        <p className="text-center text-sm sm:text-base text-white">
          The AI chooses and configures. Pulumi and OpenTofu execute. There is no
          hallucinated Terraform anywhere in the path.
        </p>
      </Callout>
    </Slide>
  );
}
