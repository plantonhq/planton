import { testimonial } from './testimonials.ts';
import type { ChapterId } from './story.ts';

/** Buyer-facing homepage, grounded in the wiki ledger in content/copywriting.
 * Sources are internal authoring metadata; components render only public copy.
 * Chapter links preserve the story vocabulary without repeating its old pitch.
 */
export const HOMEPAGE = {
  title: 'Cloud Infrastructure and Application Delivery',
  description:
    // ch 2 (what Planton is): the one sentence and its first action.
    'Planton deploys every push to GitHub into your own cloud, and shows the cost and permissions of every infrastructure change before it runs. Start by pasting one prompt into your coding agent.',
  headline: ['Set up your cloud.', 'Let your team ship.'],
  intro:
    'Give developers and their coding agents a repeatable path from repository to running service. Planton brings infrastructure, application delivery, and your team’s controls together—in your cloud.',
  caption: 'Your coding agent. Your cloud. A plan before every infrastructure change.',
  illustration: 'Illustrative workflow',
  overview: {
    eyebrow: '01 / THE WHOLE PICTURE',
    title: 'One path from infrastructure to application.',
    intro:
      'Your coding agent can do these three steps for you from one prompt. After that, every push deploys.',
    steps: [
      {
        title: 'Connect your cloud',
        label: 'YOUR FOUNDATION',
        text: 'Connect the cloud accounts your team already uses, with the access your deployment needs.',
      },
      {
        title: 'Create an environment',
        label: 'INFRA HUB',
        text: 'Define the network, runtime, and supporting services once. Reuse the architecture with each environment’s configuration.',
      },
      {
        title: 'Ship your application',
        label: 'SERVICE HUB',
        text: 'Connect your repository and configure its delivery path. Follow each build and deployment across your environments.',
      },
    ],
    sources: [
      'product.connect.connection-verification.md',
      'product.infra-hub.infra-chart.rendering-pipeline.md',
      'product.service-hub.console-creating-a-service.md',
    ],
    chapter: 'what-planton-is' as ChapterId,
  },
  agents: {
    eyebrow: 'FROM YOUR CODING AGENT',
    title: 'Your coding agent can ship\nmore than code.',
    intro:
      'Give the agent you already use a way to create cloud infrastructure, deploy your service, and inspect what happened—with Planton carrying the configuration, access controls, and deployment records.',
    setup:
      'On planton.ai, paste the Setup Prompt into Claude Code, Codex or Cursor from your app’s repository. The agent installs the CLI, connects your cloud and ships the first deployment, asking before anything costs money. Agents can also connect through MCP.',
    link: 'Connect your coding agent',
    href: '/docs/coding-agents',
    sources: [
      'product.service-hub.service-records-and-build-lane.md',
      'product.infra-hub.catalog-curation.md',
      'product.infra-hub.infra-job.md',
    ],
    chapter: 'what-planton-is' as ChapterId,
  },
  infrastructure: {
    eyebrow: '02 / INFRASTRUCTURE',
    title: 'Your stack. Your cloud.\nOne deployment workflow.',
    intro:
      'Connect managed services, Kubernetes workloads, and edge applications. A reusable infrastructure template—an Infra Chart—describes the resources and their dependencies. Planton deploys them in order.',
    paragraphs: [
      'Write the manifests yourself, or have your coding agent write them. Review the resulting configuration before you deploy. The template stays readable and editable as your requirements change.',
      'Give development and production their own configuration without rebuilding the architecture from scratch. Planton resolves dependencies and deploys the resources in order, including the connections that let workloads reach a newly created cluster.',
    ],
    note: 'See your architecture as it deploys. Explore resource dependencies, follow live deployment status, and inspect individual resources when something needs attention.',
    sources: [
      'product.infra-hub.infra-chart.rendering-pipeline.md',
      'product.infra-hub.infra-chart.one-run-cluster-composition.md',
    ],
    chapter: 'what-planton-is' as ChapterId,
  },
  delivery: {
    eyebrow: '03 / APPLICATION DELIVERY',
    title: 'From a Git push\nto a running service.',
    intro:
      'Give every Git push a repeatable path through build, development, and production—with approval where your team requires it.',
    paragraphs: [
      'Use a detected Dockerfile, a supported Buildpacks build, or a custom pipeline. Keep deployment configuration in your repository or configure it in Planton. Your team chooses the level of control it needs.',
      'Follow builds and logs, promote an artifact through your environments, and pause for approval where required. Each delivery records what shipped. When something fails, open the run and read its log and configuration.',
    ],
    note: 'Know what shipped, where it landed, and what needs attention.',
    sources: [
      'product.service-hub.console-creating-a-service.md',
      'product.service-hub.service-records-and-build-lane.md',
      'product.service-hub.console-run-view.md',
    ],
    chapter: 'services-ship-from-git' as ChapterId,
  },
  proof: {
    eyebrow: 'IN THEIR OWN WORDS',
    title: 'One customer. Both sides of the workflow.',
    quotes: [testimonial('Rakesh Kandhi'), testimonial('Balaji Borra')],
    chapter: 'proof-it-works' as ChapterId,
  },
  faq: {
    eyebrow: 'A FEW THINGS YOU MAY BE WONDERING',
    title: 'Before you start.',
    questions: [
      {
        question: 'What does Planton replace or connect?',
        answer:
          'Planton connects cloud infrastructure setup, reusable environments, and application delivery. It can provide the build-and-deploy path for a service or work with repository-authored configuration and supported external CI workflows. Your cloud accounts and source repositories stay part of the workflow.',
      },
      {
        question: 'Can we use infrastructure we already have?',
        answer:
          'Yes, for supported resource kinds and import paths. Planton can track existing infrastructure and import its IaC state without recreating it. Tracking a resource is different from deploying it; the docs list the supported kinds and import paths.',
      },
      {
        question: 'Where does Planton run?',
        answer:
          'Use hosted Planton, run the platform on your own Kubernetes cluster, or use a local Desktop instance. Desktop is an individual workspace. Shared deployments support team collaboration. Workload placement and credential options depend on your selected infrastructure and deployment setup.',
      },
      {
        question: 'How are cloud credentials handled?',
        answer:
          'Depending on the deployment, connections can use short-lived identity federation, credentials issued by Vault or OpenBAO, runner-side credentials, or managed secret references. Keyless federation requires an issuer the cloud can reach; it is not offered on a loopback-only Desktop instance.',
      },
      {
        question: 'How do I start?',
        answer:
          'Start free on planton.ai: free for up to three seats, no card. Paste the Setup Prompt into your coding agent from your app’s repository.',
      },
    ],
  },
  // ch 12 (start): hosted is the front door, and the first act is one prompt.
  close: {
    eyebrow: 'YOUR NEXT ENVIRONMENT STARTS HERE',
    title: 'Start with one\nrepository.',
    text: 'Start free on planton.ai: paste one prompt into your coding agent, and your first deployment lands in your own cloud.',
    note: 'Free for up to three seats. No card.',
  },
} as const;

