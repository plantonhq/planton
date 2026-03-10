# KubernetesPerconaMysqlOperator Pulumi Module Examples

This document provides examples of deploying the Percona Operator for MySQL using the Pulumi module.

---

## Example 1: Deploy with Default Resources

### Description

Deploy the Percona Operator for MySQL with default resource allocations using the Pulumi CLI.

### Prerequisites

- Pulumi CLI installed
- Kubernetes cluster credentials configured in Planton
- `kubectl` configured to access your cluster

### Manifest

Create a file named `percona-mysql-operator-default.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPerconaMysqlOperator
metadata:
  name: percona-mysql-operator-default
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: percona-mysql-operator
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 1000m
        memory: 1Gi
```

### Deploy

```bash
planton pulumi up --manifest percona-mysql-operator-default.yaml
```

### Verify

```bash
# Check operator pod
kubectl get pods -n percona-mysql-operator

# Check operator logs
kubectl logs -n percona-mysql-operator -l app.kubernetes.io/name=kubernetes-percona-mysql-operator -f

# Verify CRDs
kubectl get crds | grep percona
```

---

## Example 2: Deploy with Custom Resources for Large Clusters

### Description

Deploy the operator with increased resource allocations suitable for managing many MySQL clusters.

### Manifest

Create a file named `percona-mysql-operator-large.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPerconaMysqlOperator
metadata:
  name: percona-mysql-operator-large
spec:
  target_cluster:
    cluster_name: prod-gke-cluster
  namespace:
    value: percona-mysql-operator-prod
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 200m
        memory: 512Mi
      limits:
        cpu: 2000m
        memory: 2Gi
```

### Deploy

```bash
planton pulumi up --manifest percona-mysql-operator-large.yaml
```

---

## Example 3: Deploy to Development Cluster

### Description

Deploy with minimal resources for development or testing environments.

### Manifest

Create a file named `percona-mysql-operator-dev.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPerconaMysqlOperator
metadata:
  name: percona-mysql-operator-dev
spec:
  target_cluster:
    cluster_name: dev-gke-cluster
  namespace:
    value: percona-mysql-operator-dev
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 50m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
```

### Deploy

```bash
planton pulumi up --manifest percona-mysql-operator-dev.yaml
```

---

## Example 4: Deploy Using Existing Namespace

### Description

Deploy the operator into a pre-existing namespace. This is useful when namespaces are managed separately or have pre-configured policies.

### Prerequisites

Ensure the namespace exists before deployment:

```bash
# Create namespace if it doesn't exist
kubectl create namespace database-operators

# Verify namespace exists
kubectl get namespace database-operators
```

### Manifest

Create a file named `percona-mysql-operator-existing-ns.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPerconaMysqlOperator
metadata:
  name: percona-mysql-operator-existing-ns
spec:
  target_cluster:
    cluster_name: prod-gke-cluster
  namespace:
    value: database-operators
  create_namespace: false
  container:
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 1000m
        memory: 1Gi
```

### Deploy

```bash
planton pulumi up --manifest percona-mysql-operator-existing-ns.yaml
```

**Important:** With `create_namespace: false`, the namespace must exist before deployment. If the namespace doesn't exist, the deployment will fail.

---

## Example 5: Update Operator Resources

### Description

Update an existing operator deployment with new resource allocations.

### Steps

1. Modify your manifest with new resource values
2. Run the update command:

```bash
planton pulumi up --manifest percona-mysql-operator-updated.yaml
```

The Pulumi module will perform an in-place update of the Helm release.

---

## Example 6: Destroy Operator Deployment

### Description

Remove the operator from your cluster.

### Command

```bash
planton pulumi destroy --manifest percona-mysql-operator.yaml
```

This will:
1. Remove the Helm release
2. Delete the operator namespace
3. Remove all associated resources

**Note**: This does not automatically remove CRDs or any MySQL clusters deployed by the operator.

---

## Advanced Usage

### Using Custom Pulumi Module

If you need to customize the deployment beyond the standard configuration, you can reference a custom Pulumi module:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPerconaMysqlOperator
metadata:
  name: percona-mysql-operator-custom
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: percona-mysql-operator
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 1000m
        memory: 1Gi
# Custom module reference (if supported)
# pulumiModule:
#   repo: https://github.com/myorg/custom-percona-mysql-module
#   path: /
#   tag: v1.0.0
```

### Debugging Deployment Issues

Enable debug logging during deployment:

```bash
PULUMI_DEBUG_COMMANDS=true planton pulumi up --manifest percona-mysql-operator.yaml
```

Check Pulumi preview before applying:

```bash
planton pulumi preview --manifest percona-mysql-operator.yaml
```

### Local Development with Debug Script

For local development and testing:

```bash
cd apis/project/planton/provider/kubernetes/kubernetesperconamysqloperator/v1/iac/pulumi

# Set up environment
export PULUMI_STACK_INPUT=/path/to/manifest.yaml
export KUBERNETES_CREDENTIAL=/path/to/kubeconfig

# Run debug script
./debug.sh
```

---

## Post-Deployment Tasks

After successfully deploying the operator:

1. **Verify Operator Health**:
   ```bash
   kubectl get pods -n percona-mysql-operator
   kubectl describe pod -n percona-mysql-operator <operator-pod-name>
   ```

2. **Check Operator Logs**:
   ```bash
   kubectl logs -n percona-mysql-operator -l app.kubernetes.io/name=kubernetes-percona-mysql-operator -f
   ```

3. **Verify CRD Installation**:
   ```bash
   kubectl get crds
   # Look for:
   # - perconaservermysqls.ps.percona.com
   # - perconaservermysqlbackups.ps.percona.com
   # - perconaservermysqlrestores.ps.percona.com
   ```

4. **Deploy a Test MySQL Cluster**:
   ```bash
   # Use the MySQLKubernetes workload resource
   planton pulumi up --manifest mysql-test-cluster.yaml
   ```

## Common Issues and Solutions

### Issue: Operator Pod CrashLoopBackOff

**Solution**: Check resource limits and increase if necessary.

### Issue: CRDs Not Created

**Solution**: The Percona operator Helm chart automatically installs CRDs. Check Helm release status.

### Issue: Helm Release Failed

**Solution**: Check Helm release status:
```bash
kubectl get helmreleases -n percona-mysql-operator
helm list -n percona-mysql-operator
```

## Next Steps

After deploying the operator, refer to the MySQLKubernetes workload documentation to deploy actual MySQL clusters.

