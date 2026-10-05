# Locust

Declares one Locust load-testing cluster -- the open-source tool that simulates thousands of concurrent users against your own applications, with test behavior written as plain Python. The deliveryhero Helm chart (0.35.0, OCI-served from ghcr.io -- the classic index stalled in 2024 -- running the OFFICIAL locustio/locust image at 2.32.2) renders a master Deployment (the web UI and REST API on port 8089, worker coordination on 5557) and a worker fleet, each worker roughly ONE CPU CORE of load generation. Test scripts reach the pods as ConfigMaps: written inline in the spec (the module renders them and script changes ROLL THE PODS through a content-hash annotation) or referenced from ConfigMaps your CI already ships. SECURED BY DEFAULT: upstream ships the web UI OPEN -- anyone who can reach the Service can fire load at any host the cluster sees -- and this kind never deploys that: the login is ON from the first apply with a module-generated credential in the `<name>-auth` Secret, delivered through Locust's own extension seam, never as pod arguments.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Script ConfigMaps** (inline scripts) -- `<name>-locustfile` holding your locustfile as `main.py`, and `<name>-lib` holding supporting modules mounted at `lib/`; a content-hash annotation on the pod template rolls the pods when scripts change
- **The Helm release** -- the deliveryhero `locust` chart, rendering:
  - The master Deployment -- the web UI and REST API (port 8089) and the coordination endpoint workers dial (5557); exactly one master by design
  - The worker Deployment -- the load generators, at your fixed count (0 is a legal PAUSED fleet) or autoscaled
  - The master Service (ClusterIP by default) carrying the web and worker-connect port families
- **The web-login backend** (login on, non-headless) -- a module-owned ConfigMap delivering Locust's documented auth extension, plus the `<name>-auth` Secret holding the generated credential (keys `username`/`password`) -- ONE mounted source of truth; rotating the credential is a single-Secret edit
- **Autoscaling objects** (arm-dependent) -- an HPA scaling workers on CPU utilization, or a KEDA ScaledObject scaling on the live user count the master reports
- **PodDisruptionBudgets** (opt-in per role) -- `maxUnavailable 0`, so node drains never kill the master or workers mid-test

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster. Map it as the default for your environment, or pick it explicitly in the wizard.

### Kubernetes Cluster

- **Internet or mirror access from the pods** (only when `pipPackages` / `pipRequirementsConfigMap` are set) -- pip installs at EVERY pod start; a PyPI outage becomes a pod-start failure. Bake a custom image for production gates.
- **KEDA operator** (only for the KEDA autoscaling arm) -- a KubernetesKeda composes naturally. NOTE the pairing the spec enforces: KEDA's DEFAULT trigger polls the master's `/stats/requests` API, which the web-UI login locks out and headless mode never serves -- provide custom triggers, or explicitly disable the login on a non-headless run.
- **Metrics server** (only for the HPA arm) -- plus worker CPU requests, or the HPA has no denominator.
- **Same-namespace credentials** -- Secrets named by `envFromSecrets` / `envFromSecretKeys` and existing script ConfigMaps are read by the pods at RUNTIME and must live in the install namespace.

## Deploy

### Console

