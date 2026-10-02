# BigLake and Remote Models

## Use Case

The connection behind BigLake and object tables over Cloud Storage and remote models on Vertex AI: Google creates a service account for it, and you grant that account access to the bucket and the models.

## When to Use

- BigLake or object tables over Cloud Storage
- `CREATE MODEL ... REMOTE WITH CONNECTION` for Gemini and other Vertex AI models
- Remote functions on Cloud Run

## What This Creates

- A cloud-resource connection in `US` that outputs the service account to grant

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `US` | Match the datasets that will use it. |
| grants | none | Give `cloud_resource_service_account_id` `roles/storage.objectViewer` on the bucket and `roles/aiplatform.user` for remote models. |
