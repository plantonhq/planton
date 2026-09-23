/**
 * Planton Desktop's two pages as data: the landing at /desktop and the
 * download page under it. One record for one subject; the templates in
 * components/desktop/ render it and state nothing of their own.
 *
 * Every sentence here names what backs it: a chapter of the story, the
 * desktop line and its points in ./positioning.ts, the platforms and
 * commands in ./desktop-download.ts, or a page of the public documentation
 * (the CI/CD-on-your-laptop guide, the coding-agents guide, the connections
 * guide). The measured figures carry their provenance in the record and the
 * page prints it beside them; a number never appears without it.
 *
 * What this file never carries: a platform, an installer, or a command (those
 * are ./desktop-download.ts, the one source for what can be downloaded); a
 * price (./pricing.ts); a count (./platform-stats.ts); a vendor's name; a
 * promise about a platform whose installer is off.
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators without a bundler.
 */
import { AGENT_SKILLS_INSTALL_COMMAND, DESKTOP_BREW_UPGRADE_COMMAND, type DesktopPlatform } from './desktop-download.ts';
import type { DoorPair } from './doors.ts';
import { DOORS } from './doors.ts';
import { PLATFORM_COUNTS, PLATFORM_STATS } from './platform-stats.ts';
import { POSITIONING } from './positioning.ts';
import { COMMUNITY_SEAT_LIMIT, FREE_TIER_SEATS } from './pricing.ts';
import { HERO_HEADLINE, chapter } from './story.ts';

const runs = chapter('runs-where-you-decide');
/** A short Title Case title over one or two sentences. */
export interface Tile {
  title: string;
  body: string;
}

/** One measured figure and what it measures. */
export interface Measurement {
  value: string;
  label: string;
}

/** A tab a person picks: what it is for, what they paste, and what happens. */
export interface CommandTab {
  /** Title Case; the tab's name. */
  label: string;
  /** One command per line. */
  commands: readonly string[];
  /** Sentence case; what the commands do and what to expect. */
  description: string;
}

export interface DesktopLanding {
  /** Title Case sentences; the desktop line up to its dash, derived in the story data. */
  headline: string;
  /** Sentence case; the sentence under the headline. */
  lede: string;
  doors: DoorPair;
  twoWays: {
    title: string;
    ways: readonly { title: string; trade: string; body: string }[];
    /** Sentence case; the turn after the two ways. */
    turn: string;
    /** Sentence case; the concession, from the positioning file. */
    concession: string;
  };
  ask: {
    title: string;
    lede: string;
    agent: { title: string; body: string; command: string };
    assistant: { title: string; body: readonly string[] };
    /** What Planton adds to the agent, as its own beat under the two cards. */
    addsTitle: string;
    adds: readonly string[];
    /** Sentence case; the line that closes the beat. */
    close: string;
  };
  runs: { title: string; lede: string; tiles: readonly Tile[] };
  infraHub: { kicker: string; title: string; lede: string; points: readonly (Tile & { door?: { label: string; href: string } })[] };
  serviceHub: {
    kicker: string;
    title: string;
    lede: string;
    measured: { figures: readonly Measurement[]; provenance: string };
    points: readonly Tile[];
    /** Sentence case; the one Docker sentence. */
    dockerNote: string;
  };
  everywhere: { title: string; body: string; exit: string; umbrella: string };
  whyFree: { title: string; body: readonly string[]; door: { label: string; href: string } };
  /** Registered pages to read next. */
  next: readonly string[];
  close: { title: string; body: string; guide: { label: string; href: string } };
}

export interface DesktopDownload {
  /** Sentence case; the sentence under the title. */
  lede: string;
  /** The card for a platform whose installer is off, with the platform's name filled in. */
  unavailable: (platform: DesktopPlatform) => { title: string; body: string; cliDoor: { label: string; href: string }; note: string };
  afterInstall: { title: string; lede: string; beats: readonly Tile[]; thenTitle: string; thenLede: string };
  /** The three follow-on tabs; the CLI note varies by platform. */
  followOns: (platform: DesktopPlatform) => readonly CommandTab[];
  /** The lead ends where the checksums link begins; the link's words follow it. */
  verify: (platform: DesktopPlatform) => { title: string; lead: string; checksumsLabel: string; macNote?: string };
  updates: (platform: DesktopPlatform) => { title: string; body: string; homebrew?: string; tail: string };
  /** The footer's three doors: why the laptop, the CLI alone, self-hosted. */
  /** Two lines under the verify card: the landing's argument as a door, then the two other shapes. */
  alsoConsider: { landingLabel: string; question: string; cliLabel: string; cliHref: string; selfHostedLabel: string; selfHostedHref: string };
}

