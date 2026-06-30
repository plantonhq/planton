---
title: "Validation"
description: "How Planton validates manifests at three layers -- protobuf schema rules, CLI-side validation, and cloud provider APIs -- catching configuration errors before they reach production"
icon: "security"
order: 35
---

# Validation

Infrastructure misconfigurations that reach cloud provider APIs are expensive -- in time, in partial deployments that need cleanup, and sometimes in real cost. Planton validates your manifests at multiple layers before any cloud API call is made, catching the vast majority of errors locally in milliseconds.

## Three Layers of Validation

### Layer 1: Schema-Level Validation (Protobuf + buf-validate)

Every deployment component's API is defined in Protocol Buffers with validation rules embedded directly in the schema. These rules are defined using `buf.validate` annotations and are enforced at the protobuf level, before any application logic runs.

**Constant enforcement** on `apiVersion` and `kind`:

```protobuf
string api_version = 1 [(buf.validate.field).string.const = 'kubernetes.planton.dev/v1'];
string kind = 2 [(buf.validate.field).string.const = 'KubernetesPostgres'];
```

If your manifest has `apiVersion: kubernetes.planton.dev/v2` or `kind: PostgresKubernetes`, validation fails immediately with a clear error. These are not runtime checks -- they are schema constraints.

**Required fields:**

```protobuf
dev.planton.shared.CloudResourceMetadata metadata = 3 [(buf.validate.field).required = true];
KubernetesPostgresSpec spec = 4 [(buf.validate.field).required = true];
```

A manifest without `metadata` or `spec` fails validation before the CLI even considers running an IaC module.

**String patterns and numeric ranges:**

```protobuf
// Disk size must match Kubernetes resource quantity format
string disk_size = 3 [(buf.validate.field).cel = {
    id: "spec.container.disk_size.required"
    message: "Disk size value is invalid"
    expression: "this.matches('^\\\\d+(\\\\.\\\\d+)?\\\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$') && size(this) > 0"
}];

// Region must not be empty
string aws_region = 1 [(buf.validate.field).string.min_len = 1];

// Enum values must be defined (rejects unknown values)
EncryptionType encryption_type = 4 [(buf.validate.field).enum.defined_only = true];
```

**Cross-field validation using CEL expressions:**

Some validation rules depend on multiple fields. These are expressed as Common Expression Language (CEL) rules at the message level:

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

This rule enforces that if `ingress.enabled` is `true`, then `ingress.hostname` must be a non-empty string. The error message tells you exactly what is wrong and where.

**Nested message validation:**

Validation rules propagate through nested messages. If `AwsS3BucketSpec` contains a `ReplicationConfiguration` with a required `Destination`, and the destination has a required `bucket_arn` field -- all of these validations fire when you validate the top-level manifest.

### Layer 2: CLI Validation

The CLI validation layer loads your YAML manifest, deserializes it into the protobuf message type, and runs the `protovalidate` library against it. This is what happens when you run:

```bash
planton validate -f my-resource.yaml
```

The internal flow:

1. **Load manifest** -- parse the YAML file (or clipboard content, or Kustomize output)
2. **Extract spec** -- identify the resource type from `apiVersion` and `kind`, deserialize into the correct protobuf message
3. **Run protovalidate** -- apply all schema-level validation rules (constants, required fields, patterns, CEL expressions)
4. **Format errors** -- present validation failures with clear field paths and error messages

Validation also runs automatically before any deployment command. When you execute `planton pulumi up -f my-resource.yaml`, the manifest is validated before the IaC module is ever invoked. If validation fails, the command exits with an error -- no cloud resources are touched.

### Layer 3: Cloud Provider Validation

The final validation layer is the cloud provider itself. Even after your manifest passes schema validation and CLI validation, the cloud provider's API performs its own checks during deployment:

- AWS checks that subnet IDs exist and belong to the correct VPC
- GCP checks that the project ID is valid and you have the right permissions
- Kubernetes checks that the target namespace exists (or is being created)

Layers 1 and 2 catch structural and type errors. Layer 3 catches environmental errors -- resources that do not exist, permissions that are missing, quotas that are exceeded. Together, the three layers provide comprehensive validation from syntax to infrastructure reality.

## The Foreign Key System

Some spec fields need to reference other resources. For example, the `namespace` field in `KubernetesPostgresSpec` can either be a literal string value or a reference to a `KubernetesNamespace` resource.

Planton handles this through the `StringValueOrRef` type:

```protobuf
message StringValueOrRef {
    oneof literal_or_ref {
        string value = 1;
        ValueFromRef value_from = 2;
    }
}

message ValueFromRef {
    CloudResourceKind kind = 1;
    string env = 2;
    string name = 3 [(buf.validate.field).required = true];
    string field_path = 4;
}
```

In your manifest, a literal value looks like this:

```yaml
spec:
  namespace:
    value: my-namespace
```

A reference to another resource looks like this:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: shared-namespace
      fieldPath: spec.name
```

The field-level annotations on the protobuf definition specify the default kind and field path, so the CLI knows which resource type a reference should point to:

```protobuf
StringValueOrRef namespace = 2 [
    (buf.validate.field).required = true,
    (default_kind) = KubernetesNamespace,
    (default_kind_field_path) = "spec.name"
];
```

This system enables cross-resource references while maintaining type safety and validation at the schema level.

## Validation in Practice

Running validation explicitly:

```bash
# Validate from a file
planton validate -f postgres.yaml

# Validate from clipboard
planton validate --clipboard

# Validate from Kustomize output
planton validate --kustomize-dir ./k8s --overlay production
```

Validation runs automatically before deployment:

```bash
# Validation happens before the IaC module runs
planton pulumi up -f postgres.yaml --stack my-org/my-project/prod
```

When validation fails, the CLI outputs a formatted error with the field path and the rule that was violated, so you know exactly what to fix in your manifest.

## Why This Matters

Without layered validation, a misconfigured manifest would travel all the way to the cloud provider API before failing -- potentially creating partial resources, incurring costs, or failing midway through a multi-resource deployment. With Planton's validation:

- **Wrong `apiVersion` or `kind`?** Caught at layer 1, instantly.
- **Missing required field?** Caught at layer 1, before any module runs.
- **Invalid field value?** Caught at layer 1, with the exact field path and validation message.
- **Cross-field constraint violated?** Caught at layer 1, with a human-readable CEL error message.
- **YAML syntax error or wrong resource type?** Caught at layer 2, during manifest loading.
- **Wrong subnet ID or missing permissions?** Caught at layer 3, by the cloud provider.

The first two layers eliminate an entire class of deployment failures before a single network request is made.

## What's Next

- **[Deployment Components](deployment-components)** -- How validation rules are defined in each component's protobuf
- **[Manifests](manifests)** -- The manifest structure that validation operates on
- **[Dual IaC Engines](dual-iac-engines)** -- What happens after validation passes
