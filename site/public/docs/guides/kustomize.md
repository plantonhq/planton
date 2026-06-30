---
title: "Kustomize Integration"
description: "Using Kustomize with Planton for multi-environment deployments — directory structure, overlays, and workflows"
icon: "guide"
order: 40
---

# Kustomize Integration

Kustomize lets you manage variations of Planton manifests without duplication. Instead of maintaining separate manifest files for dev, staging, and production, you maintain one **base** manifest and environment-specific **overlays** that patch the base.

For the conceptual overview of manifest sources (including Kustomize), see [Manifests](../concepts/manifests). For flag details, see [CLI Reference](/docs/cli/cli-reference).

```text
manifests/database/
|-- base/
|   \-- database.yaml      # Shared configuration
\-- overlays/
    |-- dev/
    |-- staging/
    \-- prod/               # Environment-specific patches
```

Planton integrates Kustomize as a Go library (`sigs.k8s.io/kustomize`), not as an external binary. The `--kustomize-dir` and `--overlay` flags trigger Kustomize to build the final manifest at deployment time.

---

## Quick Start

### 1. Create Base Manifest

```bash
mkdir -p services/api/kustomize/base
```

**`services/api/kustomize/base/deployment.yaml`**:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
metadata:
  name: api
spec:
  container:
    image:
      repo: myapp/api
      tag: latest
    replicas: 1
    resources:
      limits:
        cpu: 500m
        memory: 512Mi
```

**`services/api/kustomize/base/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - deployment.yaml
```

### 2. Create Environment Overlay

**`services/api/kustomize/overlays/prod/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

patches:
  - path: patch.yaml
```

**`services/api/kustomize/overlays/prod/patch.yaml`**:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
metadata:
  name: api
spec:
  container:
    image:
      tag: v1.0.0
    replicas: 3
    resources:
      limits:
        cpu: 2000m
        memory: 4Gi
```

### 3. Deploy with Planton

```bash
# Deploy to production
planton pulumi up \
  --kustomize-dir services/api/kustomize \
  --overlay prod

# Deploy to dev
planton pulumi up \
  --kustomize-dir services/api/kustomize \
  --overlay dev
```

**What happens**:
1. Planton runs `kustomize build services/api/kustomize/overlays/prod`
2. Merges base + prod overlay into final manifest
3. Validates the result
4. Deploys using Pulumi or OpenTofu

---

## Directory Structure

### Standard Layout

```text
<service-name>/kustomize/
|-- base/
|   |-- kustomization.yaml          # Base kustomization config
|   \-- <resource>.yaml             # Base resource definition
\-- overlays/
    |-- dev/
    |   |-- kustomization.yaml      # Dev environment config
    |   \-- patch.yaml              # Dev-specific patches
    |-- staging/
    |   |-- kustomization.yaml
    |   \-- patch.yaml
    \-- prod/
        |-- kustomization.yaml
        \-- patch.yaml
```

### Example: Complete Service

```text
backend/services/api/kustomize/
|-- base/
|   |-- kustomization.yaml
|   |-- deployment.yaml
|   \-- database.yaml
\-- overlays/
    |-- dev/
    |   |-- kustomization.yaml
    |   |-- deployment-patch.yaml
    |   \-- database-patch.yaml
    |-- staging/
    |   |-- kustomization.yaml
    |   |-- deployment-patch.yaml
    |   \-- database-patch.yaml
    \-- prod/
        |-- kustomization.yaml
        |-- deployment-patch.yaml
        \-- database-patch.yaml
```

---

## Creating Patches

### Strategic Merge Patches

The most common approach - specify only the fields you want to change:

**Base**:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPostgres
metadata:
  name: app-database
spec:
  container:
    replicas: 1
    resources:
      limits:
        cpu: 500m
        memory: 1Gi
    diskSize: 10Gi
```

**Prod Patch** (`overlays/prod/patch.yaml`):

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPostgres
metadata:
  name: app-database
spec:
  container:
    replicas: 3              # Override
    resources:
      limits:
        cpu: 2000m           # Override
        memory: 4Gi          # Override
    diskSize: 100Gi          # Override
```

**Result**: Base + patch merged = 3 replicas, 2000m CPU, 4Gi memory, 100Gi disk.

