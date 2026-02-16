# GcpVertexAiEndpoint: Research & Design Document

## Service Overview

Vertex AI Endpoints are the serving layer in Google Cloud's AI Platform. An endpoint is a stable URL that acts as a gateway for model prediction requests. It decouples the infrastructure (the endpoint itself, its networking, and encryption) from the operational concern of deploying and routing traffic to specific models.

### What an Endpoint Does

- Provides a stable prediction URL that persists across model deployments
- Manages traffic splitting between multiple deployed model versions
- Handles auto-scaling of prediction serving infrastructure
- Supports private networking (VPC peering and PSC) for secure inference
- Enables monitoring and logging of prediction requests

### What an Endpoint Does NOT Do

- It does not train models
- It does not manage model artifacts (that's Model Registry)
- It does not run batch predictions (that's BatchPredictionJob)
- Model deployment to the endpoint is a separate operational step

## Deployment Landscape

### Terraform: `google_vertex_ai_endpoint`

```hcl
resource "google_vertex_ai_endpoint" "endpoint" {
  name         = "1234567890"
  display_name = "My Endpoint"
  location     = "us-central1"
  network      = "projects/12345/global/networks/my-vpc"

  encryption_spec {
    kms_key_name = "projects/.../cryptoKeys/my-key"
  }
}
```

Key characteristics:
- `name` is Required, numeric-only, max 10 digits (no leading zeros)
- `network` and `private_service_connect_config` are mutually exclusive
- `dedicated_endpoint_enabled` conflicts with PSC
- Labels are supported (non-authoritative)
- `encryption_spec` is ForceNew (immutable after creation)
- Provider version: `~> 6.0` required

### Pulumi: `vertex.AiEndpoint`

```go
endpoint, _ := vertex.NewAiEndpoint(ctx, "endpoint", &vertex.AiEndpointArgs{
    DisplayName: pulumi.String("My Endpoint"),
    Location:    pulumi.String("us-central1"),
})
```

Key characteristics:
- `Name` is optional (auto-generated when omitted)
- Same mutual exclusion rules as Terraform
- Labels are supported
- Package: `github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/vertex`

## 80/20 Scoping Rationale

### Included (Covers 95%+ of Use Cases)

| Feature | Rationale |
|---------|-----------|
| Display name + description | Core identification |
| Location (region) | Required for all endpoints |
| VPC peering via `network` | Standard private networking |
| Private Service Connect | Modern private networking |
| CMEK encryption | Enterprise compliance |
| Dedicated endpoint DNS | Production performance |
| Framework GCP labels | Operational consistency |
| Endpoint name (optional) | Import/migration scenarios |

### Excluded (Deferred to v2)

| Feature | Rationale |
|---------|-----------|
| `traffic_split` | Model deployment is operational, not infrastructure |
| `predict_request_response_logging_config` | Observability tuning, configured post-deployment |
| `model_deployment_monitoring_job` | Only populated after model deployment |
| `psc_automation_configs` | TF-only feature, not in Pulumi SDK |
| `enable_secure_private_service_connect` | Pulumi-only, IAM-authorized PSC |

### Design Decisions

1. **Flattened `encryption_spec`** to `kms_key_name` -- single field doesn't need a wrapper message
2. **PSC as sub-message** (not flat bool) -- `project_allowlist` is essential for access control
3. **`endpoint_name` exposed as optional** -- TF requires numeric name; auto-generated when omitted
4. **No `model_deployment_monitoring_job` output** -- always empty from IaC; only populated after model deployment
5. **`dedicated_endpoint_enabled` added** -- not in original plan but important for production ML serving

## Networking Deep Dive

### Public Endpoints (Default)

Requests go to `{region}-aiplatform.googleapis.com`. Simple but shared infrastructure.

### VPC-Peered Endpoints

The endpoint gets a private IP within the peered VPC. Requires:
- Private Services Access configured on the VPC
- `network` field set to the fully qualified network path

### Private Service Connect

Modern approach using PSC service attachments. Benefits:
- No VPC peering required
- Fine-grained access control via `project_allowlist`
- Strongest network isolation
- Cannot combine with `dedicated_endpoint_enabled`

## Provider Feature Gaps

| Feature | Terraform | Pulumi | Our Component |
|---------|-----------|--------|---------------|
| `name` (numeric ID) | Required | Optional | Optional (auto-generated) |
| `psc_automation_configs` | Yes | No | Excluded |
| `enable_secure_private_service_connect` | No | Yes | Excluded |
| Labels | Yes | Yes | Yes (framework) |
| CMEK | Yes | Yes | Yes |

## References

- [Vertex AI Endpoints Documentation](https://cloud.google.com/vertex-ai/docs/predictions/overview)
- [Private Endpoints](https://cloud.google.com/vertex-ai/docs/predictions/using-private-service-connect)
- [Terraform Resource](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/vertex_ai_endpoint)
- [Pulumi Resource](https://www.pulumi.com/registry/packages/gcp/api-docs/vertex/aiendpoint/)
