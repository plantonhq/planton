# KubernetesPerconaMongoOperator Pulumi Module Examples

This document provides examples of deploying the Percona Operator for MongoDB using the Pulumi module.

---

## Example 1: Deploy with Default Resources

### Description

Deploy the Percona Operator for MongoDB with default resource allocations using the Pulumi CLI.

### Prerequisites

- Pulumi CLI installed
- Kubernetes cluster credentials configured in Planton
- `kubectl` configured to access your cluster

### Manifest

Create a file named `percona-operator-default.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPerconaMongoOperator
metadata:
  name: percona-operator-default
spec:
  targetCluster:
    credentialId: my-k8s-cluster
  namespace: percona-operator  # Optional: defaults to "percona-operator"
  createNamespace: true
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
planton pulumi up --manifest percona-operator-default.yaml
```

### Verify

```bash
# Check operator pod
kubectl get pods -n percona-operator

# Check operator logs
kubectl logs -n percona-operator -l app.kubernetes.io/name=kubernetes-percona-mongo-operator -f

# Verify CRDs
kubectl get crds | grep percona
```

---

## Example 2: Deploy with Custom Resources for Large Clusters

### Description

Deploy the operator with increased resource allocations suitable for managing many MongoDB clusters.

### Manifest

Create a file named `percona-operator-large.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPerconaMongoOperator
metadata:
  name: percona-operator-large
spec:
  targetCluster:
    credentialId: production-k8s-cluster
  namespace: percona-operator-prod  # Custom namespace for production
  createNamespace: true
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
planton pulumi up --manifest percona-operator-large.yaml
```

---

## Example 3: Deploy to Development Cluster

### Description

Deploy with minimal resources for development or testing environments.

### Manifest

Create a file named `percona-operator-dev.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPerconaMongoOperator
metadata:
  name: percona-operator-dev
spec:
  targetCluster:
    credentialId: dev-k8s-cluster
  namespace: percona-operator-dev
  createNamespace: true
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
planton pulumi up --manifest percona-operator-dev.yaml
```

---

## Example 4: Deploy to Existing Namespace

### Description

Deploy the operator to an existing namespace managed separately (e.g., via GitOps or separate Terraform).

### Prerequisites

Ensure the namespace exists:

```bash
kubectl create namespace shared-operators
```

### Manifest

Create a file named `percona-operator-existing-ns.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPerconaMongoOperator
metadata:
  name: percona-operator-existing
spec:
  targetCluster:
    credentialId: my-k8s-cluster
  namespace: shared-operators
  createNamespace: false  # Use existing namespace
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
planton pulumi up --manifest percona-operator-existing-ns.yaml
```

**Important**: The namespace `shared-operators` must exist before running this command.

---

## Example 5: Update Operator Resources

### Description

Update an existing operator deployment with new resource allocations.

### Steps

1. Modify your manifest with new resource values
2. Run the update command:

```bash
planton pulumi up --manifest percona-operator-updated.yaml
```

The Pulumi module will perform an in-place update of the Helm release.

---

## Example 6: Destroy Operator Deployment

### Description

Remove the operator from your cluster.

### Command

```bash
planton pulumi destroy --manifest percona-operator.yaml
```

This will:
1. Remove the Helm release
2. Delete the operator namespace
3. Remove all associated resources

**Note**: This does not automatically remove CRDs or any MongoDB clusters deployed by the operator.

---

## Advanced Usage

### Using Custom Pulumi Module

If you need to customize the deployment beyond the standard configuration, you can reference a custom Pulumi module:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPerconaMongoOperator
metadata:
  name: percona-operator-custom
spec:
  targetCluster:
    credentialId: my-k8s-cluster
  createNamespace: true
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
#   repo: https://github.com/myorg/custom-percona-module
#   path: /
#   tag: v1.0.0
```

### Debugging Deployment Issues

Enable debug logging during deployment:

```bash
PULUMI_DEBUG_COMMANDS=true planton pulumi up --manifest percona-operator.yaml
```

Check Pulumi preview before applying:

```bash
planton pulumi preview --manifest percona-operator.yaml
```

### Local Development with Debug Script

For local development and testing:

```bash
cd apis/project/planton/provider/kubernetes/kubernetesperconamongooperator/v1/iac/pulumi

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
   kubectl get pods -n percona-operator
   kubectl describe pod -n percona-operator <operator-pod-name>
   ```

2. **Check Operator Logs**:
   ```bash
   kubectl logs -n percona-operator -l app.kubernetes.io/name=kubernetes-percona-mongo-operator -f
   ```

3. **Verify CRD Installation**:
   ```bash
   kubectl get crds
   # Look for:
   # - perconaservermongodbs.psmdb.percona.com
   # - perconaservermongodbbackups.psmdb.percona.com
   # - perconaservermongodbrestores.psmdb.percona.com
   ```

4. **Deploy a Test MongoDB Cluster**:
   ```bash
   # Use the MongoDBKubernetes workload resource
   planton pulumi up --manifest mongodb-test-cluster.yaml
   ```

## Common Issues and Solutions

### Issue: Operator Pod CrashLoopBackOff

**Solution**: Check resource limits and increase if necessary.

### Issue: CRDs Not Created

**Solution**: The Percona operator Helm chart automatically installs CRDs. Check Helm release status.

### Issue: Helm Release Failed

**Solution**: Check Helm release status:
```bash
kubectl get helmreleases -n percona-operator
helm list -n percona-operator
```

## Next Steps

After deploying the operator, refer to the MongoDBKubernetes workload documentation to deploy actual MongoDB clusters.

