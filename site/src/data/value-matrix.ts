/**
 * The value matrix: the granular per-plan feature enumeration rendered on
 * the pricing page (its only consumer).
 *
 * Rows that gate on a PAID capability carry the platform's entitlement key
 * (`entitlementKey`) -- the same string the offer catalog sells and the
 * server enforces -- so the displayed story and the enforced reality can be
 * checked against each other mechanically. A capability that is decided
 * packaging but not yet shipped renders as "Coming Soon", never as live.
 *
 * Numbers in cells come from the pricing-truth module (`./pricing`), never
 * from literals here.
 */

import {
  FREE_TIER_SEATS,
  COMMUNITY_SEAT_LIMIT,
  MARKETS,
  SELF_HOSTED_LICENSE_SEAT_CEILINGS,
} from './pricing.ts';

export type PlanColumnId = 'free' | 'team' | 'community' | 'license' | 'enterprise';

export interface PlanColumn {
  id: PlanColumnId;
  label: string;
  sublabel?: string;
}

export const VALUE_MATRIX_COLUMNS: PlanColumn[] = [
  { id: 'free', label: 'Free', sublabel: 'Planton.ai' },
  { id: 'team', label: 'Team', sublabel: 'Planton.ai' },
  { id: 'community', label: 'Community', sublabel: 'Your Infrastructure' },
  { id: 'license', label: 'Licensed', sublabel: 'Your Infrastructure' },
  { id: 'enterprise', label: 'Enterprise' },
];

export type MatrixCell =
  | { kind: 'included' }
  | { kind: 'not_included' }
  | { kind: 'coming_soon' }
  | { kind: 'text'; label: string };

const YES: MatrixCell = { kind: 'included' };
const NO: MatrixCell = { kind: 'not_included' };
const SOON: MatrixCell = { kind: 'coming_soon' };
const text = (label: string): MatrixCell => ({ kind: 'text', label });

const everywhere: Record<PlanColumnId, MatrixCell> = {
  free: YES,
  team: YES,
  community: YES,
  license: YES,
  enterprise: YES,
};

const unlimitedEverywhere: Record<PlanColumnId, MatrixCell> = {
  free: { kind: 'text', label: 'Unlimited' },
  team: { kind: 'text', label: 'Unlimited' },
  community: { kind: 'text', label: 'Unlimited' },
  license: { kind: 'text', label: 'Unlimited' },
  enterprise: { kind: 'text', label: 'Unlimited' },
};

export interface MatrixRow {
  feature: string;
  description?: string;
  /** The entitlement key this row gates on, when it is a paid capability. */
  entitlementKey?: string;
  cells: Record<PlanColumnId, MatrixCell>;
}

export interface MatrixCategory {
  category: string;
  rows: MatrixRow[];
}

const licenseSeatText = SELF_HOSTED_LICENSE_SEAT_CEILINGS.join(' or ');
// Seat ceilings are market-independent; any market's tier list carries them.
// Each enterprise package is named beside its own ceiling — a bare
// "100 — 250" read as a range (and as a maximum), which is neither.
const enterpriseSeatText = `Up to ${MARKETS.us.enterprise
  .map((t) => `${t.seatCeiling} (${t.name.replace('Enterprise ', '')})`)
  .join(' · ')}`;

