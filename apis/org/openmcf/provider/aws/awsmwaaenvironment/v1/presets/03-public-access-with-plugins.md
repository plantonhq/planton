# Preset: Public Access with Plugins and Custom Packages

An MWAA environment with public webserver access, custom plugins, Python requirements,
a startup script, and aggressive worker auto-scaling. Demonstrates the full breadth of
MWAA's extensibility features.

## When to Use

- Teams needing internet-accessible Airflow UI (IAM-authenticated)
- Environments with custom Airflow operators, hooks, or sensors packaged in plugins.zip
- Pipelines requiring additional Python packages beyond the base Airflow installation
- High-burst DAG workloads needing up to 25 concurrent workers

## Configuration Highlights

- **Environment class**: `mw1.large` (4 vCPU, 8 GB per component)
- **Workers**: Auto-scaling between 1 and 25 Celery workers
- **Webservers**: Auto-scaling between 2 and 5
- **Schedulers**: 3 for fast DAG parsing across a large DAG repository
- **Access**: `PUBLIC_ONLY` — Airflow UI accessible over the internet (IAM login required)
- **Plugins**: `plugins/plugins.zip` containing custom operators, hooks, and sensors
- **Requirements**: `requirements/requirements.txt` for additional Python packages
- **Startup script**: `scripts/startup.sh` for OS-level setup at environment boot
- **Config overrides**:
  - `core.default_timezone` set to `America/New_York`
  - `webserver.dag_default_view` set to `grid`
- **Logging**: All 5 modules enabled (webserver at WARNING to reduce log noise)

## Cost Estimate

Approximately **$1.96/hr** for mw1.large (~$1,410/month) base environment cost plus
worker scaling (each additional mw1.large worker adds ~$0.19/hr). CloudWatch Logs costs
apply for the 5 enabled logging modules.

## Customization

- Add `pluginsS3ObjectVersion` and `requirementsS3ObjectVersion` to pin artifact versions
- Add `kmsKeyArn` for customer-managed encryption
- Switch `webserverAccessMode` to `PRIVATE_ONLY` if public access is no longer needed
- Adjust `airflowConfigurationOptions` for celery concurrency, parallelism, or other tuning
