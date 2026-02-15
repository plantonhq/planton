---
title: "Preset: On-Demand Minimal"
description: "The simplest possible Kinesis stream for development, prototyping, or variable-throughput workloads. AWS manages all capacity automatically."
type: "preset"
rank: "01"
presetSlug: "01-on-demand-minimal"
componentSlug: "kinesis-data-stream"
componentTitle: "Kinesis Data Stream"
provider: "aws"
icon: "package"
order: 1
---

# Preset: On-Demand Minimal

## Use Case

The simplest possible Kinesis stream for development, prototyping, or variable-throughput workloads. AWS manages all capacity automatically.

## What You Get

- **Capacity**: Auto-scaling (up to 200 MB/s write, 400 MB/s read)
- **Retention**: 24 hours (AWS default)
- **Encryption**: None
- **Monitoring**: Stream-level metrics only

## When to Use

- Development and testing environments
- New projects where throughput is unknown
- Bursty or unpredictable workloads
- Getting started with Kinesis

## Cost

Pay-per-use: ~$0.08/GB written + ~$0.04/GB read. No idle cost when no data is flowing.
