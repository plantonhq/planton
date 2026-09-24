# DigitalOcean Kubernetes Cluster -- Operational Guide

Judgment calls that matter when you run DOKS clusters.

## The version field is a creation pin, not an upgrade lever

`kubernetesVersion` decides what the cluster is CREATED at. After that, both provisioners deliberately ignore drift on it, for a hard reason: the Terraform provider destroys and recreates the entire cluster whenever the configured version is LOWER than the live one, and `autoUpgrade` routinely moves the live version ahead of your pin. Without the ignore, an auto-upgraded cluster plus a stale manifest equals an accidental cluster replacement.

Practical consequences:

- Enable `autoUpgrade` and set a `maintenancePolicy` window; patch upgrades happen there.
- Minor/major upgrades are an operational action (DigitalOcean control panel or `doctl kubernetes cluster upgrade`), not a spec edit.
- Prefer a minor prefix (`"1.35"`) over a full slug -- the next section says why.

## The version must be one DigitalOcean offers today

DigitalOcean keeps only a short list of creatable Kubernetes versions -- three minors, one patch slug each -- and rotates the patch slugs every few weeks (`GET /v2/kubernetes/options` or `doctl kubernetes options versions`; on 2026-09-16 the list was `1.34.10-do.5`, `1.35.7-do.5`, `1.36.3-do.5`). A manifest that names a retired slug fails at apply with `422`, and it fails weeks after it was written and worked. Two ways to stay creatable:

- **Name the minor** (`"1.35"`). DigitalOcean resolves it to the current patch at create time, and the value stays valid for the minor's whole support window. This is what the presets and examples here do.
- **Name the full slug** only when the exact starting patch matters, and expect to bump it. The read-back is always the full slug either way, and both provisioners ignore drift on the field, so a prefix never produces a diff.

## Size the default pool once; grow with separate pools

Changing the default pool's `size` (or `gpuPartitionMode`) does not resize the pool — it REPLACES THE ENTIRE CLUSTER, workloads included. The inline pool is the cluster's foundation; treat it that way:

- Pick a size with headroom for system pods plus your steady-state base load.
- Add capacity classes (bigger nodes, GPU nodes, tainted dedicated pools) as separate `DigitalOceanKubernetesNodePool` resources — those resize and replace independently without touching the cluster.
- With `autoScale: true`, leave `nodeCount` out — the manifest is rejected if you set both. The pool starts at `minNodes` and DigitalOcean's cluster-autoscaler owns the count between `minNodes` and `maxNodes` from then on. The reason is the provider's own behavior: it writes the live count back into `node_count` on every read and re-applies a stated one on every update, so a stated count and a running autoscaler fight each other on every apply (measured: a pool that autoscaled to two nodes planned `node_count 2 -> 1` on both provisioners). A fixed pool (`autoScale` off) states `nodeCount` and scales by editing it.

## HA is a one-way door with a price tag

`highlyAvailable: true` gives the control plane multiple replicas and a real SLA, at an extra monthly cost — and it can never be turned off again. Turn it on for production clusters whose API must survive a control-plane node loss; leave it off for everything else. Note that on newer DOKS versions DigitalOcean's own default is HA ON — this component sends an explicit false when unset, so you never get a surprise HA bill.

## Firewall the control plane, but don't lock yourself out

`controlPlaneFirewall` restricts who can reach the public Kubernetes API endpoint. Two traps:

- The allowed list must include wherever `kubectl`, CI, and the provisioner itself run. Locking the provisioner out turns every subsequent apply into a timeout.
- `enabled` is explicit, so you can stage an address list with `enabled: false` and flip it on when the list is proven.

## Network placement is create-only

`vpc`, `clusterSubnet`, `serviceSubnet`, `workerSubnetUuid`, and `isolatedWorkers` are all fixed at creation. Decide them first; retrofitting means a new cluster and a workload migration. Custom pod/service CIDRs matter when the VPC peers with networks that would collide with DigitalOcean's defaults — set them then, leave them unset otherwise.

## The control plane has no public IPv4 any more

