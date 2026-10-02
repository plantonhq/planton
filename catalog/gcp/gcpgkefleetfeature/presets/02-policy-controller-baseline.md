# Policy Controller Baseline

## Use Case

Install Policy Controller on every cluster in the fleet with Google's template library and the Pod Security Standards baseline bundle, auditing every minute.

## When to Use

- A fleet that must enforce or audit Kubernetes security standards consistently
- Starting a policy program from Google's bundles before writing your own constraints

## What This Creates

- The Fleet API and the Policy Controller API on the fleet host project
- The policycontroller feature, with Policy Controller as the fleet-wide default

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `...policyContent.bundles` | `pss-baseline-v2022` | Other Google bundles, e.g. `cis-k8s-v1.5.1`, `policy-essentials-v2022`. |
| `...installSpec` | `INSTALL_SPEC_ENABLED` | `INSTALL_SPEC_SUSPENDED` to install with webhooks off while you tune. |
| `...exemptableNamespaces` | `kube-system` | Namespaces Policy Controller should skip. |
