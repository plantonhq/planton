# Background Worker Deployment

This preset deploys a background worker process without ingress. Use this for queue consumers, event processors, or any long-running process that does not serve HTTP traffic.

## When to Use

- Queue consumers (RabbitMQ, SQS, Kafka, NATS)
- Event processors or stream processors
- Background jobs that run continuously but do not expose an HTTP endpoint

## Key Configuration Choices

- **No ingress** -- worker processes do not serve external traffic
- **No ports defined** -- no Kubernetes Service is created; the worker only consumes from external sources
- **Environment variable** (`WORKER_CONCURRENCY`) -- example of how to pass configuration via env vars; replace with your app-specific variables
- **Single replica** (`minReplicas: 1`) -- scale by increasing this value or enabling HPA

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace for the deployment | Your namespace management or `KubernetesNamespace` resource |
| `<your-container-registry>/<your-image>` | Container image repository | Your container registry |
| `<your-image-tag>` | Image tag or version | Your CI/CD pipeline output |

## Related Presets

- **01-web-service** -- Web service with ingress for HTTP traffic
- **02-web-service-with-hpa** -- Production web service with autoscaling
