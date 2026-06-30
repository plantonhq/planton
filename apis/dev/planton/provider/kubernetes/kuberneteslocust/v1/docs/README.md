# Deploying Locust on Kubernetes: From Manual Manifests to Production-Ready Abstractions

## Introduction

Load testing has come a long way from the days of manually scaling virtual machines and praying your infrastructure could handle the spike. Locust, with its Python-based scripting and elegant distributed architecture, modernized load testing by making it accessible to developers. But deploying Locust *on Kubernetes* introduced a new set of challenges: managing the master-worker architecture, injecting test scripts without rebuilding containers, installing dependencies at runtime, and orchestrating rolling updates when scripts change.

The Kubernetes ecosystem has evolved a clear pattern for deploying Locust at scale. This document explores that landscape—from the anti-patterns that fail under load to the production-ready approaches that power tests generating 20,000+ requests per second. More importantly, it explains *why* Planton's LocustKubernetes resource is designed the way it is: as an opinionated abstraction that eliminates the most painful friction points in the developer workflow.

## The Locust Architecture: Master-Worker and the GIL

Before understanding deployment patterns, you need to understand *why* Locust is inherently distributed.

### The Python GIL Constraint

Locust is written in Python, which means it inherits Python's Global Interpreter Lock (GIL). The GIL ensures only one thread can execute Python bytecode at a time within a single process. No matter how many CPU cores your machine has, a single Locust process can only fully utilize *one* core.

To bypass this limitation and generate serious load, Locust employs a **master-worker architecture**:

- **Master Node**: A single process (started with `--master`) that runs the web UI, coordinates workers, and aggregates statistics. Critically, the master *does not* simulate any users itself.
- **Worker Nodes**: One or more processes (started with `--worker`) that connect to the master, receive commands, run the test scripts, and report statistics back for aggregation.

The standard practice is to run **one worker process per CPU core** to maximize load generation capacity.

### Mapping to Kubernetes Primitives

This architecture maps cleanly to Kubernetes:

- **Master**: A Deployment with `replicas: 1` and container args `["--master"]`
- **Master Service**: A ClusterIP Service exposing ports 8089 (web UI), 5557, and 5558 (worker communication). This provides a stable DNS name (e.g., `locust-master`) for worker discovery.
- **Workers**: A separate Deployment with `replicas: N` and container args `["--worker", "--master-host=locust-master"]`
- **UI Access**: An Ingress resource to expose the master's web UI externally

Both master and workers are deployed as standard **Deployments** (not StatefulSets). While the master aggregates state in memory, this state is *ephemeral*—it only contains statistics for the currently running test. If the master pod crashes, the test is over. Using a StatefulSet would provide no benefit, as the master doesn't require persistent storage or a stable pod identity. The stable identity workers need is provided by the Service's DNS name, not the pod itself.

## The Deployment Method Spectrum

Let's progress through the deployment approaches, from what doesn't work to what powers production systems.

### Level 0: The Single-Pod Anti-Pattern

Deploying Locust as a single pod is fundamentally broken:

1. **GIL Limitation**: A single Python process can only use one CPU core, severely throttling load generation.
2. **No Scalability**: The entire design philosophy of Locust is *distributed* load generation. A single pod cannot be scaled horizontally to meet high-volume requirements.

**Verdict**: Never deploy Locust as a single pod. It defeats the purpose of the tool.

### Level 1: Manual Kubernetes Manifests

The "roll your own" approach involves creating YAML manifests manually. This is the foundational pattern that all higher-level abstractions build upon.

**What you need**:
- `master-deployment.yaml`: Deployment with one master pod
- `worker-deployment.yaml`: Deployment with N worker pods
- `service.yaml`: ClusterIP Service for the master (ports 8089, 5557, 5558)
- `scripts-cm.yaml`: ConfigMap created via `kubectl create configmap --from-file=locustfile.py`

