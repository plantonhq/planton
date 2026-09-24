# Valkey and Neo4j mint their own passwords

**Date**: September 17, 2026
**Type**: Feature
**Components**: `KubernetesValkey` and `KubernetesNeo4j` (spec, both IaC modules, outputs, presets, e2e scenarios, guides), the forge and update rules, the anatomy baseline, the reference eval bank

## Summary

Declaring a Valkey ACL user used to require a password, and declaring a Neo4j server without one left the Helm chart to generate a credential it logged once and nothing could find afterwards. Both kinds now do what the operator and CloudNativePG already do: a credential the module can mint is minted. A Valkey user declared without a password gets a module-generated one; a Neo4j server declared without `auth` gets a module-generated admin password. Either way the credential lands in the `<name>-auth` Secret the kind already documents and is reported as an output, so a manifest carries no secret and no placeholder, nobody escrows a value, and clients read the Secret the outputs name.

## What Changed

- **KubernetesValkey**: `spec.auth.users[].password` is optional (still `sensitive`); a user needs only a name. Terraform adds `random_password.user` (for_each keyed by username over the password-less users; 32 letters and digits; generation shape under `ignore_changes`) and the `random` provider; the `<name>-auth` Secret's data is one `merge` of generated under declared. Pulumi adds the `random.RandomPassword` twin per username (`auth-password-<username>`). `password_secret` keeps its shape. The `minimal` e2e scenario declares `default` with no password; `behavioral-persistence` mixes one declared and one generated user. Both presets drop `<your-password>`. The import map gains the by-value companion row. The Pulumi module gains its README; the anatomy baseline loses its `missing-iac-readme` line for this kind.
- **KubernetesNeo4j**: the module materializes the `<name>-auth` Secret for every arm but `existing_secret` -- the declared password, or one `random_password.admin` / `random.RandomPassword` generates (24 letters and digits, two of each class) when `auth` is empty -- carrying the chart's `NEO4J_AUTH: neo4j/<password>` pair and a bare `password` key. `auth_secret_name` is always set; a new `password_secret` output (`KubernetesSecretKey`) names the bare-password key for the module-owned arms and is unset for `existing_secret`. The `minimal` e2e scenario declares no `auth`, and the verifier now reads the minted credential and runs the Cypher proof on it (an empty Secret name is a verifier defect, not a skip). The dev and GraphRAG presets drop their `change-me` passwords; the production preset keeps `existing_secret`. The kind gains a `GUIDE.md`.
- **Teaching**: both kinds' guides carry the credential law in their own voice; every README, catalog page, module README, preset sidecar, and control reference says the same thing; the eval bank gains a question (q23) checked against both guides and both reference pages.
- **Framework**: the forge rule gains "Credentials the module can mint are minted, never required" and the update rule gains the matching update bullet, so the next kind is born right and the next audit can grep the violation.

## Why It Matters

Every adopter who declared a Valkey with auth invented and escrowed a password per user; every Neo4j declared without one shipped a credential only a pod log ever saw. Both modules already materialized a Secret and reported it -- they were one `random_password` away from doing the whole job. Now the default shape of both kinds is the honest one, and a platform composing them into an environment declares users, never passwords.

## Compatibility

Additive. A declared password is used exactly as before on both kinds. Consumers of Valkey's `password_secret` (Superset, Airflow, RayCluster) see no shape change. Neo4j's `auth_secret_name` is now always populated; a consumer that relied on it being empty for the absent-auth arm was relying on a credential it could not read.
