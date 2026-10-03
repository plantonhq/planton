# GCP Pub/Sub Topic IAM Member - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying additive topic-scoped IAM grants using Planton's `GcpPubSubTopicIamMember` API. The module is written in Go and uses the Pulumi GCP provider to create `pubsub.TopicIAMMember` resources (backed by `google_pubsub_topic_iam_member`).

The grant is ADDITIVE: it merges one (role, member) pair into the topic's IAM policy without touching any other member's bindings, and destroy subtracts only this exact pair.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **An existing topic** — the grant is policy metadata; no billed infrastructure is created
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── README.md         # This file
└── module/
    ├── main.go        # Module coordinator
    ├── iam_member.go  # IAM member grant creation
    └── outputs.go     # Output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `iac-input.yaml` with the grant specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpPubSubTopicIamMember
  metadata:
    name: audit-sink-publisher
  spec:
    topic:
      value: projects/my-gcp-project-123/topics/audit-events
    role:
      value: roles/pubsub.publisher
    member:
      value: serviceAccount:service-123456789@gcp-sa-logging.iam.gserviceaccount.com
```

### 3. Deploy

```bash
export IAC_INPUT_FILE_PATH=iac-input.yaml
pulumi up
```

### 4. Destroy

```bash
pulumi destroy
```

Destroy removes exactly this (role, member) pair from the topic's policy — no other grant is touched, and the topic is never affected.

## Outputs

| Output | Description |
|--------|-------------|
| `topic` | The topic whose policy received the grant |
| `role` | The granted role |
| `member` | The granted member |
| `etag` | The topic IAM policy etag after the grant |

## Behavior Notes

- Every argument is immutable: any spec change replaces the grant (the IAM API has no grant update).
- There is no project argument — the provider reads the project from the topic's full name.
- No condition block is ever sent: Pub/Sub topics do not accept conditional role bindings.
- The member format is validated on literal values; a referenced `writer_identity` or `member` output already arrives in `serviceAccount:<email>` form.
