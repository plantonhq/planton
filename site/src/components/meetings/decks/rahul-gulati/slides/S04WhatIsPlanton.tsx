'use client';

import {
  Slide,
  SlideHeader,
  TwoCol,
  Card,
  CardTitle,
  Callout,
} from '@/components/deck';

export default function S04WhatIsPlanton() {
  return (
    <Slide>
      <SlideHeader
        sectionTag="What Planton Is"
        title="The Self-Service Cloud Platform"
        subtitle="AI designs the infrastructure, verifies the cost and permissions before anything is created, and publishes it as templates your whole team can deploy — in your own cloud account."
      />

      <TwoCol className="max-w-4xl mx-auto mb-6">
        <Card variant="purple" className="text-left">
          <CardTitle className="text-white mb-1">Infra Hub</CardTitle>
          <p className="text-sm text-white/70 mb-4 font-medium">
            Cursor for Cloud Infrastructure
          </p>
          <ul className="space-y-2 text-sm text-white/60">
            <li>• Describe what you need — it composes on a live canvas</li>
            <li>
              • Cost and permissions <strong className="text-white">verified before deploy</strong>
            </li>
            <li>
              • <strong className="text-white">600+ typed components</strong>{' '}
              across 8 providers, open-source modules underneath
            </li>
            <li>
              • Publish as an Infra Chart — a template your team reuses
            </li>
          </ul>
        </Card>

        <Card variant="pink" className="text-left">
          <CardTitle className="text-white mb-1">Service Hub</CardTitle>
          <p className="text-sm text-white/70 mb-4 font-medium">
            Vercel for Backend, In Your Own Cloud
          </p>
          <ul className="space-y-2 text-sm text-white/60">
            <li>• Git push to a production URL</li>
            <li>• No pipeline YAML, no Dockerfile required</li>
            <li>• Results written back into GitHub checks and Deployments</li>
            <li>• Approval gates, live logs, full audit history</li>
          </ul>
        </Card>
      </TwoCol>

      <Callout variant="highlight" className="max-w-3xl mx-auto">
        <p className="text-center text-sm sm:text-base text-white">
          <strong>Runs in the customer&apos;s cloud account</strong> — their
          credentials, their state, their audit trail. Open-source modules
          underneath, so there is no lock-in to trust us about.
        </p>
      </Callout>
    </Slide>
  );
}
