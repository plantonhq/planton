# Production Elastic EFS with Access Points

Regional, encrypted, elastic throughput, lifecycle policies (AFTER_30_DAYS IA, AFTER_1_ACCESS primary), backup, 2 access points (app-data with uid/gid 1000, logs with uid/gid 1001).

## When to Use

- Production workloads with variable or unpredictable I/O patterns
- ECS tasks or Lambda functions needing per-application root directories and POSIX identity
- Workloads where cost optimization via lifecycle policies is important
- Multi-tenant or multi-app file sharing with least-privilege access

## What It Configures

- **Elastic throughput** — Throughput scales up/down with workload; recommended for spiky or unpredictable access
- **Lifecycle policies** — Files move to IA after 30 days, to Archive after 90 days; accessed files return to Standard (AFTER_1_ACCESS)
- **Backup enabled** — Daily backups via AWS Backup
- **Access points** — `app-data` (uid/gid 1000) and `logs` (uid/gid 1001) with dedicated root directories; EFS creates paths with correct ownership if they don't exist

## What to Customize

- Replace placeholders: `<subnet-id-az-a>`, `<subnet-id-az-b>`, `<security-group-id>`
- Adjust `transitionToIa` and `transitionToArchive` (e.g., AFTER_7_DAYS, AFTER_60_DAYS) based on access patterns
- Add or modify access points for additional apps (e.g., cache, uploads)
- Change POSIX uid/gid to match your container or application user