/** Source-scoped booking copy; the form payload and existing calendar stay unchanged. */
export const DEMO_COPY = {
  eyebrow: 'LET’S MAKE IT CONCRETE',
  title: 'Your cloud.\nYour team.\nYour walkthrough.',
  intro:
    'See how Planton could connect infrastructure setup and application delivery for your team. Tell us a little about yourself, then choose a time to talk.',
  agendaTitle: 'What we’ll walk through',
  agenda: [
    'Create a reusable cloud environment.',
    'Take a service from repository to deployment.',
    'Review access, approvals, and the change record.',
  ],
  note: 'Bring your current stack and the parts of delivery you want to improve.',
  steps: ['Your details', 'Choose a time'],
  formTitle: 'Start with your details',
  submit: 'Continue to available times',
  schedulerTitle: 'Choose a time that works for you.',
  schedulerText:
    'Your details have been submitted. Select and confirm a slot below to book your meeting.',
  loading: 'Loading available times…',
  calendarFailure:
    'The calendar could not load here. Your details are saved; open the scheduling page to choose a time.',
  fallback: 'Having trouble with the calendar?',
  calendarLink: 'Open the scheduling page',
  calLink: 'swarup-donepudi/60min',
  calUrl: 'https://cal.com/swarup-donepudi/60min',
  confirmation: 'Your meeting is booked.',
  confirmationText:
    'Look for the calendar invitation in your email. We look forward to learning about your team.',
} as const;
