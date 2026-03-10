
# Example 1: Basic Grafana Deployment

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGrafana
metadata:
  name: grafana-instance
spec:
  target_cluster:
    cluster_name: "my-gke-cluster"
  namespace:
    value: "grafana"
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 500m
        memory: 512Mi
```

---

# Example 2: Grafana with Ingress Enabled

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGrafana
metadata:
  name: grafana-prod
spec:
  target_cluster:
    cluster_name: "production-gke-cluster"
  namespace:
    value: "grafana"
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 200m
        memory: 512Mi
      limits:
        cpu: 1
        memory: 1Gi
  ingress:
    enabled: true
    hostname: grafana.example.com
```

---

# Example 3: Grafana Deployment with Environment Variables

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGrafana
metadata:
  name: grafana-with-env
spec:
  target_cluster:
    cluster_name: "dev-gke-cluster"
  namespace:
    value: "grafana"
  create_namespace: true
  container:
    env:
      variables:
        GF_SECURITY_ADMIN_PASSWORD: securepassword
        GF_SECURITY_ADMIN_USER: admin
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 1
        memory: 1Gi
```

---

# Example 4: Grafana with Environment Secrets

The below example assumes that secrets are managed by Planton's [GCP Secrets Manager](https://buf.build/openmcf/apis/docs/main:ai.planton.code2cloud.v1.gcp.gcpsecretsmanager) deployment module.

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGrafana
metadata:
  name: grafana-secure
spec:
  target_cluster:
    cluster_name: "production-gke-cluster"
  namespace:
    value: "grafana"
  create_namespace: true
  container:
    env:
      secrets:
        GF_SECURITY_ADMIN_PASSWORD: ${gcpsm-my-org-prod-gcp-secrets.admin-password}
      variables:
        GF_SECURITY_ADMIN_USER: admin
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 1
        memory: 1Gi
```

---

# Example 5: Minimal Grafana Deployment

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGrafana
metadata:
  name: minimal-grafana
spec:
  target_cluster:
    cluster_name: "dev-gke-cluster"
  namespace:
    value: "grafana"
  create_namespace: true
  container: {}
```
