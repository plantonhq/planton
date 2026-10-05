# Overview

The AwsEksFargateProfile API resource declares which Kubernetes pods of
an `AwsEksCluster` run on AWS Fargate: serverless, per-pod compute with
no EC2 nodes to size, patch, or scale. Pods matching the profile's
selectors are scheduled onto Fargate; everything else keeps running on
the cluster's node groups.

## Why We Created This API Resource

A cluster typically mixes compute styles -- steady workloads on node
groups, bursty or isolation-sensitive workloads on Fargate -- and each
profile is an independent object with its own lifecycle:

- **Attach everything by reference**: the cluster
  (`status.outputs.name`), the pod execution role (an `AwsIamRole`
  carrying its own policy), and the private subnets -- the architecture
  graph shows exactly which namespaces run serverless.
- **Selector-driven placement**: namespace (with `*`/`?` wildcards) plus
  optional label matching decides what runs on Fargate -- no
  scheduler configuration inside the cluster.
- **Honest immutability**: AWS makes the entire profile create-only;
  the spec says so on every field instead of letting a change surprise
  you with a replacement.

## Key Features

### Placement Control

- **Up to 5 selectors** (the AWS limit, enforced at validation), each
  matching a namespace and optionally requiring pod labels.
- **Wildcard namespaces** ("team-*") cover namespace-per-team patterns
  with one selector.

### Honest Composition

- **Pod execution role by reference**: the role that pulls images and
  writes logs is a referenced `AwsIamRole` trusting
  `eks-fargate-pods.amazonaws.com` with
  `AmazonEKSFargatePodExecutionRolePolicy` attached on the role itself.
- **Private subnets by reference**: AWS rejects subnets with an
  internet-gateway route; outbound traffic goes through a NAT gateway.

### Operational Truths, Documented

- The whole profile replaces on any change -- run overlapping profiles
  during migrations to keep pods scheduled.
- AWS serializes profile operations per cluster (one create/delete at a
  time); both engines simply wait.

## Benefits

- **Composability**: cluster, role, and subnets attach through
  `valueFrom` references.
- **Honest constraints**: the selector count and shape are CEL-enforced
  at validation time; immutability is documented where it applies.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `fargate_profile_arn`: the profile's ARN
- `fargate_profile_name`: the profile's name
- `status`: the profile's state after provisioning ("ACTIVE")

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
