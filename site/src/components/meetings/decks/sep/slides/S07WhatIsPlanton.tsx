'use client';

import { Slide, SlideHeader, Card, CardTitle, TwoCol, Callout } from '@/components/deck';

export default function S07WhatIsPlanton() {
  return (
    <Slide>
      <SlideHeader
        title="Planton: The Self-Service Cloud Platform"
        subtitle="AI designs the infrastructure, teams deploy it as templates — in your own cloud account."
      />

      <TwoCol className="max-w-4xl mx-auto mb-6">
        <Card variant="purple" className="text-left">
          <CardTitle className="text-white mb-3">
            1. Infrastructure Hub
          </CardTitle>
          <p className="text-sm text-white/70 mb-4">
            Deploy production-grade infrastructure by filling forms (not writing
            Terraform).
          </p>
          <ul className="space-y-2 text-sm text-white/60">
            <li>• VPCs, security groups, load balancers</li>
            <li>• ECS/EKS clusters, databases, DNS</li>
            <li>• Uses Terraform/Pulumi (open source, auditable)</li>
            <li>
              • <strong className="text-white">&lt;1 hour</strong> from form to
              production
            </li>
          </ul>
        </Card>

        <Card variant="pink" className="text-left">
          <CardTitle className="text-white mb-3">2. Service Hub</CardTitle>
          <p className="text-sm text-white/70 mb-4">
            Connect Git repos, get automatic CI/CD pipelines.
          </p>
          <ul className="space-y-2 text-sm text-white/60">
            <li>• Commit → Build → Push → Deploy → URL</li>
            <li>• No GitHub Actions YAML required</li>
            <li>• Works: ECS, K8s, Cloud Run, Azure</li>
            <li>• No Dockerfile required (BuildPacks)</li>
          </ul>
        </Card>
      </TwoCol>

      <Callout variant="success" className="max-w-2xl mx-auto">
        <p className="text-[#10b981] font-semibold text-sm sm:text-base text-center">
          Everything runs in <strong>YOUR</strong> cloud account. All
          infrastructure code is open source on GitHub.
        </p>
      </Callout>
    </Slide>
  );
}
