# Basic Private Airflow Environment

A minimal MWAA environment with private webserver access, suitable for development
and small-scale DAG workloads within a VPC.

## When to Use

- Development and testing Airflow pipelines
- Small teams getting started with managed Airflow on AWS
- Environments where the Airflow UI should only be accessible within the VPC

## Configuration Highlights

- **Environment class**: `mw1.small` (1 vCPU, 2 GB per component)
- **Workers**: Auto-scaling between 1 and 5 Celery workers
- **Schedulers**: 2 (default for redundancy)
- **Access**: `PRIVATE_ONLY` — Airflow UI accessible only via VPC endpoint
- **Networking**: 2 private subnets across different AZs, 1 pre-existing security group
- **S3 source**: DAGs in `dags/` folder of the configured S3 bucket
- **Encryption**: AWS-managed default key (`aws/airflow`)

## Cost Estimate

The cost drivers are the mw1.small base environment (billed hourly around the clock, the dominant line) plus worker scaling — each additional Celery worker adds its own hourly charge while running. The verified figure for this preset lives in the kind's generated estimate at `catalog/_pricing/estimates/awsmwaaenvironment.yaml` — computed from the pinned price book, never hand-typed here.

## Customization

- Upgrade `environmentClass` to `mw1.medium` for heavier DAG workloads
- Add `loggingConfiguration` to enable CloudWatch Logs for debugging
- Add `kmsKeyArn` for customer-managed encryption
- Add `pluginsS3Path` and `requirementsS3Path` for custom operators and packages