### JSON 6902 Patches

For more complex changes:

**`overlays/prod/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

patches:
  - target:
      kind: KubernetesPostgres
      name: app-database
    patch: |-
      - op: replace
        path: /spec/container/replicas
        value: 3
      - op: add
        path: /metadata/labels/environment
        value: production
```

---

## Common Patterns

### Pattern 1: Environment-Specific Resources

Different instance sizes per environment:

**Dev** (small, cheap):

```yaml
spec:
  container:
    replicas: 1
    resources:
      limits:
        cpu: 500m
        memory: 512Mi
```

**Prod** (large, resilient):

```yaml
spec:
  container:
    replicas: 5
    resources:
      limits:
        cpu: 2000m
        memory: 4Gi
```

### Pattern 2: Environment-Specific Labels

Add labels for cost tracking:

**`overlays/prod/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

commonLabels:
  environment: production
  cost-center: engineering
  team: backend
```

### Pattern 3: Environment-Specific Images

**`overlays/dev/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

images:
  - name: myapp/api
    newTag: latest          # Dev uses latest

patches:
  - path: dev-patch.yaml
```

**`overlays/prod/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

images:
  - name: myapp/api
    newTag: v1.2.3          # Prod uses specific version

patches:
  - path: prod-patch.yaml
```

### Pattern 4: Shared Configuration + Environment Patches

**`base/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonLabels:
  app: api
  managed-by: planton

resources:
  - deployment.yaml
  - database.yaml
  - cache.yaml
```

Each resource in base defines shared configuration, overlays patch for environment needs.

---

## Workflow Examples

### Deploying to Multiple Environments

```bash
# Deploy to dev
planton pulumi up \
  --kustomize-dir services/api/kustomize \
  --overlay dev \
  --yes

# Test in dev...

# Deploy to staging
planton pulumi up \
  --kustomize-dir services/api/kustomize \
  --overlay staging \
  --yes

# Test in staging...

# Deploy to production (with review)
planton pulumi preview \
  --kustomize-dir services/api/kustomize \
  --overlay prod

# Review changes...

planton pulumi up \
  --kustomize-dir services/api/kustomize \
  --overlay prod
```

### Combining Kustomize with --set Overrides

```bash
# Kustomize overlay + runtime override
planton pulumi up \
  --kustomize-dir services/api/kustomize \
  --overlay prod \
  --set spec.container.image.tag=v1.2.4
```

**Order of precedence**:
1. Base manifest
2. Overlay patches applied
3. `--set` overrides applied last (highest priority)

### Preview Built Manifest

```bash
# See what Kustomize generates (useful for debugging)
cd services/api/kustomize
kustomize build overlays/prod

# Or let Planton build and show it
planton load-manifest \
  --kustomize-dir services/api/kustomize \
  --overlay prod
```

---

For using Kustomize in CI/CD pipelines with branch-based overlay selection, see [CI/CD Integration](./cicd-integration).

---

## Advanced Techniques

### Multiple Bases

Useful for shared components:

```text
common/
\-- base/
    |-- kustomization.yaml
    \-- shared-config.yaml

service-a/kustomize/
\-- overlays/
    \-- prod/
        |-- kustomization.yaml  # References ../../../common/base
        \-- patch.yaml

service-b/kustomize/
\-- overlays/
    \-- prod/
        |-- kustomization.yaml  # Also references ../../../common/base
        \-- patch.yaml
```

### Components (Reusable Pieces)

**`components/monitoring/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component

patches:
  - path: add-monitoring.yaml
```

**Use in overlay**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

components:
  - ../../../components/monitoring
```

### Generating ConfigMaps from Files

**`overlays/prod/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

configMapGenerator:
  - name: app-config
    files:
      - config/prod.yaml
      - config/secrets.encrypted
```

---

## Troubleshooting

### Error: "no such file or directory"

**Problem**: Kustomize can't find referenced files.

**Solution**:
```bash
# Check file paths in kustomization.yaml
# Ensure paths are relative to kustomization.yaml location

