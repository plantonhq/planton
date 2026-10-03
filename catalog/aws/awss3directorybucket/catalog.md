# AWS S3 Directory Bucket

Deploys an S3 directory bucket (S3 Express One Zone) — single-digit-millisecond object storage that lives in exactly one availability zone or Local Zone, co-located with the compute that hammers it. This is the speed tier for ML training shards, feature stores, and hot intermediate results that regular S3 is too slow for; the system of record stays in regional S3. AWS mandates the full bucket name `{base}--{zone_id}--x-s3`, so the modules derive it from `metadata.name` and `zoneId` — you name the base, and the name and location can never disagree. Everything on the spec except `forceDestroy` replaces the bucket: a directory bucket is replaced, not edited.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **S3 Directory Bucket** — the Express One Zone bucket in the zone `zoneId` names, under the derived full name `{metadata.name}--{zoneId}--x-s3`, with the declared zone type, redundancy class, and force-destroy posture
- **AWS Tags** — resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account, including S3 Express permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An Express-supported zone** — not every availability zone supports S3 Express One Zone; confirm the zone before pinning `zoneId`, because the zone is fixed for life.
- **Local Zone opt-in** (only for `zoneType: LocalZone`) — the account must have the target Local Zone enabled before a bucket can land there.
- **`s3express:CreateSession` grants for consumers** — directory buckets authenticate through the session API, not per-object IAM; a policy written for regular S3 object ARNs silently fails against them.

## Deploy

### Console

Open the deployment store, find **AWS S3 Directory Bucket**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **ML Training Scratch** preset in the [Presets](#presets) tab for the hot-shard shape: a force-destroyable bucket in the training cluster's own zone.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsS3DirectoryBucket
metadata:
  name: training-scratch
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  zoneId: use1-az4
  forceDestroy: true
```

```shell
planton apply -f directory-bucket.yaml
```

This creates a directory bucket named `training-scratch--use1-az4--x-s3` in availability zone `use1-az4`, destroyable even while it holds objects. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a directory bucket. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The zone choice IS the performance model** — `zoneId` takes a zone ID, never a letter name: `use1-az4` is the same physical zone in every AWS account, while `us-east-1a` is a per-account shuffle. Co-location means matching the compute subnet's zone ID (`aws ec2 describe-subnets` shows `AvailabilityZoneId`), not its letter — the latency win exists only for compute in the SAME zone.

**Everything except forceDestroy is a one-way door** — the zone, the zone type, the redundancy class, and the base name all replace the bucket on change. Moving zones or renaming means a new bucket and a data copy, so settle the name and zone before the first object lands.

**One zone means one zone** — Express One Zone data has no cross-AZ replica: a zone impairment takes the data offline, and a zone loss can lose it. Put reconstructible data here (training shards, caches, intermediates) and keep the system of record in a regional bucket.

**Leave dataRedundancy unset** — the redundancy class must name the zone type it survives in (SingleAvailabilityZone for availability zones, SingleLocalZone for Local Zones), and AWS derives the only valid pairing when the field is empty. The spec rejects a mismatched pair; setting it buys nothing.

**The session API is the auth model** — SDKs handle the `CreateSession` handshake transparently, but the IAM policies of every consumer must grant `s3express:CreateSession` on the bucket. This is the most common "it works in regular S3 but not here" failure.

**Directory semantics change listing habits** — objects are organized by real directories, listing is prefix-constrained, and there is no lifecycle, versioning, or replication surface. Code written against regular S3's flat-namespace conventions may need its listing paths revisited.

**Wire the bucket_name output, never the literal** — the full name encodes the zone, so consumers hardcoding it survive a zone change only by accident. Reference the output.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — the zone and redundancy fields are plain strings, so nothing is wired from other Infra Components.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `bucket_name` | The full derived name (`{base}--{zone_id}--x-s3`) — what S3 Express clients address | Application and training-job configuration |
| `bucket_arn` | The bucket's ARN | IAM policies granting `s3express:CreateSession` |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Hot training scratch beside the GPUs** — a One Zone bucket in the training cluster's zone, force-destroyable because everything in it is reconstructible from the regional source-of-truth bucket. The trade is deliberate: teardown is one command, and a zone loss costs a re-run, not data. Start from the **ML Training Scratch** preset.

**Metro-edge object storage in a Local Zone** — media and rendering workloads that need single-digit latency to on-prem or edge compute get a bucket in the metro's Local Zone, with `zoneType: LocalZone` and the redundancy class naming the zone type explicitly. Requires the Local Zone opt-in on the account. Start from the **Local Zone Edge Cache** preset.

## Works With

- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — the regional system of record; directory buckets hold the hot, reconstructible copy of data whose durable home is a general-purpose bucket

Beyond that pairing the component is standalone: it references no other Infra Components, and consumers reach it through the `bucket_name` output rather than a typed edge.
