# KubernetesGhaRunnerScaleSet - Technical Documentation

## Overview

The `KubernetesGhaRunnerScaleSet` deployment component enables declarative deployment of GitHub Actions self-hosted runners on Kubernetes clusters. It leverages the official Actions Runner Controller (ARC) Helm chart to create AutoScalingRunnerSet resources that dynamically scale based on workflow demand.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        GitHub                                    │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Repository/Organization                      │   │
│  │                                                          │   │
│  │  Workflow Job → Queue → Webhook → Controller → Runner   │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                            │
│                                                                  │
│  ┌──────────────────────┐    ┌─────────────────────────────┐   │
│  │  Controller (ARC)     │    │  Runner Scale Set           │   │
│  │  ─────────────────    │    │  ───────────────            │   │
│  │  Watches for jobs     │───▶│  AutoScalingRunnerSet       │   │
│  │  Creates runner pods  │    │  EphemeralRunner pods       │   │
│  │  Manages lifecycle    │    │  PVCs for caching           │   │
│  └──────────────────────┘    └─────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## How It Works

### Registration Flow

1. Helm chart creates an `AutoScalingRunnerSet` custom resource
2. Controller registers the scale set with GitHub via the config URL
3. GitHub associates runners with the specified repository/organization/enterprise
4. Runners appear in GitHub Settings → Actions → Runners

### Job Execution Flow

1. Workflow job is triggered with `runs-on: [self-hosted, scale-set-name]`
2. GitHub queues the job and notifies the controller
3. Controller creates an `EphemeralRunner` pod
4. Runner pod registers with GitHub, picks up the job
5. Job executes in the runner pod
6. Runner pod terminates after job completion

### Scaling Behavior

| Scenario | Behavior |
|----------|----------|
| Jobs queued | Scale up to handle queue (up to `maxRunners`) |
| Jobs complete | Scale down to `minRunners` |
| No jobs | Maintain `minRunners` (can be 0 for cost savings) |
| Surge in jobs | Parallel runners up to `maxRunners` |

## Container Modes

### DIND (Docker-in-Docker)

```yaml
containerMode:
  type: DIND
```

- Runner pod includes a privileged DinD sidecar
- Supports `docker build`, `docker run`, etc.
- Required for workflows that build/push Docker images
- **Requires**: Privileged container support in cluster

### KUBERNETES

```yaml
containerMode:
  type: KUBERNETES
  workVolumeClaim:
    storageClass: fast-ssd
    size: "50Gi"
```

- Each workflow step runs as a separate Kubernetes pod
- Native Kubernetes container execution
- Uses container hooks for orchestration
- **Requires**: Ephemeral volume support, ServiceAccount permissions

### KUBERNETES_NO_VOLUME

```yaml
containerMode:
  type: KUBERNETES_NO_VOLUME
```

- Same as KUBERNETES but without ephemeral volumes
- For clusters that don't support ephemeral volume claims
- Workspace is not persisted between steps

### DEFAULT

```yaml
containerMode:
  type: DEFAULT
```

- Direct execution on the runner pod
- No container isolation for steps
- Simple workflows that don't need Docker

## Persistent Volumes

### Purpose

PVCs persist data across runner pod restarts, enabling:
- **Dependency caching**: npm, maven, gradle, pip packages
- **Docker layer caching**: Faster image builds
- **Build artifacts**: Share between jobs

### Implementation

```yaml
persistentVolumes:
  - name: npm-cache
    size: "20Gi"
    storageClass: standard
    mountPath: /home/runner/.npm
```

Creates:
1. A `PersistentVolumeClaim` named `{release-name}-npm-cache`
2. Volume mount in the runner container spec
3. Volume reference in the pod spec

### Cache Effectiveness

For optimal caching:
- Use `minRunners >= 1` to keep at least one runner warm
- PVCs are per-scale-set, not per-runner (shared cache)
- Consider storage class with good IOPS for large caches

## Authentication

### PAT Token

Personal Access Token authentication:

**Permissions needed:**
| Scope | Repository | Organization | Enterprise |
|-------|-----------|--------------|------------|
| `repo` | Required | - | - |
| `admin:org` | - | Required | - |
| `manage_runners:enterprise` | - | - | Required |

**Secret structure:**
```yaml
github_token: ghp_xxxxxxxxxxxx
```

### GitHub App

Recommended for organizations:

**Permissions needed:**
- Repository: `actions:read`, `metadata:read`
- Organization: `self_hosted_runners:read/write`

**Private key encoding:**
The `privateKeyBase64` field must be base64 encoded. Encode your PEM file before providing it:
```bash
cat private-key.pem | base64
```

**Secret structure (created internally):**
```yaml
github_app_id: "123456"
github_app_installation_id: "654321"
github_app_private_key: |
  -----BEGIN RSA PRIVATE KEY-----
  ...
```

### Existing Secret

For secrets provisioned outside this component:

```yaml
github:
  existingSecretName: my-github-secret
```

Secret must contain either PAT or GitHub App fields.

## IaC Implementations

### Pulumi Module

Location: `iac/pulumi/module/`

Key files:
- `main.go`: Entry point, orchestrates deployment
- `locals.go`: Configuration parsing, defaults, exports
- `runner.go`: Helm release and PVC creation
- `vars.go`: Constants (chart name, repo, version)

### Terraform Module

Location: `iac/tf/`

Key files:
- `main.tf`: Namespace, PVCs, Helm release
- `locals.tf`: Value transformations
- `variables.tf`: Input variable definitions
- `outputs.tf`: Stack outputs

## Relationship with Controller

The runner scale set requires the controller to be installed:

```
KubernetesGhaRunnerScaleSetController (one per cluster)
    └── KubernetesGhaRunnerScaleSet (many per cluster)
            ├── Scale Set 1: repo runners
            ├── Scale Set 2: org runners
            └── Scale Set 3: enterprise runners
```

Controller discovery:
1. Helm chart looks for controller by label `app.kubernetes.io/part-of=gha-rs-controller`
2. If not found, specify `controllerServiceAccount` explicitly

## Troubleshooting

### Runners Not Appearing in GitHub

1. Check controller logs: `kubectl logs -n arc-system deploy/arc-controller-manager`
2. Verify GitHub credentials: Secret must have correct keys
3. Check config URL format: `https://github.com/<owner>/<repo>` or `https://github.com/<org>`

### Runners Stuck Pending

1. Check PVC binding: `kubectl get pvc -n <namespace>`
2. Verify storage class exists
3. Check node resources for pod scheduling

### Jobs Not Picked Up

1. Verify `runs-on` label matches `runnerScaleSetName`
2. Check runner group permissions in GitHub
3. Ensure maxRunners > 0

### Docker Not Working in DIND Mode

1. Verify privileged containers are allowed (PSP/PSA)
2. Check DinD sidecar logs
3. Ensure DOCKER_HOST environment is set

## Best Practices

1. **Start with minRunners: 0** for cost efficiency
2. **Set realistic maxRunners** based on cluster capacity
3. **Use persistent volumes** for dependency caching
4. **Prefer GitHub App** over PAT for organizations
5. **Use runner groups** to control repository access
6. **Monitor runner metrics** via controller metrics endpoint

