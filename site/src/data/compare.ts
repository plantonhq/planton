/**
 * The Compare page as data: chapter 11 of the story, one page. Three kinds
 * of tool sit near Planton, and each is described by what it does, never by
 * name. A category record carries what the category does, what the reader
 * keeps when they run both, what Planton does at the same moment (the
 * chapters' own proof sentences, quoted by index), and the page that proves
 * it; the template shows that page's record beside the words, so every claim
 * on this page has the same illustrated record beside it that its proving
 * page shows. The questions are the ones people actually ask when they
 * compare, answered in the chapters' words. The template renders this record
 * and states nothing of its own.
 *
 * What this file never carries, by chapter 11's own rule: a vendor's name, a
 * feature table with a competitor column, the labels "DevSecOps" or
 * "FinOps". A sentence appears on this page once; where a question would
 * repeat a category's point, it says something the category did not. Every
 * sentence names the chapter it mirrors.
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators without a bundler.
 */
import { START_DOORS, type DoorPair } from './doors.ts';
import type { ProofPoint } from './page-shapes.ts';
import { chapter } from './story.ts';

const what = chapter('what-planton-is');
const verified = chapter('verified-before-it-exists');
const rules = chapter('your-rules-hold');
const record = chapter('every-deployment-leaves-a-record');
const compares = chapter('how-it-compares');

/** One kind of tool near Planton, described by what it does. */
export interface ComparisonCategory {
  /** The section's anchor, and the target of the hero's route to it. */
  id: string;
  /** Title Case; the category, never a vendor. */
  title: string;
  /** Title Case; the short name the hero routes by ("Governance Platforms"). */
  shortTitle: string;
  /** Sentence case; what this kind of tool does, stated as a fact and not a slight. */
  theyDo: string;
  /** Sentence case; what the reader keeps when they run both, and what Planton hands it. The section's largest sentence after its title. */
  both: string;
  /** What Planton does at the same moment: Title Case labels over the chapters' sentences. */
  planton: readonly ProofPoint[];
  /** The registered page where Planton's side is proven; its record is shown beside these words. */
  provenAt: string;
}

/** A question a person comparing tools asks, answered in a chapter's words, with the page that answers it in full. */
export interface ComparisonQuestion {
  question: string;
  answer: string;
  readMore: string;
}

export interface ComparePage {
  /** Title Case; the promise, not a label. */
  headline: string;
  /** Sentence case; the page's opening claim, two sentences. */
  lede: string;
  /** Sentence case; who this page is written for, in one line. */
  forWhom: string;
  /** The one idea the page turns on, with the page whose record proves it. */
  difference: { claim: string; provenAt: string };
  categories: readonly ComparisonCategory[];
  questions: readonly ComparisonQuestion[];
  /** Two registered pages to read next. */
  next: readonly string[];
  doors: DoorPair;
}

