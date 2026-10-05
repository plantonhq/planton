# KubernetesKafkaConnector Terraform Module

Deploys one `kafka.strimzi.io/v1` `KafkaConnector` custom resource —
one declared data pipe on an existing KubernetesKafkaConnect cluster.
The typed spec renders into the CR body in `locals.tf`
(`local.connector_manifest`) — the exact twin of the Pulumi module's
`main.go` (same keys rendered and omitted, numbers as numbers).

## Module Behavior

- **One resource, deliberately**: a single `kubectl_manifest` for the
  KafkaConnector. No namespace resource — the namespace belongs to
  the KubernetesKafkaConnect resource's lifecycle (the placement
  contract: connectors live in their Connect cluster's own
  namespace).
- **The binding label rides on top of the identity labels** —
  `strimzi.io/cluster` renders from `spec.connect_cluster`; without
  it the cluster operator never picks the connector up.
- **CRs apply through `kubectl_manifest` (alekc/kubectl)** — no
  cluster connection needed at plan time, so an Infra Chart can plan
  the operator, the Connect cluster, and its connectors in one run.
- **No wait_for block** — reconciliation belongs to the cluster
  operator; a "class not found" condition surfaces on the resource
  status, never as an apply failure.
- **The offset ConfigMap targets are declarations only** — the
  list/alter actions run when the resource carries the
  `strimzi.io/connector-offsets` annotation (an operational verb
  outside this module's scope).
- **Null-prune idiom** — conditional entries are `key = cond ? value
  : null` inside one object literal and pruned, so tasksMax and
  maxRestarts stay numbers in the rendered CR.

## Resources

| Resource | Condition |
|---|---|
| `kubectl_manifest.connector` | always |

## Usage

```bash
planton tofu apply --manifest connector.yaml
```

## Local Development

```bash
terraform init
terraform validate
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Inputs

See `variables.tf` for the full variable specification (generated
from the spec proto). The spec arrives from the proto→tfvars
converter in snake_case with the `namespace`, `connect_cluster` and
offset-ConfigMap references resolved to literal strings before
Terraform runs.

## Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Namespace the KafkaConnector lives in (the Connect cluster's namespace) |
| `connector_name` | Connector name (`metadata.name`) — what the Connect REST API and `connect-<name>` consumer groups key off |

## Parity

Kept in lockstep with the Pulumi module (`../pulumi/module/`): same
`kafka.strimzi.io/v1` apiVersion, same CR body (class, tasksMax,
version, config, state, autoRestart, the offset declarations), same
binding-label rendering, same outputs.
