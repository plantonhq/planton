'use client';

import {
  Slide,
  SlideHeader,
  FlowDiagram,
  Grid,
  Card,
  CardTitle,
  CardText,
  Callout,
} from '@/components/deck';

const verifications = [
  {
    title: 'Cost, Before Deployment',
    description:
      'An itemized cloud bill for the architecture — each line citing the provider price document and the date it was verified.',
  },
  {
    title: 'Compliance Posture',
    description:
      'Controls mapped to HIPAA and CIS requirements, with evidence for what is enforced — stated plainly, never as a certification.',
  },
  {
    title: 'Least-Privilege Permissions',
    description:
      'A downloadable IAM policy scoped to exactly this architecture — ready for a security review.',
  },
];

export default function S06PavedRoadInOneSitting() {
  return (
    <Slide variant="solution">
      <SlideHeader
        sectionTag="The Self-Service Loop"
        title="Compose Once. Publish. Developers Self-Serve."
        subtitle="A platform engineer builds and verifies the architecture. Developers deploy it from a template."
      />

      <FlowDiagram
        className="mb-6"
        steps={[
          { label: 'Describe' },
          { label: 'Watch It Compose' },
          { label: 'Verify' },
          { label: 'Deploy' },
          { label: 'Publish as Template' },
        ]}
      />

      <Grid cols={3} gap="sm" className="max-w-4xl mx-auto mb-6">
        {verifications.map((item) => (
          <Card key={item.title} variant="success" className="text-left">
            <CardTitle className="mb-1 !text-sm sm:!text-base">
              {item.title}
            </CardTitle>
            <CardText>{item.description}</CardText>
          </Card>
        ))}
      </Grid>

      <Callout variant="highlight" className="max-w-3xl mx-auto">
        <p className="text-center text-sm sm:text-base text-white">
          Once published, any developer can deploy the same architecture
          through a form. No Terraform, no tickets.
        </p>
      </Callout>
    </Slide>
  );
}
