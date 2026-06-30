# Pulumi Module - GCP GKE Node Pool

This directory contains the Pulumi implementation for the **GcpGkeNodePool** resource, enabling declarative node pool management using Go.

## Overview

The Pulumi module provisions GKE node pools with production-ready defaults:
- **Autoscaling or fixed size**: Scale-to-zero support or predictable capacity
- **Spot VMs**: Cost optimization for fault-tolerant workloads
- **Auto-upgrade and auto-repair**: Enabled by default with customizable settings
- **Custom service accounts**: Least-privilege IAM
- **Flexible disk options**: Standard, SSD, or Balanced persistent disks
- **Node labels**: For Kubernetes workload scheduling

## Directory Structure

```
pulumi/
├── main.go              # Pulumi program entrypoint
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build and deployment automation
├── debug.sh             # Interactive debugging helper
├── README.md            # This file
├── overview.md          # Architecture and design documentation
└── module/
    ├── main.go          # Module orchestration
    ├── locals.go        # Local value transformations
    ├── outputs.go       # Stack outputs mapping
    └── node_pool.go     # Node pool resource provisioning
```

## Prerequisites

### Tools

- **Pulumi CLI** v3.0+ ([install](https://www.pulumi.com/docs/install/))
- **Go** 1.21+ ([install](https://go.dev/doc/install))
- **gcloud CLI** ([install](https://cloud.google.com/sdk/docs/install))

### Authentication

Authenticate with GCP:

```bash
gcloud auth application-default login
```

This creates credentials at `~/.config/gcloud/application_default_credentials.json` that Pulumi uses automatically.

### Required Permissions

The authenticated account needs:
- `roles/container.admin` (GKE cluster and node pool management)
- `roles/compute.viewer` (to lookup cluster details)
- `roles/iam.serviceAccountUser` (if using custom service accounts)

## Standalone Usage

### 1. Prepare Stack Input

Create a `stack-input.json` file with your node pool configuration:

```json
{
  "target": {
    "metadata": {
      "name": "prod-general",
      "org": "my-org",
      "env": "prod"
    },
    "spec": {
      "cluster_project_id": {
        "value": "my-gcp-project"
      },
      "cluster_name": {
        "value": "prod-cluster"
      },
      "cluster_location": {
        "value": "us-central1"
      },
      "machine_type": "n2-standard-4",
      "disk_size_gb": 100,
      "disk_type": "pd-ssd",
      "service_account": "gke-prod-sa@my-gcp-project.iam.gserviceaccount.com",
      "autoscaling": {
        "min_nodes": 3,
        "max_nodes": 10,
        "location_policy": "BALANCED"
      },
      "node_labels": {
        "workload-tier": "general",
        "environment": "production"
      },
      "management": {
        "disable_auto_upgrade": false,
        "disable_auto_repair": false
      }
    }
  }
}
```

### 2. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init prod-general-pool
```

### 3. Configure Stack

Set the GCP project and stack input:

```bash
pulumi config set gcp:project my-gcp-project
pulumi config set --path stack_input --plaintext "$(cat stack-input.json)"
```

### 4. Deploy

Preview changes:

```bash
pulumi preview
```

Deploy the node pool:

```bash
pulumi up
```

Pulumi shows a detailed diff and prompts for confirmation before applying changes.

### 5. View Outputs

```bash
pulumi stack output
```

Outputs include:
- `node_pool_name`: Name of the created node pool
- `instance_group_urls`: URLs of the managed instance groups
- `current_node_count`: Current number of nodes
- `min_nodes`: Effective minimum size
- `max_nodes`: Effective maximum size

### 6. Update Configuration

Modify `stack-input.json` and re-run:

```bash
pulumi config set --path stack_input --plaintext "$(cat stack-input.json)"
pulumi up
```

Pulumi computes the diff and applies only necessary changes.

### 7. Destroy

Remove the node pool:

```bash
pulumi destroy
```

**Warning:** This terminates all nodes and evicts all pods in the pool.

## Environment Variables

### Required

- **GCP Project**: Set via `pulumi config set gcp:project <project-id>` or `GOOGLE_PROJECT` environment variable

### Optional

- **GCP Region**: Default determined by cluster location (automatically looked up)
- **GOOGLE_APPLICATION_CREDENTIALS**: Path to service account JSON (alternative to `gcloud auth`)
- **PULUMI_BACKEND_URL**: Custom state backend (default: Pulumi Cloud)

Example using service account:

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
export GOOGLE_PROJECT=my-gcp-project
pulumi up
```

## Development and Debugging

### Interactive Debugging

Use the provided `debug.sh` script for interactive development:

```bash
./debug.sh
```

This script:
1. Reads `hack/manifest.yaml` (test manifest)
2. Converts YAML to JSON stack input
3. Runs Pulumi with the test configuration
4. Displays detailed output for debugging

### Build and Test

```bash
# Build the Pulumi program
go build -o pulumi-main

# Run tests
go test ./module/...

# Format code
go fmt ./...

# Lint code
golangci-lint run
```

### Makefile Targets

```bash
# Build
make build

# Run with debug output
make debug

# Clean build artifacts
make clean
```

## Troubleshooting

### "Cluster not found" Error

**Cause:** The parent GKE cluster doesn't exist or the name/project is incorrect.

**Solution:**
1. Verify cluster exists: `gcloud container clusters list --project=my-gcp-project`
2. Check `cluster_name` and `cluster_project_id` in stack input
3. Ensure cluster is in the expected location (region or zone)

### "Insufficient Permissions" Error

**Cause:** Authenticated account lacks required IAM permissions.

**Solution:**
1. Grant `roles/container.admin`: 
   ```bash
   gcloud projects add-iam-policy-binding my-gcp-project \
     --member="user:your-email@example.com" \
     --role="roles/container.admin"
   ```
2. Wait a few minutes for IAM propagation

### "Quota Exceeded" Error

**Cause:** GCP project quota insufficient for requested resources (CPUs, IPs, etc.).

**Solution:**
1. Check quotas: `gcloud compute project-info describe --project=my-gcp-project`
2. Request quota increase in GCP Console: IAM & Admin → Quotas
3. Use smaller `max_nodes` or less resource-intensive `machine_type`

### Node Pool Creation Hangs

**Cause:** GKE is waiting for capacity, especially for Spot VMs or GPUs.

**Solution:**
- For Spot VMs: Set `location_policy: ANY` to search all zones
- For GPUs: Try alternative zones or machine types
- Check GCP status: https://status.cloud.google.com/

### "Node Configuration Change Requires Recreation"

**Expected behavior:** Changes to `machine_type`, `disk_type`, `disk_size_gb`, or `image_type` require node recreation. GKE cordons, drains, and replaces nodes automatically.

**To minimize disruption:**
- Ensure `PodDisruptionBudgets` are configured for critical workloads
- Use `max_surge` and `max_unavailable` settings (configured in module)
- Perform updates during maintenance windows

## Integration with Planton CLI

This module is typically invoked via the Planton CLI, not standalone. The CLI:
- Manages stack input from YAML manifests
- Handles state storage and locking
- Provides unified UX across all deployment components

For most users, **use the CLI** instead of running Pulumi directly:

```bash
planton apply -f node-pool.yaml
```

## Architecture

See **[overview.md](./overview.md)** for:
- Module architecture and design patterns
- Resource relationships and dependencies
- Local value transformations
- Output mappings

## Related Documentation

- **[Component README](../../README.md)**: User-facing overview
- **[Examples](../../examples.md)**: Working YAML manifests
- **[Research Document](../../docs/README.md)**: Deep dive into GKE node pools
- **[Terraform Module](../tf/README.md)**: Terraform/OpenTofu alternative

## Support

For issues or questions:
- **Planton Repository**: https://github.com/plantonhq/planton
- **Pulumi Documentation**: https://www.pulumi.com/docs/
- **GKE Node Pools Documentation**: https://cloud.google.com/kubernetes-engine/docs/concepts/node-pools

