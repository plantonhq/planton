# GCP Shared VPC Host

Enables a Google Cloud project as a Shared VPC HOST — the project whose VPC networks and subnetworks other projects (service projects, `GcpSharedVpcServiceProject`) attach to and deploy their workloads into. Shared VPC is how an organization keeps one network team owning one set of networks while many application teams deploy from their own projects with their own IAM and billing.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Shared VPC host enablement** -- the `compute_shared_vpc_host_project` flag on the project

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module whose identity holds `roles/compute.xpnAdmin` on the ORGANIZATION (or on a folder above the project). Shared VPC administration is organization-level; a project-level role is not enough.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **The project must exist** -- declare it with `GcpProject` and reference its `project_id` output, pass the project ID as a literal, or leave `projectId` empty to enable the project the credentials are configured for.
- **A project can be a host OR a service project**, never both.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSharedVpcHost
metadata:
  name: network-host
spec:
  projectId:
    value: acme-network-host
```

```shell
planton apply -f shared-vpc-host.yaml
```

## Configuration Reference

### Required Fields

None. An empty spec enables the provider's default project.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project that becomes the host, as a `GcpProject` reference or a literal ID. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` (disable the host role; fails while service projects are attached), `PREVENT` (destroy fails), or `ABANDON` (unmanaged, stays a host). |

### Validation Rules

- **`deletionPolicy`**: `DELETE`, `PREVENT`, or `ABANDON`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `host_project_id` | `string` | The project ID enabled as the host — the resolved value; what a `GcpSharedVpcServiceProject`'s `hostProjectId` references |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Disabling a host fails while service projects are attached** — destroy the `GcpSharedVpcServiceProject` resources first; a chart that references the host from its attachments does this in order.
- **Attaching a service project is not enough by itself**: its deployers and service agents still need `roles/compute.networkUser` on the host's subnetworks, granted with the catalog's IAM kinds.
- **Cost**: enabling a host is free; the networks and their traffic bill as they always did.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpSharedVpcServiceProject](/docs/catalog/gcp/gcpsharedvpcserviceproject) — attaches a service project to this host
- [GcpProject](/docs/catalog/gcp/gcpproject) — the project being enabled
- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the networks the host shares
- [GcpProjectIamMember](/docs/catalog/gcp/gcpprojectiammember) — the `networkUser` grants service projects need

## Additional Resources

- [Shared VPC overview](https://cloud.google.com/vpc/docs/shared-vpc)
- [Provisioning Shared VPC](https://cloud.google.com/vpc/docs/provisioning-shared-vpc)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
