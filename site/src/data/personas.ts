/**
 * The five people the story is told to, as data. A persona's page at
 * /solutions/<slug> and its deck at /decks/<slug> render from the same
 * record, so the two cannot drift. A record says who the person is, the
 * wall they hit, the two or three proof points that decide it for them, the
 * chapters in the order and at the depth they need them, the objections they
 * will raise and the chapter that answers each, the people who vouch, and
 * the doors they walk through at the end.
 *
 * Copy here is the persona's framing of the story, never a new claim: every
 * fact a persona surface states traces to a chapter in ./story.ts, and a
 * beat's proof sentences are the chapter's own, chosen by index. Quotes are
 * named from ./testimonials.ts; a persona nobody on record has vouched for
 * shows the counts and no quote, never a paraphrase. Doors are named from
 * ./doors.ts; counts from ./platform-stats.ts; seat numbers from ./pricing.ts.
 *
 * The human-readable form of these records, with the reasoning behind each
 * revision, is the copywriting stage folder that produced them
 * (content/copywriting/_stage-area/2026-09-18-solutions-personas/).
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators without a bundler.
 */
import type { DoorPair } from './doors.ts';
import type { ProofPoint } from './page-shapes.ts';
import { PLATFORM_STATS } from './platform-stats.ts';
import { FREE_TIER_SEATS } from './pricing.ts';
import type { ChapterId } from './story.ts';

export type PersonaSlug =
  | 'platform-engineer'
  | 'engineering-leader'
  | 'it-consultancy'
  | 'startup-founder'
  | 'security-and-governance-leader';

/** A chapter a page may tell. The roadmap chapter is decks-only and cannot be a beat. */
export type PageChapterId = Exclude<ChapterId, 'what-is-next'>;

/**
 * One chapter, told for this person. `lead` beats are full sections on the
 * page and slides in the deck; `supporting` beats are doors on the page and
 * slides in the deck. The deck has the time for every chapter; a page that
 * gave each one a screen would not be finished by anyone. Chapter 1, the
 * wall, is never a beat: the persona's `wall` is that chapter told for this
 * person, and the page's hero and the deck's cover carry it.
 */
export interface Beat {
  chapter: PageChapterId;
  weight: 'lead' | 'supporting';
  /** Sentence case; the persona's angle on the chapter, one or two sentences. Never a claim the chapter does not make. */
  angle: string;
  /** Indices into the chapter's proof sentences, in the order this person should read them. */
  proof: readonly number[];
  /** The registered page that tells this chapter in full for this person; its record's artifact is shown beside the beat. */
  provenAt: string;
}

/** A question this person asks, answered in the words of the chapter that answers it. */
export interface Objection {
  /** Sentence case; the question as the person would ask it. */
  question: string;
  answer: string;
  chapter: PageChapterId;
  /** Where to read on; defaults to the page this persona's beat for the chapter names, when there is one. */
  readMore?: string;
}

export interface Persona {
  slug: PersonaSlug;
  /** Title Case; the page heading and the deck cover. */
  name: string;
  /** Title Case; the plural the registry titles and the deck frame use ("Planton for Platform Engineers"). */
  plural: string;
  /** Title Case; the promise in this person's words. The page's headline, under the name. */
  headline: string;
  /** Sentence case; who this person is, in one line they would recognize. */
  who: string;
  /** Sentence case; the wall they hit, in their words. Opens the page and the deck. */
  wall: string;
  /** The pair of doors this person walks through, named from the door vocabulary. */
  doors: DoorPair;
  /** The two or three proof points that decide it for this person. */
  decidingProof: readonly ProofPoint[];
  /** The chapters in this person's order, each at its weight. */
  beats: readonly Beat[];
  /** Two or three questions this person asks, answered where they arise. */
  objections: readonly Objection[];
  /** Names from ./testimonials.ts; may be empty. */
  testimonials: readonly string[];
  /** Sentence case; the first sentence the presenter says, on the deck's cover. */
  opening: string;
  /** Two other personas whose pages this one offers at its close. */
  siblings: readonly [PersonaSlug, PersonaSlug];
}

