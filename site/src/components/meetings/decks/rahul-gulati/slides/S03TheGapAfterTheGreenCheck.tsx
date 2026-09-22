'use client';

import {
  Slide,
  SlideHeader,
  Grid,
  Card,
  CardTitle,
  CardText,
  Callout,
} from '@/components/deck';

const gaps = [
  {
    title: 'Infrastructure',
    description:
      'VPCs, databases, clusters, queues — provisioned by hand, or with Terraform someone has to write and maintain forever.',
  },
  {
    title: 'The Deploy',
    description:
      'CI produces a build, not a running service. The deploy step is custom glue code every team writes on its own.',
  },
  {
    title: 'The Unknowns',
    description:
      'What will this cost next month? What permissions does it need? Would it pass a security review? Usually answered after the fact.',
  },
];

export default function S03TheGapAfterTheGreenCheck() {
  return (
    <Slide variant="problem">
      <SlideHeader
        sectionTag="The Problem"
        title="What Happens After the Merge"
        subtitle="GitHub covers everything up to the merge. For small teams, everything after it is manual."
      />

      <Grid cols={3} gap="sm" className="max-w-4xl mx-auto mb-6">
        {gaps.map((gap) => (
          <Card key={gap.title} variant="danger" className="text-left">
            <CardTitle className="mb-1 !text-sm sm:!text-base">
              {gap.title}
            </CardTitle>
            <CardText>{gap.description}</CardText>
          </Card>
        ))}
      </Grid>

      <Callout variant="highlight" className="max-w-3xl mx-auto">
        <p className="text-center text-sm sm:text-base text-white">
          AI made writing code fast. Deployment is now the bottleneck.
        </p>
      </Callout>
    </Slide>
  );
}
