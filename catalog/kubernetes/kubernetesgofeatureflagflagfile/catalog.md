# GO Feature Flag Flag File

Deploys a GO Feature Flag flag file -- typed, validated feature flags with their variations, targeting rules, percentage, progressive and scheduled rollouts -- rendered into a ConfigMap that a GO Feature Flag relay reads. A flag flip edits only this resource: the relay is never re-applied and serves the change on its next poll, with no restart.

Every rule GO Feature Flag enforces on a flag is checked at plan time -- one value type across variations, a default rule that resolves, rules naming only defined variations, rollouts that ramp forward, unique rule names, real dates. Query syntax is the one check the relay makes on load: a query that does not parse drops that flag, and its evaluations return the caller's default.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ConfigMap** -- named after the resource, holding the flag file under `key` (default `flags.goff.yaml`) as one JSON document in GO Feature Flag's flag format, which the relay's default `yaml` file format reads
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Kubernetes Cluster

- **A GO Feature Flag relay** -- a GO Feature Flag component whose `configMap` retriever names this flag file's ConfigMap and key. Put the flag file in the relay's namespace, or name this namespace in the retriever; the relay module grants the read either way.
- **An existing namespace** -- the flag file creates no namespace.

## Deploy

### Console

Open the deployment store, find **GO Feature Flag Flag File**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Release Flags** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGoFeatureFlagFlagFile
metadata:
  name: release-flags
  org: acme-corp
  env: prod
spec:
  namespace:
    value: feature-flags
  flags:
    new-checkout:
      variations:
        enabled:
          boolValue: true
        disabled:
          boolValue: false
      targeting:
        - name: first-organizations
          query: org in ["acme-corp"]
          variation: enabled
      defaultRule:
        variation: disabled
```

```shell
planton apply -f release-flags.yaml
```

This renders one ConfigMap, `release-flags`, holding a `new-checkout` flag that is on for the `acme-corp` organization and off for everyone else. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to place the flag file in a namespace managed alongside it:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: feature-flags
      fieldPath: spec.name
```

The InfraPipeline creates the namespace first, then renders the flag file into it; a relay referencing this flag file's `config_map_name` and `key` resolves against it in the same pipeline.

## Key Configuration

These are the most important decisions when configuring a GO Feature Flag flag file. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One flag file per audience of owners** -- A flag file is the unit a team edits and reviews. Several flag files can feed one relay (the retriever listed later wins on a flag both define), so split by owner rather than piling every flag into one 1 MiB ConfigMap.

**Targeting by attribute** -- `targeting[].query` matches the evaluation context the caller sends (`org in ["acme"]`, `email ew "@example.com"`, `targetingKey sw "beta-"`). Target on attributes that never change meaning -- an organization identifier that is never reissued keeps a rule correct forever.

**Percentages are sticky** -- A `percentage` split or `progressiveRollout` buckets on the targeting key, so a caller keeps its variation as the share grows. `bucketingKey` buckets on another attribute instead -- a company id gives a whole company one variation.

**Ramps without edits** -- `progressiveRollout` moves a share between two dates on its own; `scheduledRollout` applies changes at given times (widen a percentage, swap the default rule, turn the flag off), each step merged into the flag field by field. Dates are real RFC 3339 timestamps.

**Turning a flag off** -- `disable: true` makes every evaluation return the caller's default value while keeping the flag's rules in the file; deleting the flag from the file does the same and loses them. Retire a flag by deleting it only after the code that evaluates it is gone.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `config_map_name` | The rendered ConfigMap | A GO Feature Flag relay's `configMap` retriever `configMapName` |
| `key` | The data key holding the flag file | The retriever's `key` |
| `namespace` | The ConfigMap's namespace | The retriever's `namespace` when the relay runs elsewhere |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Release flags by organization** -- Boolean flags on for a named list of organizations and off for everyone else; turning a feature on for one more organization is a one-line change. Start from the **Release Flags** preset.

**Self-driving rollout** -- Staff first, then a dated ramp from 0% to 100% of callers with no one editing the flag during the ramp. Start from the **Progressive Rollout** preset.

**Configuration flags** -- String, number, object or list variations that hand services a value (a limit, a color, a region list) rather than a switch.

## Works With

- [**GO Feature Flag**](/infra-catalog/kubernetes-go-feature-flag) -- the relay that serves this flag file through its ConfigMap retriever
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace the ConfigMap renders into
- [**flagd Flag File**](/infra-catalog/kubernetes-flagd-flag-file) -- the flagd alternative, rendering flagd's JSONLogic flag format
