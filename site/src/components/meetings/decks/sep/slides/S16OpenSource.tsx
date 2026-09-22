'use client';

import { Slide, SlideHeader, Card, CardTitle, Callout, Grid } from '@/components/deck';

export default function S16OpenSource() {
  return (
    <Slide>
      <SlideHeader
        title="No Black Boxes. No Lock-In."
      />

      <Callout variant="highlight" className="max-w-2xl mx-auto mb-8">
        <h3 className="text-white font-semibold text-center mb-2">
          Planton open source: All Infrastructure Modules
        </h3>
        <p className="text-white/60 text-center text-sm">
          github.com/plantonhq/planton
        </p>
      </Callout>

      <Grid cols={3} className="max-w-4xl mx-auto text-left mb-6">
        <Card>
          <CardTitle className="text-white mb-3">Transparency</CardTitle>
          <p className="text-sm text-white/60">
            Audit every line of Terraform we run. No proprietary black boxes. No
            hidden configurations.
          </p>
        </Card>

        <Card>
          <CardTitle className="text-white mb-3">Customization</CardTitle>
          <p className="text-sm text-white/60">
            Fork any module. Customize for SEP-specific requirements. Register
            your custom modules. Platform executes{' '}
            <strong className="text-white">your</strong> modules.
          </p>
        </Card>

        <Card>
          <CardTitle className="text-white mb-3">
            Exit Strategy (BOT Model)
          </CardTitle>
          <p className="text-sm text-white/60 mb-3">
            Build, Operate, Transfer. Export all configs as YAML. Continue with
            open source CLI. Transition to GitHub Actions.
          </p>
          <p className="text-[#10b981] text-sm font-medium">
            No vendor lock-in. You take the code and go.
          </p>
        </Card>
      </Grid>

      <p className="text-center text-white/60 text-sm max-w-2xl mx-auto">
        We&apos;re confident enough in our value-add to make all infrastructure code
        public.
        <br />
        <span className="text-white">
          Compare to Terraform Enterprise or Pulumi Cloud (proprietary,
          closed-source).
        </span>
      </p>
    </Slide>
  );
}
