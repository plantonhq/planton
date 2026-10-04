# KubernetesOtelCollector Terraform Module

Deploys one operator-managed OpenTelemetry Collector: the optional
namespace and the single `opentelemetry.io/v1beta1`
`OpenTelemetryCollector` custom resource — the collector workload
(Deployment/DaemonSet/StatefulSet per mode, or sidecar registration),
the `<name>-collector` Service with receiver-derived ports, the
headless and monitoring Services, and the rendered config ConfigMap
are all operator-created from it. The CR applies through
`kubectl_manifest` (alekc/kubectl provider, server-side apply), which
needs no cluster connection at plan time — a collector can be planned
before the operator's CRDs exist, so an Infra Chart can deploy the
operator and its collectors in one run.

Prerequisite at apply time: a KubernetesOtelOperator on the cluster
(it watches every namespace).

## Module Behavior

- **The config document is parsed at plan** — the v1beta1 CR's
  `config` is a STRUCTURED object (not the v1alpha1 string), so
  `config_yaml` goes through `yamldecode` and embeds as an object; an
  unparseable document fails the plan loudly. The operator's admission
  webhook validates the collector semantics at apply.
- **The name budget fails loudly** — a lifecycle precondition rejects
  `metadata.name` past 42 characters: the operator derives child names
  by suffixing (`-collector-monitoring` is the longest stable suffix
  at 21 characters) and Kubernetes caps names at 63.
- **BACKGROUND deletion, explicitly** — the OPERATOR owns the
  collector CR's cascade; its ownership references tear down the
  workload, Services and ConfigMap. Foreground propagation would block
  the delete on children the operator keeps reconciling, so the module
  sets `delete_cascade = "Background"`.
- **No wait on the CR, deliberately** — collector readiness depends on
  the operator (webhook admission, image injection, workload rollout),
  which is not part of applying the resource; the verifier owns
  readiness.
- **Unset optionals are omitted** — the rendered CR body is
  null-pruned, so the operator's defaulting stays authoritative for
  everything the manifest leaves unsaid (mode defaults to deployment,
  an empty image gets the operator's default collector image, an empty
  service account gets an operator-created one).
- **Scaling renders only where it means something** — `replicas` and
  `autoscaler` render in deployment/statefulset modes only; the
  middleware-defaulted `replicas: 1` in daemonset/sidecar modes is
  deliberately ignored.
- **One volumes entry renders both halves** — each spec `volumes`
  entry splits into the CR's `volumes` (pod volume source) and
  `volumeMounts` (container mount), so the two lists can never
  disagree.
- **The module (not the operator) owns namespace creation** —
  `create_namespace` drives a `kubernetes_namespace_v1` resource
  carrying the standard governance labels.

## Resources

| Resource | Condition |
|---|---|
| `kubernetes_namespace_v1.namespace` | `spec.create_namespace` |
| `kubectl_manifest.otel_collector` | always |

## Usage

```bash
planton tofu apply --manifest kubernetes-otel-collector.yaml
```

## Local Development

```bash
terraform init
terraform validate
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Inputs

See `variables.tf` for the full variable specification (generated from
the spec proto). The spec arrives from the proto→tfvars converter in
snake_case with the `namespace` foreign key (KubernetesNamespace)
resolved to a literal string before Terraform runs.

## State Import

Existing deployments can be adopted into state. `kubectl_manifest`
uses the composed import ID `apiVersion//kind//name//namespace`; the
CR's name is `metadata.name`, so the component's `iac/import-map.yaml`
can derive the address blind.

## Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Namespace the collector runs in |
| `collector_name` | Name of the OpenTelemetryCollector resource (equals `metadata.name`) |
| `service` | Collector Service (`<name>-collector`) with receiver-derived ports — empty in sidecar mode |
| `otlp_grpc_endpoint` | In-cluster OTLP gRPC ingest endpoint (`<service>:4317`) — valid when the config declares the standard `otlp` receiver; empty in sidecar mode |
| `otlp_http_endpoint` | In-cluster OTLP HTTP ingest endpoint (`http://<service>:4318`) — same contract; empty in sidecar mode |
| `headless_service` | Headless Service (`<name>-collector-headless`) for per-pod addressing — empty in sidecar mode |
| `monitoring_service` | Monitoring Service (`<name>-collector-monitoring`, port 8888) — empty in sidecar mode |

## Parity

This module is the behavioral twin of the Pulumi module
(`../pulumi/module/`): same rendered CR body (null-pruned, sorted env,
defaults-on-declaration), same background-deletion posture, same name
budget, same outputs — keep them in lockstep.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
