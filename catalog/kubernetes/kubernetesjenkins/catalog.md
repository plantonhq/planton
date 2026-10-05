# Jenkins

Deploys a Jenkins automation server on any Kubernetes cluster using the official Jenkins Helm chart. Supports configurable resource allocation, auto-generated admin credentials stored in a Kubernetes Secret, Helm value overrides for plugin and agent configuration, and optional external exposure through a Gateway API Gateway with a cert-manager-issued TLS certificate.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Admin Password Secret** -- a Kubernetes Secret containing a generated admin password for the Jenkins controller (never authored in the manifest)
- **Jenkins Helm Release** -- installs the official Jenkins Helm chart with configured CPU/memory resources, admin credentials from the generated secret, and a ClusterIP service for the Jenkins web UI on port 8080
- **TLS Certificate, Gateway, and HTTPRoutes** -- created only when `ingress.enabled` is `true`: a cert-manager Certificate for the hostname (issued by the cluster's ClusterIssuer), a Gateway API Gateway using the `istio` GatewayClass in the `istio-ingress` namespace, an HTTP-to-HTTPS redirect route, and the HTTPS route to the Jenkins service
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A storage class** available for persistent volumes. Jenkins uses persistent storage for job history, plugin data, and configuration. The cluster must support dynamic PV provisioning.
- **Istio with Gateway API support and cert-manager with a ClusterIssuer** -- only when `ingress.enabled` is `true`: the module creates the Gateway with the `istio` GatewayClass in the `istio-ingress` namespace and requests the TLS certificate from the ClusterIssuer named after the hostname's DNS domain.

## Deploy

### Console

Open the deployment store, find **Jenkins**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Jenkins** preset in the [Presets](#presets) tab to pre-populate a working configuration with ingress enabled.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesJenkins
metadata:
  name: build-jenkins
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "jenkins"
  createNamespace: true
  containerResources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 2000m
      memory: 4Gi
  ingress:
    enabled: true
    hostname: "jenkins.example.com"
```

```shell
planton apply -f jenkins.yaml
```

This creates a Jenkins controller with ingress at `jenkins.example.com`, auto-generated admin credentials, and production-grade resource limits. The admin password is stored in a Kubernetes Secret available in the outputs. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Jenkins deployment to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: cicd-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions Jenkins into it.

## Key Configuration

These are the most important decisions when configuring Jenkins. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Resource allocation** -- Jenkins is memory-intensive, especially with many plugins and concurrent builds. The spec defaults are modest (`1000m` CPU / `1Gi` memory limits, `50m` / `100Mi` requests) -- fine for evaluation, undersized for real CI. Instances running many concurrent pipelines or heavy compilation jobs need explicit limits in the `2000m` / `4Gi` range or beyond.

**External access rides Istio and cert-manager** -- Set `ingress.enabled: true` with a `hostname` (e.g., `"jenkins.example.com"`) and the module creates a Gateway (class `istio`), an HTTPS-redirect route, and a cert-manager Certificate whose ClusterIssuer is named after the hostname's DNS domain -- that issuer must already exist on the cluster. When ingress is disabled, access is limited to within the cluster or via the `port_forward_command` in the outputs.

**Helm value overrides** -- The `helmValues` map passes additional key-value pairs directly to the Jenkins Helm chart for configuration not covered by the spec, such as JCasC (Jenkins Configuration as Code) settings, plugin installation, agent pod templates, or RBAC configuration. Refer to the Jenkins Helm chart documentation for available options.

**Admin credentials** -- The IaC module generates a random admin password and stores it in a Kubernetes Secret. Retrieve the password from the `password_secret` output after provisioning. The admin username is available in the `username` output.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Kubernetes namespace where Jenkins is running | Application deployment manifests |
| `service` | Kubernetes service name for Jenkins | Reverse proxy or service mesh configuration |
| `port_forward_command` | Ready-to-run `kubectl port-forward` command | Local development access when ingress is disabled |
| `kube_endpoint` | Cluster-internal FQDN for the Jenkins web UI | Webhook configuration for internal services |
| `external_hostname` | Public hostname (when ingress is enabled) | Webhook configuration for Git providers |
| `internal_hostname` | Private hostname for cluster-internal routing | Internal service-to-service communication |
| `username` | Jenkins admin username | Initial login |
| `password_secret` | Kubernetes Secret reference for the admin password | Login via `secretKeyRef` or CLI retrieval |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard CI/CD server** -- Jenkins controller with ingress enabled, production-grade resources (`100m`/`256Mi` requests, `2000m`/`4Gi` limits), and a dedicated `jenkins` namespace. Pipeline agents are dynamically provisioned as Kubernetes pods. Start from the **Standard Jenkins** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for Jenkins deployment
- [**Istio**](/infra-catalog/kubernetes-istio) -- provides the `istio` GatewayClass the ingress arm's Gateway uses
- [**Cert Manager**](/infra-catalog/kubernetes-cert-manager) -- issues the TLS certificate for the ingress hostname via the cluster's ClusterIssuer
