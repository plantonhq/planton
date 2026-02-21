# OCI API Gateway Deployment Component

**Date**: 2026-02-20
**Type**: New Feature
**Components**: `apis/org/openmcf/provider/oci/ociapigateway/v1/`

## Summary

Added the OciApiGateway deployment component -- OCI's managed API endpoint bundled with a single API deployment providing route-to-backend mapping, JWT authentication, CORS, and rate limiting. Bundles `oci_apigateway_gateway` and `oci_apigateway_deployment` into one component. Second and final resource of Phase 7 (Serverless and Functions), completing the phase.

## Problem Statement / Motivation

OpenMCF's OCI provider had 28 resources covering networking, compute, containers, databases, storage, security, and one serverless resource (OciFunctionsApplication), but no way to expose OCI Functions or HTTP backends through a managed API endpoint. OCI API Gateway provides managed REST API endpoints with built-in authentication, CORS, and rate limiting -- critical for production serverless architectures. Without this component, platform teams cannot declaratively provision the API layer that sits in front of their functions, blocking the Serverless Stack infra chart.

## Solution / What's New

A complete OciApiGateway deployment component with both Pulumi (Go) and Terraform (HCL) modules, bundling the gateway infrastructure with one API deployment.

### Proto API

- **spec.proto**: 15 messages, 7 enums, 7 CEL validation rules, covering gateway + deployment specification
- **Key gateway fields**: compartment_id (StringValueOrRef), endpoint_type (enum: public/private), subnet_id (StringValueOrRef), display_name, certificate_id, network_security_group_ids
- **Key deployment fields**: path_prefix, display_name, logging_policies, request_policies (authentication, CORS, rate_limiting), routes (path, methods, backend, authorization, per-route logging)
- **Authentication**: JWT with remote JWKS or static keys (PEM + JWK formats), verify_claims, anonymous access
- **Backends**: 3 types -- HTTP (reverse proxy), Oracle Functions (serverless), Stock Response (static)
- **CEL rules**: JWT auth requires public_keys; remote JWKS requires URI; static keys requires non-empty keys; HTTP backend requires URL; Functions backend requires function_id; PEM key requires key value; JWK key requires kty; any_of authorization requires allowed_scope
- **api.proto**: Standard KRM wrapper with const-validated api_version and kind
- **stack_outputs.proto**: 3 outputs (gateway_id, hostname, deployment_endpoint)

### Design Decisions

- **Gateway + deployment bundled**: The most common pattern is one gateway serving one API. A gateway without a deployment is useless (no routes). Users needing multiple deployments per gateway can create multiple OciApiGateway components.
- **JWT authentication only for v1**: The deployment specification supports 3 auth modes (JWT, Custom/Functions-based, Token). JWT with Remote JWKS covers the most common serverless pattern (IdP validates tokens). Custom and Token auth excluded to keep complexity manageable.
- **3 of 5 backend types**: HTTP, Oracle Functions, and Stock Response cover the core use cases. Dynamic Routing and OAuth2 Logout backends excluded (advanced patterns with disproportionate proto complexity).
- **Deployment-level authentication, route-level authorization**: Authentication is configured once at the deployment level (common for all routes). Individual routes then control authorization (anonymous, authentication_only, any_of with scopes).
- **path_prefix as required string with prefix validation**: Must start with "/" and be non-empty. ForceNew in provider.
- **certificate_id as plain string**: OCI API Gateway certificates are not modeled as OpenMCF components; plain string reference.
- **function_id as plain string**: Individual OCI functions are code artifacts deployed via fn CLI, not OpenMCF components (consistent with R28 OciFunctionsApplication design).
- **Directory name**: `ociapigateway` (per WA02 convention -- lowercased kind name).

### Excluded from v1

- Gateway: ca_bundles, ip_mode, IPv4/IPv6 address configuration, response_cache_details (external Redis), locks
- Authentication: Custom (Functions-based), Token (newer JWT variant), validation_failure_policy (OAuth2 redirect flows)
- Deployment policies: dynamic_authentication, mutual_tls, usage_plans
- Route policies: body_validation, header/query transformations/validations, response_cache_lookup/store
- Backend types: dynamic_routing, oauth2_logout
- Tags: defined_tags, system_tags, freeform_tags (auto from labels)