export const VALUE_MATRIX: MatrixCategory[] = [
  {
    // The differences lead: a reader sees what they are paying for before
    // they see everything nobody charges for.
    category: 'What Changes by Plan',
    rows: [
      {
        feature: 'Seats',
        cells: {
          free: text(`Up to ${FREE_TIER_SEATS}`),
          team: text('Per seat — grow as you go'),
          community: text(`Up to ${COMMUNITY_SEAT_LIMIT}`),
          license: text(`${licenseSeatText} by license size`),
          enterprise: text(enterpriseSeatText),
        },
      },
      {
        feature: 'Support',
        cells: {
          free: text('Community'),
          team: text('Standard'),
          community: text('Community'),
          license: text('Standard — email'),
          enterprise: text('Named business-hours (Standard) · 24×7 SLA (Plus)'),
        },
      },
      {
        feature: 'Offline & Air-Gapped Operation',
        description: 'Deployments on your infrastructure never phone home; licenses verify offline',
        entitlementKey: 'air_gap',
        cells: {
          free: NO,
          team: NO,
          community: text('Runs offline'),
          license: text('Runs offline'),
          enterprise: text('Supported air-gap posture'),
        },
      },
      {
        feature: '30-Day Full-Experience Evaluation',
        description: 'Every capability on your own cluster — no card, no call',
        cells: {
          free: NO,
          team: NO,
          community: text('Self-Serve Evaluation Key'),
          license: YES,
          enterprise: NO,
        },
      },
      {
        feature: 'Deployment Safety',
        description: 'Protected environments and two-person deployment approvals',
        entitlementKey: 'deployment_safety',
        cells: { free: NO, team: YES, community: NO, license: YES, enterprise: YES },
      },
    ],
  },
  {
    // Second deliberately (founder-directed 2026-08-20): the assistant is a
    // flagship capability AND a genuine by-plan difference — hosted plans
    // include it on prepaid credits today, self-hosted installs do not have
    // it yet — so it sits with the differences at the top, right after the
    // purchase-decision facts, never buried among the included-everywhere
    // domains. AI is deployment-capability-gated, never packaging-gated:
    // no entitlement key exists or is reserved for it. When self-hosted
    // deployments gain the in-cluster engine (bring-your-own-LLM-key), the
    // SOON cells flip to a truthful included form in the same change; if
    // packaging ever gates AI, these rows must gain that key so the
    // displayed-vs-enforced guard covers them. Credits are a Planton.ai
    // construct — a self-hosted install has no wallet, so its credits cells
    // are simply not included, never SOON.
    category: 'AI Assistant',
    rows: [
      {
        feature: 'Built-in AI Assistant',
        description:
          'On Planton.ai today; self-hosted installs will run it in-cluster with your own LLM provider key',
        cells: { free: YES, team: YES, community: SOON, license: SOON, enterprise: SOON },
      },
      {
        feature: 'Prepaid AI Credits with Spend Protection',
        description: 'Transparent balance; auto-reload only within a ceiling you set',
        cells: { free: YES, team: YES, community: NO, license: NO, enterprise: NO },
      },
    ],
  },
  {
    // The dimensions usage-based pricing usually meters -- named
    // deliberately, unlimited deliberately, on every plan. Environments
    // cost the customer's cloud, not ours; capping them would be a lever,
    // not a cost defense. (A future tightening for NEW organizations would
    // be a catalog edit that grandfathers existing ones -- founder-decided
    // posture, 2026-08-13.)
    category: 'Usage & Capacity — Unlimited on Every Plan',
    rows: [
      { feature: 'Environments', cells: unlimitedEverywhere },
      { feature: 'Cloud Account Connections', cells: unlimitedEverywhere },
      { feature: 'Cloud Resources & Components', cells: unlimitedEverywhere },
      { feature: 'Services', cells: unlimitedEverywhere },
      {
        feature: 'Automation Minutes',
        description: 'Infrastructure and CI/CD automation runs — never metered for billing',
        cells: unlimitedEverywhere,
      },
    ],
  },
  {
    category: 'Deployments & Infrastructure',
    rows: [
      {
        feature: 'Multi-Cloud Component Catalog',
        description: '700+ components across AWS, GCP, Azure, Kubernetes, and more',
        cells: everywhere,
      },
      { feature: 'Guided Deployment Wizards', cells: everywhere },
      { feature: 'Infrastructure Pipelines & IaC Automation', cells: everywhere },
      { feature: 'Environment Blueprints (Infra Charts)', cells: everywhere },
      { feature: 'Bring Your Own State Backend', cells: everywhere },
    ],
  },
  {
    // Runners are ungated on every plan today. A `runners` entitlement key
    // exists in the platform's vocabulary as RESERVED; if packaging ever
    // gates runners, these rows must gain that key so the
    // displayed-vs-enforced guard covers them.
    category: 'Self-Hosted Runners',
    rows: [
      {
        feature: 'Runners in Your Own Network',
        description:
          'Runners deploy inside your private network and connect outbound-only — no inbound firewall ports, no VPN',
        cells: unlimitedEverywhere,
      },
      {
        feature: 'Credentials Stay in Your Environment',
        description:
          'Runners use your environment\u2019s native identity; secrets are resolved just-in-time inside the runner, never at the control plane',
        cells: everywhere,
      },
    ],
  },
  {
    category: 'Services & CI/CD',
    rows: [
      { feature: 'Service Deployments with Build Pipelines', cells: everywhere },
      { feature: 'Git Integrations', cells: everywhere },
      { feature: 'Multi-Environment Promotion', cells: everywhere },
    ],
  },
  {
    category: 'Secrets & Configuration',
    rows: [
      { feature: 'Secret Backends (Bring Your Own)', cells: everywhere },
      { feature: 'Just-in-Time Secret Resolution', cells: everywhere },
      { feature: 'Variables & Configuration Management', cells: everywhere },
    ],
  },
  {
    category: 'Access Control & Identity',
    rows: [
      { feature: 'Role-Based Access Control', cells: everywhere },
      { feature: 'Teams, Service Accounts & API Keys', cells: everywhere },
      {
        feature: 'Enterprise SSO (SAML/OIDC)',
        description: 'Sign in through your own identity provider',
        entitlementKey: 'sso',
        cells: { free: NO, team: SOON, community: NO, license: SOON, enterprise: SOON },
      },
      {
        feature: 'SCIM Directory Sync',
        description: 'Provision and deprovision users from your directory',
        entitlementKey: 'scim',
        cells: { free: NO, team: NO, community: NO, license: NO, enterprise: SOON },
      },
    ],
  },
  {
    category: 'Audit & Compliance',
    rows: [
      { feature: 'Deployment History & Audit Trail', cells: everywhere },
      {
        feature: 'Access Transparency',
        description: 'A log of every time vendor staff held time-boxed access inside your organization',
        entitlementKey: 'access_transparency',
        // Decided placement, honestly worded: the cell flips to included
        // when the enterprise tier sells the key.
        cells: { free: NO, team: NO, community: NO, license: NO, enterprise: SOON },
      },
      {
        feature: 'Compliance Reporting',
        cells: { free: NO, team: NO, community: NO, license: NO, enterprise: text('Enterprise package') },
      },
    ],
  },
];
