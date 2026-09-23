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

const chapters = [
  {
    title: 'Built Inside Planton',
    description:
      'A year and a half of agent engineering — learning what it takes to make AI safe for production infrastructure — happened inside Planton first.',
  },
  {
    title: 'Spun Out to Focus',
    description:
      'In January 2026 that work became its own company. Planton stays focused on cloud and DevOps; Stigmer improves agent capabilities full-time.',
  },
  {
    title: 'Planton Is Customer Zero',
    description:
      'The hardest customer a platform can have: every Planton AI feature exercises Stigmer in production, and the feedback loop is public on GitHub.',
  },
];

export default function S10BornFromPlanton() {
  return (
    <Slide>
      <SlideHeader
        sectionTag="The Story"
        title="Born From Building Planton"
        subtitle="A year and a half of agent engineering inside Planton became its own company."
      />

      <Grid cols={3} gap="sm" className="max-w-4xl mx-auto mb-6">
        {chapters.map((chapter) => (
          <Card key={chapter.title} className="text-left">
            <CardTitle className="mb-1 !text-sm sm:!text-base">
              {chapter.title}
            </CardTitle>
            <CardText>{chapter.description}</CardText>
          </Card>
        ))}
      </Grid>

      <Callout variant="success" className="max-w-3xl mx-auto">
        <p className="text-center text-sm sm:text-base text-white">
          Two products, one thesis — AI capabilities should be adopted, not
          rebuilt.
        </p>
      </Callout>
    </Slide>
  );
}
