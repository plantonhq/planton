# Background Worker

This preset deploys a background worker that processes messages from an Azure Service Bus queue. It has no ingress (not accessible via HTTP), scales to zero when the queue is empty, and scales up to 5 replicas based on queue depth. This is the standard pattern for event-driven and queue-processing workloads.

## When to Use

- Queue consumers (Azure Service Bus, Storage Queue) that process messages asynchronously
- Background jobs that should scale to zero when idle (no cost when there's no work)
- Event-driven workloads where the trigger is a message queue, not HTTP traffic
- Workers that don't need to be accessible from the internet

## Key Configuration Choices

- **0 min replicas** (`minReplicas: 0`) -- Scale-to-zero; no cost when the queue is empty
- **5 max replicas** (`maxReplicas: 5`) -- Conservative ceiling; increase for high-throughput queues
- **0.25 vCPU / 0.5 GiB memory** -- Minimal resources for message processing; increase for CPU-intensive work
- **Custom KEDA rule** (`azure-servicebus`) -- Scales based on queue message count (10 messages per replica)
- **No ingress** -- Worker is not accessible via HTTP; processes messages only
- **Secret-backed env var** -- Queue connection string is stored as a secret and injected via `secretName`

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<container-app-environment-id>` | ARM ID of the Container App Environment | Azure portal or `AzureContainerAppEnvironment` status outputs |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `image: my-registry/my-worker:latest` | Your worker container image | Your container registry |
| `queueName: my-queue` | Service Bus queue name | Azure portal or Service Bus admin |
| `value: "Endpoint=sb://..."` | Service Bus connection string | Azure portal -> Service Bus -> Shared access policies |

## Related Presets

- **01-web-service** -- Use instead for HTTP services with external ingress
- **03-enterprise-api** -- Use instead for production APIs with identity and security controls
