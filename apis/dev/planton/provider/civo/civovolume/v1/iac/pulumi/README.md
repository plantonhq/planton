# Pulumi Module to Deploy CivoVolume

## CLI usage (Planton pulumi)

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

## Debugging

This module includes a `debug.sh` helper. To enable debugging, edit `Pulumi.yaml` and uncomment the `runtime.options.binary` line so Pulumi runs the program via the script:

```yaml
name: civo-volume-pulumi-project
runtime:
  name: go
#  options:
#    binary: ./debug.sh
```

Then make the script executable and run your command (e.g., `preview` or `update`). See project documentation for full debugging instructions.

```bash
chmod +x debug.sh
planton pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

# Civo Volume Pulumi Module

## Introduction

The Civo Volume Pulumi Module provides a standardized way to define and deploy Civo block storage volumes using a Kubernetes-like API resource model. By leveraging our unified APIs, developers can specify their volume configurations in simple YAML files, which the module then uses to create and manage Civo Volume resources through Pulumi. This approach abstracts the complexity of Civo API interactions and streamlines the deployment process, enabling consistent infrastructure management.

## Key Features

- **Kubernetes-Like API Resource Model**: Utilizes a familiar structure with `apiVersion`, `kind`, `metadata`, `spec`, and `status`, making it intuitive for developers accustomed to Kubernetes to define Civo Volume resources.

- **Unified API Structure**: Ensures consistency across different resources and cloud providers by adhering to a standardized API resource model.

- **Pulumi Integration**: Employs Pulumi for infrastructure provisioning, enabling the use of real programming languages (Go) and providing robust state management and automation capabilities.

- **Customizable Volume Configuration**: Supports detailed specification of volume attributes including:
  - Volume name (with DNS-compatible validation)
  - Region selection (LON1, NYC1, FRA1, etc.)
  - Size configuration (1-16,000 GiB)
  - Filesystem type preference (informational)
  - Snapshot restoration (CivoStack only)
  - Organizational tags

- **Credential Management**: Securely handles Civo credentials via the `civoProviderConfig` field, ensuring authenticated and authorized resource deployments without exposing sensitive information.

- **Status Reporting**: Captures and stores outputs such as volume ID and attachment information in `status.outputs`. This facilitates easy reference and integration with other systems or automation tools.

- **Provider Limitations Handling**: The module gracefully handles Civo provider limitations by logging informational messages when features like filesystem formatting, snapshots, or tags are requested but not supported by the underlying provider.

## Architecture

The module operates by accepting a CivoVolume API resource definition as input. It interprets the resource definition and uses Pulumi to interact with Civo, creating the specified block storage volume. The main components involved are:

- **API Resource Definition**: A YAML file that includes all necessary information to define a volume, following the standard API structure. Developers specify the volume's desired state in this file, including name, size, region, and optional parameters.

- **Pulumi Module**: Written in Go, the module reads the API resource and uses Pulumi's Civo SDK to provision volume resources based on the provided specifications. It abstracts the complexity of resource creation, update, and deletion.

- **Civo Provider Initialization**: The module initializes the Civo provider within Pulumi using the credentials specified by `civoProviderConfig`. This ensures that all Civo resource operations are authenticated and authorized.

- **Resource Creation**: Provisions the Civo Volume as defined in the `spec`, applying configurations such as name, region, and size. The module handles:
  - Volume creation with validated parameters
  - Informational logging for unsupported features (filesystem type, snapshots, tags)
  - Civo label application for resource tracking

- **Status Outputs**: Outputs from the Pulumi deployment, such as volume ID, are captured and stored in `status.outputs`. This information is crucial for volume attachment and management.

## Usage

Refer to the example section and the main README.md for usage instructions.

## Limitations

- **Filesystem Formatting**: The Civo provider doesn't expose filesystem formatting during volume creation. The `filesystemType` field in the spec is informational only. Users must format volumes manually after attachment using tools like `mkfs.ext4` or `mkfs.xfs`.

- **Snapshot Support**: Snapshot functionality is not available on public Civo cloud. The `snapshotId` field is reserved for CivoStack (private cloud) deployments. On public Civo, specifying a snapshot ID will log a warning and create an empty volume.

- **Tags**: The Civo Volume provider doesn't currently support tags. Tags specified in the spec are recorded in Planton metadata but not applied to the Civo resource. Use Civo labels (automatically applied) for resource organization.

- **Single-Attach Only**: Civo Volumes can only be attached to one instance at a time. For shared storage scenarios, consider using NFS or Civo Object Storage.

- **Region-Scoped**: Volumes cannot be moved between regions. They must be in the same region as any instance they attach to.

- **Resize Constraints**: Volumes can only be expanded, never shrunk. Resizing requires the volume to be detached (offline resize).

## Best Practices

1. **Start Small**: Begin with the minimum size needed (you can expand later, but never shrink).

2. **Application-Level Backups**: Since snapshots aren't available on public Civo, implement backups using application tools (`pg_dump`, `mysqldump`) or filesystem tools (`rsync`, `tar`) to export data to object storage.

3. **Automation**: Use cloud-init or configuration management tools to automate volume formatting and mounting after attachment.

4. **Tagging Strategy**: While tags aren't applied to Civo resources, use them in the spec for Planton metadata tracking, cost allocation, and organizational purposes.

5. **Monitoring**: Track volume usage and set alerts for capacity planning. Civo volumes have a fixed size until manually resized.

## Contributing

We welcome contributions to enhance the functionality of this module. Please submit pull requests or open issues to help improve the module and its documentation.