**Pros**: Full control and transparency  
**Cons**: Verbose, manual ConfigMap management, no built-in mechanism for script updates or dependency management

**Verdict**: Educational, but too manual for iterative test development.

### Level 2: Helm Charts (The Industry Standard)

Helm is the dominant abstraction for deploying Locust on Kubernetes. While there is no "official" Helm chart from the locustio project, the **deliveryhero/helm-charts** repository has become the de-facto standard. The official Locust documentation explicitly endorses it as "a good helm chart" and the "most up to date" option.

#### Why DeliveryHero's Chart Won

- **Production-proven**: Used in case studies generating 19,000+ requests per second
- **Official endorsement**: Recommended by locust.io's own documentation
- **Feature-rich**: Supports the critical `pip_packages` field (more on this below)
- **Actively maintained**: Regular updates and community contributions

#### Key Features

The chart's `values.yaml` provides a mature API:

| Feature | Configuration | Purpose |
|---------|---------------|---------|
| **Script Injection** | `loadtest.locust_locustfile_configmap` | Reference to pre-existing ConfigMap containing `locustfile.py` |
| **Library Injection** | `loadtest.locust_lib_configmap` | Reference to ConfigMap with additional Python modules |
| **Pip Dependencies** | `loadtest.pip_packages: ["boto3", "pandas"]` | Packages installed at runtime via init container |
| **Worker Scaling** | `worker.replicas: 10` | Static worker count (the 80% use case) |
| **HPA Support** | `worker.hpa.enabled: false` (default) | Optional autoscaling for advanced users |
| **Ingress** | `ingress.enabled: true` | Expose web UI externally |

#### The "Tricky" Part: Manual ConfigMap Creation

The chart's primary friction point is that it *references* ConfigMaps by name. You must run `kubectl create configmap locust-scripts --from-file=locustfile.py` *before* installing the chart. This manual pre-step is exactly the kind of workflow friction that Planton eliminates (more on that later).

**Verdict**: Production-ready, but requires manual orchestration of ConfigMaps and script updates.

### Level 3: Kubernetes Operators

The operator pattern extends Kubernetes with custom resources. The most mature option is the **locust-k8s-operator**, which introduces a `LocustTest` CRD.

**How it works**:
```yaml
apiVersion: locust-operator.io/v1
kind: LocustTest
spec:
  image: locustio/locust:latest
  workerReplicas: 3
  configMap: demo-test-map
  libConfigMap: demo-lib-map
```

The operator watches for `LocustTest` resources and generates the underlying Deployments, Services, and ConfigMaps.

**Key advantage**: Declarative, Kubernetes-native API with a `status` field that enables CI/CD integration (e.g., `kubectl wait --for=condition=Completed`).

**Planton's Approach**: As an IaC framework, Planton *is* an abstraction layer. We don't deploy third-party operators. Instead, our `LocustKubernetes` resource behaves *like* the `LocustTest` CRD—it provides a high-level, declarative API while the Planton controller directly creates and manages the Kubernetes resources.

## The Developer Experience Problem

Deploying Locust on Kubernetes is technically straightforward. The challenge is the *developer workflow*:

### Problem 1: Script Management

Test scripts must get into the pods somehow. There are three methods:

| Method | How It Works | Developer Experience |
|--------|--------------|---------------------|
| **ConfigMap** | `kubectl create cm --from-file=locustfile.py` + volumeMount | ✅ Fast iteration, but "tricky" manual step |
| **Custom Docker Image** | `COPY locustfile.py` in Dockerfile | ❌ Requires full docker build/push cycle for every change |
| **Persistent Volume** | Mount NFS or EBS volume | ❌ Massive overkill; adds stateful complexity unnecessarily |

The ConfigMap approach is the industry standard because it decouples scripts from images. However, it requires manual pre-creation and lacks a native "hot reload" mechanism.

### Problem 2: Python Dependencies

