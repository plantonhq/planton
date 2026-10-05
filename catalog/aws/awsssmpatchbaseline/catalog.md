# AWS SSM Patch Baseline

Deploys an AWS Systems Manager patch baseline: the patching policy for one operating system, defining which patches auto-approve (and after what soak period), which are explicitly blocked, and which patch groups the policy governs. The baseline can also claim the account/region default designation for its OS, so ungrouped nodes patch against your policy instead of AWS's predefined default. Patch groups and the default designation fold into this one component — they hang off the baseline's ID and have no life of their own.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Patch Baseline** — the approval rules, explicit approve/reject lists, optional global filters, and (for Linux) alternative patch source repositories, for one operating system. AWS identifies it as `pb-...`.
- **Patch Group Registrations** — one per `patchGroups` entry, binding nodes tagged `Patch Group: <name>` to this baseline.
- **Default Baseline Designation** — created only when `setAsDefaultBaseline` is true; deleting the component (or un-setting the flag) restores AWS's own predefined default baseline for the OS.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with permissions to manage SSM patch baselines. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Nothing — baselines stand alone. Patching later needs managed nodes tagged into the governed patch groups, and a scheduled `AWS-RunPatchBaseline` run to actually scan or install.

## Deploy

### Console

Open the deployment store, find **AWS SSM Patch Baseline**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields covering the operating system, approval rules, and patch groups. Start from the **AL2023 Security Baseline** preset in the [Presets](#presets) tab for the sensible Linux default: security patches after a soak period.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSsmPatchBaseline
metadata:
  name: al2023-security
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  operatingSystem: AMAZON_LINUX_2023
  description: AL2023 security patches after a 7-day soak
  approvalRules:
    - patchFilters:
        - key: CLASSIFICATION
          values:
            - Security
        - key: SEVERITY
          values:
            - Critical
            - Important
      approveAfterDays: 7
      complianceLevel: CRITICAL
  patchGroups:
    - prod-linux
```

```shell
planton apply -f patch-baseline.yaml
```

This creates an Amazon Linux 2023 baseline auto-approving Critical and Important security patches seven days after release, governing every node tagged `Patch Group: prod-linux`, and reporting missing patches as CRITICAL compliance. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a patch baseline. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Set the operating system explicitly** — `operatingSystem` unset means WINDOWS (the provider default), which is rarely what a Linux fleet intended, and changing it later forces replacement. One baseline governs exactly one OS; a mixed fleet needs one baseline per OS.

**The soak period is the risk lever** — `approveAfterDays: 7` gives a bad vendor patch a week to be pulled before your fleet takes it; `0` takes patches on release day. The alternative arm, `approveUntilDate`, freezes approval at a fixed cutoff you advance deliberately per change window — the two arms are mutually exclusive per rule. On Debian and Ubuntu, AWS does not support days-based approval: filter-only rules (neither arm) are the honest shape there.

**One group, one baseline per OS** — a patch group can be registered with only ONE baseline per operating system account-wide, and AWS state — not validation — is the referee: the second registration fails at apply. Coordinate group ownership across components before deploying.

**The default designation displaces silently** — claiming `setAsDefaultBaseline` marks whichever baseline previously held the OS default as non-default, with no warning to its owner. At most one baseline per OS should set it. The reverse is safe: un-setting the flag or deleting the component restores AWS's own predefined default for the OS.

**Rejected dependencies: quiet install or visible block** — when a rejected patch is a dependency of an approved one, `rejectedPatchesAction: ALLOW_AS_DEPENDENCY` (the default) installs it quietly; `BLOCK` refuses and reports noncompliance. Choose BLOCK when the rejection is a security decision that must not be bypassed transitively.

**Non-security breadth is Linux-only** — `enableNonSecurity` on a rule (and `approvedPatchesEnableNonSecurity` for the explicit list) widens approval beyond security updates. Windows patches are always security-classified, so these flags do nothing there.

**Alternative sources fail late** — Linux `sources` entries are stored verbatim; AWS never validates repo reachability. A wrong repo definition shows up as install failures at patch time, not at deploy.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — the approval policy, patch groups, and designation are all self-contained values. The binding to instances happens through the `Patch Group` tag on managed nodes, not through references.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `baseline_id` | The baseline's AWS-generated ID (`pb-...`) | Addressing the baseline in CLI/API operations and external patch group registrations |
| `baseline_arn` | The baseline's ARN | IAM policy statements scoping who may modify the patching policy |

`operating_system` is also echoed — the OS the baseline resolved to (WINDOWS when the spec left it unset). It is an input echo useful for confirming the default resolution, not a composition input.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Security-only with a soak** — Critical/Important security classifications, `approveAfterDays: 7`, one governed patch group. This is the first custom baseline most accounts need: AWS's predefined defaults approve with no soak, so a bad vendor patch reaches your fleet the day it ships. Pair it with a scanning AwsSsmAssociation to see what the policy would install before any install runs. Start from the **AL2023 Security Baseline** preset.

**Claiming the OS default** — `setAsDefaultBaseline: true` with a visible compliance posture (`availableSecurityUpdatesComplianceStatus: NON_COMPLIANT` on Windows) so every node — grouped or not — patches against your policy and unapproved-but-available security updates cannot hide. The trade: the designation is account-wide state that displaces silently, so it needs a single owner. Start from the **Windows Default Baseline** preset.

**Frozen cutoff per change window** — `approveUntilDate` instead of a rolling soak: approval is frozen at a date your change process advances deliberately, giving auditors an exact answer to "what was approved when". The cost is manual advancement — a forgotten cutoff quietly stops all new approvals.

## Works With

- [**AWS SSM Association**](/infra-catalog/aws-ssm-association) — schedules the recurring `AWS-RunPatchBaseline` scan that evaluates nodes against this baseline
- [**AWS SSM Maintenance Window**](/infra-catalog/aws-ssm-maintenance-window) — runs the patch install inside a declared window with rate controls and a hard cutoff
