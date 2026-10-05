# AWS GuardDuty

Deploys the region's GuardDuty threat-detection posture: one detector with its protection plans (S3 data events, EKS audit logs, runtime monitoring, RDS login events, Lambda network logs, AI protection), finding filters for noise control, trusted and threat IP lists, findings export to S3, and — for organization administrators — member-account management. This is a region singleton: AWS allows exactly one detector per account per region, the detector has no name (its AWS-assigned ID is the identity), and a second instance or a pre-existing detector fails creation with "detector already exists". Destroy deletes the detector, all its findings, and every satellite.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **GuardDuty Detector** — the region's detector with its monitoring switch and finding re-publish frequency
- **Detector Features** — one patch per `features` entry: protection plans with their agent-management sub-toggles. Features are patches onto the detector — AWS has no delete for them, so unlisted features stay exactly as AWS has them
- **Finding Filters** — one per `filters` entry: match criteria that auto-archive (or organize) findings
- **Trusted IP Lists / Threat Intel Lists** — one per `ipSets` / `threatIntelSets` entry, reading list files from S3
- **Publishing Destination** — findings export to an S3 bucket under a KMS key, created only when `publishingDestination` is set
- **Organization Administration** — delegated-admin registration, organization configuration, org-wide feature auto-enablement, and member accounts, created only when the `organization` / `members` arms are set; or an **Invitation Accepter** when `acceptInvitationFromAccountId` takes the member side

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with GuardDuty permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- No pre-existing detector in the region — console onboarding or Organizations auto-enable both create one, and AWS allows exactly one. Adopt an existing detector by import or remove it first.
- (Only for findings export) a bucket whose policy grants `guardduty.amazonaws.com` `s3:PutObject` and `s3:GetBucketLocation`, and a KMS key whose policy grants it `kms:GenerateDataKey` — AWS rejects the destination without both consents.
- (Only for trusted/threat lists) the list files in S3, readable by GuardDuty — it re-reads them on activation.
- (Only for the organization arms) the delegated GuardDuty administrator account, or the management account when delegating via `organization.adminAccountId`.

## Deploy

### Console

Open the deployment store, find **AWS GuardDuty**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, then the protection plans and satellites. Start from the **Standalone Detector** preset in the [Presets](#presets) tab for a single account, or the **Organization Admin** preset when this account administers the organization.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsGuardDuty
metadata:
  name: guardduty
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  findingPublishingFrequency: FIFTEEN_MINUTES
  features:
    - name: S3_DATA_EVENTS
    - name: RUNTIME_MONITORING
  filters:
    - name: archive-low-severity
      description: Archive findings below severity 4
      action: ARCHIVE
      rank: 1
      criteria:
        - field: severity
          lessThan: "4"
```

```shell
planton apply -f guardduty.yaml
```

This creates the region's detector with S3 protection and runtime monitoring on, updated findings re-publishing every fifteen minutes, and low-severity findings auto-archived. An Infra Job tracks the provisioning in real time.

### InfraChart

When the detector deploys alongside its findings-export bucket and key in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  features:
    - name: S3_DATA_EVENTS
  publishingDestination:
    bucketArn:
      valueFrom:
        kind: AwsS3Bucket
        name: findings-archive
        fieldPath: status.outputs.bucket_arn
    kmsKeyArn:
      valueFrom:
        kind: AwsKmsKey
        name: findings-export-key
        fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, creates the bucket and key (with their GuardDuty grants) first, then provisions the detector exporting into them.

## Key Configuration

These are the most important decisions when configuring GuardDuty. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Removing a feature reverts nothing** — Detector, organization, and member features are patches: Create and Update are the same call, and AWS keeps the last-applied state of anything you stop listing. To turn a protection plan off, list it with `enabled: false` — deleting the entry leaves it running (and billing). The same holds for the organization configuration after destroy: the posture survives as last applied.

**Protection plans are the cost surface** — The foundational data sources are always on; each protection plan bills by the volume it analyzes (S3 data events, runtime activity, RDS logins, ...). `enable: false` on the detector suspends all monitoring and billing without losing the detector, its findings, or its configuration — the pause lever, distinct from destroy.

**Findings export needs both consents at create time** — The bucket policy grant and the key's `kms:GenerateDataKey` must exist before the destination is created. When wiring `bucketArn` by reference, the composed value is the bare bucket ARN, so the bucket policy's `s3:PutObject` grant must cover the whole bucket — a prefix-scoped grant fails creation. The destination is deliberately untagged: tags force its replacement upstream, and a tag sweep would replace findings export mid-audit.

**One active trusted list per detector** — AWS enforces it; keep spare `ipSets` entries `activate: false`. Threat intel lists have no such cap. Filters are the everyday noise control: `ARCHIVE` suppresses matching findings, `NOOP` merely organizes them, and `rank` orders evaluation.

**Admin or member, never both** — `organization`/`members` and `acceptInvitationFromAccountId` are mutually exclusive (rejected at validate time). Organization features answer a different question than detector features — NEW/ALL/NONE (which members get the plan) instead of on/off — and with `autoEnableOrganizationMembers: NEW`, the `members` list is only for exceptions and non-Organizations accounts.

**Members inherit the administrator's finding frequency** — Setting `findingPublishingFrequency` on a member detector fights the organization sync forever; set it on the admin side only.

**Agent management installs software** — The `EKS_ADDON_MANAGEMENT`, `ECS_FARGATE_AGENT_MANAGEMENT`, and `EC2_AGENT_MANAGEMENT` sub-toggles deploy GuardDuty security agents into your compute. Turn them on deliberately, per environment — declare only the sub-toggles you enable; undeclared ones are sent as explicitly disabled.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsS3Bucket** | `publishingDestination.bucketArn` | `status.outputs.bucket_arn` |
| **AwsKmsKey** | `publishingDestination.kmsKeyArn` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `detector_id` | The detector's AWS-assigned ID — the region's GuardDuty identity | Addressing the detector in AWS CLI/API operations; the import key every satellite composes from |
| `detector_arn` | The detector's ARN | IAM policies scoping GuardDuty administration |

`account_id`, `ip_set_ids`, `threat_intel_set_ids`, and `publishing_destination_id` are also exported; they are import and audit echoes for the folded satellites, not composition inputs — no catalog kind consumes them via ValueFromRef.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standalone detector with noise control** — the region's detector with the protection plans the account actually uses, a severity-floor archive filter, and findings export to a hardened bucket (GuardDuty keeps findings 90 days; the export is the long-term record). Start from the **Standalone Detector** preset.

**Organization administrator** — deployed in the delegated admin account: `autoEnableOrganizationMembers: NEW` so every new account joins automatically, org-wide feature auto-enablement deciding which members get which plans. One instance administers the whole organization's posture for the region. Start from the **Organization Admin** preset.

**One instance per active region** — GuardDuty is regional; threat detection in three regions means three instances (and, for organizations, the admin posture repeated per region). Keep the feature lists aligned across regions unless a region's workload genuinely differs.

## Works With

- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — the findings-export destination, and where trusted/threat list files live
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — encrypts exported findings (required by AWS for export)
- [**AWS GuardDuty Malware Protection Plan**](/infra-catalog/aws-guard-duty-malware-protection-plan) — on-upload malware scanning for S3 buckets, a separate GuardDuty surface with no detector edge
- [**AWS EKS Cluster**](/infra-catalog/aws-eks-cluster) — the workloads EKS audit-log and runtime-monitoring plans watch
