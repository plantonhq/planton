# AwsEcsCluster

An Amazon ECS cluster: the scheduling boundary that groups services and tasks, decides where their containers run (Fargate, EC2 capacity providers, ECS Managed Instances, or a blend), and carries cluster-wide posture -- Container Insights observability, ECS Exec auditing, Fargate storage encryption, and the Service Connect default namespace.

Capacity is a spectrum. A serverless cluster associates the AWS-managed `FARGATE` / `FARGATE_SPOT` built-ins and never thinks about instances. An EC2-backed cluster defines `ec2CapacityProviders` -- each wrapping a referenced `AwsAutoScalingGroup` whose fleet ECS scales up and down through managed scaling. A managed-instances cluster defines `managedInstancesCapacityProviders` -- ECS launches and retires the EC2 instances itself from attribute-based requirements, with no auto-scaling group, AMI, or user data to own. Services blend across all of it by provider name in their `capacityProviderStrategy`. The cluster itself is free; only the tasks and instances it schedules cost money.

The cluster name comes from `metadata.name` (create-time immutable in AWS). Everything it composes with attaches by reference: the auto-scaling groups that provide EC2 capacity (`AwsAutoScalingGroup`), the IAM identities managed instances need (`AwsIamRole`, `AwsIamInstanceProfile`), the network they launch into (`AwsSubnet`, `AwsSecurityGroup`), the KMS keys that encrypt exec sessions and Fargate storage (`AwsKmsKey`), and the Cloud Map namespace Service Connect uses.

## Spec highlights

- **Container Insights, three-valued**: `enabled` (task/service-level), `enhanced` (container-level with automatic dashboards -- the production posture), `disabled`; unset keeps the account default.
- **Folded EC2 capacity providers**: per-name entries wrapping referenced auto-scaling groups, each with ECS-managed scaling (target capacity, step bounds, warmup), managed draining, and managed termination protection. Each materializes as its own provider resource and auto-joins the cluster association -- no double-listing.
- **Folded Managed Instances capacity providers**: per-name entries where ECS owns the fleet -- an infrastructure role, an instance profile, the subnets/security groups to launch into, and attribute-based instance requirements (memory, vCPUs, CPU manufacturer, accelerators, price protection) instead of a named type. Purchase model per provider (`ON_DEMAND`, `SPOT`, or `RESERVED` against capacity reservations), idle scale-in tuning, and detailed monitoring. Creating a provider launches nothing; instances appear only when a service schedules onto it.
- **Default capacity provider strategy**: base/weight distribution over any associated provider (built-ins or folded names from either list), validated against the associated set before anything deploys; only one entry may carry a non-zero base (AWS's own rule).
- **ECS Exec auditing**: session logging to the task's own configuration (`DEFAULT`), to explicit CloudWatch/S3 destinations (`OVERRIDE`), or knowingly unaudited (`NONE`), with an optional KMS key encrypting session traffic.
- **Managed storage encryption**: customer-managed KMS keys for Fargate ephemeral task storage -- the compliance posture for regulated workloads.
- **Service Connect defaults**: the Cloud Map namespace ARN every service in the cluster inherits.

## Outputs

`cluster_name`, `cluster_arn` (the join key -- `AwsEcsService.cluster_arn` references it), `capacity_provider_names` (the full strategy vocabulary: built-ins plus every folded provider), `capacity_provider_arns` (the folded providers' identities).

## How it works

Both the Terraform/OpenTofu and Pulumi modules implement the same contract: one `aws_ecs_cluster` keyed by `metadata.name`, one `aws_ecs_capacity_provider` per folded entry -- EC2 and managed-instances alike, keyed by name so edits never disturb siblings (managed-instances providers additionally bind to the cluster at create, as the AWS API requires) -- and ONE `aws_ecs_cluster_capacity_providers` association that PUTs the union of built-ins and folded names together with the default strategy -- a single association because the API replaces the whole set on every write.

## References

- [Amazon ECS clusters](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/clusters.html)
- [Capacity providers](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/cluster-capacity-providers.html)
- [ECS Managed Instances](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/managed-instances.html)
- [Container Insights](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/ContainerInsights.html)
- [ECS Exec](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-exec.html)
- [Service Connect](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-connect.html)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
