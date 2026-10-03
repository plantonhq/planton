# GCP Service Connection Policy

Deploys a service connection policy (`google_network_connectivity_service_connection_policy`) — the per-network authorization that lets Google's service connectivity automation create Private Service Connect endpoints in your VPC on a managed-service producer's behalf. PSC-first services (Memorystore for Valkey, Redis Cluster) refuse to create instances on a network until a policy for their service class exists in the instance's region: this resource is that prerequisite.

## What Gets Created

One policy per (network, service class, region) triple. When a producer instance is created, the automation places PSC forwarding rules in the subnets you list, drawing endpoint IPs from them; the instance's connection details surface on the instance itself.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A VPC network** — referenced via `network` (a `GcpVpcNetwork` resource or a literal resource path)
- **At least one subnet** in the policy's region — referenced in `pscConfig.subnetworks`
- **IAM permissions** — `networkconnectivity.serviceConnectionPolicies.create` (e.g. `roles/networkconnectivity.consumerNetworkAdmin`)

## Quick Start

Create a file `scp.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpServiceConnectionPolicy
metadata:
  name: memorystore-valkey-policy
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

Deploy:

```shell
planton apply -f scp.yaml
```

Once the policy exists, PSC-first managed services of that class can be created on the network — e.g. a `GcpMemorystoreInstance` in the same region.

## Configuration Reference

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | Project owning the network and policy. |
| `policyName` | `string` | `metadata.name` | Name of the policy resource. Immutable. |
| `location` | `string` | — (required) | Region the policy applies to. Immutable. |
| `network` | `StringValueOrRef` | — (required) | Consumer VPC network (resource path; references resolve automatically). Immutable. |
| `serviceClass` | `string` | — (required) | The producer's published class (e.g. `gcp-memorystore`). Immutable. |
| `description` | `string` | — | Free-text purpose note. |
| `labels` | `map<string,string>` | — | User labels, merged beneath platform labels. |
| `pscConfig.subnetworks` | `StringValueOrRef[]` | — (min 1 when set) | Subnets endpoint IPs come from. Mutable. |
| `pscConfig.limit` | `int32` | GCP default | Max PSC connections under this policy. Mutable. |
| `pscConfig.producerInstanceLocation` | `string` | GCP default | Producer authorization mode; `CUSTOM_RESOURCE_HIERARCHY_LEVELS` activates the allowlist. |
| `pscConfig.allowedGoogleProducersResourceHierarchyLevels` | `string[]` | — | `projects/…`, `folders/…`, `organizations/…` entries producers may live in. |
| `deletionPolicy` | `string` | `DELETE` | What destroy does: `DELETE` removes the policy (stranding existing PSC endpoints), `PREVENT` fails the destroy to protect managed-instance connectivity, `ABANDON` leaves it authorizing unmanaged. |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `policy_id` | `string` | Fully qualified policy resource path |
| `name` | `string` | Short policy name |
| `infrastructure` | `string` | Underlying connectivity mechanism (PSC) |
| `etag` | `string` | Server-computed etag (changes on every mutation) |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **One policy per triple**: GCP rejects a second policy for the same (network, service class, region). A producer in another region needs its own policy there.
- **Deploy before the instance**: PSC-first services fail instance creation with a connectivity error until the policy exists. Keep the policy alive as long as any instance depends on it — deleting it strands existing endpoints and blocks new ones.
- **Address formats are normalized**: the Service Connectivity API requires relative resource paths; both engines strip `https://` self-link prefixes from `network` and `subnetworks`, so references and literals work in either form.
- **Regular subnets work**: service connection policies draw endpoint IPs from ordinary subnets — no special PSC purpose is required.
- **The immutables**: `location`, `network`, `serviceClass`, and the policy name are ForceNew; the `pscConfig` contents, description, and labels update in place — so subnet growth and limit raises never recreate the policy.

## Related Kinds

- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the network the policy authorizes connections into
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — supplies the endpoint IP space
- [GcpMemorystoreInstance](/docs/catalog/gcp/gcpmemorystoreinstance) — the first PSC-first consumer of this policy
- [GcpServiceNetworkingConnection](/docs/catalog/gcp/gcpservicenetworkingconnection) — the older peering-based private connectivity for PSA-era services (Cloud SQL, AlloyDB)
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project

## Additional Resources

- [Service connectivity automation overview](https://cloud.google.com/vpc/docs/about-service-connectivity-automation)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
