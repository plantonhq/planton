# AWS ECR Registry Settings

Configures the registry-level ECR posture for one AWS region: what scans your images and when, where they replicate, which upstream registries pull through transparently, what auto-created repositories look like, and which automation roles stay out of pull-time metrics. AWS keeps exactly one private registry per account and region, so this is a settings singleton — deploy at most one instance per region, or two instances fight over the same registry objects. Individual repositories are the AwsEcrRepo kind; everything here governs the registry all repositories in the region share.

## What Gets Created

This component adopts the account's existing ECR registry in the target region — the registry itself is never created or destroyed — and configures its posture arm by arm:

- **Registry permissions policy** — configured only when `registryPolicy` is set: the IAM resource policy granting other accounts registry-level actions (replication in, pull-through cache sharing). Destroying this arm deletes the policy
- **Scanning configuration** — configured only when `scanning` is set: the BASIC or ENHANCED (Amazon Inspector) engine plus per-repository-pattern frequency rules. Destroying this arm resets the registry to BASIC scanning with no rules — AWS has no delete, so the modules put the empty default back
- **Replication rules** — configured only when `replicationRules` is set: cross-region and cross-account image replication, evaluated in order. Destroying this arm resets replication to none
- **Pull-through cache rules** — one per `pullThroughCacheRules` entry, keyed by prefix: pulls of `{prefix}/...` transparently fetch from the upstream and cache here. Destroying an entry deletes the rule; already-cached repositories and images remain
- **Repository creation templates** — one per `repositoryCreationTemplates` entry, keyed by prefix: settings stamped onto repositories the registry creates on your behalf (replication, cache pulls, create-on-push). Destroying an entry deletes the template; repositories it already stamped keep their settings
- **Account settings** — configured only when `accountSettings` is set: the basic-scanner version, blob mounting, and registry-policy scope toggles. These persist after destroy at their last-applied values — AWS has no reset
- **Pull-time update exclusions** — one per `pullTimeUpdateExclusions` entry: IAM principals whose pushes do not refresh pull-time metrics

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with ECR permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **For credentialed cache upstreams (Docker Hub and similar)** — the upstream's credentials in a Secrets Manager secret whose name starts with `ecr-pullthroughcache/` (AWS's required prefix), wired via `credentialArn`.
- **For cross-account replication** — the destination registry's policy must allow `ecr:ReplicateImage` from this account; the destination account id goes in the rule's `registryId`.
- **For cross-account ECR upstreams** — an IAM role this registry assumes to pull, wired via the cache rule's `customRoleArn`.

## Deploy

### Console