# Verify structure
ls -R services/api/kustomize/
```

### Error: "kustomization.yaml not found"

**Problem**: Missing kustomization.yaml in overlay.

**Solution**:
```bash
# Create kustomization.yaml
cat > services/api/kustomize/overlays/prod/kustomization.yaml <<EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base
EOF
```

### List Fields Replaced Instead of Merged

**Problem**: Your overlay adds entries to a list field (e.g., `env.variables`, `ports`), but the overlay's list replaces the base list entirely instead of merging.

**Cause**: Kustomize only knows how to merge lists by key for built-in Kubernetes types. For Planton custom resource types (`KubernetesDeployment`, `KubernetesCronJob`, etc.), it has no schema and falls back to replacing lists wholesale.

**Solution**: Use `planton kustomize init` to generate and wire up the OpenAPI schema that teaches kustomize how to merge Planton lists correctly. See [OpenAPI Schema for List Merging](#openapi-schema-for-list-merging) below.

### Patch Not Applied

**Problem**: Your patch isn't affecting the final output.

**Solution**:
```bash
# 1. Verify patch is listed in kustomization.yaml
cat overlays/prod/kustomization.yaml

# 2. Check patch targets correct resource
# - apiVersion must match
# - kind must match
# - metadata.name must match

# 3. Test kustomize build directly
cd services/api/kustomize
kustomize build overlays/prod
```

### Wrong Overlay Applied

**Problem**: Deployed dev config to prod (or vice versa).

**Solution**:
```bash
# Always verify overlay before deploying
planton pulumi preview \
  --kustomize-dir services/api/kustomize \
  --overlay prod  # Double-check this!

# Use explicit confirmation in CI/CD
if [ "$OVERLAY" != "prod" ]; then
  echo "Deploying to $OVERLAY"
  planton pulumi up --kustomize-dir ... --overlay $OVERLAY --yes
else
  echo "Production deployment - manual approval required"
  planton pulumi up --kustomize-dir ... --overlay $OVERLAY
fi
```

---

## OpenAPI Schema for List Merging

Kustomize uses **strategic merge patch** for overlays. For built-in Kubernetes types, it knows to merge list fields by key (e.g., `containers` by `name`, `env` by `name`). For Planton custom resource types, kustomize has no schema and falls back to **replacing lists entirely**.

This means an overlay that adds one environment variable will replace the entire `variables` list from the base, losing all base entries.

Planton solves this with a built-in schema generator that uses proto reflection to discover all list fields with merge keys across all 360+ cloud resource kinds.

### Generating the Schema

```bash
# Print the universal schema to stdout
planton kustomize schema

# Write to a file
planton kustomize schema -o planton-schema.json
```

The schema covers every Planton kind that has list fields which should merge by name — environment variables, secrets, ports, volume mounts, sidecars, and more. Kinds without merge-worthy fields are excluded automatically.

### Initializing a Kustomize Directory

```bash
# Initialize a single service's _kustomize directory
planton kustomize init --dir ./services/api/_kustomize

# Scan a directory tree and initialize ALL _kustomize directories
planton kustomize init --scan ./services
```

The `init` command:

1. Generates the universal schema
2. Writes `planton-schema.json` at the `_kustomize/` root
3. Adds the `openapi:` reference to every overlay `kustomization.yaml`

After initialization, overlay kustomization files look like:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

openapi:
  path: ../../planton-schema.json

resources:
  - ../../base

patches:
  - path: patch.yaml
```

### When to Re-run

Re-run `planton kustomize init` after upgrading Planton to pick up schema changes from new or modified cloud resource kinds. The command is idempotent — it regenerates the schema file and skips overlays that already have the `openapi:` reference.

### What Gets Merged

The schema declares `x-kubernetes-patch-merge-key: "name"` for list fields whose elements have a `name` field. Common examples:

| Kind | Field Path | Merge Key |
|------|-----------|-----------|
| KubernetesDeployment | `spec.container.app.env.variables` | `name` |
| KubernetesDeployment | `spec.container.app.env.secrets` | `name` |
| KubernetesDeployment | `spec.container.app.ports` | `name` |
| KubernetesDeployment | `spec.container.app.volumeMounts` | `name` |
| KubernetesDeployment | `spec.container.sidecars` | `name` |
| KubernetesCronJob | `spec.env.variables` | `name` |
| KubernetesCronJob | `spec.env.secrets` | `name` |
| KubernetesJob | `spec.env.variables` | `name` |
| KubernetesService | `spec.ports` | `name` |

