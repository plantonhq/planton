---
title: "How to Deploy Redis on Kubernetes"
date: "2026-04-02"
author:
  - name: "Planton Team"
    title: "Platform Engineering"
    bio: "Helping teams deploy infrastructure and services without the DevOps bottleneck"
tags:
  - "kubernetes"
  - "redis"
  - "cache"
  - "infra-catalog"
category: "kubernetes"
excerpt: "Deploy a persistent Redis instance on any Kubernetes cluster using a single YAML manifest and the Planton CLI."
---

# How to Deploy Redis on Kubernetes

This tutorial walks you through deploying a Redis instance on a Kubernetes cluster through Planton. You will write a YAML manifest describing the Redis configuration you want, deploy it with a single CLI command, and verify the outputs you need to connect your applications. By the end, you will have a running Redis instance with persistence, an auto-generated password, and the connection details your services need.

> **Note**: The Planton web console provides a guided creation wizard for Redis
> and other Infra Components. This tutorial uses the CLI/YAML approach for stability
> and reproducibility. The console UI evolves frequently — always check it for the
> latest experience.

## What You Will Learn

- How Kubernetes Infra Components differ from cloud-provider Infra Components
- How to write a `KubernetesRedis` manifest that deploys Redis to any connected cluster
- How to deploy with `planton apply` and monitor progress in real time
- How to retrieve deployment outputs (service endpoint, password secret) for application use
- How to enable external access when services outside the cluster need to reach Redis

## Prerequisites

- [ ] A Kubernetes provider connection configured and set as the default for your target environment. This connection tells Planton which cluster to deploy to. You can create one through the Planton web console under Connect > Kubernetes.
- [ ] The target cluster must have a default StorageClass configured if you plan to enable persistence. Most managed Kubernetes services (GKE, EKS, AKS) include one by default.
- [ ] A Planton organization and at least one environment created
- [ ] The `planton` CLI installed and authenticated (`planton auth login`)

## What Is a Kubernetes Infra Component?

A Kubernetes Infra Component deploys open-source software (like Redis, PostgreSQL, or Kafka) onto an existing Kubernetes cluster using Helm charts, managed through the same `planton apply` workflow as cloud-managed resources. Unlike provider-managed Infra Components (like GCP Cloud SQL), these deploy directly into a namespace on your cluster. For more details, see the [Infra Components documentation](/docs/infrastructure/infra-components).

## Step 1: Write the Redis Manifest

Create a file named `redis.yaml` with the following content:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRedis
metadata:
  name: app-cache
  org: your-org
  env: production
spec:
  namespace:
    value: "redis-production"
  createNamespace: true
  container:
    resources:
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 1000m
        memory: 1Gi
    persistenceEnabled: true
    diskSize: 1Gi
