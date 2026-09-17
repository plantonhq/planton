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

const state = [
  {
    title: 'The Product Is Built',
    description:
      'In production since 2023, zero churn, zero security incidents. Self-service pricing shipped: free desktop, $20/seat hosted, self-hosted licenses, published enterprise rate card.',
  },
  {
    title: 'Every Customer Came Inbound',
    description:
      'Consulting firms and their clients, all from in-person conversations. No outbound, no partner program, no startup program yet.',
  },
  {
    title: 'The Launch Plan Is Written',
    description:
      'Eight weeks of audience warm-up, then Product Hunt, Show HN, and the communities. Sequenced and resourced on paper — not yet executed.',
  },
];

export default function S11WhereWeAre() {
  return (
    <Slide variant="gradient">
      <SlideHeader
        sectionTag="Where We Are"
        title="Where We Are Today"
      />

      <Grid cols={3} gap="sm" className="max-w-4xl mx-auto mb-6">
        {state.map((item) => (
          <Card key={item.title} variant="highlight" className="text-left">
            <CardTitle className="mb-1 !text-sm sm:!text-base">
              {item.title}
            </CardTitle>
            <CardText>{item.description}</CardText>
          </Card>
        ))}
      </Grid>

      <Callout variant="success" className="max-w-3xl mx-auto">
        <p className="text-center text-sm sm:text-base text-white">
          The product and pricing are built. Distribution is what&apos;s
          missing — and it&apos;s the conversation I&apos;d like to have with
          you.
        </p>
      </Callout>
    </Slide>
  );
}
