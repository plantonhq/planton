# Overview

The **AzureDataProtectionBackupPolicy** component creates a Data Protection backup policy -- WHEN backups run (ISO-8601 repeating intervals) and HOW LONG they are kept (a default retention plus optional named rules that keep specific backups -- first of day, first of week -- longer). ONE kind covers the six datasource types as variants: blob storage, managed disks, Kubernetes (AKS) clusters, MySQL flexible servers, PostgreSQL flexible servers, and Data Lake storage. Exactly one variant block is set; the block IS the datasource type. The policy itself is a free configuration object -- cost follows the protected instances and their backup storage.

## Purpose

- **Backup schedules as declarative infrastructure**: the schedule/retention contract reviewed and versioned like everything else, one grammar across six datasources.
- **Grandfather-father-son retention**: a default keep-everything layer plus named rules that tag first-of-week/month/year backups for longer keeps.
- **Chart-ready wiring**: the policy references its vault by typed reference and publishes the `backup_policy_id` that backup instances bind to.

## Key Features

- Full azurerm v5 surface across all six policy resources, each mirrored exactly: blob's dual operational/vault retention tiers, disk's flat retention grammar, the nested life-cycle grammar of AKS and the flexible servers, and Data Lake's flat rules with order-derived priorities.
- The provider's cross-field contracts front-loaded as manifest-time validation: blob's at-least-one-tier and vault-tier-needs-intervals lattice, Data Lake's per-rule criteria contract, ISO-8601 duration and repeating-interval formats, and the Windows time-zone vocabularies exactly where the provider enforces them.
- Policy immutability recorded honestly: EVERY field on every variant is fixed at creation (the provider ships no update path) -- changing anything replaces the policy.

## Use Cases

- **AKS cluster backup**: scheduled cluster state backups with tiered retention -- the marquee modern-backup capability.
- **Managed-disk snapshots**: daily incremental snapshots with a week of default retention and longer keeps for the first of each month.
- **Blob and database protection**: continuous operational-tier blob retention, vaulted copies on schedule, and weekly full backups for flexible-server databases.

## Future Enhancements

- Backup instance kinds (the bindings that put specific disks, blobs, and clusters under this policy's protection) complete the modern Backup story as their contracts land in the catalog.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
