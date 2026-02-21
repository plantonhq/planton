---
title: "Preemptible Dev OKE Node Pool"
description: "This preset creates a cost-optimized OKE node pool using preemptible (spot) instances for development, testing, and experimentation. Preemptible nodes use the same shapes as on-demand nodes but are..."
type: "preset"
rank: "03"
presetSlug: "03-preemptible-dev"
componentSlug: "container-engine-node-pool"
componentTitle: "Container Engine Node Pool"
provider: "oci"
icon: "package"
order: 3
---

# Preemptible Dev OKE Node Pool

This preset creates a cost-optimized OKE node pool using preemptible (spot) instances for development, testing, and experimentation. Preemptible nodes use the same shapes as on-demand nodes but are significantly cheaper because OCI can reclaim them when capacity is needed. The pool uses minimal resources, a single availability domain, and omits production features like encryption, NSGs, and rolling upgrade controls. This preset pairs with the `03-development` OKE cluster preset.

## When to Use

- Development and testing environments where occasional node reclamation is acceptable
- CI/CD environments that spin up ephemeral node pools for integration testing and tear them down afterward
- Budget-constrained teams that want functional Kubernetes worker nodes at the lowest possible cost
- Learning and experimentation with OKE where fast provisioning matters more than resilience
- Batch processing workloads on Kubernetes that tolerate interruption and can reschedule pods automatically

## Key Configuration Choices

- **Preemptible instances** (`preemptibleNodeConfig`) -- Nodes can be terminated by OCI at any time when capacity is reclaimed. OKE handles this gracefully by rescheduling pods to remaining nodes. Setting `isPreserveBootVolume: false` deletes the boot volume on reclamation since dev data is ephemeral and there is no need to pay for orphaned volumes.
- **1 OCPU / 16 GB memory** (`nodeShapeConfig`) -- Minimal viable node for running a handful of development pods. The 1:16 ratio is the standard general-purpose ratio at the smallest possible size. Sufficient for most development workloads including build agents, test runners, and small services.
- **2 nodes, single AD** (`nodeConfigDetails.size: 2`, one `placementConfig`) -- Two nodes provide basic pod rescheduling capability if one node is reclaimed. Single AD avoids the complexity of multi-AD placement that is unnecessary for dev environments. If your region has only one AD (e.g., single-AD regions like Zurich or Jeddah), this works without modification.
- **No pod networking configuration** -- Omitted entirely, which means the node pool inherits the cluster's CNI behavior. This preset pairs with the `03-development` cluster that uses flannel overlay, which requires no pod subnet or pod NSG configuration at the node pool level.
- **SSH key for debugging** (`sshPublicKey`) -- Included as a placeholder because dev environments benefit from direct SSH access to nodes for troubleshooting pod scheduling, container runtime issues, and node-level debugging.
- **No NSGs, encryption, cycling, or eviction settings** -- Intentionally omitted. Dev node pools do not need network segmentation, KMS encryption, zero-downtime rolling upgrades, or graceful eviction controls. This keeps the preset minimal and fast to deploy. Use preset 01 or 02 when graduating to production.
- **No kubernetesVersion** -- Omitted so the node pool inherits the cluster's Kubernetes version automatically. In dev environments, version pinning adds friction without meaningful benefit.
- **Node label for scheduling** (`initialNodeLabels: pool=dev`) -- Enables workloads to target dev nodes specifically, useful when a cluster has both production and dev node pools (uncommon but possible in shared clusters).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the node pool will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<oke-cluster-ocid>` | OCID of the OKE cluster to attach this node pool to | OCI Console > Developer Services > Kubernetes Clusters, or `OciContainerEngineCluster` status outputs |
| `<availability-domain>` | Availability domain name (e.g., `Uocm:PHX-AD-1`) | `oci iam availability-domain list` or OCI Console |
| `<worker-subnet-ocid>` | OCID of the subnet for worker node VNICs | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<ssh-public-key>` | SSH public key content (e.g., `ssh-rsa AAAA...`) | Your local `~/.ssh/id_rsa.pub` or equivalent |

## Related Presets

- **01-standard-production** -- Use instead for on-demand production node pools with HA, NSGs, and rolling upgrades
- **02-hardened-encrypted** -- Use instead for regulated environments requiring encryption and strict operational controls
