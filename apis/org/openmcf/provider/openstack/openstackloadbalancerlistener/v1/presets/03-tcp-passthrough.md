# TCP Passthrough Listener

This preset creates a TCP listener that passes raw TCP traffic to backend pools without any protocol-level processing. Use this for databases, message queues, gRPC, or any non-HTTP service that needs load balancing at Layer 4.

## When to Use

- Database connections (PostgreSQL 5432, MySQL 3306, Redis 6379)
- Message queue protocols (AMQP 5672, Kafka 9092, NATS 4222)
- gRPC or custom binary protocols
- End-to-end TLS where the backend handles TLS termination

## Key Configuration Choices

- **TCP protocol** (`protocol: TCP`) -- Layer 4, no HTTP processing
- **Custom port** -- set to match your backend service's listening port
- **No header insertion** -- not applicable for TCP (headers are an HTTP concept)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<loadbalancer-id>` | ID of the load balancer to attach this listener to | OpenStack console or `OpenStackLoadBalancer` status outputs |
| `<port>` | TCP port number for the service (e.g., `5432` for PostgreSQL) | Your application configuration |

## Related Presets

- **01-http** -- Use instead for HTTP traffic
- **02-https-terminated** -- Use instead for HTTPS with TLS termination at the LB
