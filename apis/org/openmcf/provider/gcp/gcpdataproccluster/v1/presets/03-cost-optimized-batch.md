# Cost-Optimized Batch

An ephemeral Dataproc cluster optimized for batch Spark jobs using Spot VMs for secondary workers, with aggressive auto-delete for cost control.

## When to Use

- Batch ETL jobs that can tolerate preemption
- Nightly or scheduled data processing pipelines
- Large-scale data transformations where cost matters more than latency
- Workloads with Spark dynamic allocation enabled

## Key Configuration Choices

- **1 master + 2 primary workers**: Minimal stable capacity
- **10 Spot secondary workers**: 80% cost savings vs on-demand for burst capacity
- **Spark dynamic allocation**: Automatically scales executors based on workload
- **15-minute idle auto-delete**: Cluster self-destructs shortly after job completion
- **No Component Gateway**: Batch jobs don't need web UI access

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-gcp-project-id>` | GCP project ID | GCP Console > Project Settings |

## Important Notes

- Spot VMs may be preempted at any time. Spark's task retry mechanism handles this gracefully for batch workloads.
- Adjust `secondaryWorkerConfig.numInstances` based on your job's parallelism requirements.
- The 15-minute idle timeout means the cluster auto-deletes ~15 minutes after the last job completes.

## Related Presets

- **01-dev-jupyter**: Interactive cluster for development
- **02-ha-production**: High-availability cluster for production