Open the deployment store, find **AWS ECR Registry Settings**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, then the registry arms: scanning, replication, cache rules, creation templates, and account toggles. Start from the **Hardened Registry Posture** preset in the [Presets](#presets) tab for the security baseline, or the **Upstream Registry Caches** preset to insulate builds from upstream rate limits.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEcrRegistrySettings
metadata:
  name: us-east-1-registry-caches
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  pullThroughCacheRules:
    - ecrRepositoryPrefix: docker-hub
      upstreamRegistryUrl: registry-1.docker.io
      credentialArn:
        valueFrom:
          kind: AwsSecretsManagerSecret
          name: ecr-pullthroughcache-dockerhub
          fieldPath: status.outputs.secret_arn
    - ecrRepositoryPrefix: k8s
      upstreamRegistryUrl: registry.k8s.io
  repositoryCreationTemplates:
    - prefix: docker-hub
      description: Cached Docker Hub images
      appliedFor:
        - PULL_THROUGH_CACHE
      imageTagMutability: IMMUTABLE
      lifecyclePolicy: '{"rules":[{"rulePriority":1,"description":"expire stale cached images","selection":{"tagStatus":"any","countType":"sinceImagePushed","countUnit":"days","countNumber":90},"action":{"type":"expire"}}]}'
    - prefix: k8s
      description: Cached registry.k8s.io images
      appliedFor:
        - PULL_THROUGH_CACHE
      imageTagMutability: IMMUTABLE
      lifecyclePolicy: '{"rules":[{"rulePriority":1,"description":"expire stale cached images","selection":{"tagStatus":"any","countType":"sinceImagePushed","countUnit":"days","countNumber":90},"action":{"type":"expire"}}]}'
```

```shell
planton apply -f ecr-registry-settings.yaml
```

This configures two pull-through caches — authenticated Docker Hub under `docker-hub/` and registry.k8s.io under `k8s/` — each paired with a creation template so cached repositories arrive with immutable tags and a 90-day expiry. An Infra Job tracks the provisioning in real time.

### InfraChart

When the registry posture deploys alongside its credential secret and CI role in one chart, wire them via ValueFromRef:

```yaml
spec:
  region: us-east-1
  pullThroughCacheRules:
    - ecrRepositoryPrefix: docker-hub
      upstreamRegistryUrl: registry-1.docker.io
      credentialArn:
        valueFrom:
          kind: AwsSecretsManagerSecret
          name: ecr-pullthroughcache-dockerhub
          fieldPath: status.outputs.secret_arn
  pullTimeUpdateExclusions:
    - valueFrom:
        kind: AwsIamRole
        name: ci-deploy-role
        fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the secret and role first, then applies the registry configuration referencing them.

## Key Configuration

These are the most important decisions when configuring registry settings. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Know each arm's destroy contract before you rely on it** — Destroy is three different things here: the registry policy and the keyed collections (cache rules, templates, exclusions) genuinely delete; scanning and replication reset to empty defaults; account settings persist at their last-applied values. To "undo" an account setting, apply the value you want before destroying — removal changes nothing.

**Pair every cache rule with a creation template** — A pull-through cache mints repositories on first pull; without a template matching the same prefix they arrive with bare defaults (mutable tags, AES256, no lifecycle policy). The paired template is what makes cached repositories arrive governed: immutable tags, your KMS key, an expiry policy that keeps the cache from growing forever.

**Cache credentials do not un-set** — Once a rule carries a `credentialArn` or `customRoleArn`, clearing the field back to empty is silently not propagated by the provider — the old credential stays attached. Replace the rule to genuinely drop credentials; rotating the secret's value needs no rule change at all. Cache-rule and template prefixes themselves are fixed for life — everything else updates in place.

**Enhanced scanning is an Inspector billing decision** — `scanType: ENHANCED` hands scanning to Amazon Inspector: OS plus language packages, and CONTINUOUS_SCAN rules that re-scan as new CVEs publish — billed by Inspector per scanned image, where BASIC's scan-on-push is not. Scope continuous re-scanning to the repositories that warrant it (a `prod-*` filter) to keep that spend proportional.

**Exclude your automation roles from pull-time metrics early** — Lifecycle policies that expire by days-since-last-pull are silently defeated by automation: replication, cache refreshes, and CI pulls all refresh the metric. Register those principals in `pullTimeUpdateExclusions` from day one — retrofitting after images carry wrong last-pull stamps cannot rewrite history.

**Registry policy is not repository policy** — `registryPolicy` grants registry-level actions only (replication in, cache sharing); grants on individual repositories belong on AwsEcrRepo's repository policy. Keep `accountSettings.registryPolicyScope: V2` — V1 is the legacy semantics.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSecretsManagerSecret** | `pullThroughCacheRules[].credentialArn` | `status.outputs.secret_arn` |
| **AwsIamRole** | `pullTimeUpdateExclusions[]`, `pullThroughCacheRules[].customRoleArn`, `repositoryCreationTemplates[].customRoleArn` | `status.outputs.role_arn` |
| **AwsKmsKey** | `repositoryCreationTemplates[].encryption.kmsKey` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `registry_url` | The registry's pull URL base, `{account}.dkr.ecr.{region}.amazonaws.com` | The image prefix workloads and CI pull cached upstream images through |
| `registry_id` | The registry id — the account's 12-digit id | Cross-account registry and replication policies |

The map outputs (`pull_through_cache_rule_registry_ids`, `repository_creation_template_registry_ids`, `pull_time_update_exclusion_arns`) enumerate the configured entries for import-ID derivation and audit rather than composition.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Security-baseline posture** — Inspector-backed enhanced scanning with continuous re-scanning scoped to production repositories (everything else scans on push), account settings pinned to current-generation values, and the CI role excluded from pull-time metrics so lifecycle policies expire by real usage. Start from the **Hardened Registry Posture** preset.

**Upstream caches for build insulation** — Docker Hub (authenticated — its anonymous limits bite CI first), registry.k8s.io, and other upstreams pulled through as `{registry_url}/{prefix}/...` and cached locally, each prefix paired with a governing creation template. Builds keep working through upstream outages and rate limits. Start from the **Upstream Registry Caches** preset.

**Multi-region image distribution** — Replication rules copy images from the build region to the regions workloads run in, with `repositoryFilters` scoping which repositories travel. Pair with creation templates for the REPLICATION path so destination repositories arrive governed, and remember cross-account destinations need the destination registry's policy to allow `ecr:ReplicateImage` from this account.

## Works With

- [**AWS ECR Repository**](/infra-catalog/aws-ecr-repo) — the individual repositories this registry-level posture governs
- [**AWS Secrets Manager Secret**](/infra-catalog/aws-secrets-manager-secret) — holds upstream credentials for authenticated cache rules (named under `ecr-pullthroughcache/`)
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — pull-time exclusions, cross-account cache pulls, and template-creation roles
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption stamped onto auto-created repositories via template encryption
