# GcpVertexAiPersistentResource Guide

The judgment this guide protects: a persistent resource trades money for readiness. Every replica bills from provisioning to deletion, so keep one only when instant job starts or held accelerators are worth that, size it to concurrent demand, and decide its networking once -- because almost nothing about it can change in place.

## When a persistent resource pays off

Vertex AI custom training provisions machines per job, which takes minutes and can fail when accelerators are scarce. A persistent resource provisions the machines once and keeps them: jobs that name its id in `persistent_resource_id` start in seconds on machines already running, and GPUs or TPUs stay held between jobs. Ray on Vertex AI clusters are persistent resources too. If jobs are rare and a few minutes of startup do not matter, on-demand training is cheaper.

## Pools

Each `resourcePools[]` entry is a set of identical machines: `machineSpec` (a custom-training machine type and, optionally, an accelerator type and count Google pairs with it), a `replicaCount`, optional `autoscalingSpec` bounds (the floor must be at least 1 on a persistent resource), and optional boot disk settings. A pool's `id` is what a job's worker pool refers to; Google generates one when you leave it empty. Counts are whole numbers in the spec and travel to Google as decimal strings. Only `replicaCount` changes in place -- the machine, the disk, the autoscaling bounds, and the id are fixed.

## Networking is decided once

A job must use the same network as the resource it runs on. Two paths reach a VPC. `network` peers the resource through private services access: the network needs a `GcpServiceNetworkingConnection`, and `reservedIpRanges` can pin the machines to named ranges of it. Google wants the network as `projects/{project number}/global/networks/{name}`; a `GcpVpcNetwork` reference or a path with a project ID is resolved to the number by one project lookup at plan time. `pscInterfaceConfig` instead attaches the machines to a network attachment through a Private Service Connect interface, with optional DNS peering so jobs resolve private hostnames (the Vertex AI Service Agent needs `roles/dns.peer` on the target project). Google treats the two as alternatives. Both, like everything else here, are permanent.

## Identity and encryption

`enableCustomServiceAccount: true` requires every job on the resource to run as a user-managed service account the job names, instead of the Vertex AI Custom Code Service Agent -- the choice for least-privilege training. `kmsKeyName` encrypts the machines' disks under your key, and jobs on the resource must use the same key. Both are fixed at creation.

## Destroy

`deletionPolicy` decides what a destroy does: `DELETE` releases the machines and stops billing, `PREVENT` makes destroy fail, and `ABANDON` leaves the resource running -- and billing -- outside management. Running jobs lose their machines on delete, so drain them first.
