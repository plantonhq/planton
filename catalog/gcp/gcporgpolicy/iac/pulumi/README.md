# GCP Organization Policy - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying one Google Cloud organization policy using Planton's `GcpOrgPolicy` API. The module is written in Go and creates `orgpolicy.Policy` — the rules for one constraint at one scope — through the bridged Google provider.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: `roles/orgpolicy.policyAdmin` on the scope (project, folder, or organization)

## Directory Structure

```
iac/pulumi/
├── main.go            # Pulumi program entry point
├── Pulumi.yaml        # Pulumi project configuration
├── README.md          # This file
└── module/
    ├── main.go        # Module coordinator
    ├── org_policy.go  # The policy and its two rule sets
    ├── locals.go      # Resolved constraint and rendered parent
    └── outputs.go     # Stack output constants
```

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`.

## What the module does

- Renders the parent from the spec's `scope` arm (`projects/{id}`, `folders/{id}`, `organizations/{id}`); an empty scope resolves the provider's default project once through the provider's client config.
- Assembles the policy name, `{parent}/policies/{constraint}`, from the scope and the constraint — a predefined constraint by literal, or a custom constraint by reference to a `GcpOrgPolicyCustomConstraint`'s `constraint` output.
- Maps `policy` onto the provider's `spec` block and `dry_run_policy` onto `dry_run_spec`, sending optional scalars only when set.
- Exports `name` and `etag`.

## Parity with Terraform

Both engines send the same arguments and produce the same two outputs. One `PARITY` note applies to both: Google's API models a rule's verdict as booleans in a one-of and the provider flattens that into the strings `"TRUE"` / `"FALSE"` / unset; the spec keeps the API's shape and each module renders the string for exactly the arm that is set (`tristate` here, `locals.tf` on the Terraform side), so `enforce: false` reaches Google as `"FALSE"` and an unset arm is never sent.

## Notes

- **One policy per constraint per scope** — the name is the identity; both halves are immutable and a change to either recreates the policy.
- **Boolean versus list constraints** — a boolean constraint takes `enforce`; a list constraint takes `values`, `allow_all`, or `deny_all`. Google rejects a mismatch at apply time with a clear message; the constraint's type is not knowable offline.
- **Rules are set-compared** by the provider, so their order never causes a spurious update.
