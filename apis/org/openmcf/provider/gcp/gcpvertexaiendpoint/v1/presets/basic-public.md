# Preset: Basic Public Endpoint

**Rank**: 1 (most common)

## Use Case

The simplest Vertex AI Endpoint -- a public prediction URL with Google-managed encryption. Suitable for development, testing, or workloads where network isolation is not required.

## What This Creates

- One Vertex AI Endpoint accessible via the shared regional DNS
- Google-managed encryption (no CMEK)
- No private networking

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | Deploy closer to your users or meet data residency requirements |
| `displayName` | `My ML Endpoint` | Give your endpoint a meaningful name |
| `description` | (empty) | Add context for your team |

## Next Steps

After creating the endpoint, deploy a model to it using the Vertex AI API:

```bash
gcloud ai endpoints deploy-model ENDPOINT_ID \
  --region=us-central1 \
  --model=MODEL_ID \
  --display-name="v1"
```
