---
title: "GHA Runner Scale Set"
description: "GHA Runner Scale Set deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesgharunnerscaleset"
---

# Kubernetes GHA Runner Scale Set

Deploys a GitHub Actions Runner Scale Set on a Kubernetes cluster, providing self-hosted runners that automatically scale based on workflow demand. The module installs an AutoScalingRunnerSet custom resource via the official Helm chart. When GitHub Actions workflows request runners with matching labels, the controller creates ephemeral runner pods to execute the jobs, scaling down to a configurable minimum when idle.

## What Gets Created

When you deploy a KubernetesGhaRunnerScaleSet resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Persistent Volume Claims** — one PVC per entry in `persistentVolumes`, used to persist build caches or dependencies across job runs
- **GitHub Credentials Secret** — a Kubernetes Secret containing the PAT token or GitHub App credentials (skipped when `existingSecretName` is provided)
- **Helm Release** — the `gha-runner-scale-set` chart from `oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set`, which creates the AutoScalingRunnerSet custom resource that the controller watches

## Prerequisites

- **KubernetesGhaRunnerScaleSetController** must already be deployed in the cluster (use the [KubernetesGhaRunnerScaleSetController](/docs/catalog/kubernetes/gha-runner-scale-set-controller) component)
- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **GitHub authentication** — one of:
  - A Personal Access Token (PAT) with `repo` scope (repository-level) or `admin:org` scope (organization-level)
  - A GitHub App with an installation ID and private key
  - A pre-existing Kubernetes Secret containing the credentials

## Quick Start

Create a file `gha-runner-scale-set.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGhaRunnerScaleSet
metadata:
  name: my-runners
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesGhaRunnerScaleSet.my-runners
spec:
  namespace:
    value: gha-runners
  createNamespace: true
  github:
    configUrl: https://github.com/my-org
    patToken:
      token: ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
  containerMode:
    type: DIND
```

Deploy:

```shell
openmcf apply -f gha-runner-scale-set.yaml
```

This registers a runner scale set named `my-runners` against the `my-org` GitHub organization using Docker-in-Docker mode. Workflows targeting `runs-on: [self-hosted, my-runners]` will be picked up by the ephemeral runner pods.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace where the runner scale set is installed. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `github` | `object` | GitHub connection configuration. | Required |
| `github.configUrl` | `string` | GitHub URL to register runners against. Accepts repository (`https://github.com/org/repo`), organization (`https://github.com/org`), or enterprise (`https://github.com/enterprises/ent`) URLs. | Must start with `https://github.com/` or `https://github.` |
| `github.patToken.token` | `string` | Personal Access Token. Required when using PAT authentication. | Required (if PAT auth) |
| `github.githubApp.appId` | `string` | GitHub App ID or Client ID. | Required (if App auth) |
| `github.githubApp.installationId` | `string` | GitHub App Installation ID. | Required (if App auth) |
| `github.githubApp.privateKeyBase64` | `string` | Base64-encoded PEM private key for the GitHub App. | Required (if App auth) |
| `containerMode` | `object` | Container mode configuration for running workflows. | Required |
| `containerMode.type` | `enum` | Container mode type. Valid values: `DIND`, `KUBERNETES`, `KUBERNETES_NO_VOLUME`, `DEFAULT`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying the scale set. |
| `helmChartVersion` | `string` | `"0.13.1"` | Version of the `gha-runner-scale-set` Helm chart. Chart versions align with the runner image versions. |
| `scaling.minRunners` | `int32` | `0` | Minimum number of idle runners. Set to `0` for scale-to-zero. Must be >= 0. |
| `scaling.maxRunners` | `int32` | `5` | Maximum concurrent runners. Limits resource consumption. Must be >= 1. |
| `runnerGroup` | `string` | `"default"` | Runner group name in GitHub (organization or enterprise level). |
| `runnerScaleSetName` | `string` | `metadata.name` | Name of the scale set as it appears in GitHub. Used as the `runs-on` label in workflow YAML. |
| `containerMode.workVolumeClaim.size` | `string` | — | Size of the ephemeral work volume for Kubernetes mode (e.g., `10Gi`). Required when `containerMode.type` is `KUBERNETES`. |
| `containerMode.workVolumeClaim.storageClass` | `string` | cluster default | Storage class for the work volume. |
| `containerMode.workVolumeClaim.accessModes` | `string[]` | `["ReadWriteOnce"]` | Access modes for the work volume. |
| `runner.image.repository` | `string` | `"ghcr.io/actions/actions-runner"` | Runner container image repository. |
| `runner.image.tag` | `string` | `"2.331.0"` | Runner container image tag. |
| `runner.image.pullPolicy` | `string` | `"IfNotPresent"` | Image pull policy. Valid values: `Always`, `IfNotPresent`, `Never`. |
| `runner.resources.requests.cpu` | `string` | `"500m"` | CPU request for the runner container. |
| `runner.resources.requests.memory` | `string` | `"1Gi"` | Memory request for the runner container. |
| `runner.resources.limits.cpu` | `string` | `"2"` | CPU limit for the runner container. |
| `runner.resources.limits.memory` | `string` | `"4Gi"` | Memory limit for the runner container. |
| `runner.env` | `object[]` | — | Environment variables injected into the runner container. Each entry has `name` and `value`. |
| `runner.volumeMounts` | `object[]` | — | Additional volume mounts for the runner container. Each entry has `name`, `mountPath`, and optional `readOnly` / `subPath`. |
| `persistentVolumes` | `object[]` | — | Persistent volumes to create and mount. Each entry has `name`, `size`, `mountPath`, and optional `storageClass`, `accessModes`, `readOnly`. |
| `controllerServiceAccount.namespace` | `string` | — | Namespace where the controller is installed. Required when automatic controller discovery does not work. |
| `controllerServiceAccount.name` | `string` | — | Name of the controller's service account. |
| `imagePullSecrets` | `string[]` | — | Names of Kubernetes Secrets for pulling images from private registries. |
| `labels` | `map<string, string>` | — | Labels applied to all resources created by the scale set. |
| `annotations` | `map<string, string>` | — | Annotations applied to all resources created by the scale set. |
| `github.existingSecretName` | `string` | — | Name of a pre-existing Secret containing GitHub credentials. Must be in the same namespace. Mutually exclusive with `patToken` and `githubApp`. |

