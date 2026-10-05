# Terraform Module to Deploy AwsEcsCluster

This module provisions an AWS ECS cluster and its capacity: the cluster
itself (Container Insights, ECS Exec auditing, Fargate storage
encryption, Service Connect defaults), one `aws_ecs_capacity_provider`
per folded EC2 entry (each wrapping a referenced auto-scaling group),
and a single `aws_ecs_cluster_capacity_providers` association that PUTs
the union of Fargate built-ins and EC2 provider names onto the cluster
together with the default strategy.

Generated `variables.tf` reflects the proto schema for `AwsEcsCluster`
(generator-owned; never hand-edit).

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via IaC input (CLI), not in the manifest `spec`.

For more examples, see `../../e2e/manifest.yaml` and [`e2e/manifest.yaml`](../../e2e/manifest.yaml).
