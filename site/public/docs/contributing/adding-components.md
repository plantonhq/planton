---
title: "Adding Components"
description: "Step-by-step guide for creating a new deployment component in OpenMCF — from Protocol Buffer definitions to dual IaC modules"
icon: "integration"
order: 20
---

# Adding Deployment Components

This guide walks through creating a new deployment component for OpenMCF. A deployment component is a self-contained package that enables declarative deployment of a specific cloud resource — from an S3 bucket to a Kubernetes cluster to a Cloudflare Worker.

Every component follows the same structure: Protocol Buffer API definitions, dual IaC modules (Pulumi + Terraform), and documentation. This consistency across 360+ components and 17 providers is what makes OpenMCF predictable for users.

## Anatomy of a Component

A complete deployment component lives at `apis/org/openmcf/provider/<provider>/<component>/v1/` and contains:

```text
apis/org/openmcf/provider/aws/awss3bucket/v1/
|-- api.proto              # KRM resource model (apiVersion, kind, metadata, spec, status)
|-- spec.proto             # Configuration fields for the resource
|-- stack_input.proto      # IaC input contract (target resource + provider config)
|-- stack_outputs.proto    # IaC output contract (what the module exports)
|-- spec_test.go           # Validation tests for the spec
|-- README.md              # Component overview
|-- docs/
|   \-- README.md          # Research doc with design rationale
|-- iac/
|   |-- hack/
|   |   \-- manifest.yaml  # Development test manifest
|   |-- pulumi/
|   |   |-- main.go        # Pulumi entrypoint (loads stack input, calls module)
|   |   |-- Pulumi.yaml    # Pulumi project configuration
|   |   |-- Makefile        # Build targets for the Pulumi module
|   |   \-- module/
|   |       |-- main.go    # Resource creation and exports
|   |       |-- locals.go  # Local variables from stack input
|   |       \-- outputs.go # Output key constants
|   \-- tf/
|       |-- main.tf        # Terraform resource definitions
|       |-- variables.tf   # Input variables (metadata + spec)
|       |-- outputs.tf     # Output definitions
|       |-- provider.tf    # Provider and backend configuration
|       \-- locals.tf      # Derived local values
```

## Naming Conventions

| Element | Convention | Example |
|---------|-----------|---------|
| Folder name | `<provider><resource>` lowercase, no separators | `awss3bucket` |
| Kind name | `<Provider><Resource>` PascalCase | `AwsS3Bucket` |
| apiVersion | `<provider>.openmcf.org/v1` | `aws.openmcf.org/v1` |
| Proto package | `org.openmcf.provider.<provider>.<component>.v1` | `org.openmcf.provider.aws.awss3bucket.v1` |
| Pulumi project | `<component>-pulumi-project` | `awss3bucket-pulumi-project` |

## Step-by-Step Creation Workflow

### Phase 1: Define the API

The Protocol Buffer definitions are the foundation. Every other piece — IaC modules, CLI behavior, validation, SDKs — derives from these files.

#### 1.1 Create `spec.proto`

The spec defines the user-facing configuration fields. Design the spec around the 90/10 principle: cover the broad majority of the provider's real surface, benchmarked against the provider's own API as the floor for completeness, with sensible defaults so advanced fields stay out of the way until needed.

```protobuf
syntax = "proto3";

package org.openmcf.provider.aws.awss3bucket.v1;

import "buf/validate/validate.proto";

message AwsS3BucketSpec {
  // The AWS region where the S3 bucket will be created.
  string aws_region = 1 [(buf.validate.field).string.min_len = 1];

  // Whether the bucket should have public access.
  bool is_public = 2;

  // Enable versioning to protect against accidental deletions.
  bool versioning_enabled = 3;

  // Tags for resource governance and cost allocation.
  map<string, string> tags = 6;
}
```

Design principles for spec fields:

- **Deployment-agnostic**: Describe desired state, not implementation details
- **Use proto field names that match the cloud provider's terminology** where possible
- **Use enums for constrained choices** (encryption types, storage classes, SKU tiers)
- **Use `StringValueOrRef` for cross-resource references** (subnet IDs, VPC IDs, project IDs) — this enables the foreign key system
- **Mark fields with defaults as `optional`** and annotate with `(org.openmcf.shared.options.default)`

#### 1.2 Add Validation

Add `buf.validate` annotations to enforce constraints at the proto level:

```protobuf
string aws_region = 1 [(buf.validate.field).string.min_len = 1];

int32 allocated_storage_gb = 7 [(buf.validate.field).int32.gt = 0];

EncryptionType encryption_type = 4 [(buf.validate.field).enum.defined_only = true];
```

For cross-field validation, use CEL expressions:

```protobuf
option (buf.validate.message).cel = {
  id: "subnets_or_group"
  message: "Provide either subnet_ids (>=2) or db_subnet_group_name"
  expression: "(this.subnet_ids.size() >= 2) || has(this.db_subnet_group_name)"
};
```

