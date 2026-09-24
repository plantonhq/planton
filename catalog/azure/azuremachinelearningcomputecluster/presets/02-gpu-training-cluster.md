# GPU Training Cluster

This preset creates a scale-to-zero GPU cluster for deep-learning training -- two V100 nodes at peak, released ten minutes after their last job because GPU idle time is the most expensive waste in an ML estate.

## When to Use

- Deep-learning training that needs GPU acceleration
- Teams whose GPU quota is scarce -- scale-to-zero keeps the quota's cost honest
- Workspaces whose GPU quota lives in a different region (set `region` accordingly; the cluster's nodes may run away from the workspace)

## Key Configuration Choices

- **`Standard_NC6s_v3`** -- one V100 per node; swap for the family your quota grants (`az vm list-usage --location <region>` shows current usage)
- **`PT10M` idle duration** -- shorter than the CPU preset deliberately: warm GPU nodes bill at rates where ten idle minutes already matter
- **`maxNodeCount: 2`** -- a promise your quota must keep; the create succeeds regardless and the failure would arrive at scale-up
- **`SYSTEM_ASSIGNED` identity** -- grant data and registry access before jobs run

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-machine-learning-workspace-id>` | ARM ID of the parent workspace | `AzureMachineLearningWorkspace` status outputs (`machine_learning_workspace_id`), or reference it with valueFrom |

File the GPU family quota increase FIRST -- most subscriptions start at zero for NC/ND families.