export const DESKTOP_LANDING: DesktopLanding = {
  headline: HERO_HEADLINE,
  // Chapters 8 and 12
  lede: 'The whole Planton platform runs on your laptop and deploys to your cloud with the logins already on your machine. No account. Free for individuals, including commercial use.',
  doors: { primary: 'desktop', secondary: 'codingAgentsDocs' },
  twoWays: {
    // Chapter 1, told for the person who already has an agent and a cloud CLI.
    title: 'You Already Have Two Ways to Do This',
    ways: [
      {
        title: 'A Platform That Hides the Cloud',
        trade: 'Convenience, in exchange for control.',
        body: 'Your app runs in someone else\u2019s account, under IAM you did not write, on a bill you cannot itemize. When you outgrow it, you start over.',
      },
      {
        title: 'Your Agent and the Cloud CLI',
        trade: 'Control, in exchange for a record.',
        body: 'The agent writes the Terraform from memory and you learn at apply time what it got wrong. Permissions nobody derived. A cost you find out about next month. A database password that went through the chat to reach the shell. No pipeline. The second environment is a second conversation.',
      },
    ],
    turn: 'Planton is a platform too. It just runs on your machine, in your account, and gives the agent rules you can read.',
    concession: POSITIONING.desktop.concession,
  },
  ask: {
    // Chapter 2, proof 2; the coding-agents guide; the desktop's free device lane and the bring-your-own-key choice.
    title: 'Two Ways to Ask. The Same Skills.',
    lede: 'Ask from inside the tool you already use, or ask the assistant built in. Both run the same skills.',
    agent: {
      title: 'Your Coding Agent',
      body: 'Install the Planton skills once. From then on Cursor, Claude Code, or Codex composes infrastructure grounded in the real schema, validates it offline, and applies it through the platform on your laptop, while you stay in your editor.',
      command: AGENT_SKILLS_INSTALL_COMMAND,
    },
    assistant: {
      title: 'The Built-In Assistant',
      body: [
        'Open the app and ask. The Planton Assistant converses with zero keys and zero accounts; Planton funds it.',
        'Prefer to keep everything on your machine? Bring your own Anthropic or Cursor key in the engine chooser, and the conversation and the model traffic never leave the laptop.',
      ],
    },
    addsTitle: 'What Planton Adds to the Agent You Already Use',
    adds: POSITIONING.desktop.whatItAdds,
    close: 'Your agent and the platform\u2019s own assistant share the same skills, the same catalog, and the same rules about what never happens without your say-so.',
  },
  runs: {
    // The local runtime: the daemon supervises Postgres, Temporal, a cache, and the control plane; every downloaded artifact is verified before it is installed; tools already on the PATH are used, never replaced.
    title: 'What Actually Runs on Your Machine',
    lede: 'A full platform as ordinary processes. No containers to babysit.',
    tiles: [
      {
        title: 'The Control Plane and One Runner',
        body: 'The same control plane that serves planton.ai, running as a local process, with a single runner that is ready the moment the app is. No slug to name, no registration, no credentials to hand over.',
      },
      {
        title: 'Postgres, Temporal, and a Cache, as Native Processes',
        body: 'Downloaded once on first launch, verified, and supervised by the app. Docker is not required.',
      },
      {
        title: 'OpenTofu and Pulumi as Your Tools',
        body: 'Planton honors the tofu or pulumi already on your PATH. Only when nothing usable exists does it install one, through a version manager you already use or from the tool\u2019s official distribution, checksum-verified.',
      },
      {
        // Chapter 8, proof 2
        title: 'Your Data Stays Here',
        body: 'Secrets encrypted in the local database with the key in your OS keychain. State on your disk. Logs on your disk. Local data never leaves the laptop.',
      },
    ],
  },
  infraHub: {
    // The hub's one analogy lives here and nowhere else on the page; the connections guide backs the detection.
    kicker: POSITIONING.infraHub.analogy,
    title: `${POSITIONING.infraHub.name} on Your Laptop`,
    lede: POSITIONING.infraHub.line,
    points: [
      {
        title: 'Your Clouds, Found, Not Pasted',
        body: 'Planton reads the cloud logins already on your machine (AWS profiles, gcloud configurations, az subscriptions, kubeconfig contexts) and turns each into a ready connection. Cloudflare and DigitalOcean tokens are detected from your environment and verified against the provider before they are offered. You never paste a key into a form.',
      },
      {
        title: `${PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} Component Kinds Across ${PLATFORM_STATS.CLOUD_PROVIDER_COUNT} Providers`,
        body: `The full catalog, seeded into your local instance at first boot. Counted from the open-source repository on ${PLATFORM_COUNTS.countedOn}.`,
        door: DOORS.catalogBrowser,
      },
      {
        title: 'A Map of Everything You Run',
        body: 'Accounts, environments, projects, resources, and the services that span them: one living picture, with drill-down to each project\u2019s diagram.',
      },
    ],
  },
  serviceHub: {
    // Chapter 6 on the laptop; every sentence and figure is the CI/CD-on-your-laptop guide's.
    kicker: POSITIONING.serviceHub.analogy,
    title: `${POSITIONING.serviceHub.name} on Your Laptop`,
    lede: 'Push to GitHub. Your laptop notices, builds the commit in a pod, deploys it to the cluster or cloud you connected, and puts a status on the commit.',
    measured: {
      figures: [
        { value: '35 s', label: 'Push to Run Start' },
        { value: '30 s', label: 'Build' },
        { value: '34 s', label: 'Deploy' },
        { value: '~800 MiB', label: 'Build Cluster at Rest' },
      ],
      provenance:
        'Measured on an M-series Mac, end to end, unattended: pushed at 07:33:26, the run started at 07:34:01 (waking the stopped build cluster), then a 30 s build and a 34 s deploy; the idle build cluster sampled at about 800 MiB. A running, verified deployment inside two minutes of the push. Your numbers will differ with your machine and your Dockerfile; the shape will not.',
    },
    points: [
      {
        title: 'No Public URL, No GitHub App',
        body: 'A laptop cannot receive a webhook, so the local instance asks GitHub whether the branches it watches have moved, with your own gh sign-in and conditional requests that cost nothing while nothing changed. Each push becomes the same run a webhook would have started.',
      },
      {
        title: 'Builds in a Pod on Your Machine',
        body: 'The build runs on a small build cluster the app sets up inside Docker Desktop the first time you say yes. It sleeps after ten idle minutes and wakes on your next push. Your sign-in enters the build once and is deleted when the run ends.',
      },
      {
        title: 'Deploys Where You Point It',
        body: 'The image goes to the cluster or cloud you connected through the same modules hosted Planton uses, with the same rollout verification; the result lands on the commit as a status your pull request shows.',
      },
    ],
    dockerNote: 'Building on your own machine needs Docker Desktop. Planton detects it and never installs it. Everything else on this page runs without Docker.',
  },
  everywhere: {
    // Chapter 8, proof 3, told for the laptop.
    title: 'The Same Planton Everywhere',
    body: 'This laptop, planton.ai, or your company\u2019s cluster: pick where in the instance switcher. One data model, one set of code paths; only which capabilities run differs. Your laptop can also deploy into your team\u2019s Planton, with your cloud credentials still never leaving your machine.',
    // Chapter 8, proof 1: the exit path, said where the person deciding to install asks it.
    exit: runs.proof[1],
    umbrella: POSITIONING.umbrella.sentence,
  },
  whyFree: {
    // Chapter 12; the seat counts read the pricing constants, and prices stay on the pricing page.
    title: 'Why It Is Free',
    body: [
      'Planton Desktop is free for individuals, including commercial use. Not a trial, not a tier with the good parts removed: the full catalog, the wizards, the IaC pipelines, service CI/CD, and secrets management are the core product on every edition.',
      `Here is the business model, stated plainly. Planton makes its money when a team adopts it: hosted seats beyond the ${FREE_TIER_SEATS} free ones, self-hosted licenses beyond the community edition\u2019s ${COMMUNITY_SEAT_LIMIT} seats, and prepaid AI credits. The solo developer is not a funnel stage. You are the person this was built to delight, and if you one day bring a team, nothing you built gets redone.`,
    ],
    door: { label: 'See Full Pricing', href: '/pricing' },
  },
  next: ['/product/cli', '/product/infra-hub', '/product/service-hub'],
  close: {
    title: 'Install It. Ask It Something.',
    body: 'Download, open, pick Run on This Computer, and ask for what you need in your own words. Or install the skills and ask from your editor.',
    guide: { label: 'Read the CI/CD Guide', href: '/docs/ci-cd/on-your-laptop' },
  },
};