#### 1.3 Create `stack_outputs.proto`

Define the outputs that the IaC module will export after deployment:

```protobuf
syntax = "proto3";

package org.openmcf.provider.aws.awss3bucket.v1;

message AwsS3BucketStackOutputs {
  string bucket_id = 1;
  string bucket_arn = 2;
  string region = 3;
}
```

Outputs should include identifiers and connection information that other resources or users might need.

#### 1.4 Create `api.proto`

The API proto wires the spec and outputs into the KRM resource model:

```protobuf
syntax = "proto3";

package org.openmcf.provider.aws.awss3bucket.v1;

import "buf/validate/validate.proto";
import "org/openmcf/provider/aws/awss3bucket/v1/spec.proto";
import "org/openmcf/provider/aws/awss3bucket/v1/stack_outputs.proto";
import "org/openmcf/shared/metadata.proto";

message AwsS3Bucket {
  // Fixed apiVersion for this resource type
  string api_version = 1 [(buf.validate.field).string.const = 'aws.openmcf.org/v1'];

  // Fixed kind name
  string kind = 2 [(buf.validate.field).string.const = 'AwsS3Bucket'];

  // KRM metadata (name, labels, annotations)
  org.openmcf.shared.CloudResourceMetadata metadata = 3
    [(buf.validate.field).required = true];

  // User-defined configuration
  AwsS3BucketSpec spec = 4 [(buf.validate.field).required = true];

  // Deployment status with stack outputs
  AwsS3BucketStatus status = 5;
}

message AwsS3BucketStatus {
  AwsS3BucketStackOutputs outputs = 1;
}
```

The `api_version` and `kind` fields use `string.const` validation to enforce exact values.

#### 1.5 Create `stack_input.proto`

The stack input combines the target resource with provider-specific configuration:

```protobuf
syntax = "proto3";

package org.openmcf.provider.aws.awss3bucket.v1;

import "org/openmcf/provider/aws/awss3bucket/v1/api.proto";
import "org/openmcf/provider/aws/provider.proto";

message AwsS3BucketStackInput {
  // The target resource to deploy
  AwsS3Bucket target = 1;
  // Provider credentials and configuration
  org.openmcf.provider.aws.AwsProviderConfig provider_config = 2;
}
```

#### 1.6 Write Validation Tests

Create `spec_test.go` to verify that validation rules work correctly:

```go
package awss3bucket_v1_test

import (
    "testing"
    pb "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awss3bucket/v1"
    "github.com/bufbuild/protovalidate-go"
)

func TestAwsS3BucketSpec_Validation(t *testing.T) {
    validator, _ := protovalidate.New()

    t.Run("valid minimal spec", func(t *testing.T) {
        spec := &pb.AwsS3BucketSpec{
            AwsRegion: "us-east-1",
        }
        err := validator.Validate(spec)
        if err != nil {
            t.Errorf("expected valid, got: %v", err)
        }
    })

    t.Run("empty region rejected", func(t *testing.T) {
        spec := &pb.AwsS3BucketSpec{
            AwsRegion: "",
        }
        err := validator.Validate(spec)
        if err == nil {
            t.Error("expected validation error for empty region")
        }
    })
}
```

### Phase 2: Register the Kind

Add the new kind to the cloud resource kind enum and regenerate stubs.

#### 2.1 Add Enum Entry

Add the new kind to `apis/org/openmcf/shared/cloudresourcekind/cloud_resource_kind.proto`:

```protobuf
// AWS enum range: 1000-1999
AWS_S3_BUCKET = 1001;
```

Each provider has a reserved enum range. Add the new kind within the correct provider's range.

#### 2.2 Regenerate

```bash
# Generate Go stubs from proto definitions
make protos

# Regenerate the cloud resource kind map
make generate-cloud-resource-kind-map
```

### Phase 3: Implement the Pulumi Module

The Pulumi module translates the protobuf spec into actual cloud resources using the Pulumi Go SDK.

#### 3.1 Entrypoint (`iac/pulumi/main.go`)

```go
package main

import (
    "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awss3bucket/v1/iac/pulumi/module"
    "github.com/plantonhq/openmcf/pkg/iac/pulumi/stackinput"
    awss3bucketv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awss3bucket/v1"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &awss3bucketv1.AwsS3BucketStackInput{}
        if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
            return err
        }
        return module.Resources(ctx, stackInput)
    })
}
```

#### 3.2 Module Implementation (`iac/pulumi/module/main.go`)

The module's `Resources` function creates the actual cloud resources:

```go
package module

import (
    awss3bucketv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awss3bucket/v1"
    "github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awss3bucketv1.AwsS3BucketStackInput) error {
    locals := initializeLocals(ctx, stackInput)

    // Create the S3 bucket
    bucket, err := s3.NewBucketV2(ctx, "bucket", &s3.BucketV2Args{
        // ... resource configuration from locals
    })
    if err != nil {
        return err
    }

    // Export outputs
    ctx.Export(OpBucketId, bucket.ID())
    ctx.Export(OpBucketArn, bucket.Arn)

    return nil
}
```

