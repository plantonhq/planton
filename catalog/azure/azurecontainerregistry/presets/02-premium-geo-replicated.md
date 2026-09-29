# Premium Geo-Replicated Registry

This preset creates a Premium registry with a zone-redundant home replica, one geo-replication, and automatic purging of untagged manifests. It is the multi-region production shape: images push once and serve locally in every replicated region, and the registry keeps serving pulls if a replicated region goes down.

## When to Use

- Workloads deployed in more than one Azure region pulling the same images
- Pull-latency- or egress-cost-sensitive fleets (each replica serves its region locally, with no cross-region egress fees)
- Registries that must survive a regional outage

## Key Configuration Choices

- **`sku: PREMIUM`** -- geo-replication, zone redundancy, and the retention policy are all Premium-gated (spec validation enforces the same gates ARM does)
- **`zoneRedundancyEnabled: true`** -- both the home replica and the replication spread across availability zones; this is fixed at creation for the home replica, so decide it here
- **`georeplications`** -- one entry per additional region; the list must not contain the home region. Add or remove replicas in place as the deployment footprint changes
- **`retentionPolicyInDays: 30`** -- CI pushes constantly re-tag images and orphan the old manifests; a 30-day untagged-manifest window keeps storage bounded without risking anything still referenced

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | The home replica's region | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `yourregistry123` (`registryName`) | Globally unique registry name (5-50 lowercase alphanumerics) | Becomes `{name}.azurecr.io` |
| `<secondary-azure-region>` | The region to replicate into (differs from home) | Your regional deployment strategy |
