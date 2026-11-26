import type { CodeArtifact, Check, TableArtifact } from './artifact.ts';
import type { CloudId } from './clouds.ts';
import { PLATFORM_STATS } from './platform-stats.ts';
import { FREE_TIER_SEATS } from './pricing.ts';
import { STAND } from './story.ts';

/**
 * Home, one section per record, in the order the page tells the story (the
 * story's spine, v3.0: a company with no DevOps team runs its own cloud).
 * Each section names the chapter it mirrors; components render these records
 * and state nothing. Home is built one or two sections per session, so the
 * records here are the sections that exist today, in page order.
 */
export const HOMEPAGE = {
  title: 'Run Your Cloud Without a DevOps Team',
  description:
    // ch 1 (a new cloud account, and no DevOps team): the stand, then the first action.
    `${STAND} Paste one prompt into your coding agent, and it takes your cloud account to a running app in your own cloud.`,
  // ch 1: the stand. The team the diagram beside it draws is the rest of the hero's story.
  headline: STAND,
  intro:
    // ch 1 (one prompt to a running app), then ch 6 (every push ships through one path).
    'Paste one prompt into your coding agent, and your app runs in your own cloud account, even a brand-new one. Then every push ships through one path, with your team’s approvals.',
  // ch 12 (start): the price answered at the first door. The free tier never asks for a card (the claims ledger, the price row).
  caption: `Free for up to ${FREE_TIER_SEATS} seats. Never asks for a card.`,
  /**
   * ch 1: the founder's measure of the product, proven by the setup prompt
   * itself. Paste to running app is proven on Google Cloud Run (the claims
   * ledger, the Setup Prompt row), so no other cloud is named beside it.
   */
  firstApp: {
    // A non-breaking hyphen keeps "brand-new" whole when the heading wraps.
    title: 'From a brand\u2011new cloud account to a running app.',
    intro:
      'These are the exact words your coding agent reads. Planton writes them for you when you sign up, with your own one\u2011time code.',
    promptLabel: 'An example prompt, for Neha at Slotwren, a demo company',
    you: {
      title: 'You',
      steps: [
        'Open your app’s repository in Claude Code, Codex or Cursor, and paste the prompt.',
        'Say yes before anything changes your cloud or costs money.',
        'Install Planton’s GitHub App, so every push after this ships the same way.',
      ],
    },
    agent: {
      title: 'Your Agent',
      steps: [
        'Installs the Planton CLI and signs in as you.',
        'Connects your cloud without storing a key.',
        'Gets your app running in your own cloud account.',
        'Tells you what it set up, with a link to each thing.',
      ],
    },
  },
  /**
   * ch 2 (proof: the agent line), ch 10 (proof: the modules' proving record),
   * ch 11 (proof: what deploys was written and proven before anyone asked).
   * The artifacts are AGENT_RUNS': what a real coding agent wrote on each cloud.
   */
  infrastructure: {
    title: 'Ask your coding agent for infrastructure. The code that deploys it is already written.',
    intro:
      // ch 2, proof (the agent line); the claims ledger, "Your coding agent writes manifests" and "validated offline".
      'Your agent picks from Planton\u2019s catalog and writes a manifest that says what you want, then checks it before anything exists. It never writes the Terraform or Pulumi: the module that deploys each part is already written, in the open, and proven in a real cloud account.',
    // ch 8 (runs where you decide): the resources land in the customer's account (the claims ledger, the own-cloud row).
    next: 'When you say so, Planton deploys it into your own cloud account with those modules.',
    // The picker that sets the visitor's cloud for this section and every one below it.
    pickerLabel: 'Your Cloud',
  },
  // ch 12 (start): what the visitor keeps, then the first act.
  close: {
    title: 'Your app, running in your own cloud, on your own bill.',
    // ch 8 (runs where you decide): what the agent builds stays in the customer's account (the claims ledger, the own-cloud row).
    text: 'What your agent builds stays in your cloud account, and your cloud provider bills you directly.',
    note: `Free for up to ${FREE_TIER_SEATS} seats. Never asks for a card.`,
  },
} as const;

