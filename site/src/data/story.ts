/**
 * The Planton story as data: thirteen chapters in the order every surface
 * tells them. The landing page, the Trust pages, the persona pages and
 * decks, and llms.txt all render from this file, so a sentence about what
 * Planton is or does appears here once and nowhere else.
 *
 * The canonical narrative lives in the company repository as
 * company/marketing/positioning/the-planton-story.md; every claim below
 * names its chapter, and a chapter's `neverSay` list is the honesty rail
 * its surfaces are graded against. The umbrella and hub sentences come from
 * ./positioning.ts (the vocabulary law); numbers come from ./platform-stats.ts
 * and prices from ./pricing.ts, never from here.
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators (llms.txt) without a bundler.
 */
import { POSITIONING } from './positioning.ts';
import { COMMUNITY_SEAT_LIMIT, EVALUATION_DAYS, FREE_TIER_SEATS } from './pricing.ts';

export type ChapterId =
  | 'the-wall'
  | 'what-planton-is'
  | 'verified-before-it-exists'
  | 'your-rules-hold'
  | 'every-deployment-leaves-a-record'
  | 'services-ship-from-git'
  | 'bring-what-you-have'
  | 'runs-where-you-decide'
  | 'who-it-is-for'
  | 'proof-it-works'
  | 'how-it-compares'
  | 'start'
  | 'what-is-next';

export interface Chapter {
  /** Stable id; the landing section anchor and the deck slide group key. */
  id: ChapterId;
  /** 1-based position in the story. */
  number: number;
  /** Title Case; the heading a page or slide shows. */
  title: string;
  /**
   * The claim in one or two sentences a visitor could read on a page.
   * Sentence case, plain words, no analogy for the umbrella.
   */
  claim: string;
  /** The shipped behavior that makes the claim true, each a sentence a page can show as a proof point. */
  proof: readonly string[];
  /** Sentences this chapter's surfaces must never contain (the honesty rails, for the reviewer and the grep gate). */
  neverSay: readonly string[];
  /** True for the roadmap chapter: rendered in decks with a disclosure line, never on the website. */
  decksOnly?: boolean;
}

/** The disclosure every deck attaches to chapter 13. Never rendered on the site. */
export const ROADMAP_DISCLOSURE = 'Roadmap. Not yet shipped; shown so you know where this is going.';

/** The spine, in one sentence, for a hero or a deck cover. */
export const STORY_SPINE =
  'Proof at creation, not observation after: every change Planton deploys is priced before it exists with its coverage stated, judged against a budget, held to your rules, stamped with the controls it enforces, and left behind as an immutable record.';

/** The first sentence of every demo (chapter 1): the desktop line in full. */
export const OPENING_LINE = POSITIONING.desktop.line;

/**
 * The landing hero's headline: the opening line up to its dash, so the
 * headline is two sentences and the three words after the dash (your
 * account, your laptop, free) are said once more in the fine print under
 * the doors rather than in the headline. Derived, never retyped.
 */
export const HERO_HEADLINE = `${OPENING_LINE.split(' \u2014 ')[0]}.`;

/**
 * The hero's second sentence: where Planton sits relative to the agent and
 * the CLI the visitor already has, and the three things it adds. Answers the
 * first objection on the first screen.
 */
export const HERO_SUBHEAD =
  'It sits beside the coding agent and cloud CLI you already use. Verifiable: cost, permissions, and controls checked before anything exists. Recorded: every change kept where you can query it. Reusable: a template your team redeploys.';

