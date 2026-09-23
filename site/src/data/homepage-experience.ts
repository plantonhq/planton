/** Homepage experience: product claims mirror homepage.ts and its source ledger.
 * The real screenshot is a supplied product capture, not simulated live state. */
export const HERO = {
  id: 'hero',
  version: 1,
  title: 'Your team. One path to your cloud.',
  description:
    'Platform engineers define reusable environments and controls. Developers work in Planton or through their coding agent. GitHub supplies source changes. Planton deploys infrastructure and applications into your cloud and returns execution details.',
  team: ['Platform Engineers', 'Developers'],
  compactTeam: ['Platform', 'Engineers', 'Developers', 'Editor + Agent'],
  engineLabel: 'Infrastructure + Delivery',
  inputDetails: ['Reusable Environments', 'Editor + Coding Agent', 'Source Changes'],
  repository: 'GitHub',
  controls: 'Access & Approvals',
  cloud: 'Your Cloud',
  workloads: 'Environments · Applications · Data',
  providers: 'AWS, GCP, Azure, Cloudflare, and DigitalOcean are supported choices.',
  illustration: 'Illustrative platform workflow',
  complete: 'Your team ships. Your cloud stays yours.',
  phases: [
    { title: 'Establish the foundation.', text: 'Reusable environments and your team’s controls.' },
    { title: 'Make a change.', text: 'Work from Planton, your editor, or your repository.' },
    {
      title: 'Run through Planton.',
      text: 'Infrastructure and delivery follow the configured controls.',
    },
    {
      title: 'Bring the result back.',
      text: 'See what ran, where it landed, and what needs attention.',
    },
  ],
} as const;
export const PRODUCT_PROOF = {
  eyebrow: 'INSIDE PLANTON',
  title: 'See how your infrastructure\nfits together.',
  intro:
    'Explore resource relationships, follow deployment status, and inspect the resources that need attention—all in the context of your architecture.',
  frame: 'Architecture View',
  capture: 'Actual Planton export · Self-Hosted on GKE',
  image: '/_site/images/product/architecture-view.png',
  alt: 'Actual architecture of our self-hosted Planton instance: VPC, subnet, GKE cluster, Planton Platform, Planton Operator, gateway, certificates, DNS, and node pool.',
  captionLead: 'Dogfooding, in practice.',
  caption:
    'This is our self-hosted Planton instance on GKE, which we use to manage Planton SaaS. Its cloud foundation and Kubernetes resources appear in one architecture view.',
  notes: [
    {
      title: 'Understand Resource Boundaries',
      text: 'Explore the relationships behind your stack. In this capture, the network contains a subnet and GKE cluster, with workloads grouped by namespace.',
      detail: 'Visible here: network, subnet, cluster, and namespace boundaries.',
      focus: 'groups',
    },
    {
      title: 'See What Runs Inside',
      text: 'Planton Platform and the Planton Operator sit alongside the gateway, certificates, and DNS resources that support this instance.',
      detail: 'Visible here: Planton itself, running inside its GKE cluster.',
      focus: 'whole',
    },
    {
      title: 'Inspect Individual Resources',
      text: 'Recognize the components in your architecture by kind and name. Here, the node pool, gateway, and Planton instance show their names and configuration details.',
      detail: 'Visible here: resource kinds, names, and configuration summaries.',
      focus: 'resources',
    },
  ],
} as const;
export const CONTROL_COPY = {
  eyebrow: 'CONTROL & OWNERSHIP',
  title: 'Self-service for your team.\nControl where it matters.',
  intro:
    'Define what your team can create, require approval where needed, and keep the record of every deployment.',
  label: 'Example Protected-Environment Workflow',
  steps: [
    {
      title: 'Access',
      value: 'Within Your Permissions',
      text: 'Requests use the access of the person asking, including requests from coding agents.',
    },
    {
      title: 'Required Approval',
      value: 'Human Decision',
      text: 'Protected environments wait for the approvals your team configures.',
    },
    {
      title: 'Deployment Record',
      value: 'Configuration & Outcome',
      text: 'Return to what ran, its approval history, and the execution result.',
    },
  ],
  placement: 'Choose Where Planton Runs',
  choices: [
    { title: 'Hosted', text: 'A shared platform for your team.' },
    { title: 'Self-Hosted', text: 'Run Planton on your Kubernetes cluster.' },
    { title: 'Desktop', text: 'An individual local workspace.' },
  ],
  boundary: 'Your Cloud',
  boundaryText: 'Your environments, applications, and data.',
  note: 'Platform placement and workload placement are separate choices. Connection methods and capabilities vary by deployment.',
  adoption:
    'Start with one environment or service. Supported resource kinds can be imported without recreating them; your infrastructure manifests remain readable and editable.',
} as const;
