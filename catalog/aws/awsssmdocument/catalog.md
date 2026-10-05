# AWS SSM Document

Deploys a customer-owned AWS Systems Manager document: a reusable, versioned definition of commands to run on managed nodes or automation steps to drive against AWS APIs. State Manager associations, maintenance windows, Run Command, and Automation all execute documents — this component owns authoring them. Every content change creates a new document version and promotes it to the default, so consumers pinning `$DEFAULT` follow your releases automatically. Documents can be shared to other AWS accounts or publicly.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SSM Document** — the document body (JSON or YAML) with its declared parameters, document type, and optional target-type restriction. The document's name is `metadata.name`, and changing it forces replacement.
- **Sharing Permissions** — configured only when `shareWithAccountIds` is set; the module shares the document to the listed account IDs (or publicly via the single entry `All`) and automatically un-shares before delete.
- **Artifact Attachments** — configured only when `attachmentSources` is set, for document types that carry artifacts (Package installers). AWS never reads attachment metadata back, so the first apply after an import re-asserts them.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with permissions to manage SSM documents. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Nothing — documents stand alone. Executing them later needs managed nodes (Command, Session) or automation targets (Automation), but the document deploys without either.

## Deploy

### Console

Open the deployment store, find **AWS SSM Document**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields covering the document type, format, and content. Start from the **Shell Command Document** preset in the [Presets](#presets) tab for the workhorse shape: a parameterized shell script run on managed EC2 nodes.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSsmDocument
metadata:
  name: collect-diagnostics
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  documentType: Command
  documentFormat: YAML
  content: |
    schemaVersion: "2.2"
    description: Collect service diagnostics from managed nodes
    parameters:
      LogLines:
        type: String
        description: Number of journal lines to collect
        default: "200"
    mainSteps:
      - action: aws:runShellScript
        name: collectDiagnostics
        inputs:
          runCommand:
            - journalctl -n {{LogLines}} --no-pager
  targetType: /AWS::EC2::Instance
```

```shell
planton apply -f ssm-document.yaml
```

This creates a schema-2.2 Command document named `collect-diagnostics` with one defaulted parameter callers can override per run, restricted to EC2 instance targets. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a document. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The name is the identity** — `metadata.name` is the document name (3–128 characters of letters, digits, `_.-`), it is what associations reference, and changing it forces replacement. Downstream consumers hold the name, so treat a rename as a migration, not an edit.

**`documentType` decides the execution model** — Command documents run scripts on managed nodes; Automation documents drive multi-step AWS-API workflows (and are what maintenance window `AUTOMATION` tasks and rate-controlled association runbooks execute); Session documents configure Session Manager; Policy documents gather inventory. The type cannot express both models at once — an operation that touches AWS APIs belongs in an Automation document, not a shell script juggling the CLI.

**Content updates version, never mutate** — Every content change creates a new document version and the module promotes it to the default, which is how `$DEFAULT`-pinned associations pick up releases with no edit on their side. One AWS rule to know: schema-1.x (legacy command) documents only accept updates when the content itself changed — a no-op content rewrite is rejected.

**`versionName` labels are immutable forever** — AWS rejects reusing a label on different content, per document, permanently. Treat them like git tags: bump per release, never recycle.

**Sharing is one field, and `All` means public** — `shareWithAccountIds` takes twelve-digit account IDs, or the single entry `All` to share the document with every AWS account. Review content for account-specific values before sharing broadly; un-sharing on delete is automatic.

**Deletes are asynchronous at AWS** — A deleted document reads as `Deleting` briefly, and re-creating the same name during that window fails. Rapid destroy/recreate cycles of the same name need a wait or a fresh name.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — the document body, type, and sharing list are all self-contained values. Consumers point at the document, not the other way around.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `document_name` | The document's name (also the provider's import ID) | The `documentName` reference on AwsSsmAssociation; maintenance window task ARNs |
| `document_arn` | The document's ARN | IAM policy statements scoping who may execute or update it |
| `default_version` | The version `$DEFAULT` pins resolve to; content updates promote the new version here | Verifying which release consumers are running |

`latest_version`, `document_hash`, and `status` are also exposed — observability echoes (the newest version number, the content's Sha256 digest, and the lifecycle state) useful for auditing what deployed, not as composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Fleet scripts through an audited channel** — A parameterized Command document replaces SSH-and-run for agent installs, cert rotation, and diagnostics: IAM gates who can execute it, and every run is recorded. Pair it with an AwsSsmAssociation targeting by tag so new instances join coverage automatically. Start from the **Shell Command Document** preset.

**Automation runbooks for AWS-API procedures** — Stop-resize-restart, snapshotting, remediation: multi-step Automation documents with typed parameters, executable unattended from a maintenance window or rate-controlled across a fleet via an association's `automationTargetParameterName`. Add `aws:approve` steps for human gates on destructive operations. Start from the **Automation Runbook** preset.

**Cross-account distribution** — One account authors and versions the runbook; `shareWithAccountIds` distributes it to sibling accounts, which execute it by name. The trade against copy-per-account is a single release process — and a single blast radius when a release is bad, so pair broad sharing with deliberate `versionName` labels.

## Works With

- [**AWS SSM Association**](/infra-catalog/aws-ssm-association) — binds this document to targets on a schedule, wired via the `documentName` reference
- [**AWS SSM Maintenance Window**](/infra-catalog/aws-ssm-maintenance-window) — executes this document as a registered task inside a defined change window
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — hosts artifact files that Package documents attach via `attachmentSources`
