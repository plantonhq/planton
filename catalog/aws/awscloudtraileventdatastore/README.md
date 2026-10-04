<p align="center">
  <img src="logo.svg" alt="AWS CloudTrail Event Data Store" width="80"/>
</p>

# AWS CloudTrail Event Data Store

Manage a [CloudTrail Lake event data store](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-lake.html)
— a queryable, immutable store of AWS activity events with its own
retention and billing lifecycle, independent of any trail.

## What Gets Managed

- **The event data store** (`metadata.name` is the store name): the
  pricing mode (extendable vs fixed retention), the queryable window
  (7–2555 days), multi-region and organization ingestion, termination
  protection, and the ingestion pause switch.
- **Ingestion scope**: advanced event selectors (fine-grained field
  matching over management, data, and network-activity events). AWS
  requires every selector to carry an `eventCategory` condition; an
  omitted selector list makes AWS materialize a default
  all-management-events selector.
- **Encryption**: optional SSE-KMS via a key whose policy allows
  CloudTrail (fixed at creation — changing it replaces the store).

Destroying this component **soft-deletes the store**: AWS holds it in
`PENDING_DELETION` for 7 days (the name stays reserved) before the
purge, and **refuses the delete while termination protection is on** —
set `termination_protection_enabled: false` and apply first. Lake
bills per GB ingested (plus per GB retained on the fixed pricing
mode); scope selectors deliberately.

The trail (S3 log delivery) is deliberately NOT part of this kind
— see [AwsCloudTrail](../awscloudtrail).

See [v1alpha1/reference.md](v1alpha1/reference.md) for the full field
reference generated from the spec proto.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
