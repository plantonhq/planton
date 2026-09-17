'use client';

import { Slide, SlideHeader, Card, CardTitle, Grid } from '@/components/deck';

export default function S02WeKnowSEP() {
  return (
    <Slide>
      <SlideHeader title="What We Know About SEP" />

      <Grid cols={3} className="text-left">
        <Card>
          <CardTitle className="text-white mb-3">About SEP</CardTitle>
          <ul className="space-y-2 text-sm text-white/70">
            <li>• 300+ employees, 100% employee-owned</li>
            <li>• Westfield, Indiana (home base since 1988)</li>
            <li>• Fortune 100 to scale-up clients</li>
            <li>• Specializing in &quot;high cost of failure&quot; work</li>
          </ul>
        </Card>

        <Card>
          <CardTitle className="text-white mb-3">Industries</CardTitle>
          <ul className="space-y-2 text-sm text-white/70">
            <li>• Aerospace & Defense (safety-critical)</li>
            <li>• Healthcare & Life Sciences (HIPAA)</li>
            <li>• Fintech & Financial Technology (regulated)</li>
            <li>• Automotive & Heavy Machinery</li>
            <li>• Consumer & Industrial IoT</li>
          </ul>
        </Card>

        <Card>
          <CardTitle className="text-white mb-3">Core Values</CardTitle>
          <ul className="space-y-2 text-sm text-white/70">
            <li>
              • <strong className="text-white">&quot;Do it right&quot;</strong>
            </li>
            <li>• &quot;Talent unlocked&quot;</li>
            <li>• &quot;Excitement to share&quot;</li>
          </ul>
        </Card>
      </Grid>
    </Slide>
  );
}
