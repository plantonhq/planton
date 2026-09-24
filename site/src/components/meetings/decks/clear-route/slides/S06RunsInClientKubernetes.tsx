'use client';

import {
  Slide,
  SlideHeader,
  Grid,
  Card,
  CardTitle,
  CardText,
  Badge,
} from '@/components/deck';

const properties = [
  {
    title: 'No Gatekeeping',
    description:
      'Charts are publicly pullable. No registry login, no account, no license key, no image pull secret.',
  },
  {
    title: 'Secret-Free Connections',
    description:
      'Identity comes ambiently from the runner pod via IRSA or GKE and AKS workload identity. We never copy or store client credentials.',
  },
  {
    title: 'The Operator Does The Rest',
    description:
      'Postgres, Redis, Temporal, control plane, console, identity, gateway and runner — with per-component health in plain language.',
  },
];

export default function S06RunsInClientKubernetes() {
  return (
    <Slide>
      <SlideHeader
        sectionTag="Deployment"
        title="It Runs Inside Your Client's Kubernetes"
        subtitle="One command, into their cluster, inside their security boundary"
      />

      <div className="max-w-3xl mx-auto mb-6">
        <pre className="bg-black/40 border border-white/10 rounded-xl p-4 sm:p-5 text-xs sm:text-sm text-white/80 overflow-x-auto">
          <code>
            {'helm install planton oci://ghcr.io/plantonhq/charts/planton \\\n'}
            {'  --namespace planton --create-namespace'}
          </code>
        </pre>
      </div>

      <Grid cols={3} gap="sm" className="max-w-4xl mx-auto mb-6">
        {properties.map((property) => (
          <Card key={property.title} className="text-left">
            <CardTitle className="mb-1 !text-sm sm:!text-base">
              {property.title}
            </CardTitle>
            <CardText>{property.description}</CardText>
          </Card>
        ))}
      </Grid>

      <div className="text-center">
        <Badge variant="warning">
          Self-hosted is in active preview — we want a design partner
        </Badge>
      </div>
    </Slide>
  );
}
