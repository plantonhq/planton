/**
 * The Trust section as data: five pages, each the composition of one or two
 * chapters for the reader who signs rather than the one who deploys. A page
 * record names the chapter whose claim opens it, the proof points it lays
 * out (each a sentence from the story), the honesty statements it makes
 * (the platform's own grammar, said out loud, because a security reviewer
 * trusts a vendor that under-claims), and the doors it offers next.
 *
 * Every sentence here is either a chapter's sentence or a rule the platform
 * already follows on its own screens (the control-posture grammar). Nothing
 * is claimed for the first time on a Trust page.
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators without a bundler.
 */
import type { RecordArtifact } from './page-shapes.ts';
import { chapter, type ChapterId } from './story.ts';

export interface TrustPoint {
  /** Title Case label. */
  label: string;
  /** The chapter's sentence. */
  text: string;
}

export interface TrustPage {
  /** The route under /trust/. */
  path: string;
  /** The chapter this page renders. */
  chapter: ChapterId;
  /**
   * Sentence case; the sentence that opens the page. Defaults to the
   * chapter's claim; set when the page's angle differs from the chapter's
   * (the security posture page renders chapter 3 from the reviewer's side).
   */
  lede?: string;
  /** Sentence case; who this page is written for and what it answers, in one line. */
  forWhom: string;
  points: readonly TrustPoint[];
  /** The proof beside the claims: what the product shows, as a record of labeled facts. */
  artifact: RecordArtifact;
  /** Sentence case; the honesty statements this page makes about its own words. */
  honesty: readonly string[];
  /** The two sibling Trust pages to read next, always within the section; their titles come from the registry. */
  next: readonly string[];
  /** Where the artifact is real: the product page that shows it. */
  seeIt: { label: string; href: string };
}

const verified = chapter('verified-before-it-exists');
const rules = chapter('your-rules-hold');
const record = chapter('every-deployment-leaves-a-record');
const runs = chapter('runs-where-you-decide');

