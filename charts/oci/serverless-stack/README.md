# Serverless Stack InfraChart

This chart provisions an **event-driven serverless platform on Oracle Cloud Infrastructure**:

* Custom VCN with internet, NAT, and service gateways
* Public subnet for the API Gateway (internet-facing)
* Private subnet for OCI Functions (secure backend access)
* Functions application for deploying serverless functions
* API Gateway for HTTP routing to functions
* Object Storage bucket for artifacts and data
* Centralized log group for function and gateway logs

## Resources Created

| Resource | Kind | Condition |
|----------|------|-----------|
| Virtual Cloud Network | `OciVcn` | Always |
| Public Subnet | `OciSubnet` | Always |
| Private Subnet | `OciSubnet` | Always |
| Functions Application | `OciFunctionsApplication` | Always |
| API Gateway | `OciApiGateway` | Always |
| Object Storage Bucket | `OciObjectStorageBucket` | Always |
| Log Group | `OciLogGroup` | Always |

## Parameters

| Name | Description | Default |
|------|-------------|---------|
| `compartment_ocid` | OCI compartment OCID | — |
| `vcn_cidr` | VCN CIDR block | `10.0.0.0/16` |
| `public_subnet_cidr` | Public subnet CIDR | `10.0.0.0/24` |
| `private_subnet_cidr` | Private subnet CIDR | `10.0.1.0/24` |
| `app_name` | Functions app name | `serverless-api` |
| `functions_shape` | generic_x86 / generic_arm | `generic_x86` |
| `gateway_name` | API Gateway name | `api-gateway` |
| `bucket_name` | Artifact bucket name | `serverless-artifacts` |
| `log_retention_days` | Log retention (days) | `30` |

## Architecture

```
Internet
   │
   ▼
┌──────────────────────────────┐
│  Public Subnet               │
│  • API Gateway (HTTP routes) │
│  → Internet Gateway          │
└──────────┬───────────────────┘
           │ invokes
┌──────────▼───────────────────┐
│  Private Subnet              │
│  • Functions Application     │
│  → NAT Gateway (outbound)    │
└──────────────────────────────┘
           │ reads/writes
┌──────────▼───────────────────┐
│  Object Storage Bucket       │
│  Log Group                   │
└──────────────────────────────┘
```
