# KubernetesAgentSubstrate Terraform Module

Installs the Agent Substrate control plane of the release `spec.version` names (v1.0.0 when unset) straight from what upstream publishes: the tag's own files and definitions, and every image pinned by the digest upstream's registry serves (`images.tf`), composed the way upstream's installer composes them. A new upstream release needs nothing from Planton. The typed spec is applied onto upstream's own objects in `documents.tf`; nothing is rebuilt from scratch, so every setting left unset keeps upstream's value. The exact semantic twin of the Pulumi module (`../pulumi/module/`); a gated test plans this module and holds every document and output equal to the Pulumi composition.

## How the Documents Are Built

1. `data "http" "release_file"` reads every file the composition reads at the release's tag, upstream's database setup script included. A file the tag does not carry lands in `files_missing`; the composition then reads the module's own release (`default_release_file`) and the preflight refuses the apply, while a destroy, which evaluates no precondition, proceeds. `images.tf` maps each image placeholder to its image and reads each image's digest from the registry (an anonymous pull token, then a `HEAD` on the tag's manifest).
2. Each file's octal file-mode literals (`defaultMode: 0600`) are rewritten as decimal first, because Kubernetes reads YAML 1.1 octal and `yamldecode` follows YAML 1.2. The text is then split on separator lines, each piece getting back the newline the split consumed, and every document keyed by its composed identity, `apiVersion//kind//name[//namespace]`, which is also what `../import-map.yaml` reads.
3. The resource definitions go through the catalog's kept-definition block (`helm_crds.tf`, bundle branch, upstream's three definition files at the release's tag; the upgrade rule judges a move to an older release's definitions).
4. Typed settings patch upstream's objects (`locals.tf` resolves them with upstream's defaults). A flag upstream already passes keeps its position and takes the new value; any other flag is appended, only when it differs from upstream's default. Upstream's own components (agentgateway, cordon, size10, the Cloud SQL sidecar) apply as upstream applies them, read from the release; its text markers are filled from `splices.tf`.
5. Each workload carries the ownership label, and a digest of the spec-derived configuration it reads, so a change rolls it; the API server also carries upstream's environment digest and the installed-version stamp.

## Generated Secrets

`pools.tf` generates the five certificate authorities and the actor token signing key with the TLS provider and writes them in upstream's pool format, and the key's public key set byte for byte as upstream derives it; it is self-contained so a test applies it alone and checks its pools and key set with upstream's own code. `secrets.tf` writes them, the API server's environment and an external database's server CA as `kubernetes_secret_v1` resources. A re-apply never replaces them; a destroy does.

## Order and Refusals

Every refusal fires at plan time, before anything is written. The preconditions sit on `terraform_data.preflight` (`preflight.tf`), whose values are all known at plan, and conditions are not evaluated on destroy. No data source depends on a resource, which would defer its read to apply.

- **A namespace other than upstream's** — a `namespace` reference that resolved to another name.
- **A workload someone else created** — `data "kubernetes_resources" "owned_workload"` reads each Deployment, DaemonSet and StatefulSet by its fixed name; one without the `planton.ai/kind=KubernetesAgentSubstrate` label belongs to another installation.
- **A namespace's owner** — `ate-system` (its owner is a choice) and `podcertificate-controller-system` (always this resource's).
- **A minor-release move** — the installed-version stamp on the API server, and the minor stamp on the bundled database's volume, against upstream's upgrade policy.
- **A definition someone else owns** — `helm_crds.tf`.
- **A release this module cannot install** — a file or image the release does not publish, or an image placeholder this module does not know (`files_missing`, `images_missing`, `unknown_placeholders`).
- **A release that changed what the module edits** — the workloads, containers, start hook and component patches the module patches must be where it expects them.

Then the apply order (`main.tf`): the namespaces, the kept definitions, the Secrets, and the four phases, each awaited: prerequisites, database, system, runtime.

## Resources

| Resource | Condition |
|----------|-----------|
| `kubectl_manifest.namespace` | `ate-system` with `create_namespace`; `podcertificate-controller-system` always |
| `kubectl_manifest.helm_crds` | `crds.install` (default true); kept unless `crds.keep_on_uninstall` is false |
| `tls_private_key.ca`, `tls_self_signed_cert.ca` | Always, five each |
| `tls_private_key.actor_jwt` | Always |
| `kubernetes_secret_v1.*` | Always; `postgres_server_ca` only with an external database's server CA |
| `kubectl_manifest.prerequisites` | Always: RBAC, ConfigMaps, the admission policy, the pod-certificate controller |
| `kubectl_manifest.database` | Always: the OpenTelemetry ConfigMap; the bundled database with `database.bundled` |
| `kubectl_manifest.system` | Always: the control plane and its policies; the credential provider with `kubernetes_secrets`; the PodMonitoring on GKE |
| `kubectl_manifest.runtime` | Always: the default gVisor SandboxConfig |
| `terraform_data.preflight` | Always (writes nothing; carries the refusals) |

## Outputs

See `outputs.tf`; every output is described in `../../v1alpha1/outputs.proto`.