export const COMPARE: ComparePage = {
  // The story's spine, as the headline.
  headline: 'Proof at Creation, Not Observation After',
  // Chapter 11, with the spine's three verbs.
  lede: 'Every change Planton deploys is priced, held to your rules, and recorded before it exists. Governance tools observe after the fact, infrastructure-as-code tools work on Terraform you still write, and portals catalog what you have; Planton sits beside all three, and you keep them.',
  forWhom: 'For the platform engineer who already runs some of these, and the leader who signed for them.',
  difference: {
    // The story's own framing of the spine ("How to read this").
    claim: 'The tools near Planton either watch what exists and report what they find, or run the Terraform you still write. Planton owns the moment infrastructure is created and proves it as it goes, and every deployment leaves this record behind.',
    provenAt: '/trust/the-record',
  },
  categories: [
    {
      id: 'governance-and-posture',
      title: 'Governance and Posture Platforms',
      shortTitle: 'Governance Platforms',
      // Chapter 11
      theyDo: 'Cost, security, and operations dashboards over your whole estate. They read what exists, detect what is wrong, and tell you what to fix, after it was created.',
      // Chapter 11, proof 0
      both: 'Keep the posture tool for the estate you already have and for whatever Planton did not create. Planton hands it a smaller problem: everything that went through Planton arrives priced, within budget, inside your rules, and stamped with the controls it enforces. A complement, never a replacement.',
      planton: [
        { label: 'Priced Before It Exists', text: verified.proof[0] },
        { label: 'Controls Stated with Evidence', text: verified.proof[3] },
        { label: 'Stamped on the Record', text: record.proof[0] },
      ],
      provenAt: '/trust/security-posture',
    },
    {
      id: 'infrastructure-as-code',
      title: 'Infrastructure-as-Code Tools and Their Guardrails',
      shortTitle: 'Infrastructure-as-Code Tools',
      // Chapter 11
      theyDo: 'Orchestrators that run the Terraform your team writes, and guardrail tools that check it at plan time, field name by field name. Some estimate cost at plan time. The writing is still yours to do.',
      // Chapter 11, proof 1; chapter 8, proof 1
      both: 'Keep the orchestrator for the Terraform you already run outside Planton. Your Terraform stays yours: the modules are open-source Terraform and Pulumi, and what you already run is adopted, not rewritten.',
      planton: [
        // Chapter 11's claim; the validation the Coding Agents page already documents.
        { label: 'Typed Self-Service, Not Authoring', text: 'Every component is a typed schema over an open-source module. A person or an agent writes a short manifest, the schema validates it before anything touches your cloud, and the module that runs is the same one every time.' },
        // Chapter 11, proof 1, second sentence
        { label: 'One Vocabulary, Not a Hundred Field Names', text: 'Every covered component reports its controls against the same fixed list of 17, so what you check is one vocabulary, not each kind\u2019s field names.' },
        { label: 'Every Door Obeys the Same Rules', text: rules.claim },
      ],
      provenAt: '/product/open-source',
    },
    {
      id: 'developer-portals',
      title: 'Developer Portals and Service Catalogs',
      shortTitle: 'Developer Portals',
      // Chapter 11
      theyDo: 'A catalog of what your organization has: services, owners, documentation, scorecards. A portal tells you what you have; someone still has to build and staff everything it points at.',
      // Chapter 11
      both: 'Keep the portal as the front page of your engineering organization and let it point at Planton for the part it cannot do: creating the infrastructure and shipping the service.',
      planton: [
        { label: 'Deploys What You Need', text: compares.proof[2] },
        { label: 'A Design Becomes a Template', text: what.proof[0] },
        { label: 'Every Push Becomes a Deployment', text: what.proof[1] },
      ],
      provenAt: '/product/infra-hub',
    },
  ],
  questions: [
    {
      // Chapters 2 and 11. The engines are named as the CLI runs them: the Terraform module runs under OpenTofu or Terraform, the other module under Pulumi.
      question: 'Why not just Terraform?',
      answer: 'Planton does not compete with Terraform; it runs the modules. Every component ships with a pre-written, tested open-source module in Terraform, which OpenTofu or Terraform runs, and another for Pulumi, and you choose the engine without changing your manifest. What changes is who writes what: the manifest is yours, short and typed; the module is Planton\u2019s, and it does the rest.',
      readMore: '/product/open-source',
    },
    {
      // Chapter 11
      question: 'Is this another abstraction layer I will be fighting in six months?',
      answer: 'Each provider keeps its full native configuration; Planton does not pretend one cloud\u2019s network is another\u2019s. What is consistent is the structure and the workflow: one manifest shape for every component, one door every request goes through, and one fixed list that controls are stated against. The manifest you write is a plain document you keep in your own repository.',
      readMore: '/product/catalog',
    },
    {
      // Chapter 5
      question: 'Is it a set of operators reconciling my cluster?',
      answer: 'No. Nothing runs in your cluster watching your resources. Every change is one stack job: it plans, pauses at the gates you set, applies from a runner in your own network, and is kept and queryable with the exact configuration embedded. A control loop cannot naturally pause between the plan and the apply; a job can, and that pause is where your approvals live. The trade is that nothing self-heals: a job runs when a person, a push, or an agent asks.',
      readMore: '/trust/the-record',
    },
  ],
  next: ['/trust', '/product/open-source'],
  doors: START_DOORS,
};