Real-world test scripts have dependencies (requests, boto3, pandas, etc.). Managing these creates another workflow challenge:

| Method | How It Works | Developer Experience |
|--------|--------------|---------------------|
| **Custom Docker Image** | `RUN pip install -r requirements.txt` in Dockerfile | ❌ Requires rebuilding the image for every new package |
| **Runtime Installation** | `pip_packages: ["boto3"]` in Helm values | ✅ Packages installed via init container at pod startup |

The DeliveryHero Helm chart's `pip_packages` field is a **game-changer** for developer experience. The chart's entrypoint script reads this list and runs `pip install` before starting Locust. This completely eliminates the need to build custom Docker images for common dependencies.

The tradeoff is slower pod startup (packages install on every restart) and a runtime dependency on PyPI. But for iterative test development, this is an acceptable cost for the massive DX improvement.

### Problem 3: Script Updates

Locust doesn't auto-detect file changes. If you update a ConfigMap, the running pods won't see the change—they only read the script at startup.

**The Manual Workflow**:
1. Update `locustfile.py`
2. Update ConfigMap: `kubectl apply -f cm.yaml`
3. Manually restart: `kubectl rollout restart deployment/locust-master`

**The Superior Kustomize Pattern**:

Production case studies reveal a much better approach using Kustomize's `configMapGenerator`:

1. Kustomize generates a ConfigMap with a *content hash* in the name: `locust-scripts-a1b2c3d4`
2. The Deployment is patched to reference this hashed name
3. When you change `locustfile.py` and run `kubectl apply -k .`, Kustomize creates a *new* ConfigMap with a *new* hash
4. Updating the ConfigMap reference in the Deployment's pod template triggers an **automatic rolling update**

This hash-based rollout is the gold standard for "GitOps-native" script updates. Planton implements this logic internally—changing the `load_test` spec triggers an automatic rolling update of the managed Deployments.

## Planton's Design Philosophy

Planton's LocustKubernetes resource is not a simple wrapper around the DeliveryHero Helm chart. It's an **opinionated abstraction** that automates the most painful parts of the workflow.

### What We Solve

1. **Automatic ConfigMap Management**: You provide script content directly in the API. The Planton controller synthesizes ConfigMaps with content-hashed names, eliminating the "tricky" manual pre-step.

2. **Automatic Dependency Management**: You specify `pip_packages` in the spec. The controller injects an init container that installs these packages at runtime, removing the need for custom Docker images.

3. **Automatic Rollouts**: Any change to your test script content triggers a rolling update. No manual `kubectl rollout restart` required.

### The API Design

The `LocustKubernetesSpec` proto reflects the 80/20 principle—expose the 20% of configuration that 80% of users need:

**Essential fields** (the 80%):
- `load_test.main_py_content`: The test script itself
- `load_test.lib_files_content`: Additional Python modules
- `load_test.pip_packages`: Runtime dependencies
- `master_container.resources`: Master CPU/memory allocation
- `worker_container.replicas`: Static worker count (the primary scaling knob)
- `worker_container.resources`: Per-worker resources
- `ingress`: Web UI access configuration

**Advanced fields** (the 20%):
- `helm_values`: Escape hatch for fine-grained control (e.g., HPA, affinity, tolerations)

### Why Static Scaling is the Default

You might expect worker autoscaling (HPA) to be a core feature. It's not, and here's why:

The Horizontal Pod Autoscaler is **reactive**. It scales *after* observing high CPU. But Locust's master dispatches workload when you start the test—only to *currently connected* workers. If HPA adds a new worker pod mid-test, that worker connects to the master but sits idle. It wasn't part of the initial workload dispatch, and Locust has no protocol to rebalance users to new workers.

The production pattern is **static pre-scaling**: determine the required worker count (e.g., 50), set `worker_container.replicas: 50`, wait for all pods to be ready, *then* start the test.