const cliInstallNote = (platform: DesktopPlatform): string =>
  platform.id === 'macos'
    ? 'The Homebrew cask installs the CLI with the app. From the disk image, Planton offers to install it for you and shows the one PATH line if ~/.local/bin is not on your PATH yet.'
    : 'Planton offers to install the CLI for you on first launch and shows the one PATH line if ~/.local/bin is not on your PATH yet.';

export const DESKTOP_DOWNLOAD: DesktopDownload = {
  // Chapter 12
  lede: 'Free for individuals, including commercial use. No account, no sign-up. Pick your platform; the rest is one launch.',
  unavailable: (platform) => ({
    title: `The ${platform.name} Installer Returns with the Next Release`,
    body: `The build currently published for ${platform.name} is not one we would ask you to run, so the link is off until the next release replaces it. Two ways forward today:`,
    cliDoor: { label: 'Install the CLI on WSL', href: '/docs/cli' },
    note: 'The CLI runs the same schema lookups, validation, and deploys from a terminal; the desktop app adds the assistant, the canvas, and the local instance.',
  }),
  afterInstall: {
    // First launch as the desktop wiki and the local-runtime articles describe it; the connections guide backs the third beat.
    title: 'After You Install',
    lede: 'One launch does the rest.',
    beats: [
      {
        title: 'Pick Where Planton Runs',
        body: 'First launch asks one question: this computer, planton.ai, or a self-hosted deployment. Choose Run on This Computer. No account is asked for, then or later.',
      },
      {
        title: 'Watch It Set Itself Up',
        body: 'The app downloads its runtime once: a Java runtime, Postgres, Temporal, a cache, and the control plane. A few hundred megabytes with a real progress bar, verified against the release\u2019s checksums. Later upgrades fetch only what changed.',
      },
      {
        title: 'Your Clouds Are Already There',
        body: 'Planton finds the AWS, Google Cloud, Azure, and Kubernetes sign-ins on your machine and offers each as a ready connection. Then ask for what you need.',
      },
    ],
    thenTitle: 'Then, From Wherever You Work',
    thenLede: 'Ask from your editor, from the terminal, or build your services on this machine. Each is a minute.',
  },
  // The coding-agents guide, the CLI guide, and the CI/CD-on-your-laptop guide.
  followOns: (platform) => [
    {
      label: 'Coding Agent',
      commands: [AGENT_SKILLS_INSTALL_COMMAND],
      description:
        'Installs the Planton skills into Cursor, Claude Code, Codex, and any other agent on the machine. Then ask in your editor, in your own words: "I need a Postgres database for this service in dev." The agent composes the manifest from the real schema, validates it offline, and applies it through the platform on this laptop, with one confirmation from you.',
    },
    {
      label: 'CLI',
      commands: ['planton --version', 'planton explain aws-vpc', 'planton validate -f infrastructure/'],
      description: `${cliInstallNote(platform)} Schema lookups and validation work offline, with no account.`,
    },
    {
      label: 'Build Locally (Optional)',
      commands: ['planton local build-cluster status'],
      description:
        'Only needed to build your own services on this machine; everything else on this page runs without Docker. Say yes to "Build on this machine?" once, with Docker Desktop running. Planton creates a small build cluster inside it, sleeps it after ten idle minutes, and wakes it on your next push.',
    },
  ],
  verify: (platform) => ({
    title: `Verify the Download on ${platform.name}`,
    lead: 'Hash the file you downloaded and compare it with',
    checksumsLabel: 'the release\u2019s checksums',
    macNote: platform.id === 'macos' ? 'The next two lines ask Gatekeeper and the notarization ticket directly; a signed, stapled build passes both.' : undefined,
  }),
  updates: (platform) => ({
    title: 'Updates',
    body: 'The app checks for updates and installs them with one click.',
    homebrew: platform.id === 'macos' ? DESKTOP_BREW_UPGRADE_COMMAND : undefined,
    tail: 'Updates are signed by Planton and verified before they are applied.',
  }),
  alsoConsider: {
    landingLabel: 'Why run Planton on your laptop?',
    question: 'Want the platform without the app?',
    cliLabel: 'Install the CLI on its own',
    cliHref: '/docs/cli',
    selfHostedLabel: 'run the self-hosted edition on your own cluster',
    selfHostedHref: '/docs/self-hosting',
  },
};
