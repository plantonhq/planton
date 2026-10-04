# KubernetesKafkaUser Terraform Module

## Module Behavior

- **Single resource**: one `kubectl_manifest` applying a
  `kafka.strimzi.io/v1` KafkaUser. The credentials Secret is generated
  by the cluster's user operator, not by this module.
- **Twin locals**: `locals.tf` renders the same spec shape the Pulumi
  module produces (null-prune idiom) — identical field mappings, kept in
  lockstep, so the two engines apply byte-equivalent declarations.
- **`kubectl_manifest` apply**: the alekc/kubectl provider needs no
  cluster connection at plan time — a user can be PLANNED before the
  Strimzi CRDs exist, which is what lets an Infra Chart deploy the
  operator, the cluster, and its users in one run.
- **Placement rendered from the spec**: the CR lands in the Kafka
  cluster's own namespace with the `strimzi.io/cluster` label — without
  the label the user operator never picks the resource up.
- **No namespace resource, deliberately**: the namespace belongs to the
  KubernetesKafka resource's lifecycle.
- **No secret material, structurally**: the module renders
  authentication TYPE, ACLs, and quotas — credentials are
  OPERATOR-generated, never module-declared, so nothing in the state is
  sensitive.
- **Honest `secret_name`**: exported as the user name for
  credential-bearing users, EMPTY for `tls-external` and
  authentication-less users — no Secret is generated for either.
- **No `wait_for` block, deliberately**: reconciliation belongs to the
  user operator, not to applying the resource.

## Usage

```bash
planton tofu apply --manifest kafka-user.yaml
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
| `namespace` | Namespace the KafkaUser resource lives in (the Kafka cluster's namespace) |
| `username` | Kafka principal name (`metadata.name`) |
| `secret_name` | The operator-generated credentials Secret (empty for tls-external users — no Secret is generated) |