HPA is supported via `helm_values` for the 20% of users with advanced use cases (e.g., external state management patterns). But it's disabled by default because it's a common trap for new users.

## Production Operations

### Resource Allocation Strategy

Master and worker resource needs are asymmetric:

- **Workers**: CPU and network-bound. They execute test scripts and generate HTTP traffic. Their needs scale linearly with replica count.
- **Master**: CPU and memory-bound. The master processes and aggregates statistics from *every* worker in real-time. In a high-volume test (100 workers, 20,000 RPS), the master is processing a massive inbound stream. Under-provisioning the master is a common pitfall that leads to OOMKills mid-test.

**Recommendation**: Generous master resources (e.g., 1 CPU, 2Gi memory) for large-scale tests.

### Security Hardening

The DeliveryHero Helm chart dangerously defaults `securityContext.runAsNonRoot` to `false`. Planton enforces security by default:

- `runAsNonRoot: true`
- `runAsUser: 1000` (high UID)
- Minimal RBAC (Locust pods need zero Kubernetes API permissions)
- Optional NetworkPolicy enforcement (restrict master/worker communication to necessary ports)

### Cost Optimization: Spot Instances

Locust workers are a *perfect* fit for Spot/Preemptible instances:

- **Stateless**: Workers hold no persistent state
- **Fault-tolerant**: Losing 1 out of 50 workers is acceptable performance degradation
- **Batch workload**: Tests have a defined start and end

**Pattern**:
1. Create a Kubernetes NodePool using Spot instances
2. Taint the nodes: `workload-type=spot:NoSchedule`
3. Configure worker scheduling via `helm_values`:
   - `worker.tolerations` to tolerate the taint
   - `worker.nodeSelector` or `worker.affinity` to target Spot nodes

This can reduce compute costs by up to 90% while maintaining test reliability.

### Observability

Locust doesn't natively export Prometheus metrics. The standard solution is the **containersol/locust_exporter** sidecar, which scrapes the master's web UI and translates statistics into a Prometheus-compatible `/metrics` endpoint.

The DeliveryHero project provides a dedicated Helm chart for this exporter: `deliveryhero/prometheus-locust-exporter`.

**Recommendation**: Planton could expose a simple `metrics.prometheus.enabled` boolean that co-deploys this exporter and the necessary ServiceMonitor resources.

### CI/CD Integration

For automated performance testing pipelines, "headless" mode is essential. The challenge is: how does the CI job know when the test is complete?

The operator pattern solves this elegantly via a `status` sub-resource. The controller monitors the master pod and updates `status.phase = "Completed"` or `"Failed"` when the test finishes. This enables a robust, Kubernetes-native wait command:

```bash
kubectl wait --for=condition=Completed LocustKubernetes/my-test --timeout=30m
```

Planton's LocustKubernetes resource implements this status pattern, making CI/CD integration seamless.

## Conclusion: The Paradigm Shift

The evolution of Locust on Kubernetes reflects a broader shift in cloud-native development: from "infrastructure as YAML" to "infrastructure as high-level, opinionated APIs."

Manual manifests gave you control at the cost of verbosity. Helm charts provided reusable templates but still required manual orchestration of ConfigMaps and dependencies. Operators introduced declarative, Kubernetes-native resources but required installing and managing additional controllers.

Planton synthesizes the best of all three approaches: the declarative simplicity of operators, the production-readiness of the DeliveryHero Helm chart's `pip_packages` innovation, and the GitOps-native script management of Kustomize's content-hashed ConfigMaps—all wrapped in a single, opinionated abstraction that eliminates workflow friction.

The result is a load testing platform where you define your test in protobuf, commit it to git, and let the system handle the rest. No manual ConfigMaps. No custom Docker images. No manual rollouts. Just iterative, fast-paced load test development that scales from dev clusters to production systems generating 20,000+ requests per second.

That's the promise of truly cloud-native infrastructure.

