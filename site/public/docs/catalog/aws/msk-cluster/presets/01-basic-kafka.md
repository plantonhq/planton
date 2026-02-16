---
title: "Preset: Basic Kafka Cluster"
description: "A minimal 3-broker MSK cluster suitable for development and testing workloads."
type: "preset"
rank: "01"
presetSlug: "01-basic-kafka"
componentSlug: "msk-cluster"
componentTitle: "MSK Cluster"
provider: "aws"
icon: "package"
order: 1
---

# Preset: Basic Kafka Cluster

A minimal 3-broker MSK cluster suitable for development and testing workloads.

## When to Use

- Development and testing environments
- Small-scale event streaming prototyping
- Teams getting started with managed Kafka on AWS

## Configuration Highlights

- **Instance type**: `kafka.t3.small` (burstable, cost-effective for low throughput)
- **Brokers**: 3 across 3 AZs (minimum for high availability)
- **Authentication**: SASL/IAM (recommended, no password management)
- **Encryption**: Defaults (TLS client-broker, in-cluster encryption enabled)
- **Storage**: AWS-managed defaults per instance type

## Cost Estimate

Approximately $0.12/hr for 3 x kafka.t3.small brokers (~$90/month) plus EBS storage costs.

## Customization

- Upgrade `instanceType` to `kafka.m5.large` for production workloads
- Add `serverProperties` to tune replication factor and ISR settings
- Add `logging` for CloudWatch or S3 broker log delivery