Open the deployment store, find **Locust**, and click **Deploy**. The wizard walks the full run declaration: placement, the scripts (inline or existing ConfigMaps), the target, the test environment, dependencies, the master and worker fleet, autoscaling, sign-in, images, and the front door. Start from the **Web load test** preset in the [Presets](#presets) tab for an interactive swarm, or the **Headless CI gate** preset for a CI performance gate.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesLocust
metadata:
  name: load-test
  org: acme-corp
  env: dev
spec:
  namespace:
    value: load-test
  createNamespace: true
  loadTest:
    inline:
      locustfileContent: |
        from locust import HttpUser, task, between

        class WebsiteUser(HttpUser):
            wait_time = between(1, 2)

            @task
            def get_index(self):
                self.client.get("/")
    targetHost:
      value: http://my-app.my-namespace.svc.cluster.local:8080
  workers:
    replicas: 2
    resources:
      requests:
        cpu: 500m
        memory: 256Mi
```

```shell
planton apply -f load-test.yaml
```

This creates a one-master, two-worker Locust cluster in the `load-test` namespace with the inline script rendered as a ConfigMap and the web-UI login ON with a module-generated credential. An Infra Job tracks the provisioning in real time.

Read the generated password with:

```shell
kubectl get secret load-test-auth -n load-test \
  -o jsonpath='{.data.password}' | base64 -d
```

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the swarm to a namespace managed by another Infra Component -- and point `targetHost` at an endpoint another resource exports:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: load-test-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions the Locust cluster into it.

## Key Configuration

These are the most important decisions when configuring a Locust cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**`loadTest.scripts`** (required oneof) -- `inline` (locustfile + lib modules rendered into ConfigMaps; changes roll the pods) XOR `existingConfigMaps` (your CI ships the scripts; same-namespace rule applies).

**`loadTest.targetHost`** -- a literal URL or a reference to another resource's exported endpoint ("load-test what you declare"). Empty = the locustfile must set `host` itself; Locust refuses to start a test without one from either source.

**`loadTest.headless`** -- no web UI; the test starts as soon as the pods are up. Run shape rides `LOCUST_USERS` / `LOCUST_SPAWN_RATE` / `LOCUST_RUN_TIME` in `loadTest.environment`. The login is moot and the web-UI outputs honestly read empty.

**`loadTest.name`** -- labels every resource and lands in the Deployments' IMMUTABLE selector labels: renaming is delete-and-recreate (cheap -- the cluster is stateless). Empty = the resource name.

**`workers.replicas`** -- 0 or more; 0 is a PAUSED fleet. Ignored while an autoscaling arm owns the count. Size from the target: requests-per-second ÷ what one CPU core of your test generates.

**`workers.autoscaling`** -- fixed count (no arm) XOR `hpa` (CPU-based; the 40% default target is deliberate -- workers must scale BEFORE saturating or they distort the load they generate) XOR `keda` (the live user count; `minReplicas: 0` is legal scale-to-zero; `customTriggers` replaces the default trigger).

**`webUiAuth.enabled`** -- unset = ON with the generated credential (the open UI never ships); the explicit `false` is a recorded decision for fenced dev clusters and the KEDA default-trigger pairing. The login mechanism requires image tags >= 2.21.0 -- below the floor the chart would render credentials as literal pod arguments, which the module refuses.

**`service.type`** -- ClusterIP by default: compose real exposure (Gateway API, KubernetesIngress) over the exported service handle; the annotations map carries the cloud-LB vocabulary when you flip to LoadBalancer.

**`helmValues`** -- merged LAST over everything the spec renders; the module RE-PINS the login wiring and script delivery after the merge, so the security posture cannot be silently disabled from the hatch.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

`loadTest.targetHost` also accepts a ValueFromRef against ANY resource's exported endpoint -- load-test the services you already declare -- and same-namespace Secrets/ConfigMaps named by `envFromSecrets`, `envFromSecretKeys`, and existing script ConfigMaps are read by the pods at runtime.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the cluster runs in | Locating the swarm for diagnostics |
| `master_service` | The master Service name | The handle exposure kinds (Gateway API, Ingress) route to |
| `web_endpoint` | In-cluster web UI / REST API URL (port 8089) | Driving runs from automation; pair with the credential when the login is on |
| `master_bind_endpoint` | Worker-connect endpoint (port 5557) | Registering additional workers from other namespaces or clusters |
| `web_ui_username` | The login identity (empty when the login is off or the run is headless) | Operator sign-in |
| `web_ui_password_secret` | Secret + key holding the generated password (empty when the login is off or headless) | Operator sign-in; rotating the credential |
| `port_forward_command` | One line to reach the web UI from a workstation | Local access when no exposure is composed |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Interactive tuning swarm** -- the untouched defaults: inline scripts, the secured web UI, a small fixed fleet; drive runs from the browser and download reports before teardown. Start from the **Web load test** preset.

**CI performance gate** -- `headless: true` with run shape in `environment`; the pipeline applies the manifest, watches the master logs, and tears the namespace down. Start from the **Headless CI gate** preset.

**Scale-to-zero standing swarm** -- the KEDA arm with `minReplicas: 0` and custom cron triggers: the fleet exists only during scheduled runs.

**Isolated generator pool** -- worker `scheduling` onto a dedicated node pool so the load generator and the system under test never fight for CPU.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the placement unit this kind installs into
- [**KEDA**](/infra-catalog/kubernetes-keda) -- the operator the KEDA autoscaling arm requires
- [**Kubernetes NetworkPolicy**](/infra-catalog/kubernetes-network-policy) -- fence who can reach the master Service (defense in depth beside the login)
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) -- compose real exposure over `master_service` when engineers need the UI beyond the cluster
- [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) -- the Ingress-based alternative for exposing the web UI
