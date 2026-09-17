/**
 * The Distributions pages as data: chapter 8 ("runs where you decide"), one
 * record per shape the platform runs in. Every record renders the same
 * chapter; what differs is the shape's own facts, taken from the
 * documentation of that shape (the self-hosting guide, the licensing
 * lifecycle, the runner architecture) and never from a guess. Seats and days
 * come from ./pricing.ts; the desktop distribution's page is the desktop
 * landing, which has its own data (./desktop-download.ts, ./positioning.ts).
 *
 * The mobile companion is not a distribution here: it is not shipped, and the
 * website shows only shipped behavior.
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators.
 */
import type { DoorPair } from './doors.ts';
import type { Artifact, ProofPoint } from './page-shapes.ts';
import { COMMUNITY_SEAT_LIMIT, EVALUATION_DAYS, FREE_TIER_SEATS, SELF_SERVE_SEAT_CEILING } from './pricing.ts';
import { chapter } from './story.ts';

export interface DistributionPage {
  /** The route under /distributions/. */
  path: string;
  /** Sentence case; the shape in one or two sentences. */
  lede: string;
  /** Sentence case; who chooses this shape and why, in one line. */
  forWhom: string;
  points: readonly ProofPoint[];
  artifact: Artifact;
  /** Sentence case; what stays yours in this shape, said plainly. */
  yours: readonly string[];
  /** The other distribution pages to read next. */
  next: readonly string[];
  doors: DoorPair;
}

const runs = chapter('runs-where-you-decide');