### Bundled Resources

1. **API Gateway** -- the managed network endpoint (public or private) with optional TLS certificate and NSG controls
2. **API Deployment** -- the API specification defining routes, backends, authentication, CORS, and rate limiting

### Pulumi Module (Go)

5 files across the module package:
- `main.go` -- orchestrator with 7 enum maps (endpoint_type, log_level, backend_type, public_key_type, key_format, rate_key, authorization_type), OCI provider setup
- `locals.go` -- Locals struct with freeform tags and display name fallback
- `gateway.go` -- apigateway.NewGateway() with conditional certificate and NSGs, exports gateway_id and hostname, chains to deploymentResource()
- `deployment.go` -- apigateway.NewDeployment() with buildSpecification(), buildLoggingPolicies(), buildRequestPolicies(), buildAuthentication(), buildPublicKeys(), buildCors(), buildRateLimiting(), buildRoutes(), buildBackend(), buildRouteRequestPolicies(), buildRouteLoggingPolicies()
- `outputs.go` -- 3 output constants (gateway_id, hostname, deployment_endpoint)

### Terraform Module (HCL)

6 files:
- `main.tf` -- oci_apigateway_gateway.this with conditional certificate_id and NSG IDs
- `deployment.tf` -- oci_apigateway_deployment.this with deeply nested specification block: dynamic logging_policies, dynamic request_policies (authentication with public_keys and verify_claims, cors, rate_limiting), dynamic routes with backend, dynamic request_policies (authorization), dynamic logging_policies
- `locals.tf` -- display_name fallback, deployment_display_name, freeform tags, 7 enum conversion maps
- `variables.tf` -- Comprehensive object type definitions matching the proto spec structure
- `outputs.tf` -- 3 outputs (gateway_id, hostname, deployment_endpoint)
- `provider.tf` -- OCI provider version constraint

### Validation Tests

66 Ginkgo/Gomega tests (40 valid, 26 invalid scenarios) covering:
- Gateway: endpoint types, display_name, certificate_id, NSGs, StringValueOrRef patterns
- Deployment: path_prefix validation (leading slash required), logging policies (all log levels)
- Backend types: HTTP (with URL, timeouts, headers), Oracle Functions (with function_id), Stock Response (with status/body)
- Authentication: JWT with remote JWKS, remote JWKS with cache duration, static PEM keys, static JWK keys, verify claims, anonymous access, token header/query param/scheme, clock skew
- CORS: wildcard origins, full configuration with credentials and max_age
- Rate limiting: client_ip and total rate keys
- Route authorization: anonymous, authentication_only, any_of with scopes
- Multiple routes with mixed backend types
- Full configuration with all features combined
- Validation failures: missing required fields, CEL rule violations, enum constraints

### Kind Registration

- **Enum**: OciApiGateway = 3361
- **ID Prefix**: ociagw
- **Section**: Serverless and Functions (alongside R28 OciFunctionsApplication)

## Benefits

- Enables declarative provisioning of managed API endpoints with built-in security
- Completes Phase 7 (Serverless and Functions): both OciFunctionsApplication and OciApiGateway done
- Unblocks the Serverless Stack infra chart (Chart 4): all prerequisites now available
- Covers the primary serverless API pattern: API Gateway + JWT auth + Functions backends
- Full IaC parity between Pulumi and Terraform modules

## Impact

- **Platform teams**: Can now declaratively create API gateways with JWT authentication, CORS, rate limiting, and routing to HTTP/Functions/static backends
- **OCI provider coverage**: 29/37 resources complete (78.4%), Phase 7 complete
- **Infra charts**: Serverless Stack chart prerequisites fully met (OciFunctionsApplication + OciApiGateway)

## Validation Results

- `go build` -- clean
- `go vet` -- clean
- `go test` -- 66/66 passed
- `terraform validate` -- success

---

**Status**: Production Ready
