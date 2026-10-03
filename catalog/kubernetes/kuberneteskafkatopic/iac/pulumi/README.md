# KubernetesKafkaTopic Pulumi Module

## Module Behavior

- **Single untyped CustomResource**: the typed spec renders to one
  `kafka.strimzi.io/v1` KafkaTopic — the Strimzi CRDs type `config` with
  `x-kubernetes-preserve-unknown-fields`, which crd2pulumi cannot carry,
  so no generated package is shipped for the Kafka family (the same
  ruling as the KubernetesKafka module). The topic itself is created by
  the cluster's topic operator, not by this module.
- **Placement rendered from the spec**: the CR lands in the Kafka
  cluster's own namespace with the `strimzi.io/cluster` label — without
  the label the topic operator never picks the resource up.
- **No namespace resource, deliberately**: the namespace belongs to the
  KubernetesKafka resource's lifecycle.
- **Pinned topic name**: the exported `topic_name` output resolves
  exactly as the operator does (`spec.topic_name` when set, else
  `metadata.name`), so the handle can never drift. Twin of the
  Terraform module's locals, kept in lockstep.
- **No await machinery**: reconciliation belongs to the topic operator,
  not to applying the resource. Destroying the resource deletes the
  TOPIC AND ITS DATA (the topic operator propagates deletion).

## Usage

```bash
export IAC_INPUT=$(cat ../../e2e/manifest.yaml | base64)
pulumi up
```

## Local Development

```bash
make deps
make build
```

## Debug

```bash
bash debug.sh ../../e2e/manifest.yaml
```
