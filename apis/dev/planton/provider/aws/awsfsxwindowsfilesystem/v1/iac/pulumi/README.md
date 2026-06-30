# Pulumi Module to Deploy AwsFsxWindowsFileSystem

This Pulumi module provisions an Amazon FSx for Windows File Server file system using the Planton CLI. Use the hack manifest at `../hack/manifest.yaml` as a quick starting point.

## CLI

```bash
# Preview
planton pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
planton pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh
planton pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Destroy
planton pulumi destroy \
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
├── debug.sh             # Delve debugging wrapper
├── README.md            # This file
├── overview.md          # Architecture deep-dive
└── module/
    ├── main.go          # Orchestration: provider setup, resource creation, output export
    ├── locals.go        # Input transformation and AWS tag construction
    ├── file_system.go   # FSx Windows File System resource implementation
    └── outputs.go       # Output key constants
```

## Stack Outputs

| Output Key                        | Description                                         |
|-----------------------------------|-----------------------------------------------------|
| `file_system_id`                  | The ID of the FSx file system                       |
| `file_system_arn`                 | Full ARN for IAM policy construction                |
| `dns_name`                        | DNS name for mounting the file system                |
| `preferred_file_server_ip`        | IP address of the preferred file server              |
| `remote_administration_endpoint`  | Endpoint for remote PowerShell administration        |
| `network_interface_ids`           | Network interface IDs created for the file system    |
| `vpc_id`                          | VPC ID in which the file system was created          |
| `owner_id`                        | AWS account ID of the file system owner              |

## Examples

See `./examples.md` for sample manifests tailored to FSx for Windows File Server.

## Debugging

A helper script `debug.sh` is provided to launch the Pulumi program under Delve. To enable it, uncomment the binary option in `Pulumi.yaml`:

```yaml
runtime:
  name: go
  # options:
  #   binary: ./debug.sh
```

Then run the CLI commands above; the Pulumi engine will execute the local debug binary. For more details, see docs: docs/pages/docs/guide/debug-pulumi-modules.mdx
