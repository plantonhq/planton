'use client';

import {
  Slide,
  SlideHeader,
  QuoteBox,
  Grid,
  Card,
  CardTitle,
  CardText,
} from '@/components/deck';

const frictions = [
  {
    title: 'Waiting on a Ticket',
    description:
      'A developer needs a database. They file a request and wait days for someone on the platform team to have capacity.',
  },
  {
    title: 'Waiting on a Review',
    description:
      'The Terraform exists, but it sits in a pull request behind the one engineer who understands the module.',
  },
  {
    title: 'Rebuilding Every Time',
    description:
      'Every new client engagement re-derives the same scaffolding from scratch. The learning does not compound.',
  },
];

export default function S03BottleneckIsRouteToLive() {
  return (
    <Slide variant="problem">
      <SlideHeader
        sectionTag="The Problem"
        title="The Bottleneck Is Not Code Generation"
      />

      <QuoteBox
        attribution="ClearRoute — clearroute.io"
        className="max-w-3xl mx-auto mb-6 sm:mb-8"
      >
        Developers are using AI in ~60% of daily work but process bottlenecks
        mean they&apos;re doing little more than generating backlog faster.
      </QuoteBox>

      <Grid cols={3} gap="sm" className="max-w-4xl mx-auto">
        {frictions.map((friction) => (
          <Card key={friction.title} variant="danger" className="text-left">
            <CardTitle className="mb-1 !text-sm sm:!text-base">
              {friction.title}
            </CardTitle>
            <CardText>{friction.description}</CardText>
          </Card>
        ))}
      </Grid>
    </Slide>
  );
}
