'use client';

import {
  Slide,
  SlideHeader,
  Card,
  CardTitle,
  CardText,
  Grid,
  Callout,
} from '@/components/deck';

const steps = [
  {
    title: 'Be Our Design Partner',
    description:
      'Self-hosted Planton on one live ClearRoute engagement, with our team embedded. You shape what "enterprise-ready" means for it.',
  },
  {
    title: 'A Technical Deep Dive',
    description:
      'Your platform engineers against our architecture — the catalog model, the runner, secret handling, the MCP surface.',
  },
  {
    title: 'Then Talk Commercials',
    description:
      'Only once it has earned a place in your delivery method. We would rather prove it first.',
  },
];

export default function S09NextSteps() {
  return (
    <Slide variant="gradient">
      <SlideHeader
        sectionTag="Next Steps"
        title="What We'd Like To Ask"
      />

      <Grid cols={3} gap="sm" className="max-w-4xl mx-auto mb-6">
        {steps.map((step) => (
          <Card key={step.title} variant="highlight" className="text-left">
            <CardTitle className="mb-1 !text-sm sm:!text-base">
              {step.title}
            </CardTitle>
            <CardText>{step.description}</CardText>
          </Card>
        ))}
      </Grid>

      <Callout variant="success" className="max-w-3xl mx-auto">
        <p className="text-center text-sm sm:text-base text-white">
          Thank you. Questions — and please make them hard ones.
        </p>
      </Callout>
    </Slide>
  );
}