#### 3.3 Locals and Outputs

`module/locals.go` extracts values from the stack input into a `Locals` struct for clean access throughout the module.

`module/outputs.go` defines output key constants that match the `stack_outputs.proto` field names:

```go
package module

const (
    OpBucketId  = "bucket_id"
    OpBucketArn = "bucket_arn"
    OpRegion    = "region"
)
```

#### 3.4 Pulumi Project Config (`iac/pulumi/Pulumi.yaml`)

```yaml
name: awss3bucket-pulumi-project
runtime: go
```

### Phase 4: Implement the Terraform Module

The Terraform module provides the same deployment capability using HCL.

#### 4.1 Variables (`iac/tf/variables.tf`)

Map the proto spec fields to Terraform variable types:

```hcl
variable "metadata" {
  description = "Resource metadata"
  type = object({
    name = string
  })
}

variable "spec" {
  description = "AwsS3Bucket spec"
  type = object({
    aws_region         = string
    is_public          = optional(bool, false)
    versioning_enabled = optional(bool, false)
    tags               = optional(map(string), {})
  })
}
```

#### 4.2 Resources (`iac/tf/main.tf`)

Create the same cloud resources as the Pulumi module:

```hcl
resource "aws_s3_bucket" "this" {
  bucket        = local.bucket_name
  force_destroy = var.spec.force_destroy
  tags          = local.merged_tags
}

resource "aws_s3_bucket_versioning" "this" {
  bucket = aws_s3_bucket.this.id
  versioning_configuration {
    status = var.spec.versioning_enabled ? "Enabled" : "Suspended"
  }
}
```

#### 4.3 Outputs (`iac/tf/outputs.tf`)

Match the fields defined in `stack_outputs.proto`:

```hcl
output "bucket_id" {
  value = aws_s3_bucket.this.id
}

output "bucket_arn" {
  value = aws_s3_bucket.this.arn
}

output "region" {
  value = var.spec.aws_region
}
```

#### 4.4 Provider (`iac/tf/provider.tf`)

```hcl
terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
}

provider "aws" {
  region = var.spec.aws_region
}
```

### Phase 5: Write Documentation

#### 5.1 Component README

Create `README.md` at the component root with a concise overview of what the component deploys and its key configuration options.

#### 5.2 Research Doc

Create `docs/README.md` with deeper design rationale: why certain fields were chosen, what the 90/10 coverage trade-offs are, deployment best practices, and anti-patterns to avoid.

#### 5.3 Hack Manifest

Create `iac/hack/manifest.yaml` with a minimal working manifest for development testing:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: awss3bucket-demo
spec:
  awsRegion: us-east-1
  isPublic: false
```

### Phase 6: Build and Test

Run the full validation suite for the new component:

```bash
# Generate proto stubs (if not already done)
make protos

# Regenerate kind map
make generate-cloud-resource-kind-map

# Build validation — proto-generated Go code compiles
go build ./apis/org/openmcf/provider/aws/awss3bucket/v1/...

# Vet check
go vet ./apis/org/openmcf/provider/aws/awss3bucket/v1/iac/pulumi/...

# Run validation tests
go test -v ./apis/org/openmcf/provider/aws/awss3bucket/v1/...

# Validate Terraform module
cd apis/org/openmcf/provider/aws/awss3bucket/v1/iac/tf
terraform init && terraform validate
```

All checks must pass before submitting a pull request.

## Design Principles

### 90/10 Coverage

Cover the broad majority of the provider's real surface, benchmarked against the provider's own API as the floor -- reach the long tail an advanced user needs, with sensible defaults so the common path stays simple. Quality is the constant: every field is researched, validated, and exercised in both engines; genuinely beta or niche knobs are skipped with a recorded reason.

### Deployment-Agnostic Specs

Specs describe **what** the user wants, not **how** the IaC module implements it. A user says "I want versioning enabled" — they do not say "create an `aws_s3_bucket_versioning` resource with status Enabled." The spec is the interface; the IaC module is the implementation.

### Secure Defaults

Default to the secure option. Public access should be `false` by default. Encryption should be enabled by default. Deletion protection should require explicit opt-out. Users should have to make a conscious choice to reduce security, not to enable it.

### Dual IaC Parity

Both the Pulumi module and the Terraform module should produce the same cloud resources from the same spec. Users choose their provisioner based on preference or organizational requirements — not because one module supports features the other does not.

## What's Next

- **[Contributing Guide](/docs/contributing)** — Development environment setup, building, and testing
- **[Deployment Components](/docs/concepts/deployment-components)** — Conceptual overview of the component model
- **[Validation](/docs/concepts/validation)** — How the three-layer validation system works
- **[Cloud Resource Kinds](/docs/concepts/cloud-resource-kinds)** — Full taxonomy of component kinds and providers
