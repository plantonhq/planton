# GCP Service Connection Policy

Authorizes Google's service connectivity automation to place Private Service Connect endpoints in your VPC for a producer's service class — the prerequisite PSC-first managed services (Memorystore for Valkey, Redis Cluster, and the producers that follow them) check before creating instances in a region. Without the policy, instance creation fails with a connectivity error that says nothing about policies; with it, endpoints appear automatically in your subnets and their IPs surface on the instance. One policy exists per (network, service class, region) triple, and it must outlive every instance that depends on it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Service Connection Policy** — a `google_network_connectivity_service_connection_policy` binding one service class to one network in one region, carrying the PSC subnet address space, the optional connection limit, and the optional producer hierarchy allowlist
- **Network Connectivity API enablement** — `networkconnectivity.googleapis.com` enabled in the target project (the control plane that owns these policies; never disabled on destroy)
- **Compute Engine API enablement** — `compute.googleapis.com` enabled in the target project (the network, subnets, and the automation's forwarding rules are Compute-side objects; never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** — an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A VPC network** for the `network` field — the consumer network the automation is authorized to connect into (a GcpVpcNetwork reference or a literal `projects/{project}/global/networks/{name}` path).
- **At least one subnet in the policy's region** for `pscConfig.subnetworks` — regular-purpose subnets work; no special PSC purpose is needed (unlike PSC NAT subnets for published services).
- **The producer's published service class name** for `serviceClass` — Google publishes one identifier per service (`gcp-memorystore` for Memorystore for Valkey, `gcp-memorystore-redis` for Redis Cluster); third-party producers publish their own.

## Deploy

### Console

Open the deployment store, find **GCP Service Connection Policy**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the network, region, and service class, and the PSC subnet and limit settings. Start from the **Memorystore for Valkey Policy** preset in the [Presets](#presets) tab to pre-populate the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpServiceConnectionPolicy
metadata:
  name: memorystore-valkey-policy
  org: acme-corp
  env: prod
spec:
  location: us-central1
  network:
    value: projects/acme-prod-networking/global/networks/prod-vpc
  serviceClass: gcp-memorystore
  pscConfig:
    subnetworks:
      - value: projects/acme-prod-networking/regions/us-central1/subnetworks/prod-subnet
```

```shell
planton apply -f service-connection-policy.yaml
```

This authorizes Memorystore for Valkey to place PSC endpoints in `prod-subnet` — after which creating a Memorystore instance on this network in `us-central1` just works. An Infra Job tracks the provisioning in real time.

### InfraChart

When the network and subnets are deployed in the same chart, wire the policy to them with ValueFromRef:

```yaml
spec:
  location: us-central1
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: prod-vpc
      fieldPath: status.outputs.network_id
  serviceClass: gcp-memorystore
  pscConfig:
    subnetworks:
      - valueFrom:
          kind: GcpSubnetwork
          name: prod-subnet
          fieldPath: status.outputs.subnetwork_self_link
```

The InfraPipeline deploys the network and subnet first, then creates the policy — and any Memorystore instance later in the graph finds its authorization already in place.

## Key Configuration

These are the most important decisions when configuring a service connection policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Service class accuracy** — `serviceClass` is not free text: it must be the producer's published identifier exactly. A typo creates a valid policy that authorizes nothing, and the resulting instance-create failure looks identical to the missing-policy case — the least self-explanatory error in this corner of GCP.

**Immutable identity** — `location`, `network`, `serviceClass`, and the policy name are all create-time; changing any of them destroys and recreates the policy. Everything inside `pscConfig` (plus `description` and `labels`) updates in place — growing subnets or raising the limit never disturbs existing endpoints.

**Deploy order** — create the policy BEFORE the first instance of its service class in that region, one policy per (network, service class, region) triple — GCP rejects a second. In charts, make producer instances depend on the policy.

**PSC address space and the limit** — `pscConfig.subnetworks` is where endpoint IPs come from; the subnets must live in the policy's region and network. `pscConfig.limit` caps how many connections the automation may create: set it deliberately in shared networks so a runaway instance-creation loop hits the cap instead of exhausting subnet space.

**Producer allowlist** — `producerInstanceLocation: CUSTOM_RESOURCE_HIERARCHY_LEVELS` restricts producers to the projects, folders, or organizations listed in `allowedGoogleProducersResourceHierarchyLevels`. The two fields only work together — validation enforces both-or-neither, because an allowlist under the default location mode is silently ignored by GCP.

**Deletion policy** — deleting the policy strands every PSC endpoint created under it and blocks new instances of the class in that region, while the producer instances keep running and failing. `deletionPolicy: PREVENT` is the right posture wherever managed instances depend on the policy.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `network` | `status.outputs.network_id` |
| **GcpSubnetwork** (per entry) | `pscConfig.subnetworks` | `status.outputs.subnetwork_self_link` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `policy_id` | Fully qualified path (`projects/{project}/locations/{location}/serviceConnectionPolicies/{name}`) | Addressing the policy in gcloud and audit tooling |
| `infrastructure` | The connectivity mechanism the automation uses (`PSC`) | Confirming the mechanism without inspecting individual connections |
| `etag` | Server-computed change token, changes on every mutation | Change detection when auditing shared networks |

Producer instances do not reference these outputs — they consume the policy by its existence on the network, which is why deploy order (policy first) matters more than wiring.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Memorystore for Valkey in one region** — the minimal shape: one network, one subnet, `gcp-memorystore`. Start from the **Memorystore for Valkey Policy** preset.

**Shared VPC with a guardrail** — the policy lives with the network in the host project (explicit `projectId`), and `limit` caps how many instances service projects can attach before the platform team deliberately raises it. Start from the **Shared VPC with Connection Cap** preset.

**Producer allowlist** — custom hierarchy mode restricting which producer projects, folders, or organizations may connect into the network — the posture for regulated environments. Start from the **Producer Hierarchy Allowlist** preset.

## Works With

- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) — the consumer network the policy authorizes connections into
- [**GCP Subnetwork**](/infra-catalog/gcp-subnetwork) — the address space PSC endpoint IPs are drawn from
- [**GCP Memorystore Instance**](/infra-catalog/gcp-memorystore-instance) — the PSC-first producer that requires this policy before it can deploy
- [**GCP Project**](/infra-catalog/gcp-project) — provides the project that owns the network and the policy
