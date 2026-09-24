# Azure Machine Learning Compute Cluster -- Operational Guide

Judgment that saves real time when running compute clusters. The field reference lives in the API Explorer; this is the operational layer above it.

## Scale to zero unless you have a latency argument

`minNodeCount: 0` is the default posture worth defending: the cluster costs nothing between jobs, and most training workloads tolerate the few minutes of node spin-up. Keep a non-zero minimum only where a team genuinely iterates interactively against the cluster all day -- and remember every warm node bills around the clock at the VM's full rate. Live-proven: a scale-to-zero dedicated cluster (Standard_DS2_v2, min 0 / max 1, system identity) creates as an ARM compute object in about one to two minutes and holds zero nodes for its whole life -- the object is what you pay time for, not a VM.

## The idle duration is your real cost knob

`scaleDownNodesAfterIdleDuration` decides how long a finished node waits for the next job. `PT2M` releases nodes aggressively (cheapest, most spin-up churn); `PT30M`-`PT1H` keeps nodes warm through a working session of closely-spaced experiments. Tune it per cluster, not globally -- a nightly-batch cluster and an iteration cluster want opposite settings. It updates in place; changing it is free.

## Low-priority nodes are a contract, not a discount switch

`LOW_PRIORITY` nodes cost a fraction of dedicated but Azure evicts them at any time, taking running work with them. The contract: your training code checkpoints and resumes, or the job simply reruns. Never put an unresumable multi-hour job on a low-priority cluster to save money -- the eviction costs more than the discount saved. Priority is ForceNew; converting a cluster means replacing it.

## Almost everything replaces the cluster -- plan for it

Only `identity`, `scaleSettings`, and `tags` update in place. VM size, priority, SSH, networking, description -- all ForceNew. Replacement is cheap for the cluster itself (it holds no data), but every RUNNING job on it dies. Change ForceNew fields in a window when the cluster is idle; check the workspace's job list first.

## Quota failures arrive at scale-up, not at create

The cluster object creates happily with a `maxNodeCount` your subscription cannot honor -- the failure surfaces later, when a job forces a scale-up past your regional vCPU quota for that VM family. Before sizing a GPU cluster, check the family quota (`az vm list-usage --location <region>`) and file the increase; treat `maxNodeCount` as a promise your quota must be able to keep.

## The cluster's identity is how jobs reach data

Give the cluster a managed identity and grant it: Storage Blob Data Reader/Contributor on the datastores' storage, Key Vault access where jobs read secrets, and AcrPull on the workspace's container registry. Identity updates IN PLACE -- you can add or fix grants on a live cluster without replacing it. With `localAuthEnabled: false`, identity becomes the only path -- flip it only after the grants are proven.

## Cross-region clusters read back at the workspace region

The cluster's `region` places its NODES -- useful when GPU quota lives in a different region from the workspace. ARM, however, reports the cluster envelope at the WORKSPACE's region; do not be surprised when portal views and ARM reads show the workspace region for a cluster whose nodes run elsewhere.

## SSH is for debugging, closed by default

`sshPublicAccessEnabled` defaults to false and should stay there; enable it (with the `ssh` block's admin account) only while diagnosing node-level problems, and prefer the SSH public key over a password. The whole SSH block is ForceNew -- decide the admin account at creation, not after.