export const DISTRIBUTION_PAGES: readonly DistributionPage[] = [
  {
    path: '/distributions/hosted',
    lede: `The control plane runs at planton.ai and holds your records; everything it deploys runs in your cloud account under your own connections. Free for up to ${FREE_TIER_SEATS} seats with no card.`,
    forWhom: 'For the team that wants nothing to run and nothing to upgrade, and still wants every deploy in its own account under its own keys.',
    points: [
      { label: 'What Lives at planton.ai', text: 'The control plane: your organization\u2019s records (the manifests you declared, every stack job with its cost fact and verdicts), the console, and the identity server. Secrets are held as references and resolved on the runner at the moment of use; they can live in your own secrets manager.' },
      { label: 'What Lives in Your Account', text: 'Everything Planton deploys, under connections you own, tagged with its organization, environment, kind, and id. The bill is your cloud invoice.' },
      { label: 'Keyless Connections', text: runs.proof[0] },
      { label: 'Where the Work Runs', text: 'A deploy executes on a runner: one Planton operates, your own laptop with the cloud sign-in already on it, or a standing appliance inside your network. Credentials are resolved at the moment of use: keyless where the cloud supports it, otherwise from the secrets manager just in time.' },
      { label: 'Usage Never Becomes an Invoice Line', text: 'Runner minutes are never billed. Usage is recorded for abuse protection and nothing else.' },
      { label: 'Nothing Redone When You Grow', text: runs.proof[3] },
      { label: 'Per Seat, Self-Serve', text: `Teams pay per seat. Up to ${SELF_SERVE_SEAT_CEILING} seats, nobody talks to sales: the plan is a card away.` },
    ],
    artifact: {
      kind: 'record',
      title: 'A hosted organization',
      rows: [
        { label: 'control plane', value: 'planton.ai' },
        { label: 'cloud account', value: 'yours \u00b7 AWS 123456789012 \u00b7 keyless connection' },
        { label: 'runner', value: 'Planton\u2019s, your laptop, or an appliance in your network' },
        { label: 'state', value: 'Planton\u2019s store by default \u00b7 or a bucket you name: S3, GCS, or Azure Blob' },
        { label: 'bill', value: 'your cloud invoice \u00b7 runner minutes never billed' },
        { label: 'seats', value: `${FREE_TIER_SEATS} free \u00b7 then per seat \u00b7 no sales call up to ${SELF_SERVE_SEAT_CEILING}` },
      ],
      footer: 'The shape of a hosted organization; the identifiers are examples.',
    },
    yours: [
      'The cloud account, and every resource in it, tagged with its organization, environment, kind, and id.',
      'The credentials: keyless where the cloud supports it, otherwise held in the secrets manager and resolved at the moment of use.',
      'The state: held for you by default, or in a bucket you name; moving it to your own backend is one state operation.',
      'The manifests. If you leave, the open-source CLI keeps deploying them.',
    ],
    next: ['/distributions/self-hosted'],
    doors: { primary: 'hosted', secondary: 'pricing' },
  },
  {
    path: '/distributions/self-hosted',
    lede: `The whole platform on your own Kubernetes cluster, installed with two Helm commands and run by an operator. Free for up to ${COMMUNITY_SEAT_LIMIT} seats; a ${EVALUATION_DAYS}-day evaluation of the full edition needs only an email.`,
    forWhom: 'For the platform team that wants Planton inside the same trust boundary as the infrastructure it manages.',
    points: [
      { label: 'Two Commands, Everything', text: 'The operator first, then the platform at the release you name. From that one resource the operator runs the whole stack: the control plane, the console, the identity server, the secrets manager, the databases, and an in-cluster runner. No license key, no admin account, no database, no values file.' },
      { label: 'A License That Verifies Offline', text: 'A license key is a signed proof of purchase the deployment verifies entirely offline: no network call, no phone-home. The community edition underneath is free: the whole platform, with one limit on how many people can join.' },
      { label: 'Expiry Never Bricks', text: 'The seat limit applies to people joining, never to the people already there. An install that holds more people than the community edition allows keeps every one of them working; a lapsed key never reaches into a running deployment.' },
      { label: 'Or Through Infrastructure as Code', text: 'The same two steps exist as catalog resources, so an agent or a pipeline installs Planton the way it installs everything else: two manifests, applied in order, through OpenTofu or Pulumi.' },
      { label: 'Several Plantons, One Cluster', text: 'Platforms are namespaced and one operator serves the whole cluster, so staging and production, or one platform per team, run side by side.' },
      { label: 'Guided From the Desktop', text: 'Planton Desktop offers the same install as a guided experience: pick a cluster from your kubeconfig, preflight it, choose the front door, and watch the install converge.' },
    ],
    artifact: {
      kind: 'commands',
      title: 'Install',
      file: {
        name: 'planton.yaml',
        body: 'apiVersion: kubernetes.planton.dev/v1alpha1\nkind: KubernetesPlantonPlatform\nmetadata:\n  name: planton\n  annotations:\n    planton.dev/provisioner: tofu\nspec:\n  namespace:\n    value: planton\n  createNamespace: true\n  version: <release>',
      },
      commands: [
        'helm install planton-operator oci://ghcr.io/plantonhq/charts/planton-operator \\\n  --namespace planton --create-namespace',
        'helm install planton oci://ghcr.io/plantonhq/charts/planton \\\n  --namespace planton --set platform.spec.version=<release>',
      ],
      label: 'Copy the install commands',
      caption: 'Two Helm installs, the operator first; or declare the platform as the manifest above and apply it through the CLI. The guide lists the published releases. From',
      source: { label: 'the Self-Hosting guide', href: '/docs/self-hosting' },
    },
    yours: [
      'The cluster, and everything the operator runs in it: records, secrets, deployment history, cloud credentials. Nothing leaves.',
      'The release: you choose when to upgrade, and to which published release.',
      'The license, verified against embedded public keys, with no call home.',
      'The manifests. If you leave, the open-source CLI keeps deploying them.',
    ],
    next: ['/distributions/hosted'],
    doors: { primary: 'selfHostedDocs', secondary: 'evaluation' },
  },
] as const;

export function distributionPage(path: string): DistributionPage {
  const found = DISTRIBUTION_PAGES.find((p) => p.path === path);
  if (!found) throw new Error(`"${path}" has no record in src/data/distributions.ts`);
  return found;
}
