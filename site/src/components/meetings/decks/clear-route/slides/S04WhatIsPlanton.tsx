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
        title="A Self-Service DevOps Platform, In Two Halves"
      />

      <TwoCol className="max-w-4xl mx-auto mb-6">
        <Card variant="purple" className="text-left">
          <CardTitle className="text-white mb-3">Infra Hub</CardTitle>
          <p className="text-sm text-white/70 mb-4">
            Production infrastructure from a form, not a Terraform pull request.
          </p>
          <ul className="space-y-2 text-sm text-white/60">
            <li>
              • <strong className="text-white">600+ components</strong> across 17
              clouds
            </li>
            <li>• Every component ships a Pulumi and an OpenTofu module</li>
            <li>
              • <strong className="text-white">Infra Charts</strong> are Helm
              charts for cloud infrastructure
            </li>
            <li>
              • 17 curated charts — EKS full stack, production cluster baseline
            </li>
          </ul>
        </Card>

        <Card variant="pink" className="text-left">
          <CardTitle className="text-white mb-3">Service Hub</CardTitle>
          <p className="text-sm text-white/70 mb-4">
            Git push to production, without writing the pipeline.
          </p>
          <ul className="space-y-2 text-sm text-white/60">
            <li>• Connect a repo — build and deploy are wired for you</li>
            <li>• No CI/CD YAML, no Dockerfile required</li>
            <li>• Live logs, approval gates, re-runs</li>
            <li>• Charts provision platforms; Service Hub deploys apps</li>
          </ul>
        </Card>
      </TwoCol>

      <Callout variant="highlight" className="max-w-3xl mx-auto">
        <p className="text-center text-sm sm:text-base text-white">
          <strong>Planton is not a cloud abstraction layer.</strong> The
          consistency is structural and procedural, never semantic — we do not
          pretend an AWS network and a GCP network are the same thing.
        </p>
      </Callout>
    </Slide>
  );
}