const { DEPLOYMENT_MODULE_COUNT, CLOUD_PROVIDER_COUNT, CONTROL_COUNT, FRAMEWORK_CROSSWALK_COUNT } = PLATFORM_STATS;

export const PERSONAS: readonly Persona[] = [
  {
    slug: 'platform-engineer',
    name: 'Platform Engineer',
    plural: 'Platform Engineers',
    headline: 'Self-Service That Cannot Break Your Rules',
    who: 'You run the platform your developers and their coding agents build on, and you are the one who gets paged when something they created is wrong.',
    wall: 'Agents can already create infrastructure in your account. You cannot see what they made, what it costs, or whether it follows the rules you wrote down last quarter.',
    doors: { primary: 'hosted', secondary: 'desktop' },
    decidingProof: [
      { label: 'Rules Written Once', text: 'Budgets, protected environments, and a curated catalog hold whether the request came from the console, the CLI, or an agent at two in the morning.' },
      { label: 'A Record, Not a Transcript', text: 'Every deploy is one stack job you can query by resource, environment, time, and outcome, with the exact configuration embedded.' },
      { label: 'Your Terraform Stays Yours', text: 'Every module is open-source Terraform and Pulumi under Apache 2.0. What you already run is adopted, not rewritten, and if you leave you keep deploying your manifests with the open-source CLI.' },
    ],
    beats: [
      { chapter: 'your-rules-hold', weight: 'lead', angle: 'You write the budget, the protected environments, and the allowed catalog once. Every door reads the same rules and refuses the same things, so an agent cannot do what a person could not.', proof: [0, 1, 2, 3], provenAt: '/trust/rules-and-approvals' },
      { chapter: 'verified-before-it-exists', weight: 'lead', angle: 'Before an agent\u2019s design exists, you see its monthly cost with its coverage stated, the least-privilege policy it needs, and the controls each component enforces.', proof: [0, 2, 3], provenAt: '/trust/verified-before-deploy' },
      { chapter: 'every-deployment-leaves-a-record', weight: 'lead', angle: 'The deploy nobody watched is a stack job with the configuration embedded, the verdicts stamped, and the approver named. You read it; you do not reconstruct it.', proof: [0, 1, 3], provenAt: '/trust/the-record' },
      { chapter: 'what-planton-is', weight: 'supporting', angle: 'Two halves: Infra Hub, where a design becomes a template your team redeploys, and Service Hub, where a push becomes a deployment inside your environments\u2019 gates.', proof: [0, 1], provenAt: '/product/infra-hub' },
      { chapter: 'runs-where-you-decide', weight: 'supporting', angle: 'Start on your laptop for free and move to hosted or your own cluster with the same manifests. Connections can be keyless, and every module is open source.', proof: [0, 1, 3], provenAt: '/distributions' },
      { chapter: 'services-ship-from-git', weight: 'supporting', angle: 'Once the infrastructure exists, developers connect a repository and every push obeys the promotion order and the gates you declared.', proof: [1, 2], provenAt: '/product/service-hub' },
      { chapter: 'bring-what-you-have', weight: 'supporting', angle: 'What already exists is adopted and its state imported in one verified step, so the record covers your estate, not only what Planton created.', proof: [0, 1, 2], provenAt: '/product/import' },
    ],
    objections: [
      { question: 'Is this another abstraction layer I will be fighting in six months?', answer: 'Every component is a typed schema over an open-source Terraform or Pulumi module, and controls are stated against one fixed list rather than a hundred field names. What you already run is adopted, not rewritten, and if you leave, the modules and your manifests go with you.', chapter: 'how-it-compares', readMore: '/product/open-source' },
      { question: 'What can the agent do that a person could not?', answer: 'Nothing. It comes through the same door as the console and the CLI, reads the same rules, and gets the same refusals. A protected environment pauses for a human, and nobody approves work they initiated, the assistant included.', chapter: 'your-rules-hold', readMore: '/product/coding-agents' },
      { question: 'What happens to my environments if Planton goes away?', answer: 'Every infrastructure module is open source under Apache 2.0. You take your manifests and keep deploying them with the open-source CLI. In every shape it is your cloud account, your keys, your state, and your bill.', chapter: 'runs-where-you-decide' },
    ],
    testimonials: ['Sai Saketh', 'Rakesh Kandhi'],
    opening: 'Your developers\u2019 coding agents can already create cloud infrastructure in your account. Here is what happens when they do it through the rules you wrote.',
    siblings: ['engineering-leader', 'security-and-governance-leader'],
  },
  {
    slug: 'engineering-leader',
    name: 'Engineering Leader',
    plural: 'Engineering Leaders',
    headline: 'What Your Team Deploys, Proven Before It Exists',
    who: 'You sign for the cloud bill and answer for the outage. You will not live in a console; you want to know what was deployed, what it costs, and that the rules held.',
    wall: 'Your team says the agents are making them faster. Nobody can show you the cost of what was created before it was created, or who approved it.',
    doors: { primary: 'demo', secondary: 'pricing' },
    decidingProof: [
      { label: 'Cost Before Creation', text: 'Every deployment-changing job is born with a verified monthly cost, its coverage stated, and how much more or less it will cost than what runs today.' },
      { label: 'No New Team', text: 'Your platform engineer stays the user and writes the rules once; developers and their agents self-serve inside them. What reaches you is the proof.' },
      { label: 'A Record That Does Not Depend on Asking', text: 'Every change is one immutable stack job: configuration, cost, verdict, approver, outcome. Queryable whenever you want it, never reconstructed because you asked.' },
    ],
    beats: [
      { chapter: 'verified-before-it-exists', weight: 'lead', angle: 'Cost, permissions, and controls become properties of what gets deployed, checked before it exists rather than found after.', proof: [0, 1, 3], provenAt: '/trust/verified-before-deploy' },
      { chapter: 'your-rules-hold', weight: 'lead', angle: 'Your platform engineer writes the budget and the protected environments once. A deploy that would exceed the budget pauses, and a protected environment refuses self-approval, whoever asked.', proof: [0, 1], provenAt: '/trust/rules-and-approvals' },
      { chapter: 'every-deployment-leaves-a-record', weight: 'lead', angle: 'The record is what you get to see: what was deployed, what it cost, who approved it and why. It is there whether or not anyone asks.', proof: [0, 1, 2], provenAt: '/trust/the-record' },
      { chapter: 'who-it-is-for', weight: 'supporting', angle: 'The platform engineer is the user and you are who signs. The proof is what travels between you.', proof: [0, 1], provenAt: '/solutions' },
      { chapter: 'what-planton-is', weight: 'supporting', angle: 'One platform in your own cloud account: AI designs, the platform verifies, the design becomes a template, and services ship from Git onto it.', proof: [0, 1], provenAt: '/product' },
      { chapter: 'runs-where-you-decide', weight: 'supporting', angle: 'Hosted, self-hosted, or on a laptop; in every shape it is your account, your keys, your state, and your bill, with keyless connections and open-source modules.', proof: [0, 1], provenAt: '/trust/your-cloud-your-keys' },
    ],
    objections: [
      { question: 'Do I have to open a console to get any of this?', answer: 'No. The proof is the record: the verified cost before creation, the budget verdict and who resolved it, the controls each component enforces, and the immutable job behind every change. Your platform engineer reads it and hands it to you; when you want to look yourself, the same record is one sign-in away and reads the same for you as for them.', chapter: 'who-it-is-for' },
      { question: 'Is there a savings number I can put in a budget?', answer: 'No, and we will not invent one. What you get is the verified monthly cost of each change before it is committed, with its coverage stated, judged against the budget you set. A dollar-savings figure would be a claim nobody could check, so you will not find one here.', chapter: 'verified-before-it-exists' },
      { question: 'What does this cost?', answer: `Teams pay per seat; below the self-serve ceiling a card is all it takes, and the demo beside this page is an offer, not a gate. The hosted free tier is free for up to ${FREE_TIER_SEATS} seats with no card, and Planton Desktop is free for individuals forever, commercial use included.`, chapter: 'start', readMore: '/pricing' },
    ],
    testimonials: ['Rohit Reddy Gopu'],
    opening: 'You will not see a dashboard today. You will see what reaches you after your team\u2019s agent deploys something: the cost before it existed, the rule that held, and the record.',
    siblings: ['platform-engineer', 'security-and-governance-leader'],
  },
  {
    slug: 'it-consultancy',
    name: 'IT Consultancy',
    plural: 'IT Consultancies',
    headline: 'Repeatable Client Environments, Handed Back as Manifests',
    who: 'You stand up environments for clients who mandate a cloud your team may not know, and you hand the work back when the engagement ends.',
    wall: 'Every client starts from zero, and the last client\u2019s Terraform does not fit this one.',
    doors: { primary: 'hosted', secondary: 'desktop' },
    decidingProof: [
      { label: 'One Organization per Client', text: 'Each client is its own organization, with its own cloud connection, environments, budgets, and record.' },
      { label: 'Repeatability Is the Product', text: 'Publish the environment you built for one client as an Infra Chart and deploy it into the next. A prompt cannot be redeployed; a chart can.' },
      { label: 'Hand Back Everything', text: 'Every module is open source under Apache 2.0. When the engagement ends, the client keeps their manifests and keeps deploying them with the open-source CLI.' },
    ],
    beats: [
      { chapter: 'what-planton-is', weight: 'lead', angle: 'Describe the client\u2019s environment, verify it, deploy it, and publish it as an Infra Chart. The next client starts from the chart, not from a blank prompt.', proof: [0, 1], provenAt: '/product/infra-hub' },
      { chapter: 'runs-where-you-decide', weight: 'lead', angle: 'The client\u2019s account, the client\u2019s keys, the client\u2019s bill. Connections can be keyless, so you never hold a long-lived credential for an account you do not own.', proof: [0, 1, 3], provenAt: '/trust/your-cloud-your-keys' },
      { chapter: 'verified-before-it-exists', weight: 'supporting', angle: 'The monthly cost with its coverage stated, before the client\u2019s infrastructure exists. The estimate is the conversation with the client, not the invoice.', proof: [0, 1], provenAt: '/trust/verified-before-deploy' },
      { chapter: 'every-deployment-leaves-a-record', weight: 'supporting', angle: 'Every change in every client organization is a stack job you can hand over: what was deployed, what it cost, who approved it.', proof: [0, 1], provenAt: '/trust/the-record' },
      { chapter: 'services-ship-from-git', weight: 'supporting', angle: 'Connect the client\u2019s repositories and every push is built and deployed through the environments you declared.', proof: [0, 1], provenAt: '/product/service-hub' },
    ],
    objections: [
      { question: 'The client mandated a cloud my team has not worked in. Can we deliver?', answer: `One of the people quoted below did exactly that: the client mandated GCP, the engineer had no GCP experience, and the consultancy delivered the whole environment. The catalog is ${DEPLOYMENT_MODULE_COUNT} component kinds across ${CLOUD_PROVIDER_COUNT} providers, each a typed schema over a tested module.`, chapter: 'proof-it-works', readMore: '/product/catalog' },
      { question: 'What does the client keep when we leave?', answer: 'Everything. Their manifests, their state in their own backend, and the open-source CLI that deploys the same modules. It was their cloud account and their keys from the first day.', chapter: 'runs-where-you-decide' },
      { question: 'What does this cost us per client?', answer: `Teams pay per seat, and below the self-serve ceiling a card is all it takes. The hosted free tier is free for up to ${FREE_TIER_SEATS} seats with no card. Every number is on the pricing page, read from the same file the platform enforces.`, chapter: 'start', readMore: '/pricing' },
    ],
    testimonials: ['Rohit Reddy Gopu', 'Balaji Borra'],
    opening: 'Your last client\u2019s Terraform does not fit this one. Here is how the environment you build today becomes the template you deploy for the next.',
    siblings: ['startup-founder', 'platform-engineer'],
  },
  {
    slug: 'startup-founder',
    name: 'Startup Founder',
    plural: 'Startup Founders',
    headline: 'Ship Without an Ops Hire, and Redo Nothing Later',
    who: 'You are shipping a product with a small team and no ops hire, and every hour on infrastructure is an hour not on the product.',
    wall: 'Your agent can build the infrastructure. You are not sure what it built, what it will cost next month, or how you will do it again for staging.',
    doors: { primary: 'hosted', secondary: 'desktop' },
    decidingProof: [
      { label: 'Push to Deploy', text: 'Connect a repository; every push is built, containerized, and deployed. No pipeline YAML and no Dockerfile required.' },
      { label: 'The Monthly Cost, Before It Exists', text: 'Before your agent\u2019s design is created, you see what it will cost each month with its coverage stated honestly, never a zero that means unknown.' },
      { label: 'Free to Start, Nothing Redone', text: `The hosted free tier is free for up to ${FREE_TIER_SEATS} seats with no card. The same manifests run when you are a team, so nothing is redone.` },
    ],
    beats: [
      { chapter: 'what-planton-is', weight: 'lead', angle: 'Your agent designs through Planton instead of around it: the design is verified, deployed, and published as a template you redeploy for staging.', proof: [0, 1, 2], provenAt: '/product/infra-hub' },
      { chapter: 'services-ship-from-git', weight: 'lead', angle: 'Once the infrastructure exists, push. Every push is built and deployed, promotion follows the order you declared, and the result lands in GitHub, where you already are.', proof: [0, 1], provenAt: '/product/service-hub' },
      { chapter: 'verified-before-it-exists', weight: 'supporting', angle: 'The monthly cost of what your agent designed, before it exists, so next month\u2019s bill is not a surprise.', proof: [0, 1], provenAt: '/trust/verified-before-deploy' },
      { chapter: 'runs-where-you-decide', weight: 'supporting', angle: 'Your cloud account, your keys, your bill. Start hosted with no card, or on your laptop for free; the same manifests move with you.', proof: [1, 3], provenAt: '/distributions/hosted' },
      { chapter: 'every-deployment-leaves-a-record', weight: 'supporting', angle: 'When you hire the first engineer who asks what is running and why, the answer is a record, not a memory.', proof: [0, 1], provenAt: '/trust/the-record' },
    ],
    objections: [
      { question: 'I do not know what a multi\u2011AZ database is. Is this for me?', answer: 'The prompt can be one sentence of intent: say what you built and where you want it to run. The depth is there when you have it, and every component is a typed schema, so a wrong field fails before it touches your cloud.', chapter: 'what-planton-is' },
      { question: 'Why not let my agent write the Terraform?', answer: 'It can, and you get different Terraform every time, with nobody pricing it and nothing remembering it. Through Planton the agent writes a small validated manifest; the module that runs is pre-written, tested, and open source; and the design becomes a template you redeploy.', chapter: 'the-wall', readMore: '/product/coding-agents' },
      { question: 'What happens when there are five of us?', answer: `Nothing is redone: the same manifests and the same model run on every shape. The hosted free tier covers ${FREE_TIER_SEATS} seats with no card; after that, teams pay per seat, and below the self-serve ceiling nobody talks to sales.`, chapter: 'start', readMore: '/pricing' },
    ],
    testimonials: ['Rakesh Kandhi'],
    opening: 'Your agent can build your infrastructure tonight. Here is how to make sure you can afford it, repeat it, and hand it to your first hire.',
    siblings: ['it-consultancy', 'platform-engineer'],
  },
  {
    slug: 'security-and-governance-leader',
    name: 'Security and Governance Leader',
    plural: 'Security and Governance Leaders',
    headline: 'Rules That Hold Before Anything Exists',
    who: 'You own the posture of an estate you did not build, and the tools you have tell you what went wrong after it did.',
    wall: 'Agents create infrastructure faster than your scanners find the problems. Prevention has to happen where creation happens.',
    doors: { primary: 'demo', secondary: 'pricing' },
    decidingProof: [
      { label: 'Prevention at the Write Boundary', text: 'Budgets, protected environments, a curated catalog, and managed secrets hold before anything exists, whether the request came from a person, a script, or an agent.' },
      { label: 'Controls Stated, Never Asserted', text: `Every covered component states which of ${CONTROL_COUNT} technical controls it enforces, with evidence for each claim, and never calls itself compliant.` },
      { label: 'A Record of Every Change', text: 'Every change is one immutable stack job, queryable by resource, environment, time, and outcome, with identity tags on every resource it created.' },
    ],
    beats: [
      { chapter: 'your-rules-hold', weight: 'lead', angle: 'The rules are enforced at the one place infrastructure is created, so a coding agent cannot do what a person could not, and the refusal is identical at every door.', proof: [0, 1, 2, 3], provenAt: '/trust/rules-and-approvals' },
      { chapter: 'verified-before-it-exists', weight: 'lead', angle: 'Least-privilege permissions derived from exactly what is composed, and the technical controls each component enforces, stated with evidence before it exists.', proof: [2, 3, 4], provenAt: '/trust/security-posture' },
      { chapter: 'every-deployment-leaves-a-record', weight: 'lead', angle: 'The immutable stack job: configuration, verdicts, approver, outcome, and identity tags on every resource created. Evidence that exists whether or not anyone asked.', proof: [0, 1, 3], provenAt: '/trust/the-record' },
      { chapter: 'how-it-compares', weight: 'supporting', angle: 'Posture platforms observe after the fact; Planton prevents at creation. They complement each other, and neither replaces the other.', proof: [0, 1], provenAt: '/compare' },
      { chapter: 'runs-where-you-decide', weight: 'supporting', angle: 'Keyless connections mean no long-lived cloud credential is stored anywhere. Self-hosted runs on your cluster with a license that verifies offline.', proof: [0, 1, 2], provenAt: '/trust/your-cloud-your-keys' },
      { chapter: 'what-planton-is', weight: 'supporting', angle: 'One platform in the customer\u2019s own account, with one door every request goes through. That door is where the rules live.', proof: [0, 1], provenAt: '/product' },
    ],
    objections: [
      { question: 'Is this compliant with SOC 2 or HIPAA?', answer: `No component is ever called compliant, and no deployment gets a framework verdict. What you get is each component\u2019s stated controls with evidence, and ${FRAMEWORK_CROSSWALK_COUNT} framework crosswalks that map those controls to HIPAA, SOC 2, FedRAMP Moderate, and CIS AWS Foundations so your assessor can read them.`, chapter: 'verified-before-it-exists' },
      { question: 'Which rules hold today?', answer: 'Deployment budgets that pause a deploy exceeding them; protected environments that refuse self-approval; a catalog curated to the kinds you allow, refused identically at every door; sensitive fields that take only a managed secret. Every one is stated here as it works today.', chapter: 'your-rules-hold' },
      { question: 'Does this cover what already exists in the account?', answer: 'Infrastructure that already exists can be adopted and its live state imported and verified in one step; from then on it carries the same record as everything Planton created. Import recipes exist for S3 buckets, VPCs, security groups, and container registries, proven in a live round trip.', chapter: 'bring-what-you-have', readMore: '/product/import' },
    ],
    testimonials: [],
    opening: 'Your scanners tell you what went wrong after it did. Here is what it looks like when the rule holds before the resource exists.',
    siblings: ['engineering-leader', 'platform-engineer'],
  },
] as const;

export function persona(slug: PersonaSlug): Persona {
  const found = PERSONAS.find((p) => p.slug === slug);
  if (!found) throw new Error(`no persona "${slug}"`);
  return found;
}

/** The persona's page, derived from the slug so a link and the registry cannot disagree. */
export const personaPagePath = (slug: PersonaSlug) => `/solutions/${slug}`;

/** The persona's deck, likewise. */
export const personaDeckPath = (slug: PersonaSlug) => `/decks/${slug}`;
