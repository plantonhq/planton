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

const premises = [
  {
    title: 'Engineering Is One System',
    description:
      'QCE spans Quality Engineering, Cloud Platforms and Developer Experience because delivery breaks at the seams between them, not inside them.',
  },
  {
    title: 'Route to Live',
    description:
      'Your name for everything between an idea and production. It is also the exact surface we set out to compress.',
  },
  {
    title: 'Regulated By Default',
    description:
      'Your clients are banks, insurers and national infrastructure. Anything you introduce has to survive their security review, not just yours.',
  },
];

export default function S02HowYouWork() {
  return (
    <Slide>
      <SlideHeader
        sectionTag="Where We Are Coming From"
        title="You Build Platforms For Other People"
        subtitle="That changes what a tool has to be, and it shaped what we chose to show you"
      />

      <Grid cols={3} gap="sm" className="max-w-4xl mx-auto mb-6">
        {premises.map((premise) => (
          <Card key={premise.title} className="text-left">
            <CardTitle className="mb-1 !text-sm sm:!text-base">
              {premise.title}
            </CardTitle>
            <CardText>{premise.description}</CardText>
          </Card>
        ))}
      </Grid>

      <Callout variant="highlight" className="max-w-3xl mx-auto">
        <p className="text-center text-sm sm:text-base text-white">
          So the question is not whether Planton is useful to ClearRoute. It is
          whether Planton is useful in your clients&apos; hands.
        </p>
      </Callout>
    </Slide>
  );
}
