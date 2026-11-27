/** The hero's diagram: the team, Planton and your cloud, drawn as an illustration
 * and labelled so (`illustration`). Its words mirror homepage.ts and the story. */
export const HERO = {
  id: 'hero',
  version: 1,
  title: 'Your team. One path to your cloud.',
  description:
    // ch 9 (who it is for): one buyer, the developer who inherited ops, and their coding agent.
    'You and your coding agent push to GitHub. Planton deploys infrastructure and applications into your cloud and returns execution details.',
  team: ['Developers', 'Coding Agents'],
  compactTeam: ['Your', 'Developers', 'Coding Agents', 'In Your Editor'],
  engineLabel: 'Infrastructure + Delivery',
  inputDetails: ['Pushes and Approvals', 'In Your Editor', 'Source Changes'],
  repository: 'GitHub',
  controls: 'Access & Approvals',
  cloud: 'Your Cloud',
  workloads: 'Environments · Applications · Data',
  providers: 'AWS, GCP, Azure, Cloudflare, and DigitalOcean are supported choices.',
  /** The provider marks the diagram draws, in order; keys of WORKFLOW_ICONS. */
  providerMarks: ['aws', 'gcp', 'azure', 'cloudflare', 'digitalocean'],
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
