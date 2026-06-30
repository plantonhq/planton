# KubernetesManifest Pulumi Module

## Key Features

- **Raw YAML Deployment**  
  Deploys any valid Kubernetes manifest YAML directly to a cluster. Supports single resources and multi-document YAML separated by `---`.

- **Smart Resource Ordering**  
  Uses Pulumi's `yaml/v2` module which provides intelligent CRD-aware resource ordering. CRDs are applied before Custom Resources that depend on them.

- **Namespace Management**  
  Optionally creates the target namespace with appropriate labels, or uses an existing namespace.

- **Provider Integration**  
  Seamlessly integrates with Planton's Kubernetes provider configuration for secure cluster access.

- **Output Exports**  
  Exports the namespace where resources were deployed for downstream automation.

## Usage

See [examples.md](../../examples.md) for usage details and step-by-step examples. In general:

1. Define a YAML resource describing your manifest using the **KubernetesManifest** API.
2. Run:
   ```bash
   planton pulumi up --stack-input <your-manifest-file.yaml>
   ```

## Getting Started

1. **Prepare Your Manifest YAML**  
   Write the Kubernetes resources you want to deploy. This can be a single resource or multiple resources separated by `---`.

2. **Create the KubernetesManifest Resource**  
   Wrap your manifest YAML in a KubernetesManifest API resource, specifying the namespace and whether to create it.

3. **Apply via CLI**  
   Execute `planton pulumi up --stack-input <manifest-spec.yaml>`. The Pulumi module applies your manifest to the cluster.

4. **Validate**  
   Check that resources were created using `kubectl get` commands or the Kubernetes dashboard.

## Module Structure

1. **Initialization**  
   Reads your `KubernetesManifestStackInput` (containing cluster credentials and resource definitions), sets up local variables and labels.

2. **Provider Setup**  
   Establishes a Pulumi Kubernetes Provider for your target cluster using the provided credentials.

3. **Namespace Management**  
   Creates the Kubernetes namespace if `create_namespace: true`, with appropriate Planton labels.

4. **Manifest Application**  
   Uses `yaml/v2.ConfigGroup` to apply the manifest YAML. This provides:
   - Automatic CRD ordering (CRDs before Custom Resources)
   - Proper await behavior for resource readiness
   - Multi-document YAML support

5. **Output Exports**  
   Publishes the namespace where resources were deployed.

## Benefits

- **Zero Transformation**  
  Your YAML is applied exactly as written—no interpretation or modification.

- **CRD Safety**  
  The yaml/v2 module ensures CRDs are fully registered before Custom Resources are created, avoiding timing issues.

- **Flexibility**  
  Deploy anything from a single ConfigMap to a complete application stack with dozens of resources.

- **Consistency**  
  Uses the same provider and namespace patterns as all other Planton Kubernetes components.

## Technical Details

### Why yaml/v2?

The module uses Pulumi's `yaml/v2.ConfigGroup` instead of the older `yaml.ConfigFile` for several reasons:

1. **CRD Ordering**: Automatically applies CRDs before Custom Resources
2. **Await Behavior**: Properly waits for resource reconciliation
3. **Error Handling**: Better error messages for invalid manifests
4. **Performance**: More efficient handling of large manifest files

### Resource Dependencies

When `create_namespace: true`, the namespace is created first and all manifest resources depend on it, ensuring proper ordering.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request to add features, fix bugs, or improve documentation.

## License

This project is licensed under the [MIT License](LICENSE).

