# OciApiGateway

## Overview

OciApiGateway is an Planton component that deploys an OCI API Gateway with a bundled API deployment. It provides a single declarative manifest to create a managed API endpoint with route definitions, backend configurations, JWT authentication, CORS, and rate limiting.

## Purpose

API Gateway is OCI's managed service for exposing APIs to external or internal consumers. It handles routing, authentication, authorization, CORS, rate limiting, and logging so that backend services and serverless functions do not need to implement these concerns. This component bundles the gateway (network endpoint) and its deployment (API specification) because they are most commonly created together as a single unit.

## Key Features

- **Public and private endpoints** — internet-facing or VCN-internal gateways.
- **Three backend types** — HTTP proxy, OCI Functions invocation, and stock responses.
- **JWT authentication** — token validation via remote JWKS or static keys with configurable issuers, audiences, and custom claim verification.
- **CORS policy** — configurable origins, methods, headers, and credentials for browser-based clients.
- **Rate limiting** — per-IP or aggregate request throttling.
- **Per-route authorization** — anonymous access, authentication-only, or scope-based authorization.
- **Logging** — access and execution logging at both deployment and route levels.
- **Foreign key references** — `compartmentId`, `subnetId`, and `networkSecurityGroupIds` support `valueFrom`.

## Constraints

- `endpointType` and `subnetId` are immutable after creation (ForceNew on gateway).
- `deployment.pathPrefix` is immutable after creation (ForceNew on deployment).
- Routes are evaluated in order; first match wins.
- `publicKeys` is required for JWT authentication.
- `uri` is required when public key type is `remote_jwks`.
- `keys` must be non-empty when public key type is `static_keys`.
- `allowedScope` must be non-empty when route authorization type is `any_of`.
- `url` is required for HTTP backends; `functionId` is required for Functions backends.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Public REST API proxy | Public gateway with HTTP backends |
| Serverless API | Public gateway with OCI Functions backends |
| Internal microservice gateway | Private gateway with rate limiting |
| Browser-facing API | JWT authentication + CORS policy |
| Mixed public/private routes | Anonymous access allowed + per-route authorization |
| Health check endpoint | Stock response backend returning 200 |

## Production Features

- **Freeform tags** — automatically populated from `metadata.labels` on both the gateway and deployment resources.
- **TLS termination** — optional OCI Certificates service integration for HTTPS.
- **JWT validation** — offloads token verification from backend services.
- **Rate limiting** — protects backends from traffic spikes.
