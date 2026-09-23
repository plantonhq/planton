'use client';

import { Slide, SlideHeader, Card, CardTitle, Callout, Grid } from '@/components/deck';

export default function S11DemoIntro() {
  return (
    <Slide>
      <SlideHeader title="Let's See It In Action" />

      <Callout variant="success" className="max-w-3xl mx-auto mb-8">
        <p className="text-white text-sm sm:text-base">
          <strong>Goal:</strong> Production-ready environment + deployed service
          in <strong className="text-[#10b981]">&lt;30 minutes</strong>
        </p>
      </Callout>

      <Grid cols={3} className="max-w-4xl mx-auto mb-8">
        <Card variant="purple">
          <CardTitle className="text-white mb-2">
            Part 1: Infrastructure Hub
          </CardTitle>
          <p className="text-white font-semibold mb-2">~20 minutes</p>
          <p className="text-sm text-white/60">
            Deploy AWS ECS environment (VPC, load balancer, ECS cluster, RDS,
            DNS)
          </p>
        </Card>

        <Card variant="pink">
          <CardTitle className="text-white mb-2">
            Part 2: Service Hub
          </CardTitle>
          <p className="text-white font-semibold mb-2">~5 minutes</p>
          <p className="text-sm text-white/60">
            Connect Git repo, deploy Node.js API, get production URL
          </p>
        </Card>

        <Card variant="success">
          <CardTitle className="text-[#10b981] mb-2">
            Part 3: Governance
          </CardTitle>
          <p className="text-white font-semibold mb-2">~5 minutes</p>
          <p className="text-sm text-white/60">
            Show audit trails, change history, compliance tracking
          </p>
        </Card>
      </Grid>

      <p className="text-center text-white/70 text-sm sm:text-base">
        This is what turns &quot;30 days to 30 minutes&quot; from one-time achievement
        into{' '}
        <strong className="text-white">repeatable process</strong>.
      </p>
    </Slide>
  );
}
