# Scaleway Serverless Container Examples

Complete, copy-paste ready YAML manifests for common serverless container patterns.

---

## Example 1: Simple Public Web Service

**Use Case**: Public HTTP API deployed from a Scaleway Container Registry.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessContainer
metadata:
  name: my-api
  org: my-org
  env: production
spec:
  region: fr-par
  image:
    registry_endpoint:
      valueFrom:
        kind: ScalewayContainerRegistry
        name: my-registry
        fieldPath: status.outputs.endpoint
    name: my-api
    tag: v1.2.3
  port: 8080
  privacy: public
  description: Production API service
  memory_limit_mb: 512
  timeout_seconds: 30
  http_option: redirected
  env:
    variables:
      - name: NODE_ENV
        value: production
      - name: LOG_LEVEL
        value: info
```

**Notes:**
- Scaleway Container Registry endpoint wired via `valueFrom`
- HTTP redirected to HTTPS for security
- 512 MB memory, 30s timeout for API workloads

---

## Example 2: VPC-Connected Database Service

**Use Case**: Backend service accessing a managed database via Private Network.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessContainer
metadata:
  name: user-service
  org: my-org
  env: production
spec:
  region: fr-par
  image:
    registry_endpoint:
      valueFrom:
        kind: ScalewayContainerRegistry
        name: my-registry
        fieldPath: status.outputs.endpoint
    name: user-service
    tag: v2.1.0
  port: 8080
  privacy: public
  memory_limit_mb: 1024
  cpu_limit: 560
  timeout_seconds: 60
  http_option: redirected
  env:
    variables:
      - name: APP_ENV
        value: production
    secrets:
      - name: DATABASE_URL
        value: postgresql://dbuser:password@10.0.1.5:5432/users
      - name: JWT_SECRET
        value: super-secret-jwt-key-12345
  private_network_id:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: my-private-network
      fieldPath: status.outputs.private_network_id
  health_check:
    path: /health
    failure_threshold: 3
    interval_seconds: 15
```

**Key points:**
- Private Network for secure database access via private IP
- Explicit CPU limit (560 milliCPU) for compute-bound workloads
- HTTP health check on `/health` every 15 seconds
- Database URL uses private IP (only reachable via VPC)

---

## Example 3: gRPC Service with h2c Protocol

**Use Case**: gRPC service requiring HTTP/2 cleartext protocol.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessContainer
metadata:
  name: grpc-gateway
  org: my-org
  env: production
spec:
  region: fr-par
  image:
    registry_endpoint:
      value: ghcr.io/my-org
    name: grpc-gateway
    tag: v3.0.1
  port: 50051
  privacy: public
  protocol: h2c
  memory_limit_mb: 512
  min_scale: 1
  max_scale: 10
  timeout_seconds: 120
  http_option: redirected
  env:
    variables:
      - name: GRPC_GO_LOG_SEVERITY_LEVEL
        value: info
  scaling_option:
    concurrent_requests_threshold: 20
```

**Notes:**
- `protocol: h2c` required for gRPC backends
- Non-Scaleway registry (GHCR) with plain `value` endpoint
- `min_scale: 1` for always-warm gRPC connections
- Scaling based on concurrent requests threshold

---

## Example 4: Scheduled Background Worker with Cron

**Use Case**: Private container triggered on a schedule for data processing.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessContainer
metadata:
  name: data-processor
  org: my-org
  env: production
spec:
  region: fr-par
  image:
    registry_endpoint:
      valueFrom:
        kind: ScalewayContainerRegistry
        name: my-registry
        fieldPath: status.outputs.endpoint
    name: data-processor
    tag: v1.0.0
  port: 8080
  privacy: private
  memory_limit_mb: 2048
  cpu_limit: 1120
  timeout_seconds: 300
  min_scale: 0
  env:
    secrets:
      - name: DATABASE_URL
        value: postgresql://processor:password@10.0.1.5:5432/analytics
  private_network_id:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: my-private-network
      fieldPath: status.outputs.private_network_id
  cron_triggers:
    - name: hourly-sync
      schedule: "0 * * * *"
      args: '{"mode": "incremental"}'
    - name: nightly-full
      schedule: "0 3 * * *"
      args: '{"mode": "full", "batch_size": 5000}'
```

**Notes:**
- Private container (cron-triggered only, no public HTTP endpoint)
- 2 GB memory + 1.12 vCPU for data-intensive work
- Two cron schedules with different JSON arguments
- Scale-to-zero between invocations

---

## Example 5: Always-Warm API with Health Checks and Scaling

**Use Case**: Latency-sensitive API with production-grade observability.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessContainer
metadata:
  name: realtime-api
  org: my-org
  env: production
spec:
  region: fr-par
  image:
    registry_endpoint:
      valueFrom:
        kind: ScalewayContainerRegistry
        name: my-registry
        fieldPath: status.outputs.endpoint
    name: realtime-api
    tag: v4.2.1
  port: 3000
  privacy: public
  memory_limit_mb: 1024
  cpu_limit: 560
  min_scale: 2
  max_scale: 50
  timeout_seconds: 10
  http_option: redirected
  health_check:
    path: /healthz
    failure_threshold: 2
    interval_seconds: 10
  scaling_option:
    concurrent_requests_threshold: 50
    cpu_usage_threshold: 70
  env:
    variables:
      - name: NODE_ENV
        value: production
      - name: CACHE_TTL
        value: "60"
