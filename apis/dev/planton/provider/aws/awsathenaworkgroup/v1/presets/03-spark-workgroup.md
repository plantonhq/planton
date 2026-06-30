# Preset: Spark Workgroup

An Athena workgroup configured for Apache Spark workloads (PySpark notebooks
and Spark SQL). Requires an IAM execution role with appropriate permissions.

## What This Configures

- PySpark engine version 3 selected.
- IAM execution role for Spark job execution.
- SSE_S3 encryption for query/notebook results.
- Force destroy enabled (development convenience — named queries and prepared
  statements are cleaned up on destroy).

## When to Use

- Data science teams running PySpark notebooks on Athena.
- ETL workloads using Spark SQL on S3 data.
- ML feature engineering using PySpark DataFrames.

## Before Deploying

1. **Replace the execution role ARN** with an IAM role that has:
   - `s3:GetObject`, `s3:ListBucket` on your data buckets.
   - `s3:PutObject` on the results bucket.
   - `glue:GetDatabase`, `glue:GetTable`, `glue:GetPartitions` for catalog
     access.
   - `logs:CreateLogGroup`, `logs:CreateLogStream`, `logs:PutLogEvents` for
     CloudWatch logging.
2. **Create the S3 bucket** at `spark-results` for notebook output.
3. Consider setting `forceDestroy: false` in production to prevent accidental
   loss of named queries.
