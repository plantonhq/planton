'use client';

import { Slide, SlideHeader, Card, CardTitle, CardText, Grid, Callout } from '@/components/deck';
import type { SlideComponentProps } from '@/components/deck';

const problems = [
  {
    icon: '💰',
    title: 'Ops Expertise Is Scarce and Expensive',
    description: 'Senior DevOps talent is scarce and expensive. Most companies compete for the same small talent pool.',
  },
  {
    icon: '🔁',
    title: 'Same Patterns, Rebuilt Everywhere',
    description: '80% of cloud infrastructure is the same across companies — VPCs, databases, K8s clusters — rebuilt from scratch every time.',
  },
  {
    icon: '⚡',
    title: 'Multi-Cloud Chaos',
    description: 'AWS, GCP, Azure — each with different tools, different APIs, different expertise required.',
  },
  {
    icon: '⏱️',
    title: 'Weeks to Deploy',
    description: 'Infrastructure setup takes weeks, not minutes. Every new project starts with the same tedious bootstrap.',
  },
];

export default function S02Problem(_props: SlideComponentProps) {
  return (
    <Slide variant="problem">
      <SlideHeader
        sectionTag="The Problem"
        title="DevOps Talent Is Hard to Find and Expensive to Keep"
        subtitle="Every growing company faces the same struggle."
      />

      <Grid cols={2} gap="sm" className="mb-6 sm:mb-8">
        {problems.map((p) => (
          <Card key={p.title} className="text-left">
            <div className="text-2xl mb-2">{p.icon}</div>
            <CardTitle className="mb-1 !text-sm sm:!text-base">{p.title}</CardTitle>
            <CardText>{p.description}</CardText>
          </Card>
        ))}
      </Grid>

      <Callout variant="highlight" className="max-w-2xl mx-auto">
        <p className="text-sm sm:text-base md:text-lg text-white font-medium text-center">
          &ldquo;We set out 2 years ago to solve this with AI...&rdquo;
        </p>
      </Callout>
    </Slide>
  );
}