Scalar fields (like `replicas`, `cpu`) and map fields (like `configMaps`) are unaffected — kustomize already handles those correctly without a schema.

---

## Best Practices

### 1. **Keep Base Minimal**

```yaml
# ✅ Good: Base has common configuration
base/deployment.yaml:
  name: api
  container:
    image:
      repo: myapp/api
    # No environment-specific values

# ❌ Bad: Base has production values
base/deployment.yaml:
  name: api
  container:
    replicas: 10  # Production-specific, doesn't belong in base
```

### 2. **One Overlay Per Environment**

```text
# Good
overlays/
|-- dev/
|-- staging/
\-- prod/

# Confusing
overlays/
|-- dev-us-west/
|-- dev-eu-central/
|-- staging-us-west/
\-- ... (too many combinations)
```

### 3. **Use Descriptive Patch Names**

```text
# Good
overlays/prod/
|-- kustomization.yaml
|-- resources-patch.yaml          # Increases resources
|-- replicas-patch.yaml            # Scales replicas
\-- monitoring-patch.yaml          # Adds monitoring

# Bad
overlays/prod/
|-- kustomization.yaml
|-- patch1.yaml
|-- patch2.yaml
\-- patch3.yaml
```

### 4. **Version Control Everything**

```bash
# ✅ Good: All kustomize files in Git
git add services/api/kustomize/
git commit -m "feat: add production overlay for API"

# ❌ Bad: Generated files, temp files committed
git add services/api/kustomize/overlays/prod/output.yaml  # Generated
```

### 5. **Test Overlays Before Deploying**

```bash
# ✅ Good: Preview before applying
kustomize build overlays/prod | less  # Review output
planton pulumi preview --kustomize-dir ... --overlay prod

# ⚠️ Risky: Deploy without review
planton pulumi up --kustomize-dir ... --overlay prod --yes
```

---

## Complete Example

Here's a complete real-world example:

### Directory Structure

```text
backend/services/api/kustomize/
|-- base/
|   |-- kustomization.yaml
|   |-- deployment.yaml
|   |-- database.yaml
|   \-- redis.yaml
\-- overlays/
    |-- dev/
    |   |-- kustomization.yaml
    |   |-- deployment-patch.yaml
    |   |-- database-patch.yaml
    |   \-- redis-patch.yaml
    \-- prod/
        |-- kustomization.yaml
        |-- deployment-patch.yaml
        |-- database-patch.yaml
        \-- redis-patch.yaml
```

### Files

**`base/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonLabels:
  app: api
  managed-by: planton

resources:
  - deployment.yaml
  - database.yaml
  - redis.yaml
```

**`base/deployment.yaml`**:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
metadata:
  name: api
spec:
  container:
    image:
      repo: mycompany/api
      tag: latest
    replicas: 1
    resources:
      limits:
        cpu: 500m
        memory: 512Mi
```

**`overlays/prod/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

commonLabels:
  environment: production

images:
  - name: mycompany/api
    newTag: v1.0.0

patches:
  - path: deployment-patch.yaml
  - path: database-patch.yaml
  - path: redis-patch.yaml
```

**`overlays/prod/deployment-patch.yaml`**:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
metadata:
  name: api
spec:
  container:
    replicas: 5
    resources:
      limits:
        cpu: 2000m
        memory: 4Gi
```

### Deployment

```bash
# Deploy to production
planton pulumi up \
  --kustomize-dir backend/services/api/kustomize \
  --overlay prod
```

---

## What's Next

- [Writing Manifests](./manifests) — Manifest structure and best practices
- [CI/CD Integration](./cicd-integration) — Kustomize with GitHub Actions and GitLab CI
- [Advanced Usage](./advanced-usage) — Combining Kustomize with `--set` overrides
- [State Backends](./state-backends) — Per-environment state configuration
- [CLI Reference](/docs/cli/cli-reference) — Full `planton kustomize schema` and `planton kustomize init` usage
- [Official Kustomize Docs](https://kustomize.io/) — Kustomize reference documentation

