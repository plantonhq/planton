'use client';

import { Slide, SlideHeader, IconList, Callout } from '@/components/deck';

const fits = [
  {
    icon: '🚀',
    title: 'Day One Stops Being Expensive',
    description:
      'Stop re-deriving the same scaffolding on every engagement. Arrive with the platform already built.',
  },
  {
    icon: '📐',
    title: 'QCE Becomes Repeatable',
    description:
      'One delivery method across every client, instead of a bespoke stack per account that only its builders understand.',
  },
  {
    icon: '🏢',
    title: 'It Runs In Their Environment',
    description:
      'Deployed into the client\u2019s own Kubernetes. Their cloud, their data, their security boundary.',
  },
  {
    icon: '🎁',
    title: 'You Leave An Asset, Not A Dependency',
    description:
      'The client keeps a working platform when the engagement ends. That is a stronger renewal story than a handover document.',
  },
  {
    icon: '📈',
    title: 'Best Practice Becomes Enforceable',
    description:
      'Their platform team encodes the standard once, and every product team gets it by default. Guardrails in the tooling, not in a wiki page nobody reads.',
  },
];

export default function S05WhyFitsClearRoute() {
  return (
    <Slide variant="solution">
      <SlideHeader
        sectionTag="Why This Fits You"
        title="A Toolset For Your Business, Not Just Your Engineers"
      />

      <IconList items={fits} className="max-w-3xl mx-auto mb-6" />

      <Callout variant="success" className="max-w-3xl mx-auto">
        <p className="text-center text-sm sm:text-base text-white">
          The pitch to your client is not &quot;buy Planton.&quot; It is
          &quot;ClearRoute will stand up self-service DevOps in your cloud.&quot;
        </p>
      </Callout>
    </Slide>
  );
}
