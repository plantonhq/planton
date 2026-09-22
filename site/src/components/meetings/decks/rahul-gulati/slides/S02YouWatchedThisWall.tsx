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

const observations = [
  {
    title: 'Building Was Covered',
    description:
      'Repos, reviews, CI, Copilot — the startups you onboarded could write and merge production-quality code from week one.',
  },
  {
    title: 'Deployment Was Not',
    description:
      'Cloud accounts, environments, credentials, databases, DNS, TLS. Getting to a running URL needed a platform engineer they could not afford.',
  },
  {
    title: 'Every Team Solved It Alone',
    description:
      'Each one wrote its own Terraform or hired early ops. The same problem, solved from scratch, one startup at a time.',
  },
];

export default function S02YouWatchedThisWall() {
  return (
    <Slide>
      <SlideHeader
        sectionTag="Where We Are Coming From"
        title="What Startup Programs Solved — and What They Couldn't"
        subtitle="Startups got world-class tools for building software. Deploying it stayed their problem."
      />

      <Grid cols={3} gap="sm" className="max-w-4xl mx-auto mb-6">
        {observations.map((item) => (
          <Card key={item.title} className="text-left">
            <CardTitle className="mb-1 !text-sm sm:!text-base">
              {item.title}
            </CardTitle>
            <CardText>{item.description}</CardText>
          </Card>
        ))}
      </Grid>

      <Callout variant="highlight" className="max-w-3xl mx-auto">
        <p className="text-center text-sm sm:text-base text-white">
          Passing CI is not the same as running in production. That gap is what
          Planton covers.
        </p>
      </Callout>
    </Slide>
  );
}
