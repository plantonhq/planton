# AWS Backup Settings

Manages the account-level switches behind AWS Backup: which resource types the service protects in each region, which it fully manages, and whether cross-account backup is enabled organization-wide. This is a settings singleton — AWS keeps exactly one set of Backup preferences per region and one set of global settings per account, so this component adopts and configures those existing objects rather than creating anything new. Destroy is a no-op on both arms: whatever was last applied stays in effect indefinitely.

## What Gets Created

Nothing is created at AWS — the settings objects already exist as part of the account. The module adopts and configures:

- **Region Settings** — the region's resource-type opt-in map (which types AWS Backup protects here) and management preferences (which types Backup fully manages), configured when the `regionSettings` arm is set. Identity is the account+region pair — deploy at most one instance per region
- **Global Settings** — the account-wide settings map, in practice `isCrossAccountBackupEnabled` for organization cross-account copies, configured when the `global` arm is set. Identity is the account — set this arm in exactly one instance across all regions

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with AWS Backup permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Nothing for the region arm. Cross-account backup on the global arm is only meaningful in an AWS Organizations management context.

## Deploy

### Console

Open the deployment store, find **AWS Backup Settings**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, then the two settings arms. Start from the **Region Opt-Ins** preset in the [Presets](#presets) tab — it carries the complete resource-type map with a sensible split of opted-in types.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBackupSettings
metadata:
  name: backup-region-settings
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  regionSettings:
    resourceTypeOptInPreference:
      Aurora: true
      CloudFormation: false
      DSQL: false
      DocumentDB: false
      DynamoDB: true
      EBS: true
      EC2: true
      EFS: true
      EKS: false
      FSx: true
      Neptune: false
      RDS: true
      Redshift: false
      "Redshift Serverless": false
      S3: true
      "SAP HANA on Amazon EC2": false
      "Storage Gateway": false
      Timestream: false
      VirtualMachine: false
```

```shell
planton apply -f backup-settings.yaml
```

This adopts the region's Backup preferences and opts the common workload types (EBS, EC2, RDS, Aurora, DynamoDB, EFS, FSx, S3) into AWS Backup protection while explicitly opting the rest out. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring Backup settings. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Destroy reverts nothing** — Both arms are no-op deletes at the provider: the last-applied values stay in effect indefinitely after the component is gone. The revert lever is an apply with the desired values — to turn cross-account backup off, apply `"false"` first; never rely on destroy.

**Opt-ins gate backup plans silently** — A backup plan selection covering a type the region has not opted in simply never backs it up, with no error anywhere. When a plan "misses" resources, check this component before debugging the plan.

**List every type deliberately** — AWS returns the complete preference map on read and the provider owns the full map, so a type missing from `resourceTypeOptInPreference` shows as a perpetual difference in plans. Copy the full set from `aws backup describe-region-settings`, flip the booleans you mean, and extend the list when AWS adds a type.

**Singleton discipline is yours to keep** — One instance per region for the region arm; the global arm in exactly one instance account-wide. Two instances carrying the same arm fight over the same settings object, each apply undoing the other. `metadata.name` never reaches AWS — it is Planton-side identity only.

**Management preferences are one-way at AWS** — Once `resourceTypeManagementPreference` is set for a type, it can be flipped per type but never cleared back to unset. Omit the map entirely to leave management preferences untouched.

**Cross-account backup is an organizations decision** — Flipping `isCrossAccountBackupEnabled` affects the backup posture of every account in the organization. Treat it as a change-managed control, not a casual toggle.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — it configures account- and region-scoped settings objects that exist independently of any other catalog resource.

### What This Kind Provides

The outputs, `account_id` and `region`, echo the identities of the two settings objects (the account the global arm manages, the region the region arm manages). They are audit echoes, not composition inputs — no catalog kind consumes them via ValueFromRef.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Explicit opt-in map per active region** — one instance per region where backups run, carrying the complete type map with the workload types flipped on. Making the map explicit turns AWS's silent defaults into reviewed configuration and keeps plans clean. Start from the **Region Opt-Ins** preset.

**Organization cross-account copies** — a single instance carrying the global arm with `isCrossAccountBackupEnabled: "true"`, deployed once in the management account's tooling region. Backup plan copy actions into other accounts' vaults depend on this switch. Start from the **Cross-Account Backup** preset.

**Settings before plans** — deploy this component before the backup plans that depend on the opt-ins. There is no schema edge to enforce the ordering, so make it explicit in your rollout: a plan deployed first will silently skip types the region has not opted in yet.

## Works With

- [**AWS Backup Plan**](/infra-catalog/aws-backup-plan) — plans only protect resource types the region has opted in through this component
- [**AWS Backup Vault**](/infra-catalog/aws-backup-vault) — cross-account copy destinations become usable once the global arm enables cross-account backup
