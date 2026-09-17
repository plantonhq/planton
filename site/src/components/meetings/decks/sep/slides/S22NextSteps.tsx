'use client';

import { Slide, SlideHeader, Card, CardTitle, TwoCol, Checklist, Grid } from '@/components/deck';

export default function S22NextSteps() {
  return (
    <Slide>
      <SlideHeader title="Let's Validate This Together" />

      <Card variant="success" className="max-w-3xl mx-auto text-left mb-6">
        <CardTitle className="text-[#10b981] mb-4">
          Proposed Pilot (2 Weeks)
        </CardTitle>

        <TwoCol>
          <div>
            <h4 className="text-white font-medium mb-3">Scope Options</h4>
            <ul className="space-y-2 text-sm text-white/60">
              <li>
                <strong className="text-white">Option A:</strong> RAD Labs
                incubation project
              </li>
              <li>
                <strong className="text-white">Option B:</strong> New client
                onboarding
              </li>
              <li>
                <strong className="text-white">Option C:</strong> Hybrid (one
                internal, one client)
              </li>
            </ul>
          </div>

          <div>
            <h4 className="text-white font-medium mb-3">Success Criteria</h4>
            <Checklist
              items={[
                'Environment provisioning <1 hour',
                'Audit trails meet compliance needs',
                'Integrates with existing tools',
                'Developers can self-serve',
              ]}
            />
          </div>
        </TwoCol>
      </Card>

      <Grid cols={2} className="max-w-3xl mx-auto text-left">
        <Card>
          <CardTitle className="text-white mb-3">Support</CardTitle>
          <ul className="space-y-2 text-sm text-white/60">
            <li>• Planton engineering team available</li>
            <li>• Custom module development if needed</li>
            <li>• Integration assistance</li>
          </ul>
        </Card>

        <Card>
          <CardTitle className="text-white mb-3">Follow-Up</CardTitle>
          <ul className="space-y-2 text-sm text-white/60">
            <li>• Technical deep-dive</li>
            <li>• Architecture review</li>
            <li>• Security/compliance Q&A</li>
          </ul>
        </Card>
      </Grid>
    </Slide>
  );
}