`ipv4Address` is exported because the provider exposes it, but clusters DigitalOcean creates today report none -- measured on a single-replica cluster, so this is not only the HA case the field's history suggests. The Kubernetes API server is reached through the `apiServerEndpoint` hostname (`https://<cluster-id>.k8s.ondigitalocean.com`), which fronts it. Allowlist by that hostname, and treat an empty `ipv4Address` as normal, not as a failed deploy.

## Destroying: the cluster is gone in seconds, the network remembers for minutes

A cluster delete is accepted immediately and the cluster answers `404` within a couple of seconds -- both provisioners' destroys finish in under ten. Its worker Droplets take longer to disappear, and while they do they stay listed as members of the cluster's VPC. A VPC delete issued in that window fails `409 Can not delete VPC with members` (measured: about two minutes after the cluster was gone). Tear down in dependency order and give the network a short wait after the cluster, or let the retry that any sane teardown already has absorb it -- this is a lag, not the hours-long ghost-member class a failed database create can leave.

## The kubeconfig output is a credential

`kubeconfig` is raw YAML (not base64) carrying admin credentials; write it to a file, `chmod 600` it, and point `KUBECONFIG` at it. Credentials in it expire — `kubeconfigExpireSeconds` controls the validity (0 means DigitalOcean's 7-day default); re-fetching state mints fresh ones.

## Addon toggles: unset means DigitalOcean decides

Every addon field (`routingAgent`, `corednsAutoscaler`, the GPU device plugins and DRA drivers, `rdmaSharedDevicePlugin`, `p2pOciRegistryPlugin`) is a message with one `enabled` leaf. Leaving the field out defers to DigitalOcean's own default for that addon; setting it asserts the state, on or off (both directions are live-proven: an explicit `false` on a version whose default is on reads back off). The AMD and NVIDIA device plugins are each mutually exclusive with their DRA drivers — the manifest rejects both together before any provisioner runs.

Two addons carry version facts worth knowing before you set them. `p2pOciRegistryPlugin` needs Kubernetes 1.36.0-do.2 or later — on an older version DigitalOcean refuses the whole cluster create with a validation 422 ("p2p-oci-registry is only supported on DOKS v1.36.0-do.2 or later") and creates nothing, so a manifest that enables it must pin `kubernetesVersion` at `"1.36"` or newer. `corednsAutoscaler`'s default flips with the version — off through 1.35, on from 1.36 — so a cluster that leaves it unset changes behavior when it upgrades across that line; set it explicitly if that matters to you.

## Destroy-time cleanup: read before you set it

`destroyAllAssociatedResources: true` makes destroy also delete every load balancer, volume, and volume snapshot the cluster created. That is the right call for ephemeral clusters and the wrong one anywhere volumes outlive the cluster. It only acts at destroy; it is invisible until then.

## Choosing a provisioner

There is nothing to choose on this kind: every spec field deploys identically on Terraform and Pulumi, addon toggles included. What differs is DigitalOcean's own prerequisites, and they bind both provisioners equally: `isolatedWorkers` is refused unless the cluster's VPC has a NAT gateway attached (the nodes then run on dedicated hardware and bill accordingly, and it can only be set at creation); `workerSubnetUuid` must name a subnet inside the cluster's VPC and requires `vpc` to be set; `sso` needs a working OIDC issuer and client; the GPU device plugins, DRA drivers, and RDMA plugin are accepted only on clusters with GPU node pools, and each device-plugin/DRA-driver pair is mutually exclusive. `corednsAutoscaler` defaults ON from Kubernetes 1.36; on older versions set it explicitly.

## Importing an existing cluster

Import uses the bare cluster UUID and MUTATES the remote cluster when its default pool is untagged: the provider adds a `terraform:default-node-pool` marker tag to single-pool clusters and refuses multi-pool clusters until one pool carries the tag manually. Never put that tag in `spec.tags` — the provider owns it. Also expect `registryIntegration`, `kubeconfigExpireSeconds`, and `destroyAllAssociatedResources` to stay at their configured values after import: the API never reports them back.

## What is deliberately NOT here

Additional node pools are `DigitalOceanKubernetesNodePool` resources. Load balancers and volumes appear when Kubernetes Services and PersistentVolumeClaims create them — they are Kubernetes-driven, not cluster fields. Container registries are their own resource; `registryIntegration` only wires an existing one into the cluster.
