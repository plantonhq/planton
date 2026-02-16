---
title: "Preset: Provisioned Encrypted"
description: "A provisioned stream with predictable capacity, KMS encryption using the Kinesis-owned key, and 48-hour retention for basic reprocessing."
type: "preset"
rank: "02"
presetSlug: "02-provisioned-encrypted"
componentSlug: "kinesis-data-stream"
componentTitle: "Kinesis Data Stream"
provider: "aws"
icon: "package"
order: 2
---

# Preset: Provisioned Encrypted

## Use Case

A provisioned stream with predictable capacity, KMS encryption using the Kinesis-owned key, and 48-hour retention for basic reprocessing.

## What You Get

- **Capacity**: 2 shards (2 MB/s write, 4 MB/s read)
- **Retention**: 48 hours
- **Encryption**: KMS (Kinesis-owned key — no additional KMS cost)
- **Monitoring**: Stream-level metrics only

## When to Use

- Staging environments with steady, predictable throughput
- Workloads processing sensitive data that requires encryption
- Streams where you want to control and predict costs
- When you know your throughput fits within 2 shards (2 MB/s write)

## Cost

~$21.60/month for 2 shards + extended retention cost for hours 25-48.
