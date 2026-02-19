# OCI Container Instance Examples

This document provides practical examples for deploying Oracle Cloud Infrastructure Container Instances using the OpenMCF API. Each example demonstrates a different use case, progressing from a minimal single-container instance to a fully configured production instance with security hardening, health checks, private registry authentication, and config file injection.

## Table of Contents

- [Example 1: Minimal Single-Container Instance](#example-1-minimal-single-container-instance)
- [Example 2: Multi-Container Sidecar with Shared Volume](#example-2-multi-container-sidecar-with-shared-volume)
- [Example 3: Health Checks and Resource Limits](#example-3-health-checks-and-resource-limits)
- [Example 4: Private Registry with OCI Vault Credentials](#example-4-private-registry-with-oci-vault-credentials)
- [Example 5: Config File Injection with Custom DNS](#example-5-config-file-injection-with-custom-dns)
- [Example 6: Full-Featured Production Instance](#example-6-full-featured-production-instance)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Minimal Single-Container Instance

**Use Case:** A single nginx container for development or testing — the simplest path to running a container on OCI without managing infrastructure.

**Configuration:**
- **Shape:** CI.Standard.E4.Flex (1 OCPU, default memory)
- **Containers:** 1 (nginx)
- **Restart policy:** Default (always)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerInstance
metadata:
  name: web-server
  org: my-org
  env: dev
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciContainerInstance.web-server
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "CI.Standard.E4.Flex"
  shapeConfig:
    ocpus: 1
  containers:
    - imageUrl: "docker.io/library/nginx:latest"
  vnics:
    - subnetId:
        value: "ocid1.subnet.oc1.iad.example"
```

**Deploy with OpenMCF CLI:**

```shell
openmcf apply -f web-server.yaml
```

**What happens:**
- A container instance is created with a single nginx container on a 1-OCPU E4 Flex shape.
- OCI assigns memory based on the OCPU count (minimum 1 GB per OCPU). To set memory explicitly, add `memoryInGbs` to `shapeConfig`.
- The container uses the default restart policy (ALWAYS) — if the container exits, OCI restarts it automatically.
- DNS settings are inherited from the subnet's DHCP options.
- The container instance ID and container IDs are exported as stack outputs.
- Resource principal (v2.2) is enabled by default, allowing the container to authenticate to OCI APIs.

---

## Example 2: Multi-Container Sidecar with Shared Volume

**Use Case:** A web application container with a Fluent Bit sidecar collecting application logs via a shared emptydir volume. This demonstrates the pod-like multi-container pattern where containers share networking and storage.

**Configuration:**
- **Shape:** CI.Standard.E4.Flex (2 OCPUs, 8 GB)
- **Containers:** 2 (web app + log collector)
- **Volumes:** 1 emptydir (disk-backed)
- **Resource limits:** Per-container memory and CPU limits

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerInstance
metadata:
  name: web-with-logging
  org: acme
  env: staging
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: staging.OciContainerInstance.web-with-logging
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "CI.Standard.E4.Flex"
  shapeConfig:
    ocpus: 2
    memoryInGbs: 8
  containerRestartPolicy: always
  containers:
    - imageUrl: "ghcr.io/acme/web-app:v2.1"
      displayName: "web-app"
      resourceConfig:
        memoryLimitInGbs: 6
        vcpusLimit: 3
      environmentVariables:
        LOG_DIR: "/var/log/app"
        PORT: "8080"
      volumeMounts:
        - mountPath: "/var/log/app"
          volumeName: "shared-logs"
    - imageUrl: "docker.io/fluent/fluent-bit:latest"
      displayName: "log-collector"
      resourceConfig:
        memoryLimitInGbs: 1
        vcpusLimit: 0.5
      volumeMounts:
        - mountPath: "/var/log/app"
          volumeName: "shared-logs"
          isReadOnly: true
  vnics:
    - subnetId:
        value: "ocid1.subnet.oc1.iad.example"
  volumes:
    - name: "shared-logs"
      volumeType: emptydir
      backingStore: "EPHEMERAL_STORAGE"
```

**What happens:**
- Two containers run in the same network namespace. The web app writes logs to `/var/log/app`, and the Fluent Bit sidecar reads them from the same directory via a shared emptydir volume.
- The web app gets 6 GB memory / 3 vCPUs, and the log collector gets 1 GB / 0.5 vCPUs. The remaining 1 GB / 0.5 vCPUs are available as a buffer. (1 OCPU = 2 vCPUs, so 2 OCPUs = 4 vCPUs total.)
- The log collector mounts the shared volume read-only (`isReadOnly: true`) to prevent accidental writes.
- Both containers can communicate over `localhost` — the log collector could also receive logs via HTTP on a localhost port.
- The emptydir volume uses `EPHEMERAL_STORAGE` (disk-backed). Use `MEMORY` for tmpfs-backed volumes when you need faster I/O at the cost of consuming instance memory.

---

## Example 3: Health Checks and Resource Limits

**Use Case:** An API service with HTTP readiness and TCP liveness health checks, demonstrating how OCI monitors container health and restarts unhealthy containers.

**Configuration:**
- **Shape:** CI.Standard.E4.Flex (2 OCPUs, 8 GB)
- **Containers:** 1 (API service)
- **Health checks:** HTTP readiness + TCP liveness

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerInstance
metadata:
  name: api-service
  org: acme
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: prod.OciContainerInstance.api-service
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "CI.Standard.E4.Flex"
  shapeConfig:
    ocpus: 2
    memoryInGbs: 8
  gracefulShutdownTimeoutInSeconds: 30
  containers:
    - imageUrl: "ghcr.io/acme/api-service:v3.0"
      displayName: "api"
      resourceConfig:
        memoryLimitInGbs: 7
        vcpusLimit: 3.5
      healthChecks:
        - healthCheckType: http
          port: 8080
          path: "/healthz"
          name: "readiness"
          initialDelayInSeconds: 15
          intervalInSeconds: 10
          timeoutInSeconds: 5
          failureThreshold: 3
          successThreshold: 1
          failureAction: kill
        - healthCheckType: tcp
          port: 8080
          name: "liveness"
          intervalInSeconds: 30
          failureThreshold: 5
          failureAction: kill
  vnics:
    - subnetId:
        value: "ocid1.subnet.oc1.iad.example"
```

**What happens:**
- The HTTP readiness check probes `/healthz` on port 8080 every 10 seconds, starting 15 seconds after the container starts. If the check fails 3 times consecutively, OCI kills the container.
- The TCP liveness check verifies port 8080 is accepting connections every 30 seconds. If the check fails 5 times consecutively, OCI kills the container.
- With `containerRestartPolicy` defaulting to `always`, OCI restarts the container after a health-check-triggered kill.
- The 30-second graceful shutdown timeout gives the application time to drain in-flight requests and close database connections before OCI sends SIGKILL.

**Health check design notes:**
- Use a short `initialDelayInSeconds` (10-30s) for applications with fast startup, longer (60-120s) for JVM or framework-heavy applications.
- Set `failureThreshold` high enough to avoid false positives from transient issues (3-5 is typical).
- The `failureAction: none` option logs the failure without killing the container — useful for monitoring-only health checks during initial rollout.

---

## Example 4: Private Registry with OCI Vault Credentials

**Use Case:** Pulling a container image from OCI Container Registry (OCIR) using OCI Vault-based credentials, avoiding plaintext credentials in manifests. Also demonstrates basic auth for a secondary registry.

**Configuration:**
- **Image pull secrets:** 2 (Vault-based for OCIR, basic auth for Docker Hub)
- **Registries:** OCIR + Docker Hub

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerInstance
metadata:
  name: private-app
  org: acme
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: prod.OciContainerInstance.private-app
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "CI.Standard.E4.Flex"
  shapeConfig:
    ocpus: 1
    memoryInGbs: 4
  containers:
    - imageUrl: "us-ashburn-1.ocir.io/acme-tenancy/backend-service:v1.5"
      displayName: "backend"
    - imageUrl: "docker.io/acme-private/metrics-exporter:latest"
      displayName: "metrics"
      isResourcePrincipalDisabled: true
  vnics:
    - subnetId:
        value: "ocid1.subnet.oc1.iad.example"
  imagePullSecrets:
    - registryEndpoint: "us-ashburn-1.ocir.io"
      secretType: vault
      secretId:
        value: "ocid1.vaultsecret.oc1.iad.example"
    - registryEndpoint: "docker.io"
      secretType: basic
      username: "YWNtZS1wcml2YXRl"
      password: "c2VjcmV0LXRva2VuLTEyMw=="
```

**What happens:**
- The backend container image is pulled from OCIR using credentials stored in an OCI Vault secret. The secret OCID is referenced directly — the IaC module passes it to the container instance API, which retrieves the credentials at pull time.
- The metrics exporter image is pulled from a private Docker Hub repository using base64-encoded username and password.
- Resource principal is explicitly disabled for the metrics exporter container (`isResourcePrincipalDisabled: true`) since it does not need access to OCI APIs.
- Each `imagePullSecrets` entry applies to its `registryEndpoint`. OCI matches the image URL to the appropriate secret based on the registry hostname.

**Vault secret format:**
The OCI Vault secret for OCIR authentication should contain the registry credentials in the format expected by the OCI Container Instances API. Create the secret using the OCI Console or CLI:

```shell
oci vault secret create-base64 \
  --compartment-id "ocid1.compartment.oc1..example" \
  --vault-id "ocid1.vault.oc1.iad.example" \
  --key-id "ocid1.key.oc1.iad.example" \
  --secret-name "ocir-credentials" \
  --secret-content-content "$(echo -n '{"username":"acme-tenancy/user@acme.com","password":"auth-token"}' | base64)"
```

---

## Example 5: Config File Injection with Custom DNS

**Use Case:** Injecting application configuration files into a container via a configfile volume, with custom DNS settings for internal service resolution. Demonstrates how to pass structured configuration without building it into the container image.

**Configuration:**
- **Volumes:** 1 configfile (with 2 files), 1 emptydir (tmpfs for temp data)
- **DNS:** Custom nameservers and search domains

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerInstance
metadata:
  name: config-driven-app
  org: acme
  env: staging
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: staging.OciContainerInstance.config-driven-app
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:PHX-AD-1"
  shape: "CI.Standard.E4.Flex"
  shapeConfig:
    ocpus: 1
    memoryInGbs: 4
  containers:
    - imageUrl: "ghcr.io/acme/data-processor:v2.0"
      displayName: "processor"
      command:
        - "/app/processor"
      arguments:
        - "--config=/etc/app/app.yaml"
        - "--logging=/etc/app/logging.yaml"
      workingDirectory: "/app"
      volumeMounts:
        - mountPath: "/etc/app"
          volumeName: "app-config"
          isReadOnly: true
        - mountPath: "/tmp"
          volumeName: "tmp-dir"
  vnics:
    - subnetId:
        value: "ocid1.subnet.oc1.iad.example"
  dnsConfig:
    nameservers:
      - "10.0.0.53"
    searches:
      - "internal.acme.com"
      - "svc.acme.com"
    options:
      - "ndots:5"
  volumes:
    - name: "app-config"
      volumeType: configfile
      configs:
        - fileName: "app.yaml"
          data: "ZGF0YWJhc2U6CiAgaG9zdDogZGIuaW50ZXJuYWwuYWNtZS5jb20KICBwb3J0OiA1NDMyCiAgbmFtZTogYXBwZGIKY2FjaGU6CiAgaG9zdDogcmVkaXMuaW50ZXJuYWwuYWNtZS5jb20KICBwb3J0OiA2Mzc5"
        - fileName: "logging.yaml"
          data: "bGV2ZWw6IGluZm8Kb3V0cHV0OiBzdGRvdXQKZm9ybWF0OiBqc29u"
    - name: "tmp-dir"
      volumeType: emptydir
      backingStore: "MEMORY"
```

**What happens:**
- Two configuration files (`app.yaml` and `logging.yaml`) are injected into the container at `/etc/app/` via a configfile volume. The `data` field contains base64-encoded YAML content.
- The configfile volume is mounted read-only — the application cannot modify its own configuration files at runtime.
- A tmpfs-backed emptydir volume is mounted at `/tmp` for temporary data. The `MEMORY` backing store provides faster I/O than disk but consumes instance memory.
- Custom DNS settings override the subnet's DHCP options: a private DNS server at `10.0.0.53` resolves `internal.acme.com` and `svc.acme.com` domains. The `ndots:5` option ensures short names are searched with the configured domains before falling back to absolute DNS lookups.

**Base64-encoding configuration:**

```shell
cat app.yaml | base64
cat logging.yaml | base64
```

---

## Example 6: Full-Featured Production Instance

**Use Case:** A production API service combining all features — security hardening, health checks, private registry, config file injection, multi-container composition, custom DNS, and infrastructure references via `valueFrom`. This represents the full configuration surface for a hardened production deployment.

**Configuration:**
- **Shape:** CI.Standard.E4.Flex (4 OCPUs, 16 GB)
- **Containers:** 2 (API service + Envoy proxy sidecar)
- **Health checks:** HTTP readiness + TCP liveness
- **Security:** Non-root, read-only rootfs, dropped capabilities
- **Volumes:** 2 (configfile + tmpfs emptydir)
- **Registry:** OCIR with Vault credentials
- **DNS:** Custom internal resolution

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerInstance
metadata:
  name: prod-api
  org: acme
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: prod.OciContainerInstance.prod-api
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Uocm:PHX-AD-1"
  faultDomain: "FAULT-DOMAIN-1"
  shape: "CI.Standard.E4.Flex"
  shapeConfig:
    ocpus: 4
    memoryInGbs: 16
  containerRestartPolicy: always
  gracefulShutdownTimeoutInSeconds: 45
  containers:
    - imageUrl: "us-ashburn-1.ocir.io/acme-tenancy/api-service:v3.2"
      displayName: "api"
      command:
        - "/app/server"
      arguments:
        - "--config=/etc/app/config.json"
        - "--port=8080"
      environmentVariables:
        APP_ENV: "production"
        LOG_LEVEL: "info"
        OTEL_EXPORTER_ENDPOINT: "http://localhost:9411"
      resourceConfig:
        memoryLimitInGbs: 12
        vcpusLimit: 6
      securityContext:
        isNonRootUserCheckEnabled: true
        isRootFileSystemReadonly: true
        runAsUser: 1000
        runAsGroup: 1000
        capabilities:
          dropCapabilities:
            - "ALL"
          addCapabilities:
            - "NET_BIND_SERVICE"
      healthChecks:
        - healthCheckType: http
          port: 8080
          path: "/healthz"
          name: "readiness"
          initialDelayInSeconds: 20
          intervalInSeconds: 10
          timeoutInSeconds: 5
          failureThreshold: 3
          successThreshold: 1
        - healthCheckType: tcp
          port: 8080
          name: "liveness"
          intervalInSeconds: 30
          failureThreshold: 5
          failureAction: kill
      volumeMounts:
        - mountPath: "/etc/app"
          volumeName: "app-config"
          isReadOnly: true
        - mountPath: "/tmp"
          volumeName: "tmp-dir"
    - imageUrl: "docker.io/envoyproxy/envoy:v1.28-latest"
      displayName: "envoy-proxy"
      resourceConfig:
        memoryLimitInGbs: 2
        vcpusLimit: 1
      securityContext:
        isNonRootUserCheckEnabled: true
        runAsUser: 101
        runAsGroup: 101
        capabilities:
          dropCapabilities:
            - "ALL"
          addCapabilities:
            - "NET_BIND_SERVICE"
      healthChecks:
        - healthCheckType: http
          port: 9901
          path: "/ready"
          name: "envoy-readiness"
          initialDelayInSeconds: 5
          intervalInSeconds: 15
          failureThreshold: 3
      volumeMounts:
        - mountPath: "/etc/envoy"
          volumeName: "app-config"
          isReadOnly: true
          subPath: "envoy"
  vnics:
    - subnetId:
        valueFrom:
          kind: OciSubnet
          name: private-subnet
          fieldPath: status.outputs.subnetId
      displayName: "primary"
      isPublicIpAssigned: false
      hostnameLabel: "prod-api"
      nsgIds:
        - valueFrom:
            kind: OciNetworkSecurityGroup
            name: api-nsg
            fieldPath: status.outputs.networkSecurityGroupId
  imagePullSecrets:
    - registryEndpoint: "us-ashburn-1.ocir.io"
      secretType: vault
      secretId:
        value: "ocid1.vaultsecret.oc1.iad.example"
  dnsConfig:
    nameservers:
      - "10.0.0.53"
    searches:
      - "internal.acme.com"
      - "svc.acme.com"
    options:
      - "ndots:5"
  volumes:
    - name: "app-config"
      volumeType: configfile
      configs:
        - fileName: "config.json"
          data: "eyJkYXRhYmFzZSI6eyJob3N0IjoiZGIuaW50ZXJuYWwuYWNtZS5jb20ifX0="
        - fileName: "envoy/envoy.yaml"
          path: "envoy"
          data: "c3RhdGljX3Jlc291cmNlczoKICBsaXN0ZW5lcnM6IFtdCiAgY2x1c3RlcnM6IFtd"
    - name: "tmp-dir"
      volumeType: emptydir
      backingStore: "MEMORY"
```

**What happens:**
- Two containers run in the same network namespace: an API service on port 8080 and an Envoy proxy on port 9901. The API service sends traces to Envoy's local endpoint via `localhost`.
- Both containers are hardened: non-root user enforcement, all capabilities dropped except `NET_BIND_SERVICE`, and the API service has a read-only root filesystem.
- Container-level resource limits subdivide the 4-OCPU / 16-GB instance: 12 GB / 6 vCPUs for the API and 2 GB / 1 vCPU for Envoy, leaving 2 GB / 1 vCPU as headroom.
- Health checks monitor both containers independently. If either container becomes unhealthy, OCI kills and restarts it.
- The configfile volume injects application config and Envoy config. Envoy uses the `subPath` mount to access only the `envoy/` subdirectory of the volume.
- Infrastructure references use `valueFrom` to compose with OciCompartment, OciSubnet, and OciNetworkSecurityGroup — enabling infra-chart patterns.
- The 45-second graceful shutdown timeout allows both containers to drain connections before forced termination.

---

## Common Operations

### Get Container Instance Status

After deploying a container instance, check the instance ID and container IDs from stack outputs:

```shell
# Pulumi
pulumi stack output container_instance_id
pulumi stack output container_ids

# Terraform
terraform output container_instance_id
terraform output container_ids
```

### View Container Instance Details

Use the OCI CLI to inspect the container instance:

```shell
INSTANCE_ID=$(pulumi stack output container_instance_id)

oci container-instances container-instance get \
  --container-instance-id "$INSTANCE_ID" \
  --query 'data.{state:"lifecycle-state",shape:"shape",containers:"containers[*].{name:display-name,state:lifecycle-state}"}' \
  --output table
```

### View Container Logs

Retrieve logs from a specific container within the instance:

```shell
CONTAINER_ID="ocid1.container.oc1.iad.example"

oci container-instances container get \
  --container-id "$CONTAINER_ID"

# Stream logs (requires logging configured on the compartment)
oci logging search \
  --search-query "search \"$COMPARTMENT_ID\" | where source = '$CONTAINER_ID'" \
  --time-start "2026-01-01T00:00:00Z"
```

### Restart the Container Instance

Stop and start the instance to restart all containers:

```shell
INSTANCE_ID=$(pulumi stack output container_instance_id)

oci container-instances container-instance restart \
  --container-instance-id "$INSTANCE_ID"
```

### Use Outputs in Downstream Resources

The `container_instance_id` output can be referenced by downstream resources:

```yaml
spec:
  containerInstanceId:
    valueFrom:
      kind: OciContainerInstance
      name: prod-api
      fieldPath: status.outputs.containerInstanceId
```

### Update Container Configuration

To change containers, environment variables, or volumes, update the manifest and re-apply:

```shell
openmcf apply -f prod-api.yaml
```

Container instance updates that change containers, shape, or volumes require instance recreation. OCI Container Instances do not support in-place updates for these fields.

---

## Best Practices

### Shape Selection

| Workload Type | Recommended Shape | OCPU / Memory | Rationale |
|---------------|------------------|---------------|-----------|
| Lightweight service | `CI.Standard.E4.Flex` | 1 / 2-4 GB | Minimum viable allocation for a single container. |
| Multi-container sidecar | `CI.Standard.E4.Flex` | 2 / 8-16 GB | Room for primary container + 1-2 sidecars with resource limits. |
| Memory-intensive processing | `CI.Standard.E4.Flex` | 2 / 32-64 GB | OCI allows up to 64 GB per OCPU on flex shapes. |
| Cost-optimized batch | `CI.Standard.E3.Flex` | 1 / 4 GB | E3 Flex shapes are typically cheaper than E4. |

**Always set `memoryInGbs` explicitly for production.** When omitted, OCI assigns a minimum based on OCPU count. The default may not be sufficient for memory-intensive workloads.

### Resource Limits

- **Always set `resourceConfig` in multi-container instances.** Without per-container limits, a single container can consume all instance resources, starving other containers.
- **Leave headroom.** Allocate 10-20% less than the instance total across all containers. OCI system processes consume a small amount of resources, and headroom prevents OOM situations during traffic spikes.
- **1 OCPU = 2 vCPUs.** The `vcpusLimit` field uses logical CPUs. A 2-OCPU instance has 4 vCPUs available. Set `vcpusLimit` accordingly.

### Security Hardening

For production containers:

| Setting | Recommended Value | Rationale |
|---------|------------------|-----------|
| `isNonRootUserCheckEnabled` | `true` | Prevents containers from running as root, reducing privilege escalation risk. |
| `isRootFileSystemReadonly` | `true` | Prevents runtime modifications to the container image. Mount writable volumes for temp data. |
| `capabilities.dropCapabilities` | `["ALL"]` | Drop all Linux capabilities by default. |
| `capabilities.addCapabilities` | Only what's needed | Add back only the capabilities the application requires (e.g., `NET_BIND_SERVICE` for port < 1024). |
| `runAsUser` / `runAsGroup` | Non-zero UID/GID | Explicit UID/GID avoids relying on the image's default user. |

**When using `isRootFileSystemReadonly: true`,** mount a tmpfs emptydir volume at `/tmp` (and any other paths where the application writes temporary data). Without a writable path, applications that write to the filesystem will fail.

### Volume Types

| Volume Type | Backing Store | Use Case | Notes |
|-------------|--------------|----------|-------|
| `emptydir` | `EPHEMERAL_STORAGE` | Log sharing between containers, temporary data | Disk-backed, survives container restarts, lost on instance delete |
| `emptydir` | `MEMORY` | High-speed scratch space, temp files | tmpfs-backed, consumes instance memory, faster I/O |
| `configfile` | — | Application config, certificates, static files | Base64-encoded file injection, always read-only |

- **Prefer `EPHEMERAL_STORAGE` for log volumes.** Disk-backed volumes do not consume instance memory and can handle larger volumes of log data.
- **Use `MEMORY` for `/tmp` mounts.** When the application writes small amounts of temporary data, tmpfs provides faster I/O.
- **Configfile volumes are immutable at runtime.** To update configuration, change the `data` field in the manifest and re-deploy. The container instance is recreated with the new configuration.
- **Volume limit is 32.** Plan volume allocation across all containers in the instance. Each named volume counts once regardless of how many containers mount it.

### Restart Policies

| Policy | Behavior | Use Case |
|--------|----------|----------|
| `always` (default) | Restart containers regardless of exit code | Long-running services, web servers, APIs |
| `on_failure` | Restart only on non-zero exit code | Batch jobs that should retry on failure but not on success |
| `never` | Never restart containers | One-shot tasks, data migrations, test runs |

The restart policy applies to all containers in the instance. If you need different restart behaviors, use separate container instances.

### Health Check Tuning

| Parameter | Dev/Test | Production | Notes |
|-----------|----------|------------|-------|
| `initialDelayInSeconds` | 5-10 | 15-60 | Longer for JVM, framework-heavy apps |
| `intervalInSeconds` | 30 | 10-15 | More frequent in production for faster detection |
| `timeoutInSeconds` | 10 | 3-5 | Tight timeouts in production catch hung processes |
| `failureThreshold` | 1-2 | 3-5 | Higher in production to avoid false positives |
| `failureAction` | `none` | `kill` | Use `none` during initial rollout for monitoring without disruption |

- **Use HTTP health checks for application-level health.** HTTP checks verify the application can process requests, not just that the process is alive.
- **Use TCP health checks for port-level liveness.** TCP checks verify the port is accepting connections — useful as a fallback when the application does not expose an HTTP health endpoint.
- **Start with `failureAction: none` during initial deployment.** This logs health check failures without killing the container, letting you tune thresholds before enabling automatic restarts.

### Multi-Container Patterns

| Pattern | Primary Container | Sidecar Container | Shared Resource |
|---------|------------------|-------------------|-----------------|
| **Log collection** | Application (writes logs to volume) | Fluent Bit / Fluentd (reads logs) | Emptydir volume |
| **Metrics export** | Application (exposes metrics on localhost) | Prometheus exporter (scrapes localhost) | Localhost networking |
| **Reverse proxy** | Application (listens on internal port) | Nginx / Envoy (forwards to localhost) | Localhost networking |
| **Config reload** | Application (reads config from volume) | Config watcher (updates config files) | Configfile volume |

- **Set resource limits on every container.** Without limits, the sidecar can starve the primary container or vice versa.
- **Mount shared volumes read-only where possible.** The consumer container (e.g., log collector reading logs) should mount the volume read-only.
- **Use `displayName` on every container.** Named containers are easier to identify in OCI Console, CLI output, and logs.
