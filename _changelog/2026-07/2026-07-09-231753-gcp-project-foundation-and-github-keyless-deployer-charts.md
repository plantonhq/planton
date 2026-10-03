# GCP Foundation Charts: Project Landing Zone + GitHub Actions Keyless Deployer

**Date**: July 9, 2026
**Type**: Feature
**Components**: Infra Charts, GCP Provider

## Summary

The GCP chart catalog gains its foundation layer: `gcp/project-foundation`
deploys the landing zone every workload should start on (custom-mode VPC,
right-sized regional subnet, Cloud NAT egress, deny-by-default firewall
posture with IAP-only SSH, private-services-access peering, and optionally
the project itself and a private DNS zone), and
`gcp/github-actions-keyless-deployer` deploys keyless CI/CD from GitHub
Actions (a Workload Identity Federation pool with a GitHub OIDC provider
locked to one organization, per-role additive deploy grants bound directly
to the GitHub principal, and an Artifact Registry repository with
repo-scoped push access). Both charts pass `planton chart validate` on
their defaults and on every feature-toggle arm.

## Problem Statement / Motivation

### Pain Points

- Every GCP deployment needs the same invisible first layer — network,
  egress, firewall posture, PSA peering — and getting it wrong (an
  auto-mode VPC, SSH open to the internet, PSA retrofitted into a crowded
  address plan) is expensive to undo once workloads land on it. The catalog
  had no chart that made those decisions for the user.
- The standard way teams wire GitHub Actions to GCP is still a
  service-account key pasted into a repository secret — a long-lived,
  copyable credential. The catalog carries first-class Workload Identity
  Federation kinds, but composing them correctly (attribute mapping, the
  org-scoping condition, principalSet member strings built on the project
  NUMBER) is exactly the hard wiring a chart exists to remove.

## Solution / What's New

### `gcp/project-foundation`

| Resource | Kind | Condition |
|----------|------|-----------|
| Project | `GcpProject` | `createProject` (default off) |
| VPC network | `GcpVpcNetwork` | always |
| Workload subnet | `GcpSubnetwork` | always (flow logs via `flowLogsEnabled`) |
| Cloud NAT + router | `GcpRouterNat` | always |
| allow-internal rule | `GcpFirewallRule` | always |
| IAP-only SSH rule | `GcpFirewallRule` | `iapSshEnabled` |
| PSA reserved range | `GcpGlobalAddress` | `privateServicesAccessEnabled` |
| PSA connection | `GcpServiceNetworkingConnection` | `privateServicesAccessEnabled` |
| Private DNS zone | `GcpDnsZone` | `privateDnsEnabled` |

Posture decisions the chart makes (and teaches inline): custom subnet mode
only; Private Google Access on (workloads have no external IPs); NAT covers
all current and future subnets in the region with errors-only logging; the
internal firewall rule is scoped to the subnet CIDR rather than 10/8; SSH
is reachable exclusively from Google's IAP range (35.235.240.0/20); the PSA
range defaults to a /16 because producers cannot use fragmented leftovers.

The chart defaults to bring-your-own-project: creating a project requires
an organization/folder parent and a billing account, which have no safe
defaults — `createProject: true` flips on the full vending-machine mode and
repoints every resource's `projectId` at the chart-created project via
`valueFrom`, ordering the whole foundation after the project exists.

### `gcp/github-actions-keyless-deployer`

| Resource | Kind | Condition |
|----------|------|-----------|
| Federation pool | `GcpWorkloadIdentityPool` | always |
| GitHub OIDC provider | `GcpWorkloadIdentityPoolProvider` | always |
| Deploy grants | `GcpProjectIamMember` × N (one per role in the `deployer_roles` list param) | always |
| Image repository | `GcpArtifactRegistryRepo` + repo-scoped writer grant | `registryEnabled` |

The chart grants roles DIRECTLY to the federated principal
(`principalSet://…/attribute.repository/<org>/<repo>` by default;
organization-wide when `github_repo` is blanked) — the modern keyless path
with no intermediate service account, no long-lived credential, and nothing
to rotate. The provider's `attributeCondition` pins trust to one GitHub
organization and the template comments teach why it must never be removed
(GitHub's issuer signs tokens for every repository on github.com). The
member strings are assembled from the project number and pool ID because
IAM principals for federated identities are a composition no single
resource outputs.

The README closes the loop with a complete workflow example
(`google-github-actions/auth` + `id-token: write`, docker push, Cloud Run
deploy) and the per-branch tightening recipe.

## Implementation Details

- Both charts follow the chart authoring standard: typed documented params,
  `{% if … | bool %}` toggles whose references are guarded by the same
  condition, `valueFrom` references that rely on annotated composition keys
  (name-only refs), numeric manifest fields as string params rendered
  unquoted, inline comments to the spec-proto bar, 5-section READMEs with
  Mermaid architecture diagrams.
- The deploy grants use a `list` param with a `{% for %}` loop — one
  additive `GcpProjectIamMember` per role, named from the role
  (`gha-deployer-run-admin`), so adding or removing a role never disturbs
  the others.
- The authoring rule's banned-tag list gained `from` (the validator's
  source scan always banned it; the rule prose now matches the code) and
  Chart.yaml guidance now includes verifying the `iconUrl` resolves, with
  the kind-logo URL convention documented.
- Site stats regenerated (44 charts, 444 kinds).

## Validation

- `planton chart validate` (built from the working tree) green on:
  - `project-foundation`: defaults (7 resources), `createProject=true`
    (8), `privateDnsEnabled=true` (8), all-optional-off (4), everything-on
    (9), and the standalone-project arm (`parent_id`/`billing_account_id`
    blanked).
  - `github-actions-keyless-deployer`: defaults (5 resources), org-wide
    principal (`github_repo=`), `registryEnabled=false` (4), single- and
    triple-role `deployer_roles` overrides.
- Tree-wide `charts/ make validate`: 12/44 pass — all 4 GCP charts green;
  the 32 failures are pre-existing schema drift in other providers' legacy
  charts (each provider's own catalog rebuild addresses them).
- Live chart deploys remain gated on the platform's server-side
  `chart build` after the next release, as with the state-backend charts.

## Benefits

- A team's first hour on GCP through Planton is now two chart deploys: the
  landing zone, then keyless CI/CD — with no service-account key ever
  created and the network posture decided correctly from day one.
- The keyless chart eliminates the most common GCP CI credential
  anti-pattern (exported SA keys in repo secrets) by making the correct
  pattern the path of least resistance.

## Related Work

- The GCP chart catalog rebuild opener and `planton chart validate`
  (2026-07-09) — the standard and offline gate these charts are built on.
- The GCP Workload Identity Federation kinds (2026-07-03) — the pool and
  provider these charts compose; the provider's GitHub preset is the
  proven configuration the federation template mirrors.

---

**Status**: ✅ Production Ready
