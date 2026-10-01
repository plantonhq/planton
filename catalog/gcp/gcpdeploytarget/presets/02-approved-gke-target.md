# Approved GKE Target

## Use Case

Deploy to a production GKE cluster whose control plane is private, behind a human approval, with a dedicated deployer identity and private build machines.

## When to Use

- Production clusters with a private control-plane endpoint
- Teams that separate who releases from who approves production

## What This Creates

- The Cloud Deploy API on the delivery project
- A GKE target that requires approval, reaches the cluster on its internal IP, and runs render, deploy, and verify jobs on a private worker pool as a dedicated service account

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `gke.cluster` | a `GcpGkeCluster` reference | The production cluster. |
| `gke.internalIp` | `true` | Remove for a cluster with a public endpoint, or use `dnsEndpoint` instead. |
| `executionConfigs[].workerPool` | a `GcpCloudBuildWorkerPool` reference | A pool peered into the cluster's network. |
| `executionConfigs[].serviceAccount` | a `GcpServiceAccount` reference | Grant it `roles/clouddeploy.jobRunner` and `roles/container.developer`. |
| `executionConfigs[].executionTimeout` | `1800s` | Between `600s` and `86400s`. |