/**
 * The setup prompt a new organization's Home hands over, as the server writes
 * it (`SetupAgentPrompt.render` in the platform repository), for Slotwren, a
 * demo company. The page shows the words and never a working prompt: the real
 * one carries a personal single-use sign-in code, so `{code}` renders as a
 * marked placeholder and there is no copy button. When the server's wording
 * changes, this example changes with it.
 */
export const SETUP_PROMPT_EXAMPLE = {
  code: '<her one-time code>',
  /** The sentence the page sets in full ink: the agent asks before it costs anything. */
  emphasis:
    'Ask me before anything that changes my cloud or costs money; connecting a cloud account counts.',
  lines: [
    "Set up Planton for this repository and take it to its first deployment, in my organization slotwren on planton.ai. I'm neha@slotwren.com.",
    '',
    '1. Install the Planton CLI and sign it in as me. The code works once, within 15 minutes:',
    '   curl -fsSL https://planton.ai/install.sh -o /tmp/planton-install.sh && sh /tmp/planton-install.sh',
    '   ~/.planton/bin/planton login --grant {code} --instance planton.ai',
    '   Call the CLI by that full path in every command.',
    '2. Fetch https://planton.ai/agents.md with curl, read it as written, and follow it. Skip what you already did in step 1.',
    "3. Follow the planton skill's first-deployment guide for this repository in slotwren. Ask me before anything that changes my cloud or costs money; connecting a cloud account counts. Tick each Getting Started step as you finish it with ~/.planton/bin/planton setup done <step> (run it with no step to list them).",
    '4. When you finish, tell me what you set up, with a link to each thing.',
  ],
} as const;

/** One cloud's real run, rendered by Home's infrastructure section when the visitor picks that cloud. */
export interface AgentRun {
  askLabel: string;
  /** The person's ask, excerpted with marked ellipses; the whole ask is in the run's record. */
  ask: string;
  /** The set's main file, verbatim. */
  manifest: CodeArtifact;
  /** The check the agent ran on every file, as the CLI printed it. */
  validation: CodeArtifact;
  /**
   * Every kind in the set. Each kind's catalog/<provider>/<kind>/e2e/profile.yaml
   * in this repository marks it green on both engines: its lane deploys it into
   * a real test account, checks it, destroys it and checks it clean.
   */
  proven: TableArtifact;
  /** How a reader proves it: the module, the test that deploys and destroys it, the fields offline. */
  checks: readonly Check[];
  reach: string;
}

/**
 * What a real coding agent wrote when a person at each cloud's demo company
 * asked it for a Postgres database: headless Claude Code with the released
 * Planton CLI and skills, signed in to nothing (2026-11-25; every ask, command,
 * file and check is kept in the platform repository's website research).
 * Slotwren carries Google Cloud, 4Corner Pay AWS, Northmere 1887 Azure. When a
 * kind's fields or the CLI's output change, the run is repeated and its record
 * replaced, never edited by hand.
 */
