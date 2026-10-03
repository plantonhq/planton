# AWS CloudTrail Event Data Store

Deploys a CloudTrail Lake event data store: a queryable, immutable store of AWS activity events you interrogate with SQL — security investigations and audits without shipping logs to S3 or standing up a SIEM. The store ingests events matched by its advanced event selectors, keeps them queryable for a retention window of 7 to 2555 days, and needs no trail and no bucket — Lake owns its own storage and billing. One hard constraint gates everything: AWS has closed CloudTrail Lake to new customers, so store creation fails on any account that has never created an event data store ("CloudTrail Lake is no longer accepting new customers") — deploy this component only on an account already using Lake, and use AWS CloudTrail's trail delivery everywhere else.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Event Data Store** — the Lake store with its pricing mode, retention window, multi-region and organization ingestion, advanced event selectors, termination protection, ingestion pause switch, and optional SSE-KMS encryption

Destroying the component soft-deletes the store: AWS holds it in `PENDING_DELETION` for 7 days (the name stays reserved, and the store is restorable via the console) before the purge.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with CloudTrail permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An account grandfathered into CloudTrail Lake — one that already holds (or previously created) an event data store. There is no known exception process, and nothing in this kind's spec can route around the account-level wall.
- (Only for SSE-KMS) a KMS key whose policy allows CloudTrail to use it. AWS recommends multi-region keys for multi-region stores, and losing the key makes the store unreadable.

## Deploy

### Console

Open the deployment store, find **AWS CloudTrail Event Data Store**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, then the pricing mode, retention, and ingestion selectors. Start from the **Security Investigation Store** preset in the [Presets](#presets) tab for the management-events investigation store.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudTrailEventDataStore
metadata:
  name: security-investigations
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  advancedEventSelectors:
    - name: Management events
      fieldSelectors:
        - field: eventCategory
          equals: ["Management"]
```

```shell
planton apply -f event-data-store.yaml
```

This creates a multi-region store (the AWS default) ingesting all management events, queryable with Lake SQL for the default 2555-day retention, with termination protection on. An Infra Job tracks the provisioning in real time.

### InfraChart

When the store deploys alongside its encryption key in one chart, wire the key reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: lake-key
      fieldPath: status.outputs.key_arn
  advancedEventSelectors:
    - name: Management events
      fieldSelectors:
        - field: eventCategory
          equals: ["Management"]
```

The InfraPipeline resolves the dependency graph, creates the key first, then provisions the store encrypted with it.

## Key Configuration

These are the most important decisions when configuring an event data store. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Ingestion is the bill** — Lake charges on what the store ingests, and an unscoped store (no selectors) makes AWS materialize a default selector ingesting EVERY management event. Scope `advancedEventSelectors` to what investigations actually need; AWS requires every selector to carry an `eventCategory` condition — a server-side rule the provider does not pre-check, so a selector without one fails at apply.

**Pricing mode is a one-way door in practice** — `EXTENDABLE_RETENTION_PRICING` (the default) bills ingestion once with long included retention; `FIXED_RETENTION_PRICING` bills less on ingest but charges for retention and caps it at 7 years. AWS allows changing fixed → extendable only — pick deliberately at creation. `retentionPeriodDays` accepts 7–2555 through the provider regardless of what the console offers.

**The teardown is two steps by AWS design** — With `terminationProtectionEnabled` on (the AWS default when unset), destroy FAILS. Apply `terminationProtectionEnabled: false` first, then destroy. This is deliberate AWS behavior protecting audit history, not a module defect.

**Pause instead of delete** — `suspend: true` stops ingestion while already-ingested events stay queryable (retention keeps counting) — the right lever for stores kept for occasional investigations. AWS never reports the flag back, so it is asserted on every apply and invisible to imports.

**The KMS key is fixed at creation** — Changing `kmsKeyId` replaces the store and its ingested history. Pick the key before first ingestion, prefer a multi-region key for a multi-region store, and treat the key's lifecycle as part of the store's: losing the key makes the history unreadable.

**Multi-region and organization scope** — Multi-region ingestion is the AWS default (unset = enabled); set `multiRegionEnabled: false` only for deliberately region-scoped stores. `organizationEnabled` ingests from every member account but only works from the management account or delegated CloudTrail administrator with all-features Organizations enabled.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** | `kmsKeyId` | `status.outputs.key_arn` |

### What This Kind Provides

The single output, `event_data_store_arn`, is an identity echo rather than a composition input — no catalog kind consumes it via ValueFromRef. It is the provider's import ID and what `aws cloudtrail start-query` SQL addresses in its FROM clause; the store's real product is query results in the Lake editor, not wiring for downstream resources.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**The security investigation store** — all management events, multi-region, long retention on extendable pricing. The standing capability that answers "who touched this resource in the last two years" in one SQL query, without log-pipeline archaeology. Start from the **Security Investigation Store** preset.

**Scoped data-events store** — object-level writes on specific buckets via `resources.ARN` prefix matching, short retention on fixed pricing. Data events at full volume are the fastest way to a runaway ingestion bill; a prefix-scoped, 90-day store keeps the forensic signal affordable. Start from the **Scoped Data Events Store** preset.

**Lake beside the trail** — the trail delivers the durable, cheap archive to S3; the store keeps a scoped, SQL-queryable working set for investigations. They ingest independently — the store is not fed by the trail — so their selectors can (and usually should) differ.

## Works With

- [**AWS CloudTrail**](/infra-catalog/aws-cloud-trail) — the complementary surface: file delivery to S3 for the archive, Lake for SQL investigation
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — SSE-KMS encryption of the stored events, fixed at creation
