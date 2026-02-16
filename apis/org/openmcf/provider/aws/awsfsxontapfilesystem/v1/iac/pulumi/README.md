# Pulumi Module to Deploy AwsFsxOntapFileSystem

This Pulumi module provisions an Amazon FSx for NetApp ONTAP file system using the OpenMCF CLI. Use the hack manifest at `../hack/manifest.yaml` as a quick starting point.

## CLI

```bash
# Preview
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh
openmcf pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Destroy
openmcf pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

## File Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint — loads stack input, delegates to module
├── Pulumi.yaml          # Pulumi project metadata
├── Makefile             # Build/tidy/lint helpers
├── debug.sh             # Debug script for stack output inspection
├── README.md            # This file
├── overview.md          # Architecture deep-dive
└── module/
    ├── main.go          # Orchestration: provider setup, resource creation, output export
    ├── locals.go        # Input transformation and AWS tag construction
    ├── file_system.go   # FSx ONTAP File System resource implementation
    └── outputs.go       # Output key constants
```

## Stack Outputs

| Output Key | Description |
|------------|-------------|
| `file_system_id` | The ID of the FSx ONTAP file system |
| `file_system_arn` | Full ARN for IAM policy construction |
| `dns_name` | DNS name for the file system |
| `management_dns_name` | DNS for ONTAP CLI (SSH) and REST API access |
| `management_ip_addresses` | Management endpoint IP addresses |
| `intercluster_dns_name` | DNS for SnapMirror replication |
| `intercluster_ip_addresses` | Intercluster endpoint IP addresses |
| `network_interface_ids` | Network interface IDs created for the file system |
| `vpc_id` | VPC ID in which the file system was created |
| `owner_id` | AWS account ID of the file system owner |

## Examples

See [../../examples.md](../../examples.md) for sample manifests tailored to FSx for ONTAP.
