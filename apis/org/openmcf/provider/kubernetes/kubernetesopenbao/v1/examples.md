# KubernetesOpenBao API-Resource Examples

Below are examples demonstrating how to configure and deploy the `KubernetesOpenBao` API resource. Follow the instructions to create and apply each YAML configuration using the Planton CLI.

---

## Example 1: Basic Standalone Deployment

### Description

This example demonstrates a basic standalone deployment of OpenBao within a Kubernetes cluster. It sets up a single OpenBao pod with default resource allocations and the UI enabled.

### Create and Apply

1. **Create a YAML file** using the example below.
2. **Apply the configuration** using the following command:

    ```shell
    planton apply -f <yaml-path>
    ```

### YAML Configuration

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesOpenBao
metadata:
  name: dev-openbao
spec:
  target_cluster:
    cluster_name: "my-gke-cluster"
  namespace:
    value: openbao
  create_namespace: true
  server_container:
    replicas: 1
    data_storage_size: "10Gi"
    resources:
      requests:
        cpu: "100m"
        memory: "128Mi"
      limits:
        cpu: "500m"
        memory: "256Mi"
  ui_enabled: true
```

---

## Example 2: High Availability Deployment

### Description

This example illustrates how to deploy OpenBao in high-availability mode using Raft integrated storage. It configures 3 replicas for fault tolerance and leader election.

### Create and Apply

1. **Create a YAML file** using the example below.
2. **Apply the configuration** using the following command:

    ```shell
    planton apply -f <yaml-path>
    ```

### YAML Configuration

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesOpenBao
metadata:
  name: prod-openbao
spec:
  target_cluster:
    cluster_name: "production-cluster"
  namespace:
    value: openbao-prod
  create_namespace: true
  server_container:
    replicas: 3
    data_storage_size: "50Gi"
    resources:
      requests:
        cpu: "250m"
        memory: "256Mi"
      limits:
        cpu: "1000m"
        memory: "512Mi"
  high_availability:
    enabled: true
    replicas: 3
  ui_enabled: true
```

---

## Example 3: OpenBao with Ingress and Agent Injector

### Description

This example shows a production-ready deployment with external ingress access and the Agent Injector enabled for automatic secret injection into application pods.

### Create and Apply

1. **Create a YAML file** using the example below.
2. **Apply the configuration** using the following command:

    ```shell
    planton apply -f <yaml-path>
    ```

### YAML Configuration

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesOpenBao
metadata:
  name: enterprise-openbao
spec:
  target_cluster:
    cluster_name: "production-cluster"
  namespace:
    value: openbao-enterprise
  create_namespace: true
  server_container:
    replicas: 5
    data_storage_size: "100Gi"
    resources:
      requests:
        cpu: "500m"
        memory: "512Mi"
      limits:
        cpu: "2000m"
        memory: "1Gi"
  high_availability:
    enabled: true
    replicas: 5
  ingress:
    enabled: true
    hostname: "secrets.company.com"
    ingress_class_name: "nginx"
    tls_enabled: true
    tls_secret_name: "openbao-tls"
  injector:
    enabled: true
    replicas: 2
  ui_enabled: true
  tls_enabled: true
```

---

## Example 4: Minimal Development Setup

### Description

This example provides a minimal configuration suitable for local development using minikube or kind clusters.

### Create and Apply

1. **Create a YAML file** using the example below.
2. **Apply the configuration** using the following command:

    ```shell
    planton apply -f <yaml-path>
    ```

### YAML Configuration

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesOpenBao
metadata:
  name: local-openbao
spec:
  namespace:
    value: openbao-dev
  create_namespace: true
  server_container:
    replicas: 1
    data_storage_size: "1Gi"
    resources:
      requests:
        cpu: "50m"
        memory: "64Mi"
      limits:
        cpu: "200m"
        memory: "128Mi"
  ui_enabled: true
```

---

## Example 5: HA Deployment with GCP KMS Auto-Unseal

### Description

This example shows a production HA deployment with GCP Cloud KMS auto-unseal and GKE Workload Identity. Pods automatically unseal on restart without manual intervention.

### Create and Apply

1. **Create a YAML file** using the example below.
2. **Apply the configuration** using the following command:

    ```shell
    planton apply -f <yaml-path>
    ```

### YAML Configuration

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesOpenBao
metadata:
  name: prod-openbao
spec:
  target_cluster:
    cluster_name: "production-gke-cluster"
  namespace:
    value: openbao-prod
  create_namespace: true
  server_container:
    replicas: 3
    data_storage_size: "50Gi"
    resources:
      requests:
        cpu: "250m"
        memory: "256Mi"
      limits:
        cpu: "1000m"
        memory: "512Mi"
  high_availability:
    enabled: true
    replicas: 3
  auto_unseal:
    gcp_kms:
      project:
        value: my-gcp-project
      region: asia-south1
      key_ring:
        value: openbao-unseal
      crypto_key:
        value: openbao-auto-unseal-key
      workload_identity_service_account:
        value: openbao@my-gcp-project.iam.gserviceaccount.com
  ui_enabled: true
```

---

## Post-Deployment Steps

After deploying OpenBao, you'll need to:

1. **Initialize OpenBao**: Run `bao operator init` to generate the root token (and unseal keys if using Shamir)
2. **Unseal OpenBao** (Shamir only): Use the unseal keys to unseal. If `auto_unseal` is configured, this step is handled automatically by the KMS provider on every pod startup.
3. **Configure Auth Methods**: Set up Kubernetes auth for pod authentication
4. **Create Policies**: Define access policies for secrets
5. **Store Secrets**: Begin storing and managing secrets

For HA deployments, ensure you initialize only one node and join the others to the cluster.
