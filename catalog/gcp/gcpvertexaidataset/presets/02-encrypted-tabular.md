# Encrypted Tabular

## Use Case

A tabular dataset for AutoML tabular or custom training on regulated data: rows imported from BigQuery or Cloud Storage, encrypted under your own key, and protected from an accidental destroy.

## When to Use

- Training data classified confidential or subject to key-custody rules
- Tabular models (churn, fraud, forecasting inputs) whose training set must stay reproducible
- Any dataset a destroy must never delete

## What This Creates

- An empty tabular dataset in `us-central1` with Google's `tabular_1.0.0.yaml` schema
- Encryption under the `GcpKmsKey` named `vertex-ai-key` (the Vertex AI Service Agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on it)
- `deletionPolicy: PREVENT`, so a destroy fails instead of deleting the data

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `kmsKeyName` | `vertex-ai-key` reference | Your key, in the dataset's region (fixed at creation). |
| `labels` | `data-classification: confidential` | Your classification scheme. |
| `deletionPolicy` | `PREVENT` | `ABANDON` to let the block leave management while the data stays. |