```

Replace these placeholder values with your own:

- **`metadata.name`**: A name for this Redis instance. Planton uses it to derive Kubernetes resource names (the master Service will be named `app-cache-master`, the password Secret will be `app-cache-password`).
- **`metadata.org`**: Your Planton organization slug.
- **`metadata.env`**: The environment this Redis instance belongs to (e.g., `production`, `staging`, `dev`).
- **`spec.namespace.value`**: The Kubernetes namespace where Redis will be deployed.

Here is what each section of the spec configures.

### Namespace

The `namespace` field uses a nested `value` key because it supports two modes: a literal string (shown here) or a reference to another Infra Component's outputs using `valueFrom`. For this tutorial, a literal value is all you need.

When `createNamespace` is `true`, Planton creates the namespace if it does not already exist. Set this to `false` if the namespace is managed separately or already exists on your cluster.

### Container resources

The `resources` block sets CPU and memory requests and limits for the Redis pod. Requests are what Kubernetes guarantees to the pod; limits are the maximum it can consume before being throttled (CPU) or terminated (memory).

The values shown here match the defaults: 50m CPU request with a 1000m limit, and 100Mi memory request with a 1Gi limit. For production workloads handling significant traffic, increase the memory limit to match your working set size -- Redis stores everything in memory, so the memory limit effectively caps how much data Redis can hold.

### Persistence

When `persistenceEnabled` is `true`, Redis backs up its in-memory data to a persistent volume. If the pod restarts, Redis restores data from this volume instead of starting empty. The `diskSize` field sets the size of this volume using Kubernetes quantity notation (e.g., `1Gi`, `5Gi`, `10Gi`).

When persistence is disabled, Redis operates as a purely ephemeral cache -- data is lost on pod restart.

**Important**: The `diskSize` value cannot be changed after the initial deployment. This is a Kubernetes limitation on StatefulSet persistent volume claims. If you need more storage later, you will need to deploy a new Redis instance and migrate data.

## Step 2: Deploy with `planton apply`

Run the following command to deploy Redis. The `-t` flag streams the deployment progress to your terminal in real time.

```bash
planton apply -f redis.yaml -t
```

Planton validates the manifest, creates a deployment job, and begins provisioning Redis on your Kubernetes cluster. The terminal output shows four phases:

1. **init**: Configures the Kubernetes provider using your connection credentials (a few seconds)
2. **refresh**: Checks for any existing state (a few seconds)
3. **preview**: Plans the changes -- shows the Kubernetes resources that will be created (several seconds)
4. **update**: Creates the namespace (if requested), password Secret, and installs the Redis Helm chart (typically 1-3 minutes)

Redis deployments are faster than cloud-managed database deployments because they run entirely within the Kubernetes cluster -- there is no cloud provider API to wait on.

If you prefer to deploy without streaming, omit the `-t` flag:

```bash
planton apply -f redis.yaml
```

The CLI prints the deployment job ID immediately. You can check on it later with:

```bash
planton follow <infra-job-id>
```

## Step 3: Verify the Deployment

After the deployment completes, retrieve the Infra Component to see its status and outputs:

```bash
planton get KubernetesRedis app-cache -o yaml
```

The `status.outputs` section contains the values you need to connect your applications to Redis:

| Output | Description | Example |
|---|---|---|
| `namespace` | Kubernetes namespace where Redis is running | `redis-production` |
| `service` | Kubernetes Service name for the Redis master | `app-cache-master` |
| `kube_endpoint` | Full in-cluster DNS address | `app-cache-master.redis-production.svc.cluster.local` |
| `port_forward_command` | Command for local port-forwarding access | `kubectl port-forward ...` |
| `username` | Redis username | `default` |
| `password_secret.name` | Name of the Kubernetes Secret containing the password | `app-cache-password` |
| `password_secret.key` | Key within the Secret that holds the password | `password` |

To list all deployment jobs for this resource:

```bash
planton infra job list <infra-component-id>
```

The Infra Component ID is in the `metadata.id` field of the `planton get` output.

## Step 4: Connect to Redis from Your Application

### In-cluster access

Applications running in the same Kubernetes cluster can reach Redis using the `kube_endpoint` output. The address follows the pattern `{name}-master.{namespace}.svc.cluster.local` and Redis listens on port 6379.

For a Node.js application using ioredis:

```javascript
const Redis = require('ioredis');
const redis = new Redis({
  host: 'app-cache-master.redis-production.svc.cluster.local',
  port: 6379,
  password: process.env.REDIS_PASSWORD,
});
```

### Retrieving the password

The deployment creates a Kubernetes Secret containing the auto-generated password. To retrieve it:

```bash
kubectl get secret app-cache-password -n redis-production \
  -o jsonpath='{.data.password}' | base64 -d
```

To inject this password into your application pods, mount the Secret as an environment variable in your deployment manifest:

```yaml
env:
  - name: REDIS_PASSWORD
    valueFrom:
      secretKeyRef:
        name: app-cache-password
        key: password
```

### Local access via port-forwarding

To connect to Redis from your local machine for debugging or ad-hoc queries, forward the Redis port:

```bash
kubectl port-forward -n redis-production svc/app-cache-master 6379:6379
```

Then connect with `redis-cli`:

```bash
redis-cli -h localhost -p 6379 -a "$(kubectl get secret app-cache-password -n redis-production -o jsonpath='{.data.password}' | base64 -d)"
```

Verify the connection with:

```text
localhost:6379> PING
PONG
```

## Enabling External Access (Optional)

If you need to reach Redis from outside the Kubernetes cluster -- for example, from external monitoring tools or services running in a different cluster -- you can enable the ingress configuration.

This requires [external-dns](https://github.com/kubernetes-sigs/external-dns) to be running on your target cluster. external-dns watches for Kubernetes Services with hostname annotations and automatically creates DNS records pointing to the Service's external IP.

Add the `ingress` section to your manifest:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRedis
metadata:
  name: app-cache
  org: your-org
  env: production
spec:
  namespace:
    value: "redis-production"
  createNamespace: true
  container:
    resources:
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 1000m
        memory: 1Gi
    persistenceEnabled: true
    diskSize: 1Gi
  ingress:
    enabled: true
    hostname: "redis-prod.example.com"
```

