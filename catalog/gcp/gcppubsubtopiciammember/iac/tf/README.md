# GcpPubSubTopicIamMember - Terraform Module

This Terraform module provisions a single additive topic-scoped IAM grant (`google_pubsub_topic_iam_member`). It is the Terraform-side implementation of the Planton `GcpPubSubTopicIamMember` resource kind and has feature parity with the Pulumi module.

## Overview

The module merges one (role, member) pair into the target topic's IAM policy without touching any other member's bindings; destroy subtracts only this exact pair. Every argument is immutable (ForceNew) — IAM grants have no update, so any change replaces the grant. There is no project input: the provider reads the project from the topic's full name.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcppubsubtopiciammember/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpPubSubTopicIamMember spec | — |

The `spec` object includes: `topic` (the topic's full name, `projects/<project>/topics/<topic>`), `role` (predefined or full custom role name), and `member` (IAM member format). There is no `condition`: Pub/Sub topics do not accept conditional role bindings.

## Outputs

| Name | Description |
|------|-------------|
| `topic` | The topic whose policy received the grant |
| `role` | The granted role |
| `member` | The granted member |
| `etag` | The topic IAM policy etag after the grant |
