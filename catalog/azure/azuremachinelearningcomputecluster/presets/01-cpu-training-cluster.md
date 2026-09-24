# CPU Training Cluster

This preset creates a scale-to-zero dedicated CPU cluster -- the everyday shared training pool: free between jobs, four general-purpose nodes at peak, a system identity for credential-free data access.

## When to Use

- The first compute a new workspace needs -- one pool the whole team submits jobs to
- Everyday CPU training, data prep, and pipeline steps
- Any workload where a few minutes of node spin-up latency is acceptable

## Key Configuration Choices

- **`minNodeCount: 0`** -- the cluster costs nothing while idle; nodes provision on demand
- **`PT30M` idle duration** -- finished nodes stay warm half an hour for closely-spaced experiments before releasing
- **`DEDICATED` priority** -- nodes stay up until scaled down; no eviction risk
- **`SYSTEM_ASSIGNED` identity** -- grant it Storage Blob Data Reader on training data and AcrPull on the registry BEFORE jobs run; identity updates in place, so grants can be fixed on a live cluster

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-machine-learning-workspace-id>` | ARM ID of the parent workspace | `AzureMachineLearningWorkspace` status outputs (`machine_learning_workspace_id`), or reference it with valueFrom |

The `region` and `vmSize` carry realistic examples (`eastus`, `Standard_DS3_v2`) -- set your own region and check the family's regional vCPU quota covers `maxNodeCount`.
