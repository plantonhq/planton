---
title: "Deployment Components"
description: "The atomic unit of OpenMCF: a self-contained package combining API definition, IaC implementation, and documentation for deploying a specific cloud resource"
icon: "package"
order: 20
---

# Deployment Components

A deployment component is the atomic unit of OpenMCF. It is a self-contained package that combines everything needed to deploy and manage a specific type of cloud resource: a Protocol Buffer API definition, dual IaC module implementations (Pulumi and OpenTofu/Terraform), and auto-generated documentation.

OpenMCF ships with 362 deployment components spanning 17 cloud providers. Each component follows the same structural contract, which means once you understand how one component works, you understand them all.

## What a Component Contains

Every deployment component lives at a predictable path in the repository:

```text
apis/org/openmcf/provider/{provider}/{component}/v1/
```

Inside that directory, every component contains the same set of files:

```text
apis/org/openmcf/provider/kubernetes/kubernetespostgres/v1/
|-- api.proto              # Resource envelope: apiVersion, kind, metadata, spec, status
|-- spec.proto             # Configuration fields with types and validation rules
|-- stack_input.proto      # IaC input contract: the resource + provider credentials
|-- stack_outputs.proto    # IaC output contract: what the deployment produces
|-- iac/
|   |-- pulumi/            # Pulumi module (Go)
|   |   |-- main.go
|   |   |-- module/        # Resource implementation
|   |   \-- Pulumi.yaml
|   \-- tf/                # OpenTofu/Terraform module (HCL)
|       |-- main.tf
|       |-- variables.tf
|       |-- provider.tf
|       \-- outputs.tf
\-- docs/
    \-- README.md          # Auto-generated documentation
```

This structure is not a convention. It is a contract. Every one of the 362 components follows it exactly.

## The Four-File Protocol Buffer Contract

The four `.proto` files define the complete API surface of a deployment component. Together, they specify what the resource looks like, what configuration it accepts, what the IaC modules receive as input, and what they produce as output.

### api.proto -- The Resource Envelope

The `api.proto` file defines the top-level resource message. It follows the Kubernetes Resource Model (KRM) pattern: `apiVersion`, `kind`, `metadata`, `spec`, and `status`.

Here is the `api.proto` for `KubernetesPostgres`:

```protobuf
message KubernetesPostgres {
  string api_version = 1 [(buf.validate.field).string.const = 'kubernetes.openmcf.org/v1'];
  string kind = 2 [(buf.validate.field).string.const = 'KubernetesPostgres'];
  org.openmcf.shared.CloudResourceMetadata metadata = 3 [(buf.validate.field).required = true];
  KubernetesPostgresSpec spec = 4 [(buf.validate.field).required = true];
  KubernetesPostgresStatus status = 5;
}
```

Two things to notice. First, `api_version` and `kind` are enforced as constants using `buf.validate` -- this means a manifest with a wrong `apiVersion` or `kind` value will fail validation before any cloud API is ever called. Second, `metadata` and `spec` are required, while `status` is optional and populated by the system after deployment.

Every component's `api.proto` follows this exact pattern. The only things that change are the provider group in `api_version` (e.g., `aws.openmcf.org/v1`, `gcp.openmcf.org/v1`), the `kind` value, and the spec/status message types.

### spec.proto -- The Configuration Surface

The `spec.proto` file defines every configurable field for the resource. This is where the real depth of a component lives -- field types, nested messages, enums, default values, and validation rules.

For `KubernetesPostgres`, the spec includes:

```protobuf
message KubernetesPostgresSpec {
  KubernetesClusterSelector target_cluster = 1;
  StringValueOrRef namespace = 2 [(buf.validate.field).required = true];
  bool create_namespace = 3;
  KubernetesPostgresContainer container = 4;
  KubernetesPostgresIngress ingress = 5;
  KubernetesPostgresBackupConfig backup_config = 6;
  repeated KubernetesPostgresDatabase databases = 7;
  repeated KubernetesPostgresUser users = 8;
}
```

The `container` field has its own message type with replicas, CPU/memory resources, and disk size. The `ingress` field controls external access with a hostname. Validation rules are embedded directly -- for example, the ingress message enforces that a hostname is required when ingress is enabled:

```protobuf
message KubernetesPostgresIngress {
  bool enabled = 1;
  string hostname = 2;

  option (buf.validate.message).cel = {
    id: "spec.ingress.hostname.required"
    expression: "!this.enabled || size(this.hostname) > 0"
    message: "hostname is required when ingress is enabled"
  };
}
```

Compare this with the `AwsS3Bucket` spec, which exposes entirely different fields appropriate to its platform -- encryption type, storage class, lifecycle rules, replication, CORS, and logging:

```protobuf
message AwsS3BucketSpec {
  string aws_region = 1 [(buf.validate.field).string.min_len = 1];
  bool is_public = 2;
  bool versioning_enabled = 3;
  EncryptionType encryption_type = 4;
  string kms_key_id = 5;
  map<string, string> tags = 6;
  repeated LifecycleRule lifecycle_rules = 7;
  ReplicationConfiguration replication = 8;
  LoggingConfiguration logging = 9;
  CorsConfiguration cors = 10;
  bool force_destroy = 11;
}
```

This is the provider-specific design philosophy in action. `KubernetesPostgres` and `AwsS3Bucket` share the same structural envelope (apiVersion, kind, metadata, spec, status), but their specs expose the full, native capability of their respective platforms. OpenMCF does not abstract these differences away.

