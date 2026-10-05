# KubernetesExternalDns Terraform Module

Terraform/OpenTofu module for the KubernetesExternalDns kind: installs
ExternalDNS from the official Helm chart (`external-dns` at
https://kubernetes-sigs.github.io/external-dns/) with the DNS provider fully
decoupled from the host cluster.

## Module Behavior

- **One Helm release named after `metadata.name`** — multiple ExternalDNS
  instances per cluster (one per DNS provider / zone set, separated by TXT
  owner IDs) are a first-class pattern, so nothing here is a fixed chart
  name. The chart's fullname is pinned to the release name too, so every
  chart object (Deployment, Service, RBAC, ServiceAccount) carries a
  deterministic, manifest-derived name.
- **The typed spec renders into chart values** (`locals.helm_values`), and
  the spec's `helm_values` escape hatch is passed as a SECOND values
  document that the provider merges over the first with Helm `-f`
  semantics — the exact semantic twin of the Pulumi module.
- **The `dns_provider` arm selects the upstream provider**
  (`aws` / `google` / `azure` / `azure-private-dns` / `cloudflare` /
  `webhook` / `inmemory`) and assembles that arm's CLI flags (`extraArgs`)
  in a fixed order, its credential env wiring, and — for GCP keys and
  Azure — the credential file mounts.
- **Declared credentials materialize as Kubernetes Secrets** with
  deterministic names, so the credential itself never appears in chart
  values or pod specs:
  - `<name>-cloudflare-credentials` (`api-token` → `CF_API_TOKEN`)
  - `<name>-aws-credentials` (`access-key-id`/`secret-access-key` → env)
  - `<name>-gcp-credentials` (`credentials.json`, mounted with
    `GOOGLE_APPLICATION_CREDENTIALS` pointing at it)
  - `<name>-azure-config` (the rendered `azure.json`, mounted at the
    controller's default config path — created whenever the `azure_dns`
    arm is set, since even keyless Azure modes read identity from it)
- **Keyless authentication** rides `workload_identity`: the chart-owned
  ServiceAccount gets the per-cloud annotation (GKE Workload Identity,
  EKS IRSA, AKS Workload Identity — AKS also gets the required pod label).
- **The install waits for the controller Deployment to become Available**
  (`wait`/`atomic`/`cleanup_on_fail`). Note the controller validates
  provider CREDENTIALS at its first sync, not at startup — a live install
  with wrong credentials still installs green and surfaces in controller
  logs, matching upstream behavior.

## Resources

| Resource | Condition |
|---|---|
| `kubernetes_namespace_v1.external_dns` | `spec.create_namespace` |
| `kubernetes_secret_v1.cloudflare_credentials` | `cloudflare` arm set |
| `kubernetes_secret_v1.aws_credentials` | `aws_route53` arm with static keys |
| `kubernetes_secret_v1.gcp_credentials` | `google_cloud_dns` arm with a key |
| `kubernetes_secret_v1.azure_config` | `azure_dns` arm set |
| `helm_release.external_dns` | always |

## Outputs

| Output | Meaning |
|---|---|
| `namespace` | Namespace ExternalDNS is installed in |
| `release_name` | Helm release name (= `metadata.name`) |
| `service_account_name` | Controller ServiceAccount — the subject cloud-side keyless bindings reference |

## Parity

Kept in lockstep with the Pulumi module (`iac/pulumi/module/`): same chart
identity, same values rendering (byte-identical `extraArgs` ordering, env
wiring, and `azure.json` document), same credential Secret names and keys,
same outputs. Conditional objects use the null-prune idiom throughout so
numbers and booleans keep their types in the rendered values.
