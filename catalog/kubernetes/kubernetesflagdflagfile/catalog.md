# flagd Flag File

Declares a flagd flag definition file -- typed flags with variants, a default variant and JSONLogic targeting, plus shared evaluators -- and renders it into a ConfigMap that a flagd deployment mounts as a source. Every flag is validated when the manifest is planned, so a malformed flag never reaches the daemon. A flag flip edits only this resource: flagd is never re-applied and picks the change up once the kubelet syncs its mounted volume, typically within one to two minutes.

## What Gets Created

- **ConfigMap** -- named after the resource, holding the flag definitions as flagd JSON (with the flagd `$schema`) under the data key `key`, default `flags.flagd.json`.

## Before You Deploy

### Planton Setup

A Kubernetes provider connection in the Connect module targeting the cluster. The file is served by a separate flagd (KubernetesFlagd) resource.

### Kubernetes Cluster

- The flagd that serves the file runs in the same namespace: pod volumes cannot cross namespaces.

## Deploy

### Console

Open the deployment store, find **flagd Flag File**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Release Flags** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesFlagdFlagFile
metadata:
  name: release-flags
  org: acme-corp
  env: prod
spec:
  namespace:
    value: feature-flags
  flags:
    new-checkout:
      state: ENABLED
      variants:
        "on":
          boolValue: true
        "off":
          boolValue: false
      defaultVariant: "off"
      targeting:
        if:
          - in:
              - var: org
              - - acme-corp
          - "on"
          - "off"
```

```shell
planton apply -f release-flags.yaml
```

This renders the `release-flags` ConfigMap with one flag, `new-checkout`, which is `on` for the organization `acme-corp` and `off` for everyone else. An Infra Job tracks the provisioning in real time.

### InfraChart

Deploy the file into the namespace its flagd runs in:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: feature-flags
      fieldPath: spec.name
```

The InfraPipeline renders the ConfigMap before the flagd that mounts it, so the daemon's first sync finds the file.

## Key Configuration

These are the most important decisions when configuring a flagd Flag File. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The default variant is the safe answer.** `defaultVariant` is what every caller gets when targeting does not choose a variant. For a release flag, make it the variant that hides the feature. Leaving it empty hands the decision to each caller's code default (reason `DEFAULT`, no value) -- useful when services ship their own fallbacks, risky when they disagree.

**One type per flag.** Every variant of a flag must hold the same value type -- booleans, strings, numbers or objects -- and the plan fails otherwise. A flag cannot change type later without breaking every caller that evaluates it with a typed SDK method; add a new flag instead.

**Targeting is JSONLogic.** `targeting` returns a variant name; `null` falls back to `defaultVariant`, and a name the flag does not define is an evaluation error. Use `fractional` for percentage rollouts (bucketed on the targeting key, so callers keep their variant), `sem_ver` for version gates, and `starts_with` / `ends_with` for string matching. Shared conditions belong in `evaluators`, referenced with `{"$ref": "<name>"}`; an evaluator cannot reference another.

**DISABLED is not off.** `state: DISABLED` makes evaluations return the caller's default value with reason `DISABLED`, ignoring `defaultVariant` and targeting. To switch a feature off for everyone, point targeting or `defaultVariant` at the off variant instead, so the answer stays the same for every caller.

**One file per owner.** A daemon can mount many flag files; when two define the same flag key, the source listed later in flagd's spec wins. Separate files per team keep ownership and review clean, and `metadata.flagSetId` lets SDKs and gRPC sync selectors take just one file's flags.

**The data key ends in .json.** flagd chooses its parser from the file extension, so `key` must end in `.json`; the flagd source's `key` must name the same value.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `config_map_name` | Name of the rendered ConfigMap | A flagd `configMap` source's `configMapName` |
| `key` | ConfigMap data key holding the definitions | The same source's `key` |
| `namespace` | Namespace of the rendered ConfigMap | Placing the flagd that mounts it (pod volumes cannot cross namespaces) |

## Common Patterns

**Release flags by organization.** An on/off flag that defaults to off and turns on for a list of organizations read from the evaluation context. Widening the release is a one-line edit. Start from the **Release Flags** preset.

**Percentage rollout with staff first.** A string or number flag where a shared evaluator always gives staff the new variant and `fractional` gives a slice of everyone else. Start from the **Percentage Rollout** preset.

**A file per service.** Each service's flags in their own file with a `flagSetId`, all mounted by one flagd, so each team reviews only its own changes.

## Works With

- [**flagd**](/infra-catalog/kubernetes-flagd) -- the daemon that mounts and serves this file through a `configMap` source
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the namespace the file shares with its flagd
- [**GO Feature Flag File**](/infra-catalog/kubernetes-go-feature-flag-flag-file) -- the alternative engine's flag file: query-language targeting, percentage and progressive rollouts, and scheduled changes, read through the Kubernetes API within one polling interval
