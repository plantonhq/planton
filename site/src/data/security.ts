/**
 * The Security page: the questions an engineering leader or a security
 * reviewer asks before anyone connects a cloud, each answered in plain words,
 * with the real thing behind the answer and a way to check it.
 *
 * Every sentence names the story chapter it mirrors and the claims-ledger row
 * it rests on (planton-platform `company/sales/_ledger/`, "Security answers"
 * and "Governance"). Components render these records and add no sentence of
 * their own; scripts/generate-llms.mjs renders the same records into the
 * page's Markdown twin, so a person and their coding agent read one page.
 *
 * Relative imports carry their `.ts` extension so Node can execute this file
 * for the build-time generators without a bundler.
 */
import type { ChapterId } from './story.ts';

/** A real file, policy or command output, shown flat and readable. */
export interface SecurityCode {
  kind: 'code';
  /** What the reader is looking at and where it comes from. */
  label: string;
  language: 'yaml' | 'json' | 'shell';
  lines: readonly string[];
}

/** A small table of facts (a permission and its use, a role and whether it may act). */
export interface SecurityTable {
  kind: 'table';
  label: string;
  columns: readonly string[];
  rows: readonly (readonly string[])[];
}

export type SecurityArtifact = SecurityCode | SecurityTable;

/** How a reader proves an answer without taking the page's word for it. */
export interface SecurityCheck {
  text: string;
  /** A command they can run, shown in monospace. */
  command?: string;
  /** A page that shows the same thing. */
  href?: string;
}

export interface SecuritySection {
  /** The anchor and the Markdown twin's heading id. */
  id: string;
  /** Sentence case; the reader's own question, answered. */
  title: string;
  chapter: ChapterId;
  paragraphs: readonly string[];
  artifact?: SecurityArtifact;
  check?: SecurityCheck;
}

/** The example connection every artifact uses, so one name runs through the page. */
const EXAMPLE = { org: 'acme', connection: 'aws-prod' } as const;

