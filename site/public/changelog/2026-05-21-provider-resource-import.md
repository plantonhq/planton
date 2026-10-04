---
title: "Import Existing Provider Resources into Planton"
date: 2026-05-21
category: feature
tags:
  - infra-hub
  - console
  - cli
excerpt: "Bring existing cloud infrastructure under Planton management without recreating it — a guided wizard, provisioner-specific CLI commands, and real-time progress tracking across Pulumi, Terraform, and OpenTofu."
author:
  - name: Swarup Donepudi
    title: Founder
---

You can now import existing provider resources into Planton without recreating them. If you have a DNS zone that was auto-created when you purchased a domain, a VPC that was provisioned manually before your team adopted Planton, or a database that already exists and just needs to be tracked — import brings it under management by writing to its Infra Component's IaC state file. Your actual cloud infrastructure is never modified. Nothing is created, changed, or destroyed.

## Importing from the Console

A new import wizard walks you through the process in three steps. The first step explains what import does and shows exactly which state backend will be written to — so you know where state changes land before you proceed. The second step collects the resource identifiers your IaC provisioner needs (resource type, name, and provider ID for Pulumi; resource address and provider ID for Terraform and OpenTofu). The third step shows a command preview and lets you confirm before the import runs.

To start an import, click the **Import** button in the action bar on any Infra Component detail page. The wizard is available in both the web console and the desktop app.

## Importing from the CLI

Three provisioner-specific commands give you the same capability from the terminal:

```bash
# Pulumi: specify resource type, logical name, and provider ID
planton pulumi import <infra-component> \
  --type cloudflare:index/zone:Zone --name main --id <zone-id>

# Terraform / OpenTofu: specify resource address and provider ID
planton terraform import <infra-component> --address cloudflare_zone.main --id <zone-id>
planton tofu import <infra-component> --address cloudflare_zone.main --id <zone-id>
```

All three commands support `--dry-run`, which exercises the full resolution path — Infra Component lookup, provisioner detection, credential verification — and prints the equivalent native CLI command without creating an Infra Job. Use it to verify your arguments before committing.

If you run an import command that does not match the Infra Component's actual provisioner, the CLI warns you. The warning is informational — the platform determines the actual provisioner from the Infra Component's configuration, not the CLI command path.

## Progress Tracking

Import runs as an Infra Job with three stages — initialize, import, and refresh — each reporting real-time status through the same progress interface used by every other infrastructure operation. You can monitor progress from the Infra Job detail page in the console or by streaming events from the CLI.

## Why This Matters

- **Safe adoption** — import writes to the IaC state file only; your cloud infrastructure is never modified, and failed imports leave state unchanged
- **Three provisioners** — Pulumi, Terraform, and OpenTofu are all supported with the correct identifiers for each
- **Preview before commit** — `--dry-run` shows exactly what will happen before an Infra Job is created
- **Familiar workflow** — import Infra Jobs behave like every other Infra Job on the platform, with the same progress tracking, logs, and audit trail
- **After import** — run an Update operation to reconcile Planton's configuration with the actual resource state, and the full lifecycle (updates, drift detection, teardown) applies from that point forward
