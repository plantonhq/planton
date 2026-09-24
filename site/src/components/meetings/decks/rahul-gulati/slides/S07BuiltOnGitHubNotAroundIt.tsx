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

const points = [
  {
    title: 'Installed as a GitHub App',
    description:
      "One org-level install through GitHub's own consent screen. No PATs, no stored secrets — tokens are minted per operation and expire in an hour.",
  },
  {
    title: 'Results Land Back in GitHub',
    description:
      "Commit checks turn green, deployments appear in GitHub's own Environments view. Developers never leave GitHub to know what shipped.",
  },
  {
    title: 'Actions Stays If You Love It',
    description:
      'Teams keep Actions for CI and call our CLI from a workflow. We handle everything after the merge — the infrastructure and the deploy — not the CI.',
  },
];

export default function S07BuiltOnGitHubNotAroundIt() {
  return (
    <Slide>
      <SlideHeader
        sectionTag="The Ecosystem"
        title="Built On GitHub, Not Around It"
        subtitle="GitHub covers everything up to the merge. Planton handles what comes after: the infrastructure and the deployment."
      />

      <Grid cols={3} gap="sm" className="max-w-4xl mx-auto mb-6">
        {points.map((point) => (
          <Card key={point.title} className="text-left">
            <CardTitle className="mb-1 !text-sm sm:!text-base">
              {point.title}
            </CardTitle>
            <CardText>{point.description}</CardText>
          </Card>
        ))}
      </Grid>

      <Callout variant="highlight" className="max-w-3xl mx-auto">
        <p className="text-center text-sm sm:text-base text-white">
          Open standards underneath, end to end: CNCF build tooling, Kubernetes
          resource model, OpenFGA authorization — and every infrastructure
          module Apache-2.0 on GitHub.
        </p>
      </Callout>
    </Slide>
  );
}
