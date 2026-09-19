# GCP Shared VPC Service Project

Attaches a Google Cloud project to a Shared VPC host (`GcpSharedVpcHost`) as a SERVICE project, so its workloads can be placed in the host's subnetworks. The application team keeps its own project, IAM, quotas, and bill; the network team keeps the one set of networks every team shares. One resource per attachment.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Shared VPC attachment** -- the `compute_shared_vpc_service_project` binding a service project to its host

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module whose identity holds `roles/compute.xpnAdmin` on the organization (or a folder above both projects).
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Shared VPC Host and Project

- **The host must be enabled** -- declare it with `GcpSharedVpcHost` and reference its `host_project_id` output, or pass the host's project ID as a literal.
- **The service project must exist** -- declare it with `GcpProject` and reference its `project_id` output, or pass the ID as a literal. Both projects are required.
- **A project can be a service project of at most one host.**

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSharedVpcServiceProject
metadata:
  name: payments-attach
spec:
  hostProjectId:
    value: acme-network-host
  serviceProjectId:
    value: acme-payments-prod
```

```shell
planton apply -f shared-vpc-service-project.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `hostProjectId` | `StringValueOrRef` | The host, as a `GcpSharedVpcHost` reference or the host's project ID. Immutable. |
| `serviceProjectId` | `StringValueOrRef` | The project being attached, as a `GcpProject` reference or a literal ID. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `deletionPolicy` | `string` | detach | Empty detaches on destroy (fails while resources in the service project still use a host subnetwork); `ABANDON` leaves the attachment in place. This resource accepts no other value. |

### Validation Rules

- **`deletionPolicy`**: empty or `ABANDON`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `service_project_id` | `string` | The attached service project's ID |
| `host_project_id` | `string` | The host project's ID |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Both projects are immutable** — moving a service project to another host is a detach and an attach.
- **Attaching is not enough by itself**: the service project's deployers and Google's service agents still need `roles/compute.networkUser` on the host's subnetworks (or the host project). Grant that with `GcpProjectIamMember` on the host project or subnetwork-level IAM — it is per team and has its own lifecycle.
- **Cost**: the attachment is free.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpSharedVpcHost](/docs/catalog/gcp/gcpsharedvpchost) — the host this project attaches to
- [GcpProject](/docs/catalog/gcp/gcpproject) — the service project
- [GcpProjectIamMember](/docs/catalog/gcp/gcpprojectiammember) — the `networkUser` grants on the host
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — the host's subnetworks the service project uses

## Additional Resources

- [Shared VPC overview](https://cloud.google.com/vpc/docs/shared-vpc)
- [Provisioning Shared VPC: attaching service projects](https://cloud.google.com/vpc/docs/provisioning-shared-vpc#create-shared)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
