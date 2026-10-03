# AzureStorageObjectReplication - Pulumi Module

Pulumi implementation for the AzureStorageObjectReplication deployment
component.

## Architecture

```
storage.ObjectReplication (single resource -- the two-sided policy pair)
```

## Key Design Decisions

- **One resource IS the pair** -- Azure materializes the policy on both
  accounts under one server-assigned GUID; the provider creates the
  destination side first (assigning rule IDs) and mirrors it onto the
  source, and destroy removes both.
- **`policy_id` is parsed from the destination-side ARM id** (the
  authoritative copy), identically to the Terraform module, so the
  monitoring-facing GUID output is byte-identical across engines.
- **`prefix_match` maps onto the provider's
  `filterOutBlobsWithPrefixes`** -- the spec uses ARM's own name
  because these are INCLUDE filters despite the provider attribute's
  "filter out" wording.
- **`copy_blobs_created_after` is presence-guarded** -- IaC inputs do
  not materialize proto defaults, so unset falls through to the
  provider's own OnlyNewObjects default on both engines.
- **`metrics_enabled` is a recorded skip** -- pulumi-azure v6.38 has
  not bridged it; a one-engine-only input would ship silent-drop
  divergence (see ../../GUIDE.md for the re-enable trigger).
- **No Azure tags**: ARM does not support tags on
  objectReplicationPolicies; the platform's identity tags live on the
  two accounts.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
