# KubernetesGhaRunnerScaleSet

Deploy self-hosted GitHub Actions runners on Kubernetes with automatic scaling.

## Overview

`KubernetesGhaRunnerScaleSet` creates a pool of ephemeral GitHub Actions runners that scale based on workflow demand. Runners spin up when jobs are queued and terminate after job completion.

## Prerequisites

1. **Controller Installed**: Deploy `KubernetesGhaRunnerScaleSetController` first
2. **GitHub Authentication**: PAT token or GitHub App credentials
3. **Kubernetes Cluster**: With appropriate storage class for persistent volumes

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGhaRunnerScaleSet
metadata:
  name: my-runners
spec:
  namespace:
    value: gha-runners
  createNamespace: true
  
  github:
    configUrl: https://github.com/myorg/myrepo
    patToken:
      token: ghp_xxxxxxxxxxxx
  
  containerMode:
    type: DIND
  
  scaling:
    minRunners: 0
    maxRunners: 10
```

Deploy:

```bash
planton pulumi up --manifest runner.yaml --stack org/project/env
```

Use in workflows:

```yaml
jobs:
  build:
    runs-on: [self-hosted, my-runners]
    steps:
      - uses: actions/checkout@v4
      - run: echo "Hello from self-hosted runner!"
```

## Features

### Container Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `DIND` | Docker-in-Docker with privileged sidecar | Building/pushing Docker images |
| `KUBERNETES` | Each step runs as a Kubernetes pod | Container-native workflows |
| `KUBERNETES_NO_VOLUME` | Kubernetes mode without ephemeral volumes | Clusters without ephemeral volume support |
| `DEFAULT` | Direct execution without containers | Simple workflows |

### Persistent Volumes

Cache build dependencies across job runs:

```yaml
persistentVolumes:
  - name: npm-cache
    size: "20Gi"
    mountPath: /home/runner/.npm
  - name: maven-cache
    size: "50Gi"
    mountPath: /home/runner/.m2
  - name: docker-cache
    size: "100Gi"
    mountPath: /var/lib/docker
```

### Authentication Options

**PAT Token** (simple, for personal use):
```yaml
github:
  configUrl: https://github.com/myorg/myrepo
  patToken:
    token: ghp_xxxxxxxxxxxx
```

**GitHub App** (recommended for organizations):
```yaml
github:
  configUrl: https://github.com/myorg
  githubApp:
    appId: "123456"
    installationId: "654321"
    privateKey: |
      -----BEGIN RSA PRIVATE KEY-----
      ...
      -----END RSA PRIVATE KEY-----
```

**Existing Secret** (pre-provisioned):
```yaml
github:
  configUrl: https://github.com/myorg
  existingSecretName: github-credentials
```

## Configuration Reference

### Spec Fields

| Field | Description | Required |
|-------|-------------|----------|
| `namespace` | Kubernetes namespace | Yes |
| `createNamespace` | Create namespace if missing | No |
| `github.configUrl` | GitHub URL (repo/org/enterprise) | Yes |
| `github.patToken` | PAT authentication | One of auth |
| `github.githubApp` | GitHub App authentication | One of auth |
| `github.existingSecretName` | Pre-existing secret | One of auth |
| `containerMode.type` | Runner execution mode | Yes |
| `scaling.minRunners` | Minimum idle runners (default: 0) | No |
| `scaling.maxRunners` | Maximum runners (default: 5) | No |
| `runnerGroup` | Runner group name | No |
| `runnerScaleSetName` | Name for runs-on label | No |
| `runner.image` | Custom runner image | No |
| `runner.resources` | CPU/memory limits | No |
| `persistentVolumes` | PVCs for caching | No |

### Stack Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Deployment namespace |
| `runner_scale_set_name` | Name to use in `runs-on` |
| `github_config_url` | Connected GitHub URL |
| `pvc_names` | Created PVC names |
| `min_runners` | Configured minimum |
| `max_runners` | Configured maximum |

## Examples

See [examples.md](examples.md) for detailed deployment scenarios.

