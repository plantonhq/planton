# Encrypted With Experiments

## Use Case

A TensorBoard for a model line whose metrics are sensitive and long-lived: encrypted under your key, with the standing experiments and a production baseline run declared so dashboards and evaluation pipelines can point at them by id from day one.

## When to Use

- Metrics or sample images that fall under key-custody rules
- Evaluation pipelines that compare every candidate against a fixed baseline run
- Any TensorBoard a destroy must never delete

## What This Creates

- A TensorBoard in `us-central1` encrypted under the `GcpKmsKey` named `vertex-ai-key` (the Vertex AI Service Agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on it)
- Experiments `ranking-v2` (with a `production-baseline` run) and `ranking-v3`
- `deletionPolicy: PREVENT`, fanned to every experiment and run

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `kmsKeyName` | `vertex-ai-key` reference | Your key, in the TensorBoard's region (fixed at creation). |
| `experiments` | two experiments, one baseline run | Your standing experiments; ids are fixed once created. |
| `deletionPolicy` | `PREVENT` | `ABANDON` to let the block leave management while the history stays. |
