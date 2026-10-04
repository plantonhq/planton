# KubernetesKafkaTopic Terraform Module

## Module Behavior

- **Single resource**: one `kubectl_manifest` applying a
  `kafka.strimzi.io/v1` KafkaTopic. The topic itself is created by the
  cluster's topic operator, not by this module.
- **Twin locals**: `locals.tf` renders the same spec shape the Pulumi
  module produces (null-prune idiom) — identical field mappings, kept in
  lockstep, so the two engines apply byte-equivalent declarations.
- **`kubectl_manifest` apply**: the alekc/kubectl provider needs no
  cluster connection at plan time — a topic can be PLANNED before the
  Strimzi CRDs exist, which is what lets an Infra Chart deploy the
  operator, the cluster, and its topics in one run.
- **Placement rendered from the spec**: the CR lands in the Kafka
  cluster's own namespace with the `strimzi.io/cluster` label — without
  the label the topic operator never picks the resource up.
- **No namespace resource, deliberately**: the namespace belongs to the
  KubernetesKafka resource's lifecycle.
- **No `wait_for` block, deliberately**: reconciliation belongs to the
  topic operator, not to applying the resource. Destroying the resource
  deletes the TOPIC AND ITS DATA (the topic operator propagates
  deletion).

## Usage

```bash
planton tofu apply --manifest kafka-topic.yaml
```

## Local Development

```bash
terraform init
terraform validate
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Inputs

See `variables.tf` for the full variable specification.

## Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Namespace the KafkaTopic resource lives in (the Kafka cluster's namespace) |
| `topic_name` | The actual Kafka topic name (`spec.topic_name` when set, otherwise `metadata.name`) |
