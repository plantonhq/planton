---
title: "Preset: Production Encrypted Airflow with Full Logging"
description: "A production-grade MWAA environment with customer-managed KMS encryption, all five Airflow logging modules enabled, graceful worker replacement, and a defined maintenance window."
type: "preset"
rank: "02"
presetSlug: "02-production-encrypted-logging"
componentSlug: "mwaa-environment"
componentTitle: "MWAA Environment"
provider: "aws"
icon: "package"
order: 2
---

# Preset: Production Encrypted Airflow with Full Logging

A production-grade MWAA environment with customer-managed KMS encryption, all five
Airflow logging modules enabled, graceful worker replacement, and a defined
maintenance window.

## When to Use

- Production data pipeline orchestration
- Compliance environments requiring customer-managed encryption keys
- Teams needing comprehensive CloudWatch Logs for debugging and auditing
- Workloads that cannot tolerate task interruption during environment updates

## Configuration Highlights

- **Environment class**: `mw1.medium` (2 vCPU, 4 GB per component)
- **Workers**: Auto-scaling between 2 and 10 Celery workers
- **Webservers**: Auto-scaling between 2 and 4 for high-availability UI access
- **Schedulers**: 3 (increased from default 2 for faster DAG parsing throughput)
- **Access**: `PRIVATE_ONLY` — Airflow UI accessible only via VPC endpoint
- **Encryption**: Customer-managed KMS key for data at rest
- **Logging**: All 5 modules enabled at INFO level:
  - DAG processing, scheduler, task, webserver, worker
- **Maintenance**: Tuesday 03:30 UTC weekly maintenance window
- **Updates**: `GRACEFUL` worker replacement (waits for running tasks to finish)

## Infra Chart Composition

This preset uses `valueFrom` references to compose with:
- **AwsS3Bucket** (DAGs source bucket)
- **AwsIamRole** (MWAA execution role)
- **AwsVpc** (subnets and VPC ID)
- **AwsSecurityGroup** (managed security group source)
- **AwsKmsKey** (encryption at rest)

## Cost Estimate

Approximately **$0.98/hr** for mw1.medium (~$710/month) base environment cost plus
worker scaling (each additional mw1.medium worker adds ~$0.10/hr) and CloudWatch Logs
ingestion/storage costs.

## Customization

- Add `airflowConfigurationOptions` to tune Airflow settings (e.g., parallelism, concurrency)
- Add `pluginsS3Path` for custom Airflow operators and hooks
- Increase `maxWorkers` beyond 10 for burst-heavy DAG workloads
- Change `logLevel` to `DEBUG` on specific modules for troubleshooting
