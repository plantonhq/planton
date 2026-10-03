# Kubernetes: Wrong Reference-Picker Defaults Fixed on Airflow, Kafka UI, and Keycloak

**Date**: August 7, 2026
**Type**: Fix
**Provider**: Kubernetes
**Kind(s)**: `catalog/kubernetes/kubernetesairflow`, `catalog/kubernetes/kuberneteskafkaui`, `catalog/kubernetes/kuberneteskeycloak`

## Summary

The console's reference quick-picks derive from each field's foreign-key
default annotation, and three fields carried defaults pointing at kinds
that cannot produce the referenced value. All three are fixed at the
spec level, so every deploy wizard and manifest author now gets the
correct suggestion.

- **KubernetesAirflow** — the log search backend's password Secret
  (`spec.logging.elasticsearch` / `spec.logging.opensearch`) offered a
  KubernetesPostgres application Secret. The field reused the database
  block's password-secret message, inheriting its Postgres default. It
  now has its own message defaulting to a **KubernetesOpenSearch**
  resource's operator-generated admin-credentials Secret
  (`status.outputs.admin_credentials_secret_name`, keys
  `username`/`password`) — the same grain as the sibling `host` field,
  which already defaulted to the OpenSearch service. The field comments
  teach when that output is empty (custom security config) and that the
  elasticsearch arm, having no in-catalog kind, takes an explicit Secret
  name.
- **KubernetesKafkaUi** — the schema-registry and Kafka Connect HTTP
  Basic password fields offered a KubernetesKafkaUser credential Secret.
  No catalog kind produces those credentials (Karapace and KafkaConnect
  export endpoints, not auth Secrets), so those two contexts now use a
  dedicated selector with **no default** — name the Secret you created.
  The Kafka SASL password keeps its correct KafkaUser default.
- **KubernetesKeycloak** — `spec.additional_options[].secret`, the
  arbitrary server-option escape hatch, offered a KubernetesPostgres
  application Secret. Server options are arbitrary by nature, so the
  field now uses a dedicated selector with **no default**. The database
  username/password selectors keep their correct Postgres defaults.

All retypes are wire-compatible (identical field numbers and shapes);
no manifest changes are needed — every shipped preset and example sets
these fields with explicit values.

## Why the class existed

Foreign-key defaults live on message fields, and messages get reused: a
message whose default is right in one context silently carries that
default into every other context that reuses it. Annotation inventories
cannot see the defect — it hides in the reuse, not in any single
annotation. The update workflow now teaches the class: a shared
FK-defaulting message may be reused only where its default kind is the
natural source in every consuming context; otherwise the message is
split.

## Validation

Spec tests green for all three kinds (including a new
composed-from-OpenSearch reference-shape case on the Airflow log
backend); `planton validate-refs --check` machine-verifies the new
OpenSearch edge against its outputs; `planton secret-coverage --check`
green repo-wide; targeted package and Pulumi entrypoint builds green;
catalog reference pages regenerated (the FK tables and the referenced
kinds' inbound-edge sections follow the fix). Terraform variable
surfaces are unchanged — the password-secret object shapes are
identical before and after.
