import { testimonial } from './testimonials.ts';
import type { ChapterId } from './story.ts';

/** Buyer-facing homepage, grounded in the wiki ledger in content/copywriting.
 * Sources are internal authoring metadata; components render only public copy.
 * Chapter links preserve the story vocabulary without repeating its old pitch.
 */
export const HOMEPAGE = {
  title: 'Cloud Infrastructure and Application Delivery',
  description:
    'Set up cloud infrastructure, deploy applications from Git, and give your team a repeatable way to ship. See how Planton works in your cloud.',
  eyebrow: 'THE SELF-SERVICE CLOUD PLATFORM',
  headline: ['Set up your cloud.', 'Let your team ship.'],
  intro:
    'Give developers and their coding agents a repeatable path from repository to running service. Planton brings infrastructure, application delivery, and your team’s controls together—in your cloud.',
  caption: 'AI-assisted infrastructure. Git-based delivery. Your team’s controls.',
  seeHow: 'See how it works',
  illustration: 'Illustrative workflow',
  overview: {
    eyebrow: '01 / THE WHOLE PICTURE',
    title: 'One path from infrastructure to application.',
    intro:
      'Platform engineers establish the foundation. Developers use it to get their applications running.',
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
    eyebrow: 'SELF-SERVICE DEVOPS / FROM YOUR EDITOR',
    title: 'Your coding agent can ship\nmore than code.',
    intro:
      'Give the agent you already use a way to create cloud infrastructure, deploy your service, and inspect what happened—with Planton carrying the configuration, access controls, and deployment records.',
    setup:
      'Use Planton skills with the CLI in Cursor, Claude Code, or Codex. Supported agents can also connect through MCP. Connect your account and cloud before deploying; your team’s access and approval rules still apply.',
    link: 'Connect your coding agent',
    href: '/docs/coding-agents',
    sources: [
      'product.service-hub.service-records-and-build-lane.md',
      'product.infra-hub.catalog-curation.md',
      'product.infra-hub.stack-job.md',
    ],
    chapter: 'what-planton-is' as ChapterId,
  },
  infrastructure: {
    eyebrow: '02 / INFRASTRUCTURE',
    title: 'Your stack. Your cloud.\nOne deployment workflow.',
    intro:
      'Connect managed services, Kubernetes workloads, and edge applications. A reusable infrastructure template—an Infra Chart—describes the resources and their dependencies. Planton deploys them in order.',
    paragraphs: [
      'Use the assistant to help compose the architecture, or work directly with its manifests. Review the resulting configuration before you deploy. The template stays readable and editable as your requirements change.',
      'Give development and production their own configuration without rebuilding the architecture from scratch. Planton resolves dependencies and deploys the resources in order, including the connections that let workloads reach a newly created cluster.',
    ],
    note: 'See your architecture as it deploys. Explore resource dependencies, follow live deployment status, and inspect individual resources when something needs attention.',
    sources: [
      'product.infra-hub.infra-chart.rendering-pipeline.md',
      'product.infra-hub.infra-chart.one-run-cluster-composition.md',
      'product.infra-hub.infra-chart.platform-catalog.md',
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
      'Follow builds and logs, promote an artifact through your environments, and pause for approval where required. Each delivery records what shipped. When something fails, open the run and ask the assistant for help with the actual configuration and repository context.',
    ],
    note: 'Know what shipped, where it landed, and what needs attention.',
    sources: [
      'product.service-hub.console-creating-a-service.md',
      'product.service-hub.service-records-and-build-lane.md',
      'product.service-hub.console-run-view.md',
      'product.service-hub.assistant-repository-access.md',
    ],
    chapter: 'services-ship-from-git' as ChapterId,
  },
  controls: {
    eyebrow: '04 / CONTROL & VISIBILITY',
    title: 'Give developers freedom.\nKeep changes accountable.',
    intro:
      'Self-service works when the boundaries are clear. Define which cloud components your organization can create, require approval for protected environments, and keep a record of the changes that run.',
    points: [
      {
        title: 'A catalog shaped by your team',
        text: 'Make approved component kinds available for new resources. The platform enforces creation restrictions at the API boundary, including requests from the CLI and agents.',
      },
      {
        title: 'Evidence with its limits visible',
        text: 'Review available cost, permission, and technical-control information. Proven, declared, and not-evaluable claims remain distinct; estimates are not your cloud bill.',
      },
      {
        title: 'A record you can return to',
        text: 'Inspect the captured configuration, execution outcome, and approval history. Later edits do not rewrite what an earlier deployment recorded.',
      },
    ],
    note: 'The assistant works within the access of the person asking. Repository changes go through a reviewed pull request.',
    sources: [
      'product.infra-hub.catalog-curation.md',
      'product.infra-hub.control-posture.md',
      'product.infra-hub.stack-job.md',
      'product.service-hub.assistant-repository-access.md',
    ],
    chapter: 'your-rules-hold' as ChapterId,
  },
  proof: {
    eyebrow: 'IN THEIR OWN WORDS',
    title: 'One customer. Both sides of the workflow.',
    quotes: [testimonial('Rakesh Kandhi'), testimonial('Balaji Borra')],
    chapter: 'proof-it-works' as ChapterId,
  },
  adoption: {
    eyebrow: '06 / YOUR CLOUD, YOUR CHOICE',
    title: 'Fits the cloud—\nand the way—you work.',
    intro:
      'You can begin with an existing cloud account. Supported resources can be imported into infrastructure state without recreating them. Check the supported kind and import path before bringing a resource under management.',
    paragraphs: [
      'Choose hosted Planton for a shared platform, self-host it on your Kubernetes cluster, or use Desktop for an individual workspace. Connection methods and capabilities vary by deployment; the underlying resource model stays consistent.',
      'The infrastructure modules are open source. Your manifests remain readable, and you can deploy them with the standalone open-source CLI. Your adoption path can start with one environment or service.',
    ],
    sources: [
      'product.infra-hub.cloud-resource-import.md',
      'product.infra-hub.planton.md',
      'product.desktop.feature-availability.md',
      'product.self-hosted.install-and-upgrade-lifecycle.md',
      'product.connect.connection-methods-by-deployment.md',
    ],
    chapter: 'runs-where-you-decide' as ChapterId,
  },
  faq: {
    eyebrow: 'A FEW THINGS YOU MAY BE WONDERING',
    title: 'Before we talk.',
    questions: [
      {
        question: 'What does Planton replace or connect?',
        answer:
          'Planton connects cloud infrastructure setup, reusable environments, and application delivery. It can provide the build-and-deploy path for a service or work with repository-authored configuration and supported external CI workflows. Your cloud accounts and source repositories stay part of the workflow.',
      },
      {
        question: 'Can we use infrastructure we already have?',
        answer:
          'Yes, for supported resource kinds and import paths. Planton can track existing infrastructure and import its IaC state without recreating it. Tracking a resource is different from deploying it; we can use the demo to examine the right adoption path for your stack.',
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
        question: 'What does the AI assistant do?',
        answer:
          'It helps compose infrastructure and investigate deployments using the context of the screen and the access of the person asking. It can read the relevant repository and propose changes through a pull request after review. Authorization and deployment approval requirements still apply.',
      },
      {
        question: 'What will we see in the demo?',
        answer:
          'A walkthrough of infrastructure setup, application delivery, and deployment controls, guided by your current stack and team’s needs. Share your details, then choose an available time in the calendar. A meeting is booked only after you confirm a slot.',
      },
    ],
  },
  close: {
    eyebrow: 'YOUR NEXT ENVIRONMENT STARTS HERE',
    title: 'See how Planton would\nwork for your team.',
    text: 'Bring your current stack. We’ll walk through infrastructure setup, application delivery, and the controls your team needs to ship with confidence.',
    note: 'A technical walkthrough, shaped around your team.',
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