```

**Notes:**
- `min_scale: 2` keeps 2 instances always warm (no cold starts)
- Health check with aggressive settings (10s interval, 2 failures)
- Dual scaling triggers: concurrent requests AND CPU usage
- **Cost implication**: `min_scale > 0` incurs continuous billing

---

## Example 6: Docker Hub Image (External Registry)

**Use Case**: Deploy a standard Docker Hub image without Scaleway registry.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessContainer
metadata:
  name: nginx-proxy
  org: my-org
  env: staging
spec:
  region: nl-ams
  image:
    registry_endpoint:
      value: docker.io/library
    name: nginx
    tag: "1.25-alpine"
  port: 80
  privacy: public
  memory_limit_mb: 128
  http_option: redirected
```

**Notes:**
- External registry (Docker Hub) with plain `value` endpoint
- Minimal resource footprint (128 MB)
- No Scaleway Container Registry dependency

---

## Example 7: Command Override with Custom Entrypoint

**Use Case**: Override container CMD for a different execution mode.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessContainer
metadata:
  name: worker
  org: my-org
  env: production
spec:
  region: fr-par
  image:
    registry_endpoint:
      valueFrom:
        kind: ScalewayContainerRegistry
        name: my-registry
        fieldPath: status.outputs.endpoint
    name: multi-mode-app
    tag: v2.0.0
  port: 8080
  privacy: private
  commands:
    - node
    - worker.js
  args:
    - "--queue"
    - "high-priority"
    - "--workers"
    - "4"
  memory_limit_mb: 512
  env:
    secrets:
      - name: REDIS_URL
        value: redis://10.0.1.10:6379
```

**Notes:**
- `commands` overrides the image's default CMD
- `args` provides arguments to the command
- Same image used for different execution modes (API vs worker)

---

## Example 8: Custom Domain via DNS Record

**Use Case**: Expose a container at `api.example.com`.

Deploy both resources -- the container and a CNAME record pointing to it:

```yaml
# Container
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessContainer
metadata:
  name: my-api
  org: my-org
  env: production
spec:
  region: fr-par
  image:
    registry_endpoint:
      valueFrom:
        kind: ScalewayContainerRegistry
        name: my-registry
        fieldPath: status.outputs.endpoint
    name: my-api
    tag: v1.0.0
  port: 8080
  privacy: public
  http_option: redirected
---
# DNS Record pointing to the container
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: api-dns
  org: my-org
  env: production
spec:
  zone_name:
    value: "example.com"
  name: "api"
  type: CNAME
  data:
    valueFrom:
      kind: ScalewayServerlessContainer
      name: my-api
      fieldPath: status.outputs.domain_name
  ttl: 300
```

**Notes:**
- DNS record creates DAG edge: DNS -> Container
- `domain_name` output automatically wired via `valueFrom`

---

## Example 9: Minimal Container (Development/Testing)

**Use Case**: Simplest possible container for learning or testing.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessContainer
metadata:
  name: hello-world
spec:
  region: fr-par
  image:
    registry_endpoint:
      value: docker.io/library
    name: nginx
    tag: alpine
  privacy: public
```

**Notes:**
- Minimal required fields only
- Defaults: port 8080, 256 MB memory, 300s timeout, 0-20 scaling
- No environment variables, networking, or cron

---

## Common Patterns Summary

| Use Case | Memory | CPU | Timeout | Protocol | Scaling | Key Features |
|----------|--------|-----|---------|----------|---------|--------------|
| Public API | 512-1024 MB | auto | 10-30s | http1 | 0-20 | Health check, env vars |
| Database API | 1024 MB | 560 | 30-60s | http1 | 1-20 | VPC, secrets, health check |
| gRPC Service | 512 MB | auto | 60-120s | h2c | 1-10 | Always-warm, scaling opts |
| Background Worker | 1024-2048 MB | 1120 | 300s | http1 | 0-5 | Cron triggers, private |
| Proxy/Sidecar | 128-256 MB | auto | 10s | http1 | 0-5 | External registry |

---

## Production Best Practices

### 1. Use Immutable Image Tags

**Do:**
```yaml
image:
  tag: v1.2.3   # Semantic version
  # or
  tag: sha-abc1234  # Git commit SHA
```

**Don't:**
```yaml
image:
  tag: latest  # Mutable, unpredictable deployments
```

### 2. VPC for Database Access

Always use Private Network when accessing databases:

```yaml
spec:
  private_network_id:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: my-network
      fieldPath: status.outputs.private_network_id
  env:
    secrets:
      - name: DATABASE_URL
        value: postgresql://10.0.1.5:5432/db
```

### 3. Health Checks for Production

Always configure health checks for production containers:

```yaml
spec:
  health_check:
    path: /healthz
    failure_threshold: 3
    interval_seconds: 15
```

### 4. Scaling Configuration

- **Start with defaults**: 0-20 scaling covers most use cases
- **min_scale > 0**: Only for latency-critical services (costs money 24/7)
- **Scaling options**: Use `concurrent_requests_threshold` for HTTP APIs, `cpu_usage_threshold` for compute-heavy workloads

---

## Further Reading

- **Component Overview**: See [README.md](./README.md)
- **Scaleway Containers Docs**: [Official Documentation](https://www.scaleway.com/en/docs/serverless/containers/)
- **Container Limitations**: [Scaleway Docs](https://www.scaleway.com/en/docs/serverless-containers/reference-content/containers-limitations/)