> **Note on `StringValueOrRef`:** The `namespace` field accepts either an inline `value` or a `valueFrom` reference that resolves the value from another OpenMCF resource's output at deploy time.

## Examples

### Minimal — Organization Runners with PAT and DinD

Registers a Docker-in-Docker runner scale set against a GitHub organization using a Personal Access Token:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGhaRunnerScaleSet
metadata:
  name: org-runners
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesGhaRunnerScaleSet.org-runners
spec:
  namespace:
    value: gha-runners
  createNamespace: true
  github:
    configUrl: https://github.com/my-org
    patToken:
      token: ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
  containerMode:
    type: DIND
```

Use in a workflow:

```yaml
jobs:
  build:
    runs-on: [self-hosted, org-runners]
```

### Kubernetes Mode with GitHub App and Custom Scaling

Uses GitHub App authentication, Kubernetes container mode with an ephemeral work volume, and custom scaling limits suitable for a busy CI pipeline:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGhaRunnerScaleSet
metadata:
  name: ci-runners
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesGhaRunnerScaleSet.ci-runners
spec:
  namespace:
    value: gha-runners
  createNamespace: true
  helmChartVersion: "0.13.1"
  github:
    configUrl: https://github.com/my-org
    githubApp:
      appId: "123456"
      installationId: "78901234"
      privateKeyBase64: LS0tLS1CRUdJTi4uLg==
  scaling:
    minRunners: 2
    maxRunners: 20
  runnerGroup: ci-group
  runnerScaleSetName: ci-runners
  containerMode:
    type: KUBERNETES
    workVolumeClaim:
      size: 50Gi
      storageClass: gp3
  runner:
    resources:
      requests:
        cpu: "1"
        memory: 2Gi
      limits:
        cpu: "4"
        memory: 8Gi
```

### Full — Persistent Cache, Custom Image, Existing Secret, and Target Cluster

Deploys a runner scale set on a specific GKE cluster with a persistent build cache volume, a custom runner image, environment variables, an existing credentials secret, and resource annotations:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGhaRunnerScaleSet
metadata:
  name: build-runners
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesGhaRunnerScaleSet.build-runners
spec:
  targetCluster:
    clusterKind: GcpGkeCluster
    clusterName: prod-cluster
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: gha-runners-ns
      field: spec.name
  github:
    configUrl: https://github.com/my-org/my-repo
    existingSecretName: github-runner-creds
  scaling:
    minRunners: 1
    maxRunners: 10
  runnerScaleSetName: build-runners
  containerMode:
    type: DIND
  runner:
    image:
      repository: my-registry.example.com/custom-runner
      tag: "2.331.0-custom"
      pullPolicy: Always
    resources:
      requests:
        cpu: "2"
        memory: 4Gi
      limits:
        cpu: "4"
        memory: 8Gi
    env:
      - name: DOCKER_BUILDKIT
        value: "1"
      - name: RUNNER_TOOL_CACHE
        value: /home/runner/.cache/tools
    volumeMounts:
      - name: build-cache
        mountPath: /home/runner/.cache
  persistentVolumes:
    - name: build-cache
      size: 100Gi
      storageClass: gp3
      mountPath: /home/runner/.cache
  imagePullSecrets:
    - my-registry-secret
  controllerServiceAccount:
    namespace: arc-system
    name: arc-gha-rs-controller
  labels:
    team: platform
  annotations:
    cost-center: engineering
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Namespace where the runner scale set is deployed |
| `releaseName` | `string` | Name of the Helm release |
| `chartVersion` | `string` | Version of the deployed Helm chart |
| `runnerScaleSetName` | `string` | Name of the scale set as registered with GitHub (use in `runs-on` labels) |
| `githubConfigUrl` | `string` | GitHub URL the runners are connected to |
| `githubSecretName` | `string` | Name of the Kubernetes Secret containing GitHub credentials |
| `pvcNames` | `string[]` | Names of PVCs created for persistent volumes |
| `minRunners` | `string` | Configured minimum runners |
| `maxRunners` | `string` | Configured maximum runners |
| `containerMode` | `string` | Container mode type in use (`dind`, `kubernetes`, `kubernetes-novolume`, or empty for default) |

## Related Components

- [KubernetesGhaRunnerScaleSetController](/docs/catalog/kubernetes/gha-runner-scale-set-controller) — required prerequisite; deploys the controller that watches AutoScalingRunnerSet resources and manages runner pod lifecycle
- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — provides the target namespace via `valueFrom` reference
- [KubernetesHelmRelease](/docs/catalog/kubernetes/helm-release) — generic Helm release component; use KubernetesGhaRunnerScaleSet instead for GitHub Actions runners since it provides typed configuration and validation
- [KubernetesSecret](/docs/catalog/kubernetes/secret) — can be used to manage the GitHub credentials secret independently when using `existingSecretName`