export const AGENT_RUNS: Record<CloudId, AgentRun> = {
  aws: {
    askLabel:
      'Megan, a software engineer on payments at 4Corner Pay, a demo company, to Claude Code:',
    ask: '“…We’re adding stored payment receipts, so our payments API needs a Postgres database on AWS for our staging environment… Don’t sign in and don’t deploy anything yet; I’ll review what you write first.”',
    manifest: {
      kind: 'code',
      label:
        'infrastructure/payments-db/database/payments-db.yaml, written by Claude Code from Megan’s ask (one of six files)',
      language: 'yaml',
      wrap: true,
      lines: [
        '# Staging Postgres for the payments API (stored payment receipts).',
        '# Single-AZ, small Graviton instance, private subnets only, encrypted at rest.',
        '# The master password is generated and held by AWS Secrets Manager -- it never',
        '# appears in this file or in IaC state. Apps read it via the',
        '# masterUserSecretArn output.',
        'apiVersion: aws.planton.dev/v1alpha1',
        'kind: AwsRdsInstance',
        'metadata:',
        '  name: staging-payments-db',
        '  org: 4cornerpay',
        '  env: staging',
        'spec:',
        '  region: us-east-1',
        '  subnetIds:',
        '    - valueFrom:',
        '        kind: AwsSubnet',
        '        name: staging-payments-private-a',
        '        fieldPath: status.outputs.subnet_id',
        '    - valueFrom:',
        '        kind: AwsSubnet',
        '        name: staging-payments-private-b',
        '        fieldPath: status.outputs.subnet_id',
        '  securityGroupIds:',
        '    - valueFrom:',
        '        kind: AwsSecurityGroup',
        '        name: staging-payments-db-sg',
        '        fieldPath: status.outputs.security_group_id',
        '',
        '  # Engine and size',
        '  engine: postgres',
        '  instanceClass: db.t4g.small',
        '  allocatedStorageGb: 20',
        '  maxAllocatedStorageGb: 100',
        '  storageType: gp3',
        '',
        '  # Database and credentials',
        '  dbName: payments',
        '  username: payments_admin',
        '  manageMasterUserPassword: true',
        '',
        '  # Security',
        '  storageEncrypted: true',
        '  publiclyAccessible: false',
        '',
        '  # Availability and data protection',
        '  multiAz: false',
        '  backupRetentionPeriod: 7',
        '  backupWindow: "07:00-08:00"',
        '  maintenanceWindow: "sun:08:30-sun:09:30"',
        '  copyTagsToSnapshot: true',
        '  skipFinalSnapshot: false',
        '  finalSnapshotIdentifier: staging-payments-db-final',
        '  deletionProtection: true',
        '',
        '  # Operations',
        '  performanceInsightsEnabled: true',
        '  autoMinorVersionUpgrade: true',
        '  engineLifecycleSupport: open-source-rds-extended-support-disabled',
      ],
    },
    validation: {
      kind: 'code',
      label: 'The check it ran on all six files, before anything exists',
      language: 'shell',
      wrap: true,
      lines: [
        '$ planton validate infrastructure/payments-db/database/payments-db.yaml',
        '✓ manifest is valid',
        '$ planton validate infrastructure/payments-db/network/private-subnet-a.yaml',
        '✓ manifest is valid',
        '$ planton validate infrastructure/payments-db/network/private-subnet-b.yaml',
        '✓ manifest is valid',
        '$ planton validate infrastructure/payments-db/network/vpc.yaml',
        '✓ manifest is valid',
        '$ planton validate infrastructure/payments-db/security/payments-api-sg.yaml',
        '✓ manifest is valid',
        '$ planton validate infrastructure/payments-db/security/payments-db-sg.yaml',
        '✓ manifest is valid',
      ],
    },
    proven: {
      kind: 'table',
      label:
        'What deploys them: each kind’s own open-source module, deployed and torn down in a real AWS account by Planton’s tests',
      columns: ['Kind', 'Proven On'],
      rows: [
        ['AwsRdsInstance', 'Terraform, Pulumi'],
        ['AwsVpc', 'Terraform, Pulumi'],
        ['AwsSubnet', 'Terraform, Pulumi'],
        ['AwsSecurityGroup', 'Terraform, Pulumi'],
      ],
    },
    checks: [
      {
        text: 'Read the module that deploys AwsRdsInstance',
        href: 'https://github.com/plantonhq/planton/tree/main/catalog/aws/awsrdsinstance/iac',
      },
      {
        text: 'Read the test that deploys AwsRdsInstance into a real AWS account and destroys it',
        href: 'https://github.com/plantonhq/planton/tree/main/catalog/aws/awsrdsinstance/e2e',
      },
      {
        text: 'List every field the agent chose from, offline:',
        command: 'planton explain AwsRdsInstance',
      },
    ],
    // ch 10: the count, only from platform-stats.ts; the example names what a kind is.
    reach: `AwsRdsInstance is one of ${PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} kinds in the catalog, across ${PLATFORM_STATS.CLOUD_PROVIDER_COUNT} providers, from AWS, Google Cloud and Azure to Cloudflare and Kubernetes.`,
  },
  gcp: {
    askLabel: 'Neha, co-founder and CTO of Slotwren, a demo company, to Claude Code:',
    ask: '“…We’re adding stored bookings, so the API needs a Postgres database on Google Cloud for our staging environment… Don’t sign in and don’t deploy anything yet; I’ll review what you write first.”',
    manifest: {
      kind: 'code',
      label:
        'infrastructure/booking-db/instance.yaml, written by Claude Code from Neha’s ask (one of three files)',
      language: 'yaml',
      wrap: true,
      lines: [
        '# Cloud SQL for PostgreSQL instance backing the booking API in staging.',
        '# Sized for staging: one dedicated vCPU, single zone, small SSD that grows on demand.',
        'apiVersion: gcp.planton.dev/v1alpha1',
        'kind: GcpCloudSql',
        'metadata:',
        '  name: staging-booking-db',
        '  org: slotwren',
        '  env: staging',
        'spec:',
        '  instanceName: slotwren-staging-booking-db',
        '  region: us-central1',
        '  databaseEngine: POSTGRESQL',
        '  databaseVersion: POSTGRES_16',
        '  edition: ENTERPRISE',
        '  tier: db-custom-1-3840',
        '  availabilityType: ZONAL',
        '  disk:',
        '    type: PD_SSD',
        '    sizeGb: 10',
        '    autoResize: true',
        '    autoResizeLimit: 50',
        '  # No network block: public IP with no authorized networks, so the only way in',
        '  # is the Cloud SQL Auth Proxy / connectors (IAM-checked, always TLS).',
        "  # Cloud Run's built-in Cloud SQL connection uses exactly that path.",
        '  connectorEnforcement: REQUIRED',
        '  backup:',
        '    enabled: true',
        '    startTime: "08:00"',
        '    pointInTimeRecoveryEnabled: true',
        '    transactionLogRetentionDays: 7',
        '    retainedBackups: 7',
        '  maintenanceWindow:',
        '    day: 7',
        '    hour: 9',
        '    updateTrack: stable',
        '  insightsConfig:',
        '    queryInsightsEnabled: true',
        "  # The postgres superuser's password; the API connects as its own user (app-user.yaml).",
        '  rootPassword: $secret/@staging/booking-db-root-password',
        '  deletionProtection: true',
      ],
    },
    validation: {
      kind: 'code',
      label: 'The check it ran on all three files, before anything exists',
      language: 'shell',
      wrap: true,
      lines: [
        '$ planton validate infrastructure/booking-db/instance.yaml',
        '✓ manifest is valid',
        '$ planton validate infrastructure/booking-db/database.yaml',
        '✓ manifest is valid',
        '$ planton validate infrastructure/booking-db/app-user.yaml',
        '✓ manifest is valid',
      ],
    },
    proven: {
      kind: 'table',
      label:
        'What deploys them: each kind’s own open-source module, deployed and torn down in a real Google Cloud account by Planton’s tests',
      columns: ['Kind', 'Proven On'],
      rows: [
        ['GcpCloudSql', 'Terraform, Pulumi'],
        ['GcpCloudSqlDatabase', 'Terraform, Pulumi'],
        ['GcpCloudSqlUser', 'Terraform, Pulumi'],
      ],
    },
    checks: [
      {
        text: 'Read the module that deploys GcpCloudSql',
        href: 'https://github.com/plantonhq/planton/tree/main/catalog/gcp/gcpcloudsql/iac',
      },
      {
        text: 'Read the test that deploys GcpCloudSql into a real Google Cloud account and destroys it',
        href: 'https://github.com/plantonhq/planton/tree/main/catalog/gcp/gcpcloudsql/e2e',
      },
      {
        text: 'List every field the agent chose from, offline:',
        command: 'planton explain GcpCloudSql',
      },
    ],
    // ch 10: the count, only from platform-stats.ts; the example names what a kind is.
    reach: `GcpCloudSql is one of ${PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} kinds in the catalog, across ${PLATFORM_STATS.CLOUD_PROVIDER_COUNT} providers, from AWS, Google Cloud and Azure to Cloudflare and Kubernetes.`,
  },
  azure: {
    askLabel: 'Arjun, a software engineer at Northmere 1887, a demo company, to Claude Code:',
    ask: '“…We’re adding saved quotes, so our quote service needs a Postgres database on Azure for our dev environment… Don’t sign in and don’t deploy anything yet; I’ll review what you write first.”',
    manifest: {
      kind: 'code',
      label:
        'infrastructure/quote-db.yaml, written by Claude Code from Arjun’s ask (one of two files)',
      language: 'yaml',
      wrap: true,
      lines: [
        "# Postgres for the quote service's saved quotes (dev).",
        'apiVersion: azure.planton.dev/v1alpha1',
        'kind: AzurePostgresqlFlexibleServer',
        'metadata:',
        '  name: dev-quote-db',
        '  org: northmere1887',
        '  env: dev',
        'spec:',
        '  region: uksouth',
        '  resourceGroup:',
        '    valueFrom:',
        '      kind: AzureResourceGroup',
        '      name: dev-quote-service-rg',
        '      fieldPath: status.outputs.resource_group_name',
        '  # Globally unique across Azure: becomes <serverName>.postgres.database.azure.com',
        '  serverName: northmere1887-dev-quote-db',
        '  version: "16"',
        "  # Burstable 1 vCPU / 2 GiB: the cheapest tier, right for dev. No HA (Burstable can't have it).",
        '  skuName: B_Standard_B1ms',
        '  storageMb: 32768',
        '  autoGrowEnabled: false',
        '  backupRetentionDays: 7',
        '  geoRedundantBackupEnabled: false',
        '  administratorLogin: quoteadmin',
        '  # The password never lives in this file: create the secret before deploying',
        "  # (planton secret set quote-db-admin-password --string '<value>' --env dev).",
        '  administratorPassword:',
        '    value: $secret/@dev/quote-db-admin-password',
        '  # Public endpoint, but the only firewall rule admits Azure-internal services',
        '  # (0.0.0.0-0.0.0.0), not the internet. Add a rule for a developer IP if needed.',
        '  publicNetworkAccessEnabled: true',
        '  firewallRules:',
        '    - name: allow-azure-services',
        '      startIpAddress: 0.0.0.0',
        '      endIpAddress: 0.0.0.0',
        '  databases:',
        '    - name: saved_quotes',
        '  tags:',
        '    team: motor-quote-and-buy',
        '    service: quote-service',
      ],
    },
    validation: {
      kind: 'code',
      label: 'The check it ran on both files, before anything exists',
      language: 'shell',
      wrap: true,
      lines: [
        '$ planton validate infrastructure/quote-db.yaml',
        '✓ manifest is valid',
        '$ planton validate infrastructure/resource-group.yaml',
        '✓ manifest is valid',
      ],
    },
    proven: {
      kind: 'table',
      label:
        'What deploys them: each kind’s own open-source module, deployed and torn down in a real Azure account by Planton’s tests',
      columns: ['Kind', 'Proven On'],
      rows: [
        ['AzurePostgresqlFlexibleServer', 'Terraform, Pulumi'],
        ['AzureResourceGroup', 'Terraform, Pulumi'],
      ],
    },
    checks: [
      {
        text: 'Read the module that deploys AzurePostgresqlFlexibleServer',
        href: 'https://github.com/plantonhq/planton/tree/main/catalog/azure/azurepostgresqlflexibleserver/iac',
      },
      {
        text: 'Read the test that deploys AzurePostgresqlFlexibleServer into a real Azure account and destroys it',
        href: 'https://github.com/plantonhq/planton/tree/main/catalog/azure/azurepostgresqlflexibleserver/e2e',
      },
      {
        text: 'List every field the agent chose from, offline:',
        command: 'planton explain AzurePostgresqlFlexibleServer',
      },
    ],
    // ch 10: the count, only from platform-stats.ts; the example names what a kind is.
    reach: `AzurePostgresqlFlexibleServer is one of ${PLATFORM_STATS.DEPLOYMENT_MODULE_COUNT} kinds in the catalog, across ${PLATFORM_STATS.CLOUD_PROVIDER_COUNT} providers, from AWS, Google Cloud and Azure to Cloudflare and Kubernetes.`,
  },
};

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
