/**
 * The Product pages as data: the platform engineer's half of the story, one
 * record per page under /product. Each record names the chapters it renders;
 * its proof points are the chapters' own sentences wherever a chapter states
 * the fact, and documented shipped behavior (the docs and the open-source
 * repository) where a page needs a fact the story only summarizes. Commands
 * are quoted from the documentation, never invented. Numbers come from
 * ./platform-stats.ts and prices from ./pricing.ts.
 *
 * Product pages lead with the user's doors (start free, download the desktop
 * app) because their reader installs; each page also names the Trust pages
 * where its claims are proven for the reader who signs.
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators.
 */
import type { DoorPair } from './doors.ts';
import { illustratedFooter, type Artifact, type HeroStrip, type ProofPoint, type Step } from './page-shapes.ts';
import { PLATFORM_COUNTS, PLATFORM_STATS } from './platform-stats.ts';
import { POSITIONING } from './positioning.ts';
import { chapter, type ChapterId } from './story.ts';

export interface ProductPage {
  /** The route under /product/. */
  path: string;
  /** The chapters this page renders, the first being the one its lede comes from. */
  chapters: readonly ChapterId[];
  /** The hub's one analogy, shown under the title; only the two hubs carry one. */
  analogy?: string;
  /** Sentence case; the page's opening claim. */
  lede: string;
  /** Sentence case; who this page is written for and what it answers, in one line. */
  forWhom: string;
  /** A strip of facts under the hero: the providers' marks, or the platform's counts. */
  strip?: HeroStrip;
  /** How it works, in the order a person experiences it; omitted where a page is a thing rather than a loop. */
  steps?: readonly Step[];
  points: readonly ProofPoint[];
  artifact: Artifact;
  /** The Trust pages where these claims are proven for the reader who signs. */
  provenAt: readonly string[];
  /** The two sibling Product pages to read next. */
  next: readonly string[];
  /** The doors this page leads with; the user's pair unless the page has a better first step. */
  doors?: DoorPair;
}

const what = chapter('what-planton-is');
const verified = chapter('verified-before-it-exists');
const rules = chapter('your-rules-hold');
const record = chapter('every-deployment-leaves-a-record');
const git = chapter('services-ship-from-git');
const bring = chapter('bring-what-you-have');
const runs = chapter('runs-where-you-decide');
const proof = chapter('proof-it-works');