export const SECURITY = {
  eyebrow: 'Security',
  // ch 8 (runs where you decide): the page's promise. Ledger: "No stored cloud keys", "They cut it off in their own cloud, any time".
  headline: 'One role in your cloud. One App in your GitHub.',
  lede: 'In your cloud and your GitHub, Planton works through nothing else. Remove both, and Planton is out.',
  sections: [
    {
      id: 'what-planton-can-touch',
      title: 'What Planton can touch',
      chapter: 'verified-before-it-exists',
      paragraphs: [
        // Ledger: "The role can do only what the kinds they choose need" (scoped by action; never Owner, Editor or Contributor).
        'Only what you let it deploy. When you connect a cloud, you create one role in your own account: an IAM role on AWS, a service account on Google Cloud, an app registration on Azure. You choose what Planton may deploy, everything in its catalog for that cloud or only the types of infrastructure you pick, and the role gets only the actions those types publish. It is never Owner, Editor or Contributor.',
        // Ledger: "Every infrastructure change is a plan…", "Before anything exists: a bill preview, the runner's exact permission policy".
        'Your coding agent never holds cloud credentials. It writes a file describing the infrastructure and asks Planton to deploy it. Before anything is created, you see the plan: what it will cost each month and the exact permissions it will use.',
      ],
      artifact: {
        kind: 'code',
        label: 'The permissions one type of infrastructure, an S3 bucket, publishes in Planton’s open-source catalog (the first of six statements)',
        language: 'yaml',
        // Read from plantonhq/planton catalog/aws/awss3bucket/iac/permissions.yaml, statement ManageBucket.
        lines: [
          '# catalog/aws/awss3bucket/iac/permissions.yaml',
          '- sid: ManageBucket',
          '  actions:',
          '    - s3:CreateBucket',
          '    - s3:DeleteBucket',
          '    - s3:ListBucket',
          '    - s3:GetBucketTagging',
          '    - s3:PutBucketTagging',
          '    - s3:GetBucketObjectLockConfiguration',
          '    - s3:PutBucketObjectLockConfiguration',
          '  resources:',
          "    - arn:aws:s3:::*",
        ],
      },
      check: {
        text: 'Every type’s permission list, in the open-source catalog',
        href: 'https://github.com/plantonhq/planton/tree/main/catalog',
      },
    },
    {
      id: 'no-stored-keys',
      title: 'No stored keys',
      chapter: 'runs-where-you-decide',
      paragraphs: [
        // Ledger: "No stored cloud keys…", "The role trusts only one named connection of one organization"; docs security/what-your-cloud-trusts.
        'Planton stores no cloud keys. Your cloud trusts one issuer, https://planton.ai, and only for this one connection of your organization. To act, Planton presents a token that expires in 15 minutes and receives cloud access that expires within an hour.',
      ],
      artifact: {
        kind: 'code',
        label: `The trust the setup writes in AWS, for a connection named ${EXAMPLE.connection}`,
        language: 'json',
        // The statement AwsConnectionScripts.trustPolicy renders, with the example account and connection.
        lines: [
          '{',
          '  "Effect": "Allow",',
          '  "Principal": {',
          '    "Federated": "arn:aws:iam::123456789012:oidc-provider/planton.ai"',
          '  },',
          '  "Action": "sts:AssumeRoleWithWebIdentity",',
          '  "Condition": {',
          '    "StringEquals": {',
          '      "planton.ai:aud": "sts.amazonaws.com",',
          `      "planton.ai:sub": "planton:v1:org:${EXAMPLE.org}:conn:${EXAMPLE.connection}:provider:aws"`,
          '    }',
          '  }',
          '}',
        ],
      },
      check: {
        text: 'Read the issuer your cloud trusts:',
        command: 'curl https://planton.ai/.well-known/openid-configuration',
      },
    },
    {
      id: 'where-the-work-runs',
      title: 'Where the work runs, and what Planton keeps',
      chapter: 'runs-where-you-decide',
      paragraphs: [
        // docs runner/planton-hosted-runners ("What a Planton-hosted runner can access").
        'Each deploy runs on a Planton-hosted runner that holds no standing access. It gets credentials for that one job, and they expire within an hour.',
        // Ledger: "Their infrastructure state can live in their own bucket…", "Their secrets can stay in their own…" (both start on Planton's side).
        'Your infrastructure’s state and your app’s secrets (a database password, an API token) start in Planton. You can keep both in your own cloud instead: the state in your own bucket under your own key, the secrets in your own AWS Secrets Manager, Google Secret Manager or Azure Key Vault.',
      ],
      artifact: {
        kind: 'table',
        label: 'Where each thing lives',
        columns: ['What', 'Where it lives'],
        // Ledger: "No stored cloud keys…", the state and secrets rows (their conditions: both start on Planton's side).
        rows: [
          ['Cloud keys', 'Nowhere. Planton stores none'],
          ['Each deploy', 'A Planton-hosted runner, with credentials for that one job'],
          ['Infrastructure state', 'Planton, under a key only its runners hold; or your own bucket, under your own key'],
          ['App secrets', 'Planton’s vault; or your own AWS Secrets Manager, Google Secret Manager or Azure Key Vault'],
        ],
      },
      check: {
        text: 'How state moves to your own bucket',
        href: '/docs/connections/state-backends',
      },
    },
    {
      id: 'what-the-github-app-sees',
      title: 'What the GitHub App sees',
      chapter: 'services-ship-from-git',
      paragraphs: [
        // wiki product.integrations.github; the App's permissions as GitHub reports them (ledger, "A GitHub connection's Verify…", 2026-11-25).
        'Only the repositories you choose on GitHub’s own install screen. GitHub gives Planton tokens that expire within an hour, so there is nothing to paste or rotate. It reads your code to build it. What it writes is a pull request it opens for you (its own branch and commits), a check on each build, a record of each deploy, and a re-run or cancel of a GitHub Actions run you ask for.',
      ],
      artifact: {
        kind: 'table',
        label: 'What the App asks GitHub for, and why',
        columns: ['Permission', 'Access', 'What Planton does with it'],
        // GithubConnectionCapability, the product's one list of what an App is granted.
        rows: [
          ['Contents', 'Read and write', 'Clone your code to build it; write the branch of a pull request it opens for you'],
          ['Pull requests', 'Read and write', 'Build a preview for each pull request; open pull requests for you'],
          ['Checks', 'Write', 'Report each build and preview as a check on its commit'],
          ['Deployments', 'Write', 'Record each deploy as a GitHub deployment'],
          ['Actions', 'Read and write', 'Show your GitHub Actions runs; re-run or cancel them for you'],
          ['Metadata', 'Read', 'List the repositories you installed it on'],
        ],
      },
      check: {
        text: 'The same permissions, on the App’s page at GitHub',
        href: 'https://github.com/apps/planton-ai',
      },
    },
    {
      id: 'who-approves-what',
      title: 'Who approves what',
      chapter: 'your-rules-hold',
      paragraphs: [
        // Ledger: "Seven built-in roles, including a Release Approver who cannot deploy what they approve"; outside a protected environment no second approval is asked ("Production can require…", only on a protected environment).
        'You decide who may deploy to each environment, with built-in roles such as Developer, Platform Engineer and Release Approver. In an environment you haven’t protected, the people you let deploy, and their coding agents, deploy without waiting for anyone else.',
        // Ledger: "In a protected environment…", "Production can require a second person's approval…" (Team plan).
        'On the Team plan, you can protect an environment. Then every infrastructure change and every deploy into it waits for a teammate’s approval, and the person who asked can’t give it.',
      ],
      artifact: {
        kind: 'table',
        label: 'Who can approve a change in a protected environment',
        columns: ['Who', 'Can approve'],
        // Ledger: "Neither the Assistant nor a coding agent can approve a run for its person"; SeparationOfDuty.
        rows: [
          ['The person who asked', 'No'],
          ['A coding agent, theirs or anyone’s', 'No'],
          ['A teammate you gave approval on that environment', 'Yes'],
        ],
      },
    },
    {
      id: 'the-record',
      title: 'The record',
      chapter: 'every-deployment-leaves-a-record',
      paragraphs: [
        // Ledger: "Every change names the person and the coding agent, kept for the life of the organization; access changes… export as CSV".
        'Every infrastructure change is kept for the life of your organization: its exact configuration, who asked for it, the coding agent that made it, what it cost and who approved it. Every change to who can do what is kept too, and you can export it.',
      ],
      artifact: {
        kind: 'table',
        label: 'What each infrastructure change keeps',
        columns: ['Kept', 'What it is'],
        // wiki product.infra-hub.infra-job ("every Infra Job is kept with its frozen configuration, who triggered it, its cost fact and its approvals"); the activity feed names the agent.
        rows: [
          ['Configuration', 'The exact manifest it applied, frozen'],
          ['Asked by', 'The person who asked for it'],
          ['Coding agent', 'The agent that made it, such as Claude Code'],
          ['Cost', 'The plan’s monthly cost'],
          ['Approved by', 'Who approved it, in a protected environment'],
          ['Revision', 'Its numbered revision of the stack'],
        ],
      },
      check: {
        text: 'Read what Claude Code changed this week:',
        // The CLI's own example (`planton activity --help`).
        command: 'planton activity --agent claude --since 7d',
      },
    },
    {
      id: 'cut-planton-off',
      title: 'Cut Planton off',
      chapter: 'runs-where-you-decide',
      paragraphs: [
        // Ledger: "They cut it off in their own cloud, any time", "Leaving never touches their cloud…" (state Planton holds is deleted with the organization).
        'Delete the connection’s trust in your cloud and Planton holds nothing it can use there; uninstall the App in GitHub and it can’t reach your code. Your infrastructure keeps running, and the files that describe it, and the open-source modules that deploy them, stay yours. If Planton holds your state or secrets, move them to your own cloud first: they are deleted with your organization.',
      ],
      artifact: {
        kind: 'table',
        label: 'What to delete in your cloud',
        columns: ['Cloud', 'Delete'],
        // docs security/what-your-cloud-trusts, "How to Revoke".
        rows: [
          ['AWS', 'The connection’s IAM role'],
          ['Google Cloud', 'The connection’s OIDC provider'],
          ['Azure', 'The connection’s federated credential on its app registration'],
        ],
      },
      check: {
        text: 'Print the exact removal commands for this connection:',
        command: `planton connect aws cleanup ${EXAMPLE.connection}`,
      },
    },
  ] satisfies readonly SecuritySection[],
  /** The docs that go deeper, in the order a reviewer would open them, beside the page's two doors. */
  more: {
    title: 'Go deeper',
    // The reviewer's next step is a question, not a sign-up: the address Planton already publishes for support (legal pages, pricing FAQ).
    contact: { text: 'Questions from your security review:', address: 'support@planton.ai' },
    docs: [
      { label: 'What Your Cloud Trusts', href: '/docs/security/what-your-cloud-trusts' },
      { label: 'Keyless Cloud Connections', href: '/docs/connections/keyless-cloud-connections' },
      { label: 'Planton-Hosted Runners', href: '/docs/runner/planton-hosted-runners' },
      { label: 'Audit Trails', href: '/docs/security/audit-trails' },
    ],
  },
} as const;