### stack_input.proto -- The IaC Input Contract

The `stack_input.proto` file defines what the IaC modules receive when they run. It always contains two fields: the `target` resource (the full manifest including metadata and spec) and a `provider_config` (credentials and connection details for the cloud provider).

```protobuf
message KubernetesPostgresStackInput {
  KubernetesPostgres target = 1;
  org.openmcf.provider.kubernetes.KubernetesProviderConfig provider_config = 2;
}
```

For an AWS component, the provider config is different:

```protobuf
message AwsS3BucketStackInput {
  AwsS3Bucket target = 1;
  org.openmcf.provider.aws.AwsProviderConfig provider_config = 2;
}
```

The stack input is the bridge between the manifest you write and the IaC module that provisions the resource. The CLI loads your manifest, constructs the stack input, and passes it to the IaC engine.

### stack_outputs.proto -- The IaC Output Contract

The `stack_outputs.proto` file defines what the IaC module produces after deployment. These are the values you need to connect to or reference the deployed resource.

For `KubernetesPostgres`, the outputs include connection details:

```protobuf
message KubernetesPostgresStackOutputs {
  string namespace = 1;
  string service = 2;
  string port_forward_command = 3;
  string kube_endpoint = 4;
  string external_hostname = 5;
  KubernetesSecretKey username_secret = 8;
  KubernetesSecretKey password_secret = 9;
}
```

Outputs are provider-specific. An S3 bucket produces a bucket ARN and endpoint URL. A GCP Cloud SQL instance produces a connection name and IP address. Each component defines exactly what its deployment produces.

## Dual IaC Implementation

Every deployment component ships with two IaC module implementations that achieve the same result using different engines.

### Pulumi Module (Go)

The Pulumi module is a Go program that uses the Pulumi SDK to provision resources. The entry point loads the stack input and delegates to a `Resources` function:

```go
func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &kubernetespostgresv1.KubernetesPostgresStackInput{}

        if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
            return errors.Wrap(err, "failed to load stack-input")
        }

        return module.Resources(ctx, stackInput)
    })
}
```

The `module/` package contains the actual resource creation logic, split across focused files: `namespace.go`, `main.go` (the PostgreSQL operator resources), `outputs.go`, and so on.

### OpenTofu/Terraform Module (HCL)

The Terraform module uses standard HCL configuration files. The `variables.tf` file mirrors the protobuf spec structure:

```hcl
variable "metadata" {
  type = object({
    name = string
    org  = optional(string)
    env  = optional(string)
  })
}

variable "spec" {
  type = object({
    namespace        = object({ value = string })
    create_namespace = optional(bool, false)
    container = optional(object({
      replicas  = optional(number, 1)
      resources = optional(object({ ... }))
      disk_size = optional(string, "1Gi")
    }))
  })
}
```

Both implementations receive the same input (the manifest's metadata and spec) and produce the same outputs. The choice between Pulumi and OpenTofu/Terraform is yours -- see [Dual IaC Engines](dual-iac-engines) for guidance on when to use each.

## Provider-Specific by Design

OpenMCF deliberately does not create abstraction layers across cloud providers. There is no `GenericDatabase` component that magically works on AWS, GCP, and Kubernetes. Instead, there are specific components for each platform:

| Need | AWS | GCP | Kubernetes |
|------|-----|-----|------------|
| PostgreSQL | `AwsRdsInstance` | `GcpCloudSql` | `KubernetesPostgres` |
| Object Storage | `AwsS3Bucket` | `GcpGcsBucket` | -- |
| DNS Zone | `AwsRoute53Zone` | `GcpDnsZone` | -- |
| Kubernetes Cluster | `AwsEksCluster` | `GcpGkeCluster` | -- |

This is intentional. Each cloud provider has different capabilities, pricing models, operational characteristics, and configuration options. An S3 bucket supports lifecycle rules, versioning policies, and cross-region replication. A GCS bucket has different storage classes and different access control models. Abstracting these into a common interface would either lose capabilities or create a leaky abstraction.

What OpenMCF provides instead is consistency at the structural level:

- Every component uses the same manifest format (KRM)
- Every component is validated using the same protobuf validation framework
- Every component is deployed using the same CLI commands
- Every component follows the same four-file contract

The workflow is identical. The configuration is provider-specific.

## A Real Example

Here is a complete manifest for deploying PostgreSQL on Kubernetes:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: session-store
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: production.KubernetesPostgres.session-store
spec:
  namespace:
    value: session-store
  container:
    replicas: 1
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 2000m
        memory: 2Gi
    diskSize: 1Gi
  ingress:
    enabled: false
```

Every field in this manifest traces directly to a protobuf definition. The `apiVersion` matches the const in `api.proto`. The `kind` matches the const in `api.proto`. The `metadata` fields match `CloudResourceMetadata`. The `spec` fields match `KubernetesPostgresSpec`. The labels configure the Pulumi state backend. Nothing is invented. Nothing is ambiguous.

## What's Next

- **[Manifests](manifests)** -- Deep dive into the KRM manifest structure, metadata fields, and manifest sources
- **[Cloud Resource Kinds](cloud-resource-kinds)** -- The full taxonomy of 362 components across 17 providers
- **[Dual IaC Engines](dual-iac-engines)** -- How the Pulumi and OpenTofu/Terraform implementations work
- **[Component Catalog](/docs/catalog)** -- Browse the documentation for every deployment component