When ingress is enabled, Planton creates a LoadBalancer Service with an `external-dns.alpha.kubernetes.io/hostname` annotation. external-dns picks up this annotation and creates a DNS record pointing `redis-prod.example.com` to the load balancer's external IP address. Redis remains accessible on port 6379 through this hostname.

Replace `redis-prod.example.com` with a hostname within a DNS zone managed by your external-dns configuration.

After deploying with ingress enabled, the `external_hostname` output will show the configured hostname:

```bash
planton get KubernetesRedis app-cache -o yaml
```

Connect from outside the cluster:

```bash
redis-cli -h redis-prod.example.com -p 6379 -a "<your-password>"
```

**Security consideration**: Exposing Redis to the public internet is a significant security decision. Redis is designed as an internal service. If you enable external access, ensure your cluster's network policies, firewall rules, and load balancer security groups restrict access to trusted IP ranges. The auto-generated password provides authentication, but network-level controls are the primary line of defense.

## Development Configuration

For development and testing environments where speed and simplicity matter more than durability, use a lighter configuration:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRedis
metadata:
  name: app-cache-dev
  org: your-org
  env: dev
spec:
  namespace:
    value: "redis-dev"
  createNamespace: true
  container:
    resources:
      requests:
        cpu: 25m
        memory: 64Mi
      limits:
        cpu: 250m
        memory: 256Mi
    persistenceEnabled: false
```

Here is what changed from the production configuration and why:

- **Smaller resource limits**: 250m CPU and 256Mi memory. Adequate for development workloads and leaves more cluster resources available for other services.
- **`persistenceEnabled: false`**: Redis operates as a purely ephemeral cache. No persistent volume is created, which means faster startup and no storage costs. Data is lost when the pod restarts. This is acceptable for development caches that can be repopulated.
- **No `diskSize`**: When persistence is disabled, `diskSize` is not needed. The validation rules allow it to be omitted.
- **No `ingress`**: Development instances are typically accessed through port-forwarding or in-cluster connections. No external access is needed.

Deploy the development configuration the same way:

```bash
planton apply -f redis-dev.yaml -t
```

## Common Patterns and Tips

### Standalone architecture

The current `KubernetesRedis` implementation deploys Redis in standalone mode -- a single master instance. This is appropriate for caching workloads, session stores, and applications that can tolerate brief unavailability during pod restarts. If your use case requires Redis Sentinel (automatic failover with read replicas) or Redis Cluster (data sharding across nodes), consider deploying a self-managed Helm release using the `KubernetesHelmRelease` catalog kind, which gives you full control over the Helm chart values.

### Persistence sizing

Choose a `diskSize` that accounts for your expected data set plus overhead for Redis persistence files (RDB snapshots and AOF logs). Redis's memory-to-disk ratio varies by workload, but allocating disk space equal to or greater than your memory limit is a reasonable starting point. Remember that `diskSize` cannot be changed after the initial deployment.

### Multiple Redis instances

You can deploy multiple Redis instances in the same namespace by using distinct `metadata.name` values. Each instance gets its own Service, Secret, and persistent volume. For example, you might run one Redis for application caching and another for session storage:

```bash
planton apply -f redis-cache.yaml
planton apply -f redis-sessions.yaml
```

The derived resource names (Services, Secrets) use the `metadata.name` as a prefix, so `app-cache` and `app-sessions` will not conflict.

## What to Do Next

Your Redis instance is running on Kubernetes. From here:

- **Connect a backend service** to Redis. If you have not deployed a service yet, see [How to Deploy Your First Service with Zero-Config CI/CD](/tutorials/how-to-deploy-your-first-service-with-zero-config-cicd) -- the environment variable and Secret mounting patterns from Step 4 above apply directly to services deployed through Planton.
- **Explore other Kubernetes catalog kinds** in the Infra Catalog. The same `planton apply` workflow works for PostgreSQL, Kafka, MongoDB, Elasticsearch, and dozens of other open-source tools that can be deployed onto your Kubernetes clusters.
- **Consider managed alternatives** for production-critical workloads. Planton's Infra Catalog also includes managed Redis options like GCP Memorystore (`GcpRedisInstance`), AWS ElastiCache (`AwsRedisElasticache`), and AWS Serverless ElastiCache (`AwsServerlessElasticache`). These trade the flexibility of running on your own cluster for fully managed operations, automated patching, and SLA-backed availability.
