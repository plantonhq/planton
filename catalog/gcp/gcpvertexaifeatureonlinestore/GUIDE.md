# GcpVertexAiFeatureOnlineStore Guide

The judgment this guide protects: an online store is standing capacity. Pick the storage kind for the serving pattern, size the floor honestly, and let feature views carry the per-model decisions -- because the storage kind is permanent and the floor bills every hour.

## Bigtable or Optimized

A new store serves from Bigtable. Google no longer creates stores with Optimized serving: it refuses with `Creating FeatureOnlineStore with Optimized serving is no longer supported.`

`bigtable` gives the store a managed Cloud Bigtable instance that autoscales between `minNodeCount` and `maxNodeCount` (Google allows at most ten times the floor) toward a CPU target (10-80 percent, Google's default 50). It suits large feature sets and high request volumes served by entity key, and `enableDirectBigtableAccess` lets clients read the instance directly.

`optimized: true` describes an Optimized store that already exists: Google's Optimized infrastructure behind a dedicated endpoint. Such stores keep serving and stay manageable -- labels, the dedicated endpoint, the key, and the views update in place -- so the flag is there to adopt and manage them; Google's optimized block has no settings, so it is a flag. Exactly one of the two is set, and switching replaces the store and every view -- for an Optimized store, the replacement must be Bigtable.

## The dedicated endpoint

Only an Optimized store has a dedicated endpoint, public by default; the `public_endpoint_domain_name` output is where its clients call. On an existing Optimized store, set `dedicatedServingEndpoint.privateServiceConnectConfig` with `enablePrivateServiceConnect: true` to serve only over Private Service Connect: Google publishes a service attachment (the `service_attachment` output, populated once a view has synced) that consumer projects in `projectAllowlist` target with a forwarding rule. Leave the block unset to keep Google's default. Private Service Connect is an Optimized-store feature, so the spec refuses it on a Bigtable store -- which every new store is. A Bigtable store serves through the shared Vertex AI endpoint under IAM.

## Feature views are the product

A feature view is what a model queries: one set of features keyed by entity ID. A registry-sourced view (`featureRegistrySource`) selects features by id from `GcpVertexAiFeatureGroup` blocks in the same location -- one definition shared with training, the recommended path. A BigQuery-sourced view (`bigQuerySource`) materializes a table directly with one entity ID column. `syncConfig.cron` refreshes on a schedule (prefix `CRON_TZ=...` to pin a time zone); `continuous: true` streams changes. Every sync reads BigQuery as the Vertex AI Service Agent, which needs `roles/bigquery.dataViewer` on the sources.

## IDs use underscores

Store and view ids allow up to 60 lowercase letters, digits, and underscores, not starting with a digit, so they are set explicitly rather than derived from the resource name. Both are fixed at creation.

## Encryption and destroy

`kmsKeyName` encrypts both the online and the offline copies under your key. `deletionPolicy` fans to every view: `DELETE` removes the views, then the store (and Google deletes the managed Bigtable instance); `PREVENT` makes destroy fail; `ABANDON` leaves the store serving -- and billing. Google refuses to delete a store that still holds views the block does not manage unless `forceDestroy` is set.