export const TRUST_PAGES: readonly TrustPage[] = [
  {
    path: '/trust/verified-before-deploy',
    chapter: 'verified-before-it-exists',
    forWhom: 'For the person who signs the cloud bill and wants to know the number before it is committed.',
    points: [
      { label: 'The Monthly Cost, With Its Coverage', text: verified.proof[0] },
      { label: 'The Catalog Release and the Delta', text: verified.proof[1] },
      { label: 'Least-Privilege Permissions', text: verified.proof[2] },
      { label: 'The Controls Each Component Enforces', text: verified.proof[3] },
      { label: 'One Statement on Every Surface', text: verified.proof[4] },
    ],
    honesty: [
      'The figure is a verified monthly cost from the catalog\u2019s pricing rules, not a bill. Its coverage is always beside it: exact with line items, a range, or plainly unpriced.',
      'A component that the pricing rules cannot price says so. A zero never stands in for unknown.',
      'Planton does not tell you what you will save. It tells you what this will cost before it exists.',
    ],
    artifact: {
      kind: 'record',
      title: 'Before deploy',
      rows: [
        { label: 'verified cost', value: `~$172/mo est.${'\u00a0\u00b7\u00a0'}5 of 6 components priced, 1 usage-based` },
        { label: 'catalog release', value: '2026.09.2 \u00b7 prices verified against provider documents' },
        { label: 'against today', value: '+$16/mo est. \u00b7 the load balancer is new' },
        { label: 'permissions', value: 'least-privilege policy \u00b7 14 actions \u00b7 ready to download' },
        { label: 'controls', value: 'encryption at rest \u00b7 encryption in transit \u00b7 no public exposure' },
        { label: 'uncovered', value: 'none of the 6 components is without a control profile' },
      ],
      footer: 'An illustration of the shape. Figures marked est. are examples; a real deploy carries its own.',
    },
    next: [
      '/trust/rules-and-approvals',
      '/trust/the-record',
    ],
    seeIt: { label: 'See It in Infra Hub', href: '/product/infra-hub' },
  },
  {
    path: '/trust/rules-and-approvals',
    chapter: 'your-rules-hold',
    forWhom: 'For the platform team that writes the rules once, and the leader who wants to know they held.',
    points: [
      { label: 'Deployment Budgets', text: rules.proof[0] },
      { label: 'Protected Environments', text: rules.proof[1] },
      { label: 'A Curated Catalog', text: rules.proof[2] },
      { label: 'Secrets Stay Secret', text: rules.proof[3] },
    ],
    honesty: [
      'These are the rules that ship today. A rule over the content of a configuration (\u201cno public data stores in production\u201d) is not one of them yet, and this page does not pretend it is.',
      'Only creation is governed by catalog curation. Updates, redeploys, and destroys of what already exists are never blocked, so disabling a kind can never strand live infrastructure.',
      'A pause is a pause for a person. Nothing here approves itself.',
    ],
    artifact: {
      kind: 'record',
      title: 'Deploy paused',
      rows: [
        { label: 'requested by', value: 'coding agent, on behalf of s.rao' },
        { label: 'environment', value: 'prod \u00b7 protected \u00b7 budget $250/mo' },
        { label: 'verified cost', value: '~$312/mo est.' },
        { label: 'verdict', value: 'over budget by ~$62/mo \u00b7 paused for a decision' },
        { label: 'who may approve', value: 'anyone with approve access on prod \u00b7 never the requester' },
        { label: 'resolution', value: 'approved by a.patel \u00b7 2026-09-17 09:14 \u00b7 \u201cthe replica is intentional\u201d' },
      ],
      footer: 'An illustration of the shape. Figures marked est. are examples; a real deploy carries its own.',
    },
    next: [
      '/trust/the-record',
      '/trust/verified-before-deploy',
    ],
    seeIt: { label: 'See It in Infra Hub', href: '/product/infra-hub' },
  },
  {
    path: '/trust/the-record',
    chapter: 'every-deployment-leaves-a-record',
    forWhom: 'For anyone who has to answer, months later, what was deployed, by whom, at what cost, and who said yes.',
    points: [
      { label: 'The Configuration, Frozen', text: record.proof[0] },
      { label: 'Kept and Queryable', text: record.proof[1] },
      { label: 'One Stream, Every Surface', text: record.proof[2] },
      { label: 'Tagged in Your Cloud', text: record.proof[3] },
    ],
    honesty: [
      'The record is a fact about each deployment. It is not a compliance dashboard and it does not declare a deployment audit-ready for any framework.',
      'Drift detection is not shipped. The record tells you what was deployed; it does not yet tell you what changed behind Planton\u2019s back.',
      'This page states no retention period, because the platform does not publish one: every job is kept and queryable today, and a retention policy is not a setting you can change.',
    ],
    artifact: {
      kind: 'record',
      title: 'Deploy record',
      rows: [
        { label: 'deploy', value: 'production environment \u00b7 prod \u00b7 succeeded' },
        { label: 'requested by', value: 'coding agent, on behalf of s.rao' },
        { label: 'approved by', value: 'a.patel \u00b7 2026-09-17 09:14 \u00b7 \u201cthe replica is intentional\u201d' },
        { label: 'configuration', value: 'embedded at creation; never changes' },
        { label: 'verified cost', value: '~$312/mo est. \u00b7 catalog 2026.09.2' },
        { label: 'phases', value: 'init \u00b7 refresh \u00b7 preview \u00b7 apply \u00b7 capture' },
        { label: 'snapshot', value: '7 resources \u00b7 tagged planton.ai/environment=prod' },
      ],
      footer: 'An illustration of the shape. Every deploy leaves one of these.',
    },
    next: [
      '/trust/security-posture',
      '/trust/rules-and-approvals',
    ],
    seeIt: { label: 'See It in Infra Hub', href: '/product/infra-hub' },
  },
  {
    path: '/trust/security-posture',
    chapter: 'verified-before-it-exists',
    lede: 'Every component states which technical controls it enforces, with evidence, and how those controls serve the frameworks your reviewers ask about. Nothing is called compliant, because nothing should be.',
    forWhom: 'For the security reviewer who will read every word here looking for the one that over-claims.',
    points: [
      { label: 'Seventeen Controls, Six Categories', text: 'The platform has one vocabulary of technical controls: data protection, network exposure, identity and access, observability, resilience, and change safety. Every claim a component makes cites one of them.' },
      { label: 'Evidence, Per Component', text: verified.proof[3] },
      { label: 'Four Framework Crosswalks', text: 'The HIPAA Security Rule, SOC 2 Trust Services Criteria, FedRAMP Moderate, and the CIS AWS Foundations Benchmark are mapped onto the control vocabulary, requirement by requirement, quoted from the published source.' },
      { label: 'Computed, Never Stored', text: 'A component\u2019s framework posture is computed when you look at it, from its controls and the crosswalks. No verdict is ever stored on a component.' },
      { label: 'Least-Privilege Runners', text: 'The deploy runs under an identity built from the component kind\u2019s own permissions file: exactly the actions that kind needs to create and manage itself, and nothing else.' },
    ],
    artifact: {
      kind: 'record',
      title: 'Control profile \u00b7 S3 bucket',
      rows: [
        { label: 'encryption at rest', value: 'enforced \u00b7 every new object is encrypted at rest by default; an unencrypted object is not representable' },
        { label: 'customer-managed keys', value: 'exposed \u00b7 set encryption.kms_key_id' },
        { label: 'encryption in transit', value: 'enforced \u00b7 TLS on every request' },
        { label: 'no public exposure', value: 'enforced \u00b7 public access blocked at the bucket' },
        { label: 'HIPAA 164.312(a)(2)(iv)', value: 'served by encryption at rest, customer-managed keys' },
        { label: 'SOC 2 CC6.1', value: 'served by no public exposure, role-based access control' },
        { label: 'verdict', value: 'none' },
      ],
      footer: 'The shape of a control profile and two crosswalk rows; evidence is quoted from the component\u2019s profile. There is no verdict that says compliant, because a component enforces controls and is never called compliant.',
    },
    honesty: [
      'Posture never becomes \u201ccompliant.\u201d The correct sentence is: this component enforces these controls, which serve these requirements. SOC 2 reports on an organization\u2019s controls, not a product\u2019s; HIPAA compliance is a property of an entity\u2019s program.',
      '\u201cFedRAMP authorized\u201d is never said of any component, chart, or deployment. The crosswalk is evidence an assessment consumes.',
      'Requirements only organizational process can serve (incident response, physical security, personnel screening) are left out of the crosswalks on purpose, and the page says which ones rather than blurring them.',
      'A component without a control profile is honestly uncovered, and every surface says so.',
    ],
    next: [
      '/trust/your-cloud-your-keys',
      '/trust/the-record',
    ],
    seeIt: { label: 'The Control Catalog on GitHub', href: 'https://github.com/plantonhq/planton/tree/main/catalog/_compliance' },
  },
  {
    path: '/trust/your-cloud-your-keys',
    chapter: 'runs-where-you-decide',
    lede: 'Planton does not need a long-lived key to your cloud. Keyless connections exchange a short-lived token for temporary credentials at the moment of use; where a credential is stored, it is a managed secret resolved just in time. Every module is open source, and your manifests leave with you.',
    forWhom: 'For the team that will not hand a vendor a long-lived credential, and wants to know what leaves their account.',
    points: [
      { label: 'Keyless Connections', text: runs.proof[0] },
      { label: 'Open Source, and the Way Out', text: runs.proof[1] },
      { label: 'Secrets on the Desktop', text: runs.proof[2] },
      { label: 'Hosted, Self-Hosted, or Desktop: One Model', text: runs.proof[3] },
    ],
    artifact: {
      kind: 'record',
      title: 'Connection \u00b7 AWS production',
      rows: [
        { label: 'mode', value: 'keyless' },
        { label: 'stored credential', value: 'none' },
        { label: 'at deploy time', value: 'short-lived identity token \u00b7 scoped to this connection \u00b7 exchanged by AWS for temporary credentials' },
        { label: 'trust', value: 'an IAM role in your account that trusts Planton\u2019s issuer \u00b7 yours to revoke' },
        { label: 'rotation', value: 'nothing to rotate' },
        { label: 'modules', value: 'open source \u00b7 Apache 2.0 \u00b7 github.com/plantonhq/planton' },
      ],
      footer: 'The shape of a keyless connection. Broker-issued credentials (your own vault) are the other keyless mode.',
    },
    honesty: [
      'Your account, your keys, your state, your bill. Planton holds the record and the manifests; the infrastructure is yours in your provider\u2019s console the whole time.',
      'Where a connection stores a credential, it is a managed secret: resolved on the runner at the moment of use, never copied into a manifest, a record, or a log.',
      'The exit is concrete: your manifests, and the open-source CLI that deploys them. It is not a badge.',
    ],
    next: [
      '/trust/verified-before-deploy',
      '/trust/security-posture',
    ],
    seeIt: { label: 'The Open-Source Modules', href: '/product/open-source' },
  },
] as const;

export function trustPage(path: string): TrustPage {
  const found = TRUST_PAGES.find((p) => p.path === path);
  if (!found) throw new Error(`no Trust page at "${path}"`);
  return found;
}
