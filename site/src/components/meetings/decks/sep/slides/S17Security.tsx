'use client';

import { Slide, SlideHeader, Card, CardTitle, TwoCol, Callout } from '@/components/deck';

export default function S17Security() {
  return (
    <Slide>
      <SlideHeader
        title="Runs in YOUR Cloud. You Control Security."
      />

      <TwoCol className="max-w-4xl mx-auto text-left mb-6">
        <Card variant="default" className="border-white/20">
          <CardTitle className="text-white mb-2">
            1. SaaS (Multi-Tenant)
          </CardTitle>
          <p className="text-[#10b981] text-sm font-medium mb-4">
            Fastest to Start
          </p>
          <ul className="space-y-2 text-sm text-white/60">
            <li>• Connect Planton to your AWS/Azure/GCP</li>
            <li>• Scoped IAM roles (no blanket access)</li>
            <li>
              • Infrastructure runs in <strong className="text-white">your cloud</strong>
            </li>
            <li>
              • State stored in <strong className="text-white">your backend</strong>
            </li>
            <li>• Terraform executes in isolated runners</li>
            <li>• All modules auditable (open source)</li>
          </ul>
        </Card>

        <Card variant="purple">
          <CardTitle className="text-white mb-2">
            2. Self-Hosted (Single-Tenant)
          </CardTitle>
          <p className="text-[#10b981] text-sm font-medium mb-4">
            Maximum Security
          </p>
          <ul className="space-y-2 text-sm text-white/60">
            <li>
              • Planton runs on{' '}
              <strong className="text-white">your Kubernetes cluster</strong>
            </li>
            <li>• You control IAM roles, we never touch credentials</li>
            <li>
              • All Terraform execution within{' '}
              <strong className="text-white">your AWS boundary</strong>
            </li>
            <li>• Control plane only receives status</li>
          </ul>
          <p className="text-[#a0a0a0] mt-4 text-xs">
            Best for: Healthcare (HIPAA), Aerospace (safety-critical), Fintech
            (regulated)
          </p>
        </Card>
      </TwoCol>

      <Callout variant="highlight" className="max-w-2xl mx-auto">
        <p className="text-white/70 text-sm text-center">
          Your VP of IT/Security/Compliance would appreciate the self-hosted
          option.
        </p>
      </Callout>
    </Slide>
  );
}