export const PRODUCT_PAGES: readonly ProductPage[] = [
  {
    path: '/product/infra-hub',
    chapters: ['what-planton-is', 'verified-before-it-exists', 'your-rules-hold', 'every-deployment-leaves-a-record'],
    analogy: POSITIONING.infraHub.analogy,
    lede: POSITIONING.infraHub.line,
    forWhom: 'For the platform engineer who publishes the templates, and the developer or coding agent who deploys them.',
    steps: [
      { title: 'Describe', text: 'Say what you need: in the console, from the CLI, or through your coding agent. Every component is a typed schema, so a wrong field fails before it touches your cloud.' },
      { title: 'Compose', text: 'Watch it compose on a live canvas: the components, how they connect, what each needs from the others.' },
      { title: 'Verify', text: 'See the monthly cost with its coverage stated, the least-privilege policy, and the controls each component enforces. Nothing exists yet.' },
      { title: 'Deploy', text: 'One stack job, kept: the exact configuration, the cost fact, the verdicts, who approved, and what exists afterward.' },
      { title: 'Publish', text: 'Publish it as an Infra Chart, a template your team redeploys into the next environment. A prompt cannot be redeployed; a chart can.' },
    ],
    points: [
      { label: 'The Monthly Cost, With Its Coverage', text: verified.proof[0] },
      { label: 'Least-Privilege Permissions', text: verified.proof[2] },
      { label: 'A Budget That Pauses a Deploy', text: rules.proof[0] },
      { label: 'One Record per Change', text: record.proof[1] },
    ],
    artifact: {
      kind: 'record',
      title: 'An environment, composed',
      rows: [
        { label: 'components', value: 'vpc \u00b7 eks cluster \u00b7 rds postgres \u00b7 s3 bucket \u00b7 iam role \u00b7 dns record' },
        { label: 'verified cost', value: '~$412/mo est.\u00a0\u00b7\u00a05 of 6 components priced, 1 usage-based' },
        { label: 'permissions', value: 'least privilege \u00b7 derived from the 6 components composed' },
        { label: 'controls', value: '6 of 6 components carry a control profile with evidence' },
        { label: 'budget', value: 'within the environment\u2019s $600/mo est. deployment budget' },
        { label: 'published as', value: 'Infra Chart production-baseline \u00b7 redeployable into the next environment' },
      ],
      footer: illustratedFooter({ figures: true }),
    },
    provenAt: ['/trust/verified-before-deploy', '/trust/rules-and-approvals', '/trust/the-record'],
    next: ['/product/service-hub', '/product/catalog'],
  },
  {
    path: '/product/service-hub',
    chapters: ['services-ship-from-git', 'your-rules-hold', 'every-deployment-leaves-a-record'],
    analogy: POSITIONING.serviceHub.analogy,
    lede: POSITIONING.serviceHub.line,
    forWhom: 'For the developer who wants a push to mean deployed, and the platform engineer who wants that push to obey the environment\u2019s gate.',
    steps: [
      { title: 'Connect', text: 'Connect a repository once. From then on every push carries the record.' },
      { title: 'Push', text: 'Every push is built, containerized, and deployed. No pipeline YAML, no Dockerfile required.' },
      { title: 'Promote', text: 'Promotion follows the order your environments declare; a protected environment carries a manual gate and refuses self-approval.' },
      { title: 'Written Back', text: 'The result lands in GitHub as checks and deployments, so the person who pushed learns what happened where they already are.' },
    ],
    points: [
      { label: 'Nothing to Write for the Build', text: 'The build runs through Buildpacks from the repository as it is: no pipeline YAML and no Dockerfile to write or keep.' },
      { label: 'Promotion in Your Environments\u2019 Order', text: git.proof[1] },
      { label: 'A Protected Environment Refuses', text: git.proof[2] },
      { label: 'Usage Never Becomes an Invoice Line', text: 'Runner minutes are never billed. Usage is recorded for abuse protection and nothing else.' },
    ],
    artifact: {
      kind: 'record',
      title: 'A deployment, from a push',
      rows: [
        { label: 'trigger', value: 'push to main \u00b7 4f2a9c1 \u00b7 "checkout: retry on timeout"' },
        { label: 'build', value: 'buildpacks \u00b7 no Dockerfile \u00b7 2m 41s est.' },
        { label: 'promotion', value: 'dev \u2192 staging \u2192 prod (protected: manual gate)' },
        { label: 'approval', value: 'prod gate resolved by a reviewer \u00b7 never the initiator' },
        { label: 'github', value: 'check passed \u00b7 deployment active' },
        { label: 'record', value: 'one run, kept with its log and its revision' },
      ],
      footer: illustratedFooter({ figures: true }),
    },
    provenAt: ['/trust/rules-and-approvals', '/trust/the-record'],
    next: ['/product/infra-hub', '/product/coding-agents'],
  },
  {
    path: '/product/coding-agents',
    chapters: ['what-planton-is', 'the-wall', 'your-rules-hold'],
    lede: what.proof[2],
    forWhom: 'For the developer with Cursor, Claude Code, or Codex already open, and the platform engineer who wants that agent inside the same rules as everyone else.',
    steps: [
      { title: 'Install the Skills', text: 'One command installs the two Planton skills into every coding agent on your machine, in the open Agent Skills format.' },
      { title: 'Facts, Not Memory', text: 'Without the skills, an agent writes infrastructure from memory. With them, every fact comes from the component\u2019s reference page at answer time.' },
      { title: 'The Same Door as Everyone', text: 'The agent deploys through the same door as a person in the console or a script on the CLI: the same catalog, the same rules, the same record.' },
      { title: 'The Platform Over MCP', text: 'An agent that wants to build, apply, deploy, or read your organization\u2019s estate directly reaches those operations through the hosted MCP server, with the CLI or without it.' },
    ],
    points: [
      { label: 'Validated Before It Is Handed Over', text: 'The agent wires resources by reference instead of pasting IDs, validates the manifest, tells you the monthly cost, and asks before it applies.' },
      { label: 'The Same List, the Same Refusals', text: rules.proof[2] },
      { label: 'Lines the Agent Never Crosses', text: 'The skills say so in plain words: never change anything without asking, and never work outside your repository.' },
      { label: 'One Assistant, Every Surface', text: 'The same skills power the Planton Assistant in Planton Desktop and the web console, so your agent and the platform\u2019s own assistant work from the same instructions.' },
    ],
    artifact: {
      kind: 'commands',
      title: 'Teach your agent',
      file: {
        name: '.cursor/mcp.json',
        body: '{\n  "mcpServers": {\n    "planton": {\n      "type": "http",\n      "url": "https://mcp.planton.ai/",\n      "headers": { "Authorization": "Bearer ${env:PLANTON_API_KEY}" }\n    }\n  }\n}',
      },
      commands: ['npx skills add plantonhq/skills', 'brew install plantonhq/tap/planton'],
      label: 'Copy the install commands',
      caption: 'The first line installs the two skills into every agent on the machine; the second is the CLI they drive; the file above is the optional MCP door. All three are from',
      source: { label: 'the Coding Agents guide', href: '/docs/coding-agents' },
    },
    provenAt: ['/trust/rules-and-approvals', '/trust/verified-before-deploy'],
    next: ['/product/cli', '/product/infra-hub'],
    doors: { primary: 'codingAgentsDocs', secondary: 'hosted' },
  },
  {
    path: '/product/cli',
    chapters: ['runs-where-you-decide', 'your-rules-hold', 'every-deployment-leaves-a-record'],
    lede: 'Everything Planton does, from your terminal: validate a manifest before it touches your cloud, deploy one component or a whole directory in dependency order, stream the stack job as it runs, install an Infra Chart, and bring in what already exists.',
    forWhom: 'For the engineer who lives in a shell, and the pipeline that runs without one.',
    steps: [
      { title: 'Install', text: 'One Homebrew line on macOS; direct downloads for Linux and Windows.' },
      { title: 'Write the Manifest', text: 'The shape you already know: apiVersion, kind, metadata, spec. Validation catches a wrong field in seconds, before anything is created.' },
      { title: 'Apply', text: 'Deploy one manifest or a directory of them in dependency order, the way you already apply Kubernetes manifests, on every cloud.' },
      { title: 'Watch the Record', text: 'The stack job streams as it runs. The same event stream drives the console and the audit log, so every surface tells one story.' },
    ],
    points: [
      { label: 'The Exit Path Is the Same CLI', text: runs.proof[1] },
      { label: 'The Same Rules at Every Door', text: rules.claim },
      { label: 'The Same Statement on Every Surface', text: 'The verified cost, the permission policy, and the controls print in your terminal word for word as they show in the console: one statement, stamped by the server, on every surface.' },
      { label: 'Inside GitHub Actions', text: 'The install and login actions put the same CLI inside a workflow, so a pipeline deploys the way a person does.' },
    ],
    artifact: {
      kind: 'commands',
      title: 'Install, write, apply',
      file: {
        name: 'postgres.yaml',
        body: 'apiVersion: kubernetes.planton.dev/v1alpha1\nkind: KubernetesPostgres\nmetadata:\n  name: my-first-postgres\nspec:\n  namespace:\n    value: my-first-postgres\n  createNamespace: true\n  container:\n    replicas: 1\n    diskSize: 1Gi',
      },
      commands: ['brew install plantonhq/tap/planton', 'planton apply -f postgres.yaml'],
      label: 'Copy the CLI commands',
      caption: 'Install, write the manifest above, apply. Validation fails a wrong field before anything touches your cloud. The manifest and the command are from',
      source: { label: 'the repository README', href: 'https://github.com/plantonhq/planton#the-cli' },
    },
    provenAt: ['/trust/the-record', '/trust/rules-and-approvals'],
    next: ['/product/coding-agents', '/product/open-source'],
    doors: { primary: 'cliDocs', secondary: 'hosted' },
  },
  {
    path: '/product/catalog',
    chapters: ['proof-it-works', 'verified-before-it-exists', 'your-rules-hold'],
    lede: `${PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} component kinds across ${PLATFORM_STATS.CLOUD_PROVIDER_COUNT} providers, each a typed schema with its own fact sheet: what it costs, which controls it enforces, and the least permissions its runner needs.`,
    strip: 'providers',
    forWhom: 'For the platform engineer choosing what their organization may deploy, and the reviewer who wants the fact sheet before the deploy.',
    points: [
      { label: 'A Fact Sheet on Every Component', text: verified.proof[3] },
      { label: 'Priced From a Named Release', text: verified.proof[1] },
      { label: 'Curated to What You Allow', text: rules.proof[2] },
      { label: 'A Schema, Not a Wiki Page', text: 'Each component is a Protocol Buffer definition in the Kubernetes Resource Model shape with field-level validation, and SDKs generated in Go, Python, TypeScript, and Java.' },
      { label: 'Open Source, Every Module', text: runs.proof[1] },
    ],
    artifact: {
      kind: 'record',
      title: 'A component\u2019s fact sheet',
      rows: [
        { label: 'kind', value: 'AwsRdsCluster \u00b7 aws.planton.dev/v1alpha1' },
        { label: 'schema', value: 'typed spec \u00b7 field-level validation \u00b7 SDKs in 4 languages' },
        { label: 'cost', value: 'priced \u00b7 3 presets \u00b7 catalog release named' },
        { label: 'controls', value: `6 of ${PLATFORM_STATS.CONTROL_COUNT} enforced \u00b7 evidence attached \u00b7 absence stated` },
        { label: 'permissions', value: 'least privilege \u00b7 validated against the provider\u2019s own inventory' },
        { label: 'modules', value: 'Pulumi \u00b7 Terraform \u00b7 Apache 2.0' },
      ],
      footer: illustratedFooter({ note: 'The kind is real; the figures are examples.' }),
    },
    provenAt: ['/trust/verified-before-deploy', '/trust/security-posture'],
    next: ['/product/infra-hub', '/product/open-source'],
    doors: { primary: 'catalogBrowser', secondary: 'hosted' },
  },
  {
    path: '/product/import',
    chapters: ['bring-what-you-have', 'every-deployment-leaves-a-record'],
    lede: bring.claim,
    forWhom: 'For the team that already has an account full of infrastructure and no intention of redeploying it.',
    steps: [
      { title: 'Adopt', text: 'Register the resource in Planton without deploying anything. Nothing in your cloud is touched, and the page says so: adopted, not yet deployed.' },
      { title: 'Import', text: 'From the resource\u2019s page or the CLI, name the cloud resource that already exists. Import only writes state; the cloud resource itself is never modified.' },
      { title: 'Verify', text: 'A wrong import fails before it lands. Import never writes a configuration it did not apply.' },
      { title: 'Continue', text: 'From then on the resource carries the same record as everything Planton created: every change one stack job, kept.' },
    ],
    points: [
      { label: 'Proven in a Live Round Trip', text: bring.proof[2] },
      { label: 'Nothing Is Touched Until You Say', text: 'Adopting runs no job and changes nothing. The import itself only writes state; the cloud resource is never modified.' },
      { label: 'From the Terminal Too', text: 'The CLI imports with the same verification, and a dry run prints the native command it would run without creating a job.' },
      { label: 'Tagged Like Everything Else', text: record.proof[3] },
    ],
    artifact: {
      kind: 'record',
      title: 'An import, verified',
      rows: [
        { label: 'resource', value: 'AwsS3Bucket \u00b7 media-archive \u00b7 adopted, not yet deployed' },
        { label: 'operation', value: 'state import \u00b7 identifier derived by recipe' },
        { label: 'verify', value: 'refresh, then preview \u00b7 0 changes pending' },
        { label: 'state', value: 'written once, only what was applied' },
        { label: 'record', value: 'stack job kept \u00b7 queryable by resource, environment, time' },
      ],
      footer: illustratedFooter(),
    },
    provenAt: ['/trust/the-record'],
    next: ['/product/infra-hub', '/product/cli'],
  },
  {
    path: '/product/open-source',
    chapters: ['runs-where-you-decide', 'proof-it-works'],
    lede: runs.proof[1],
    forWhom: 'For the engineer who wants to read what will run in their account before it runs, and the one planning the exit before the entry.',
    points: [
      { label: 'The Catalog', text: `${PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} component kinds across ${PLATFORM_STATS.CLOUD_PROVIDER_COUNT} providers, each with its Pulumi and Terraform module, its cost fact sheet, its control posture with evidence, and its least-privilege permissions, all in one repository.` },
      { label: 'The Charts', text: `${PLATFORM_STATS.INFRA_CHART_COUNT} Infra Charts: whole environments composed from those components and installed in one command.` },
      { label: 'The CLI and the Engine', text: 'The open-source CLI validates manifests and runs the modules that ship with every component. It is the same engine the platform drives.' },
      { label: 'Machine-Checked Facts', text: 'The cost data is priced from pinned price books, the control posture carries framework crosswalks, and the permission manifests are validated against the providers\u2019 own published inventories.' },
      { label: 'Counted, Not Claimed', text: proof.proof[0] },
    ],
    artifact: {
      kind: 'record',
      title: 'plantonhq/planton',
      rows: [
        { label: 'license', value: 'Apache 2.0 \u00b7 the name and logo are trademarks' },
        { label: 'component kinds', value: `${PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} across ${PLATFORM_STATS.CLOUD_PROVIDER_COUNT} providers` },
        { label: 'infra charts', value: PLATFORM_STATS.INFRA_CHART_COUNT },
        { label: 'controls', value: `${PLATFORM_STATS.CONTROL_COUNT} in ${PLATFORM_STATS.CONTROL_CATEGORY_COUNT} categories \u00b7 ${PLATFORM_STATS.FRAMEWORK_CROSSWALK_COUNT} framework crosswalks` },
        { label: 'engines', value: 'Pulumi \u00b7 OpenTofu and Terraform' },
        { label: 'in production since', value: PLATFORM_STATS.IN_PRODUCTION_SINCE },
      ],
      footer: `Kinds, charts, controls, and crosswalks counted from the open-source repository on ${PLATFORM_COUNTS.countedOn}.`,
    },
    provenAt: ['/trust/your-cloud-your-keys'],
    next: ['/product/catalog', '/product/cli'],
    doors: { primary: 'github', secondary: 'hosted' },
  },
] as const;

export function productPage(path: string): ProductPage {
  const found = PRODUCT_PAGES.find((p) => p.path === path);
  if (!found) throw new Error(`"${path}" has no record in src/data/product.ts`);
  return found;
}
