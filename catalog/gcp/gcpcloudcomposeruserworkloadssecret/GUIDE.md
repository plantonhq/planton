# GcpCloudComposerUserWorkloadsSecret Guide

The judgment this guide protects: this is how credentials reach DAGs
without living in DAG code, environment variables, or the Airflow UI.
The Secret is managed as infrastructure — created, rotated, and
destroyed through IaC — while Composer materializes it as a Kubernetes
Secret in the environment's workloads namespace.

## Base64 is the contract, not a suggestion

Every `data` value must be base64-encoded
(`echo -n 'postgresql://...' | base64`) — the API rejects raw values,
and the spec validates the encoding pre-deploy. The comment trap:
`echo` WITHOUT `-n` encodes a trailing newline into the credential,
which then fails authentication at 2 a.m. with no visible difference
in the manifest.

## Rotation is a data edit, in place

`data` is the only mutable field — rotating a credential is a value
change and an apply; consumers pick it up on their next pod start or
connection re-read. Name, environment, region, and project are all
immutable: a rename is a new Secret plus a DAG reference update, so
name Secrets after their ROLE (`orders-db-credentials`), never their
current value or owner.

## Consumption happens by name, in Airflow's own terms

KubernetesPodOperator tasks mount or env-inject the Secret by its
`secretName`; Airflow connections read it through the secrets backend
convention (`connections/<name>` keys). The `secret_name` output is
the value downstream wiring should reference — chart compositions pass
it, never a re-typed literal.

## PREVENT protects pipelines, not data

The Secret's material exists in your secret source of truth; what a
destroy breaks is every DAG consuming it. `deletionPolicy: PREVENT`
fits Secrets live pipelines depend on — the destroy fails instead of
silently breaking the night's runs. `ABANDON` leaves the Kubernetes
Secret in the cluster for a management handover.

## What is deliberately absent

No `google_project_service` enablement — the Composer API is
necessarily already on in the environment's project (a Secret cannot
exist without an environment). The decoded material never appears in
outputs. Pulumi holds the data as a state secret, and OpenTofu
encrypts it with the whole state file when given an encryption key.