export const CHAPTERS: readonly Chapter[] = [
  {
    // Chapter 1
    id: 'the-wall',
    number: 1,
    title: 'What the Agent Leaves Behind',
    claim:
      'Your coding agent can already create cloud infrastructure. What it creates is unverified, unrecorded, and unrepeatable: nobody priced it, nobody checked the permissions, nothing remembers what was made, and the next environment starts from a blank prompt.',
    proof: [
      'Developers still wait and platform teams still queue, and the agent that was meant to end the wait has added a new kind of risk.',
    ],
    neverSay: ['agents are bad at infrastructure', 'eliminate DevOps engineers', 'any waiting-time or salary figure'],
  },
  {
    // Chapter 2
    id: 'what-planton-is',
    number: 2,
    title: POSITIONING.umbrella.tagline,
    claim: POSITIONING.umbrella.sentence,
    proof: [
      `${POSITIONING.infraHub.name}: ${POSITIONING.infraHub.line}`,
      `${POSITIONING.serviceHub.name}: ${POSITIONING.serviceHub.line}`,
      'Your coding agent reaches Planton through the Planton skills and the Planton MCP server, or through the CLI. It sits beside the tools you already use; nothing about how you work changes.',
    ],
    neverSay: ['an analogy for the umbrella', 'either hub analogy about the whole product', '"Template" capitalized as a product name'],
  },
  {
    // Chapter 3
    id: 'verified-before-it-exists',
    number: 3,
    title: 'Verified Before It Exists',
    claim:
      'Before anything is created, Planton tells you what it will cost each month with its coverage stated honestly, which permissions it needs and no more, and which technical controls the components enforce. Proof at creation, not detection after.',
    proof: [
      'Every deployment-changing job is born with a verified monthly cost: an exact figure with line items when the pricing rules can derive one, a range otherwise, and plainly \u201cunpriced\u201d when neither is possible. A zero never stands in for unknown.',
      'The cost names the catalog release its prices came from and, when both can be priced exactly, how much more or less this will cost each month than what is deployed today.',
      'The least-privilege permission policy is derived from exactly what is composed, per component kind.',
      'Every covered component states which of a fixed list of 17 technical controls it enforces, with evidence for each claim.',
      'The console, the CLI, and the assistant render the same server-stamped statement word for word.',
    ],
    neverSay: ['saves you $X', 'compliant', 'an estimate presented as a bill', 'estimate versus actual (not shipped)'],
  },
  {
    // Chapter 4
    id: 'your-rules-hold',
    number: 4,
    title: 'Your Rules Hold No Matter Who Asked',
    claim:
      'A platform team writes the rules once and every request obeys them, whether it came from a person in the console, a script on the CLI, or a coding agent at two in the morning.',
    proof: [
      'An environment can carry a deployment budget. A deploy whose verified cost exceeds it pauses for a human decision, and who approved, when, and why is stamped on the record.',
      'Protected environments pause before anything deploys, and nobody approves work they initiated, the assistant included.',
      'The catalog can be curated to the component kinds your organization allows. The console, the CLI, and the agent all see the same list and refuse the same things, because one answer serves both.',
      'A field the schema marks sensitive takes a managed secret. There is no way to type a raw secret into it.',
    ],
    neverSay: ['guardrails over spec content as shipped', 'policy as code', 'prevents all misconfiguration'],
  },
  {
    // Chapter 5
    id: 'every-deployment-leaves-a-record',
    number: 5,
    title: 'Every Deployment Leaves a Record',
    claim:
      'Every change to infrastructure runs as one stack job, and every stack job is kept: the exact configuration that was deployed, the cost fact, the budget verdict, who approved and why, who triggered it, what happened in every phase, and a snapshot of what exists afterward.',
    proof: [
      'The full configuration is embedded into the job when it is created, and the job is immutable: the resource may change later; the job never does.',
      'Every job is retained and queryable by resource, organization, environment, time, and outcome.',
      'One event stream drives the console, the CLI, and the audit log, so every surface tells the same story.',
      'Every cloud resource Planton creates carries identity tags naming its organization, environment, kind, and id.',
    ],
    neverSay: ['audit-ready for SOC 2', 'any framework verdict about a deployment', 'drift detection (a future consideration)', 'compliance dashboard'],
  },
  {
    // Chapter 6
    id: 'services-ship-from-git',
    number: 6,
    title: 'Services Ship From Git',
    claim:
      'Once the infrastructure exists, your services ship onto it straight from Git. Connect a repository; every push is built, containerized, and deployed, with the result written back into GitHub as checks and deployments.',
    proof: [
      'No pipeline YAML and no Dockerfile required.',
      'Builds can be manual, from a branch, or from an exact commit; promotion follows the order your environments declare; protected environments carry a manual gate.',
      'A rejected environment is never touched, and a service with deployments in a protected environment refuses deletion outright.',
      'Runner minutes are never billed.',
    ],
    neverSay: ['"Vercel for backend" about Planton', 'image scanning, SBOMs, or signing (not shipped)', 'post-deploy health checks (not shipped)'],
  },
  {
    // Chapter 7
    id: 'bring-what-you-have',
    number: 7,
    title: 'Bring What You Already Have',
    claim:
      'You do not start from an empty account. Infrastructure that already exists can be brought under Planton\u2019s record without being redeployed: it is adopted, its live state is imported and verified in one step, and from then on it carries the same record as everything Planton created.',
    proof: [
      'Create and update can track a resource without deploying anything; it reads honestly as adopted, not yet deployed.',
      'A state import is verified atomically: a wrong import fails before it lands, and import never writes a configuration it did not apply.',
      'Import recipes exist for S3 buckets, VPCs, security groups, and container registries, proven in a live round trip.',
    ],
    neverSay: ['scan your whole account and map it with AI (not released)', 'import anything'],
  },
  {
    // Chapter 8
    id: 'runs-where-you-decide',
    number: 8,
    title: 'Runs Where You Decide',
    claim:
      'Planton runs where you decide: hosted at planton.ai, self-hosted on your own Kubernetes cluster with a license that verifies offline, or free on your laptop as Planton Desktop. In every shape it is your cloud account, your keys, your state, and your bill.',
    proof: [
      'Connections can be keyless: a short-lived, connection-scoped identity token is minted at the moment of use and exchanged by the cloud for temporary credentials. Nothing is stored, nothing is rotated, and the trust is yours to revoke.',
      'Every infrastructure module is open source under Apache 2.0. If you leave, you take your manifests and keep deploying them with the open-source CLI.',
      'On the desktop, secrets are encrypted locally with the key in your OS keychain and resolved on the runner at the moment of use.',
      'The same manifests and the same model run on all three shapes, so nothing is redone when a person becomes a team.',
    ],
    neverSay: ['zero lock-in as a badge (say the exit path)', 'a mobile store link', 'air-gapped as a shipped enterprise feature beyond the offline key'],
  },
  {
    // Chapter 9
    id: 'who-it-is-for',
    number: 9,
    title: 'Who It Is For',
    claim:
      'The platform engineer is the user: they write the rules, curate the catalog, publish the Infra Charts, and hand developers and agents a self-service platform that cannot break those rules. The engineering leader is who signs: what reaches them is the proof.',
    proof: [
      'A junior DevOps engineer served a seven-person team without learning AWS from scratch.',
      'A consultancy delivered GCP infrastructure with no GCP expertise on the team, and runs one organization per client.',
      'Developers create and deploy services to dev, staging, and production without waiting on anyone.',
    ],
    neverSay: ['built for developers as the headline', 'replaces your DevOps team', 'a persona we have not sold to described as a customer'],
  },
  {
    // Chapter 10
    id: 'proof-it-works',
    number: 10,
    title: 'Proof It Works',
    claim:
      'Teams have run production on Planton since 2023. Here is what the people running it say, in their own words.',
    proof: [
      'Component kinds, providers, Infra Charts, controls, and crosswalks are counted from the open-source tree, and the date of the count is printed beside the numbers.',
      'Every testimonial is verbatim and attributed to the person who said it.',
      'Planton runs on Planton: its own infrastructure and the pipelines that ship it go through the platform.',
    ],
    neverSay: ['any savings figure', '50+ charts', 'an exact kind count in prose', 'a paraphrased or company-attributed quote'],
  },
  {
    // Chapter 11
    id: 'how-it-compares',
    number: 11,
    title: 'How It Compares',
    claim:
      'Governance platforms observe your estate after the fact and tell you what to fix; Planton prevents at the moment of creation and stamps the proof on the record, alongside those tools rather than instead of them. Infrastructure-as-code tools work on Terraform you still write, field name by field name; Planton gives you typed self-service with rules written once over a fixed list of controls. Developer portals catalog what you have; Planton deploys what you need.',
    proof: [
      'Prevent at creation versus observe after: a complement to posture tools, never a replacement.',
      'Your Terraform stays yours: the modules are open-source Terraform and Pulumi, and what you already run is adopted, not rewritten. Rules are written once over a handful of controls, not over a hundred field names.',
      'The execution layer behind the catalog: it deploys, with the record attached.',
    ],
    neverSay: ['any vendor name', 'DevSecOps platform', 'FinOps', 'a feature table with a competitor column'],
  },
  {
    // Chapter 12
    id: 'start',
    number: 12,
    title: 'Start',
    claim: `Start free today. Planton Desktop is free for individuals forever, commercial use included. The hosted free tier is free for up to ${FREE_TIER_SEATS} seats with no card. The self-hosted community edition is free for up to ${COMMUNITY_SEAT_LIMIT} seats, and a ${EVALUATION_DAYS}-day evaluation of the full edition needs only an email.`,
    proof: [
      'Every number on the pricing page reads from one data file and is checked against the enforced entitlements at build time.',
      'Teams pay per seat; below the self-serve ceiling, nobody talks to sales.',
    ],
    neverSay: ['a price as a literal in prose', 'a per-minute runner rate', 'a monthly free AI allowance as live'],
  },
  {
    // Chapter 13
    id: 'what-is-next',
    number: 13,
    title: 'What Is Next',
    decksOnly: true,
    claim:
      'Coming, and not yet shipped: rules over the content of a spec, written once over the control vocabulary and shown to the coding agent before it authors; a record of what was refused and why; estimate-versus-actual cost read back by Planton\u2019s own tags; service health on the deploy record; image scanning in the build tracks; AI-assisted import of an existing estate; a mobile companion for approvals.',
    proof: ['Designed and in some cases proven in development. None is released.'],
    neverSay: ['any of this on the website', 'any of this in a deck without the disclosure line'],
  },
] as const;

/** Look up a chapter by id; throws at build time if a page names a chapter that does not exist. */
export function chapter(id: ChapterId): Chapter {
  const found = CHAPTERS.find((c) => c.id === id);
  if (!found) throw new Error(`the story has no chapter "${id}"`);
  return found;
}

/** The chapters a website page may render (everything but the roadmap). */
export const SITE_CHAPTERS: readonly Chapter[] = CHAPTERS.filter((c) => !c.decksOnly);
