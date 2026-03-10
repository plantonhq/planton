Here are a few examples for the `KubernetesArgocd` API resource, demonstrating how to configure and deploy Argo CD on a Kubernetes cluster using Planton's unified API structure.

---

# Example 1: Basic Argo CD Deployment with Namespace Creation

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesArgocd
metadata:
  name: argocd-instance
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: argocd
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 50m
        memory: 256Mi
      limits:
        cpu: 1
        memory: 1Gi
```

---

# Example 2: Argo CD with Ingress Enabled

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesArgocd
metadata:
  name: argocd-prod
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: argocd-prod
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 100m
        memory: 512Mi
      limits:
        cpu: 2
        memory: 2Gi
  ingress:
    enabled: true
    hostname: argocd.example.com
```

---

# Example 3: Argo CD Deployment with Custom Resources

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesArgocd
metadata:
  name: argocd-custom
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: argocd-custom
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 200m
        memory: 1Gi
      limits:
        cpu: 3
        memory: 4Gi
```

---

# Example 4: Production Argo CD with High Availability

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesArgocd
metadata:
  name: argocd-ha
  labels:
    env: production
    team: platform
spec:
  target_cluster:
    cluster_name: prod-gke-cluster
  namespace:
    value: argocd-system
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        cpu: 4
        memory: 8Gi
  ingress:
    enabled: true
    hostname: argocd.prod.example.com
```

---

# Example 5: Argo CD with Pre-existing Namespace

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesArgocd
metadata:
  name: argocd-existing-ns
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: platform-tools  # Must exist before deployment
  create_namespace: false  # Use existing namespace
  container:
    resources:
      requests:
        cpu: 100m
        memory: 512Mi
      limits:
        cpu: 2
        memory: 2Gi
  ingress:
    enabled: true
    hostname: argocd.platform.example.com
```

