# GCP Vertex AI Workbench Instance -- Research and Design

## Deployment Landscape

### What Is Vertex AI Workbench?

Vertex AI Workbench is Google Cloud's managed notebook service. It provides JupyterLab instances running on Compute Engine VMs, pre-configured with popular ML frameworks, GPU drivers, and access to GCP data services. Workbench is the successor to AI Platform Notebooks (now deprecated).

### Resource Lineage

Google has iterated through three generations of managed notebooks:

1. **AI Platform Notebooks** (`google_notebooks_instance`) -- v1 API, deprecated
2. **Vertex AI Workbench Managed Notebooks** -- intermediate generation, also deprecated
3. **Vertex AI Workbench Instances** (`google_workbench_instance`) -- v2 API, current

Planton targets the current v2 API via `google_workbench_instance` (Terraform) and `workbench.Instance` (Pulumi).

### Key Terraform/Pulumi Resources

| Method | Resource | API Version |
|--------|----------|-------------|
| Terraform | `google_workbench_instance` | v2 |
| Pulumi | `workbench.Instance` | v2 |
| Terraform (deprecated) | `google_notebooks_instance` | v1 |

## Methods of Deployment

### 1. Google Cloud Console

The Workbench UI in the Cloud Console provides a wizard for creating instances. It's the most accessible method but doesn't support infrastructure-as-code workflows.

### 2. gcloud CLI

```bash
gcloud workbench instances create my-notebook \
  --location=us-central1-a \
  --machine-type=e2-standard-4
```

Good for ad-hoc creation but lacks state management.

### 3. Terraform

```hcl
resource "google_workbench_instance" "notebook" {
  name     = "my-notebook"
  location = "us-central1-a"
  
  gce_setup {
    machine_type = "e2-standard-4"
  }
}
```

The standard IaC approach. Our Terraform module follows this pattern.

### 4. Pulumi

```go
workbench.NewInstance(ctx, "notebook", &workbench.InstanceArgs{
    Location: pulumi.String("us-central1-a"),
    GceSetup: &workbench.InstanceGceSetupArgs{
        MachineType: pulumi.StringPtr("e2-standard-4"),
    },
})
```

### 5. Planton (This Component)

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVertexAiNotebook
metadata:
  name: my-notebook
spec:
  projectId:
    value: my-gcp-project
  location: us-central1-a
  machineType: e2-standard-4
```

Planton provides a declarative YAML interface with cross-resource composition, framework labels, and dual IaC backend support.

## Feature Coverage: 80/20 Analysis

### What We Cover (80% of Use Cases)

| Feature | Coverage | Rationale |
|---------|----------|-----------|
| Machine type selection | Full | Core configuration |
| GPU accelerators | Full (10 types) | Essential for ML training |
| Boot disk (type, size, CMEK) | Full | Storage foundation |
| Data disk (type, size, CMEK) | Full | Notebook data storage |
| VPC networking | Full | Enterprise security |
| Disable public IP | Full | Private deployments |
| Service account | Full | VM identity |
| VM images | Full (project, family, name) | Pre-built ML environments |
| Container images | Full (repo, tag) | Custom environments |
| Shielded VM | Full (Secure Boot, vTPM, integrity) | Security hardening |
| Desired state (ACTIVE/STOPPED) | Full | Cost management |
| Instance owners | Full | Access control |
| Metadata | Full | Custom configuration |
| Network tags | Full | Firewall integration |
| Labels | Framework-managed | Consistent labeling |

### What We Exclude (Niche/v2 Candidates)

| Feature | Exclusion Rationale |
|---------|-------------------|
| Confidential VM (SEV) | Very niche -- requires specific machine types and workloads |
| Reservation affinity | Reservation management is a separate concern |
| Third-party identity | Niche IdP integration for notebooks |
| Managed end-user credentials | Auto-enabled in newer versions |
| Access configs (explicit external IP) | Ephemeral IP is sufficient; BYOIP for notebooks is extremely rare |
| IP forwarding | Included but expected to be rarely used |

### Immutable Fields (ForceNew)

These fields cannot be changed after creation without destroying and recreating the instance:

- `location`, `instance_name`, `disable_proxy_access`
- `network_interface` (network, subnet, nic_type)
- `disable_public_ip`, `enable_ip_forwarding`
- `service_account`, `tags`
- `vm_image`, `container_image`
- `boot_disk.disk_type`, `boot_disk.kms_key`
- `data_disk.disk_type`, `data_disk.kms_key`

Disk sizes (boot and data) CAN be resized without recreation.

## Design Decisions

### 1. Flattened gce_setup

The Terraform/Pulumi providers nest all VM configuration under a `gce_setup` block. We flatten these to the top level of the spec because:

- The component IS a workbench instance -- the `gce_setup` wrapper adds no semantic value
- Matches the GcpComputeInstance pattern (boot_disk, network_interfaces at top level)
- Simpler YAML for users

### 2. Singular Sub-Messages

The providers use repeated fields for accelerator_configs, data_disks, network_interfaces, and service_accounts -- all with MaxItems: 1. We use singular messages for clarity:

- `accelerator_config` (not `accelerator_configs`)
- `data_disk` (not `data_disks`)
- `network_interface` (not `network_interfaces`)
- `service_account` (not `service_accounts`)

### 3. Int32 for Disk Sizes and Core Count

The providers use strings for `disk_size_gb` and `core_count`. We use `int32` because:

- Enables proto-level range validation (10-64000 for disks)
- Better developer experience (no quotes around numbers)
- IaC modules convert to strings internally

### 4. Derived disk_encryption

Instead of exposing a `disk_encryption` field, we derive it from the presence of `kms_key`:

- If `kms_key` is set → CMEK
- If `kms_key` is not set → GMEK (default)

This eliminates a redundant field and prevents inconsistent configurations.

### 5. Service Account as StringValueOrRef

Rather than a sub-message with `email` and `scopes` fields, we use a flat `StringValueOrRef` because:

- `scopes` is always `["cloud-platform"]` (computed, not configurable)
- The only user-facing field is the email address
- Flat StringValueOrRef enables direct infra-chart composition with GcpServiceAccount

## Best Practices

### Cost Management

- Use `desired_state: STOPPED` to suspend notebooks when not in use
- Compute charges stop; storage charges continue
- GPU instances are expensive -- always stop when not training

### Security

- Set `disable_public_ip: true` for production notebooks
- Configure a dedicated service account (don't use the default compute SA)
- Use CMEK encryption for regulated workloads
- Enable Shielded VM for defense-in-depth

### Networking

- Deploy in the same VPC as your data sources (BigQuery, GCS, etc.)
- Use Private Google Access for accessing GCP APIs without public IP
- Apply network tags for firewall rule targeting
