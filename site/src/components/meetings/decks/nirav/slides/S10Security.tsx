'use client';

import { Slide, SlideHeader, Checklist, TwoCol, Card, CardTitle, Callout } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

export default function S10Security(_props: SlideComponentProps) {
  return (
    <Slide>
      <SlideHeader
        sectionTag="Security"
        title="Zero-Trust by Design"
        subtitle="Security is not bolted on — it is the architecture"
      />

      <TwoCol className="mb-6 sm:mb-8 max-w-4xl mx-auto">
        <Card className="text-left !border-[#10b981]/20">
          <CardTitle className="mb-3 !text-base sm:!text-lg">Data Protection</CardTitle>
          <Checklist
            items={[
              'All sensitive data encrypted at rest AND on the wire',
              'Bring Your Own Vault — HashiCorp Vault, AWS Secrets Manager, or ours',
              'Secrets stored as references, never as plaintext in the database',
              'Full audit trail on every secret access',
            ]}
          />
        </Card>

        <Card className="text-left !border-[#10b981]/20">
          <CardTitle className="mb-3 !text-base sm:!text-lg">Execution Isolation</CardTitle>
          <Checklist
            items={[
              'Just-in-Time Secret Resolution — references become values only at execution time',
              'Deploy-time secrets are resolved only in the runner, inside your network',
              'Secret values are never stored in the control plane',
              'Runner operates with outbound-only connectivity',
            ]}
          />
        </Card>
      </TwoCol>

      <Callout variant="success" className="max-w-3xl mx-auto">
        <div className="text-center">
          <p className="text-sm sm:text-base md:text-lg text-white font-semibold mb-2">
            The control plane orchestrates. The runner executes. Secrets stay in your network.
          </p>
          <p className="text-xs sm:text-sm text-white/50">
            Enterprise customers can deploy runners in completely private networks with no inbound connections.
            The same security model applies whether you use SaaS, hybrid, or fully self-hosted.
          </p>
        </div>
      </Callout>
    </Slide>
  );
}
