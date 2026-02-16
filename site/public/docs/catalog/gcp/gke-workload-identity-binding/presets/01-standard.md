---
title: "Standard Workload Identity Binding"
description: "This preset creates a Workload Identity binding that allows a Kubernetes ServiceAccount (KSA) to impersonate a Google ServiceAccount (GSA). This is the GCP-recommended way for GKE pods to..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "gke-workload-identity-binding"
componentTitle: "GKE Workload Identity Binding"
provider: "gcp"
icon: "package"
order: 1
---

# Standard Workload Identity Binding

This preset creates a Workload Identity binding that allows a Kubernetes ServiceAccount (KSA) to impersonate a Google ServiceAccount (GSA). This is the GCP-recommended way for GKE pods to authenticate to Google Cloud APIs without managing service account keys.

## When to Use

- Any GKE pod that needs to access GCP services (Cloud SQL, Secret Manager, Cloud Storage, Pub/Sub)
- cert-manager pods that need DNS01 challenge resolution via Cloud DNS
- external-dns pods that need to manage Cloud DNS records
- Application workloads that read/write to GCS buckets or Pub/Sub topics

## Key Configuration Choices

- **No service account key** -- Workload Identity eliminates the need for exported JSON keys, reducing secret sprawl
- **One binding per KSA-GSA pair** -- each Kubernetes ServiceAccount maps to exactly one Google ServiceAccount
- **Namespace-scoped** -- the binding is specific to a Kubernetes namespace, following least-privilege principles

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project hosting the GKE cluster | `GcpProject` outputs |
| `<gsa-email>` | Email of the Google Service Account (e.g., `my-app@project.iam.gserviceaccount.com`) | `GcpServiceAccount` status outputs |
| `<kubernetes-namespace>` | Kubernetes namespace where the KSA lives (e.g., `default`, `cert-manager`) | Your Kubernetes cluster |
| `<kubernetes-service-account>` | Name of the Kubernetes ServiceAccount (e.g., `cert-manager`, `my-app`) | Your Kubernetes cluster |
