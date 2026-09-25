/** Shared editorial source for the film, native player, transcript, and analytics.
 * Bump the version when publishing different bytes; CDN objects are immutable. */
export const OVERVIEW_VIDEO = {
  id: 'homepage-overview',
  version: '2026-09-25-v1',
  title: 'From repository to running service.',
  label: 'See Planton in 60 seconds.',
  description: 'One service. A reusable foundation, a repeatable release, and a clear view of your infrastructure.',
  poster: '/_site/images/product/homepage-overview.webp',
  base: 'https://assets.planton.ai/videos/homepage/2026-09-25-v1',
  duration: 60,
  fps: 30,
} as const;

export const OVERVIEW_CHAPTERS = [
  {
    start: 0, end: 6, label: 'The Outcome',
    title: 'From repository to running service.',
    subtitle: 'Infrastructure, delivery, and your team’s controls. Together.',
    transcript: 'Take your service from a repository to a running application in your cloud. Planton brings infrastructure, application delivery, and your team’s controls together.',
  },
  {
    start: 6, end: 17, label: 'The Foundation',
    title: 'Create the foundation once.',
    subtitle: 'Turn the infrastructure your service needs into a reusable environment.',
    transcript: 'Platform engineers define a reusable infrastructure template. Planton creates the network, runtime, and database in dependency order. These arrows show deployment prerequisites, not application traffic.',
  },
  {
    start: 17, end: 29, label: 'The Delivery Path',
    title: 'Give your team a repeatable delivery path.',
    subtitle: 'Your code. Your coding agent, if you use one. One configured pipeline.',
    transcript: 'A developer, optionally assisted by a coding agent, prepares the application code. A matching GitHub change triggers Planton to build an artifact and deploy it to development with that environment’s configuration.',
  },
  {
    start: 29, end: 40, label: 'The Production Gate',
    title: 'Keep production under your control.',
    subtitle: 'Build once. Promote the same artifact after the required approval.',
    transcript: 'In this illustrative pipeline, production requires human approval. The production stage waits for a person, then deploys the same artifact using production configuration. Planton records the deployment.',
  },
  {
    start: 40, end: 53, label: 'The Real Product',
    title: 'See how it fits together.',
    subtitle: 'Dogfooding, in practice.',
    transcript: 'This is an actual architecture export of our self-hosted Planton instance on GKE, which we use to manage Planton SaaS. See the cloud foundation, the workloads inside Kubernetes, and the individual resources in one view. This is a product capture, not a live deployment.',
  },
  {
    start: 53, end: 60, label: 'Your Next Step',
    title: 'Your team ships.',
    subtitle: 'Your cloud stays yours.',
    transcript: 'Your team ships. Your cloud stays yours. Planton supports AWS, Google Cloud, Azure, Cloudflare, and DigitalOcean. See it with your stack. Book a demo at planton.ai.',
  },
] as const;
