/**
 * The five people the story is told to, as data. A persona's page at
 * /solutions/<slug> and its deck at /decks/<slug> render from the same
 * record, so the two cannot drift. Each record says who the person is, the
 * wall they hit, the chapters in the order they need to hear them, the proof
 * points that matter most to them, and the door they walk through at the end.
 *
 * Copy here is the persona's framing of the story, never a new claim: every
 * fact a persona surface states must trace to a chapter in ./story.ts.
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators without a bundler.
 */
import type { ChapterId } from './story.ts';

export type PersonaSlug =
  | 'platform-engineer'
  | 'engineering-leader'
  | 'it-consultancy'
  | 'startup-founder'
  | 'security-and-governance-leader';

export interface Persona {
  slug: PersonaSlug;
  /** Title Case; the page heading and the deck cover. */
  name: string;
  /** Sentence case; who this person is, in one line they would recognize. */
  who: string;
  /** Sentence case; the wall they hit, in their words. Opens the page and the deck. */
  wall: string;
  /** The chapters in the order this person needs them; the first is the opening beat. */
  chapters: readonly ChapterId[];
  /** Sentence case; the two or three proof points that decide it for this person. */
  decidingProof: readonly string[];
  /** The door at the end: label (Title Case) and where it goes. */
  callToAction: { label: string; href: string };
}

export const PERSONAS: readonly Persona[] = [
  {
    slug: 'platform-engineer',
    name: 'Platform Engineer',
    who: 'You run the platform your developers and their coding agents build on, and you are the one who gets paged when something they created is wrong.',
    wall: 'Agents can already create infrastructure in your account. You cannot see what they made, what it costs, or whether it follows the rules you wrote down last quarter.',
    chapters: ['the-wall', 'your-rules-hold', 'verified-before-it-exists', 'every-deployment-leaves-a-record', 'what-planton-is', 'runs-where-you-decide', 'services-ship-from-git', 'bring-what-you-have', 'proof-it-works', 'start'],
    decidingProof: [
      'Your rules, your repository, your CI: budgets, protected environments, and a curated catalog hold no matter who asked.',
      'Every deploy is a record you can query, not a transcript you have to scroll.',
      'The desktop app is free, and a team plan is a card away when the team arrives.',
    ],
    callToAction: { label: 'Download Planton Desktop', href: '/features/desktop/download' },
  },
  {
    slug: 'engineering-leader',
    name: 'Engineering Leader',
    who: 'You sign for the cloud bill and answer for the outage. You do not want to open a console; you want to know what was deployed, what it costs, and that the rules held.',
    wall: 'Your team says the agents are making them faster. Nobody can show you the cost of what was created before it was created, or who approved it.',
    chapters: ['the-wall', 'verified-before-it-exists', 'your-rules-hold', 'every-deployment-leaves-a-record', 'who-it-is-for', 'what-planton-is', 'runs-where-you-decide', 'proof-it-works', 'start'],
    decidingProof: [
      'Cost, permissions, and controls are properties of what gets deployed, stamped before it exists.',
      'No new team: your platform engineer stays the user and hands you the proof.',
      'The record is what you get to see, and it does not depend on how often you ask.',
    ],
    callToAction: { label: 'Book a Demo', href: '/book-demo' },
  },
  {
    slug: 'it-consultancy',
    name: 'IT Consultancy',
    who: 'You stand up environments for clients who mandate a cloud your team may not know, and you hand the work back when the engagement ends.',
    wall: 'Every client starts from zero, and the Terraform you wrote for the last one does not fit the next.',
    chapters: ['the-wall', 'what-planton-is', 'runs-where-you-decide', 'verified-before-it-exists', 'every-deployment-leaves-a-record', 'services-ship-from-git', 'proof-it-works', 'start'],
    decidingProof: [
      'A client environment from a published Infra Chart, one organization per client.',
      'Repeatability is the product: what you built for one client becomes the template for the next.',
      'Hand back everything as manifests the client can keep running with the open-source CLI.',
    ],
    callToAction: { label: 'Start Free', href: '/signup' },
  },
  {
    slug: 'startup-founder',
    name: 'Startup Founder',
    who: 'You are shipping a product with a small team and no ops hire, and every hour on infrastructure is an hour not on the product.',
    wall: 'Your agent can build the infrastructure. You are not sure what it built, what it will cost next month, or how you will do it again for staging.',
    chapters: ['the-wall', 'what-planton-is', 'services-ship-from-git', 'verified-before-it-exists', 'runs-where-you-decide', 'proof-it-works', 'start'],
    decidingProof: [
      'Push to deploy, without writing pipeline files or a Dockerfile.',
      'The monthly cost, before it exists.',
      'Free to start, and nothing is redone when you become a team.',
    ],
    callToAction: { label: 'Start Free', href: '/signup' },
  },
  {
    slug: 'security-and-governance-leader',
    name: 'Security and Governance Leader',
    who: 'You own the posture of an estate you did not build, and the tools you have tell you what went wrong after it did.',
    wall: 'Agents create infrastructure faster than your scanners find the problems. Prevention has to happen where creation happens.',
    chapters: ['the-wall', 'your-rules-hold', 'verified-before-it-exists', 'every-deployment-leaves-a-record', 'how-it-compares', 'runs-where-you-decide', 'what-planton-is', 'proof-it-works', 'start'],
    decidingProof: [
      'Prevention at the write boundary: rules hold before anything exists, whoever asked.',
      'Every component states which controls it enforces, with evidence, and never calls itself compliant.',
      'A complement to your posture tools, never a replacement for them.',
    ],
    callToAction: { label: 'Book a Demo', href: '/book-demo' },
  },
] as const;

export function persona(slug: PersonaSlug): Persona {
  const found = PERSONAS.find((p) => p.slug === slug);
  if (!found) throw new Error(`no persona "${slug}"`);
  return found;
}
