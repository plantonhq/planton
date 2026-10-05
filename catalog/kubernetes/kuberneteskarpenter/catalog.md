# Karpenter

Installs the Karpenter node-provisioning controller from the official OCI-served Helm charts. Karpenter watches for unschedulable pods and launches RIGHT-SIZED machines for them in seconds — no pre-created node groups, no per-ASG scaling policies: it picks instance types from the live catalog to fit the pending pods, consolidates under-used nodes away, and handles spot interruptions. This kind installs the AWS provider — the canonical implementation. The engine alone provisions nothing: the fleet is declared with Karpenter EC2 Node Class (the machine template) and Karpenter Node Pool (the constraints). One installation per cluster.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **CRD Helm Release** (`karpenter-crd`) -- the NodePool, NodeClaim, and EC2NodeClass definitions as their own release — upstream's supported path for keeping CRDs upgradable — annotated to survive uninstall by default, so removing the release does not cascade-delete every fleet declaration in the cluster
- **Controller Helm Release** (`karpenter`) -- the controller Deployment (chart default 2 replicas: leader plus warm standby), RBAC, and the `karpenter` service account; the chart pins controller pods away from Karpenter-provisioned nodes so the controller never disrupts its own machine
- **Namespace** (optional) -- created with standard governance labels when `createNamespace` is true; a pre-existing `kube-system` is upstream's recommended home

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS / Cluster

- An EKS (or EKS-compatible) cluster with capacity Karpenter does NOT manage for the controller itself — a managed node group or Fargate.
- AWS-side identity for the controller: an IRSA role (trust policy against the `karpenter` service account) or an EKS Pod Identity association, carrying upstream's controller IAM policy.
- For interruption handling: the SQS queue and EventBridge rules from upstream's guidance — wire them before running spot at scale so nodes drain ahead of interruptions.
- With the Prometheus ServiceMonitor: the Prometheus operator CRDs — the release fails to install without them.

## Deploy

### Console

Open the deployment store, find **Karpenter**, and click **Deploy**. The creation wizard walks you through placement, the chart pin and CRD lifecycle, the cluster identity, AWS integration (IRSA, interruption queue), availability, observability, and scheduling. Start from the **EKS Standard** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarpenter
metadata:
  name: karpenter
  org: acme-corp
  env: prod
spec:
  namespace:
    value: kube-system
  cluster:
    name: prod-eks
    eksControlPlane: true
  aws:
    irsaRoleArn:
      value: arn:aws:iam::111111111111:role/karpenter-controller
    interruptionQueue: karpenter-interruptions
```

```shell
planton apply -f karpenter.yaml
```

This installs the CRD and controller releases into `kube-system`, wired to the `prod-eks` control plane with IRSA identity and interruption handling; declare the fleet next (an EC2NodeClass and at least one NodePool) and Karpenter starts launching nodes for pending pods. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the controller's IAM identity from the IAM role that carries upstream's controller policy:

```yaml
spec:
  namespace:
    value: kube-system
  cluster:
    name: prod-eks
    eksControlPlane: true
  aws:
    irsaRoleArn:
      valueFrom:
        kind: AwsIamRole
        name: karpenter-controller-role
        fieldPath: status.outputs.role_arn
```

The InfraPipeline creates the IAM role first, then installs the controller with the resolved ARN — the reference is also the deploy-ordering edge.

## Key Configuration

These are the most important decisions when configuring Karpenter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The controller needs unmanaged capacity** -- Karpenter cannot run on nodes it provisions (it would disrupt its own machine). Keep a small managed node group or Fargate profile for the controller pods.

**Cluster identity** -- the cluster name (and the EKS control-plane flag) tell the controller which cluster to provision for; the AWS arm carries the IRSA role and the interruption queue.

**CRDs are their own release** -- the `karpenter-crd` chart keeps the fleet definitions upgradable independently of the controller, and kept CRDs mean uninstalling the controller never deletes the cluster's NodePool and EC2NodeClass declarations.

**Pick one fleet controller** -- Karpenter and Cluster Autoscaler must not manage the same capacity. Karpenter suits clusters that want just-in-time, right-sized nodes; Cluster Autoscaler suits pre-defined node groups.

**Interruption handling before spot at scale** -- the SQS queue lets the controller drain nodes ahead of spot reclaims and scheduled maintenance instead of losing them mid-flight.

**The escape hatch** -- `helmValues` carries additional chart values as a YAML document, merged LAST over everything the typed fields render — never the substitute for typed fields, never a place for secrets.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **AwsIamRole** | `aws.irsaRoleArn` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Installation namespace | Debugging and composition |
| `release_name` | Controller Helm release name (always `karpenter`) | Debugging the release (`helm status`) |
| `crd_release_name` | CRD Helm release name (`karpenter-crd`; empty when CRDs are managed elsewhere) | Verifying CRD ownership |
| `service_account_name` | Always `karpenter` | The subject IRSA trust policies and EKS Pod Identity associations are written against |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**EKS standard** -- `kube-system` home, IRSA identity, interruption queue wired. Start from the **EKS Standard** preset.

**Isolated VPC** -- private clusters that pull images through a mirror and reach AWS APIs through VPC endpoints. Start from the **EKS Isolated VPC** preset.

**Load-bearing provisioning** -- explicit controller sizing, priority, and Prometheus telemetry for provisioning-latency and disruption-decision visibility. Start from the **HA Tuned** preset.

## Works With

- [**Karpenter EC2 Node Class**](/infra-catalog/kubernetes-karpenter-ec2-node-class) -- the machine template (AMI, networking, storage, IAM) the fleet launches from; declare it right after the controller
- [**Karpenter Node Pool**](/infra-catalog/kubernetes-karpenter-node-pool) -- the fleet constraints (instance families, capacity types, limits, disruption policy); one class typically serves several pools
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the controller's IRSA identity; wire `irsaRoleArn` from its `role_arn` output
- [**Cluster Autoscaler**](/infra-catalog/kubernetes-cluster-autoscaler) -- the alternative fleet controller for pre-defined node groups; never both on the same capacity
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) -- pending pods from any workload trigger provisioning; no per-workload wiring needed
