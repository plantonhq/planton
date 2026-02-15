---
title: "Fargate Cost-Optimized Cluster"
description: "This preset creates an ECS cluster with both Fargate and Fargate Spot capacity providers, using a weighted strategy that runs approximately 80% of scaled tasks on Spot for significant cost savings..."
type: "preset"
rank: "02"
presetSlug: "02-fargate-cost-optimized"
componentSlug: "ecs-cluster"
componentTitle: "ECS Cluster"
provider: "aws"
icon: "package"
order: 2
---

# Fargate Cost-Optimized Cluster

This preset creates an ECS cluster with both Fargate and Fargate Spot capacity providers, using a weighted strategy that runs approximately 80% of scaled tasks on Spot for significant cost savings (up to 70% cheaper than on-demand). One on-demand task is always guaranteed as a stability baseline.

## When to Use

- Production workloads that can tolerate occasional Spot interruptions (stateless web services, API servers, workers)
- Cost-sensitive environments where reducing compute spend is a priority
- Services with multiple replicas where losing one task to a Spot reclamation is acceptable

## Key Configuration Choices

- **Fargate + Fargate Spot** -- Enables mixed pricing with a guaranteed on-demand baseline
- **Base 1 on-demand** (`base: 1` on FARGATE) -- At least one task always runs on on-demand Fargate, ensuring minimum availability even if all Spot capacity is reclaimed
- **80/20 Spot weighting** (`weight: 4` for Spot vs `weight: 1` for on-demand) -- For every 5 scaled tasks beyond the base, 4 use Spot and 1 uses on-demand
- **Container Insights enabled** -- Full observability for monitoring task placement and Spot interruptions

## Placeholders to Replace

This preset has no placeholders. Deploy as-is and then create `AwsEcsService` resources targeting this cluster. Services inherit the cluster's default capacity provider strategy unless they override it.

## Related Presets

- **01-fargate-standard** -- Use instead when Spot interruptions are not acceptable (stateful services, critical single-replica workloads)
