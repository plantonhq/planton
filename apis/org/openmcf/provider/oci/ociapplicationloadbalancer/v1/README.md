# Overview

The **OCI Load Balancer API Resource** provides a consistent and standardized interface for deploying and managing Oracle Cloud Infrastructure Application Load Balancers (Layer 7). This component bundles the load balancer with its backend sets, backends, listeners, certificates, hostnames, and rule sets into a single atomic deployment unit with the standard OpenMCF KRM pattern. The spec defines 17 nested message types with 3 enums, covering the full configuration surface of OCI's Application Load Balancer API.

## Purpose

This API resource streamlines the deployment of OCI Application Load Balancers by wrapping 7 tightly coupled sub-resource types into one declarative manifest. An OCI load balancer requires at minimum the load balancer itself, one backend set with a health checker, and one listener — creating these as separate OpenMCF components would require users to manage 3+ YAML manifests for even the simplest setup, with error-prone cross-references between them. The bundled approach enables users to:

- **Deploy a Complete Load Balancer Stack Atomically**: A single manifest creates the load balancer, backend sets, backends, listeners, certificates, hostnames, and rule sets in the correct dependency order. The Pulumi module creates resources in sequence (LB → certificates → backend sets → hostnames → rule sets → listeners) and wires up explicit `DependsOn` relationships.
- **Distribute Traffic Across Backends**: Three load balancing policies (round robin, least connections, IP hash) control how traffic reaches backend servers. Per-backend weights enable weighted distribution, and backup backends provide failover when primary backends are unhealthy.
- **Terminate SSL at the Load Balancer**: Upload PEM certificates and configure SSL on listeners for HTTPS termination. Backend SSL (re-encryption) is also supported for end-to-end encryption between clients, the load balancer, and backend servers.
- **Route by Virtual Hostname**: Hostnames enable multiple domains to share a single load balancer IP. Each listener can filter requests by hostname, routing `api.example.com` and `app.example.com` to different backend sets.
- **Manipulate Requests and Responses with Rule Sets**: 11 rule set actions cover HTTP redirects, header injection/removal, access control by HTTP method, IP-based connection limits, and header size configuration. The most common use case — HTTP-to-HTTPS redirect — is a single rule set item.
- **Monitor Backend Health**: HTTP and TCP health checks with configurable intervals, timeouts, retries, and response matching automatically remove unhealthy backends from rotation.
- **Manage Session Affinity**: Two mutually exclusive session persistence modes (LB-managed cookie and application-managed cookie) pin client sessions to specific backends. The LB-managed cookie supports HttpOnly, Secure, domain, path, and max-age attributes.
- **Compose with Other OCI Resources**: Reference OciCompartment, OciSubnet, and OciSecurityGroup outputs via `StringValueOrRef` for declarative, cross-resource dependency chains in infra charts.

## Key Features

- **Consistent Interface**: Aligns with the OpenMCF pattern for deploying cloud infrastructure across providers.
- **7 Sub-Resource Types Bundled**: Load balancer, backend sets, backends, listeners, certificates, hostnames, and rule sets are created as a single unit with correct dependency ordering.
- **Flexible Bandwidth**: The recommended `"flexible"` shape supports configurable minimum and maximum bandwidth from 10 to 8000 Mbps. Deprecated fixed shapes are accepted for backward compatibility.
- **4 Listener Protocols**: HTTP, HTTP/2, TCP, and gRPC listeners on any port from 1 to 65535.
- **SSL Termination and Re-encryption**: Client-facing SSL on listeners and backend-facing SSL on backend sets. Supports TLS protocol version pinning, cipher suite selection, server order preference, and mutual TLS with client certificate verification.
- **Cookie-Based Session Persistence**: LB-managed cookies (the load balancer injects a Set-Cookie header) or application-managed cookies (the load balancer reads an existing application cookie). Mutually exclusive per backend set via proto `oneof`.
- **Health Checking**: HTTP and TCP health probes with configurable port, URL path, expected return code, response body regex, interval, timeout, and retry thresholds.
- **Virtual Hostname Routing**: Multiple FQDNs per load balancer. Listeners filter by hostname for host-based routing without multiple load balancers.
- **11 Rule Set Actions**: `add_http_request_header`, `add_http_response_header`, `extend_http_request_header_value`, `extend_http_response_header_value`, `remove_http_request_header`, `remove_http_response_header`, `redirect`, `allow`, `control_access_using_http_methods`, `http_header`, `ip_based_max_connections`.
- **Delete Protection**: `isDeleteProtectionEnabled` prevents accidental deletion of production load balancers.
- **Request Tracing**: `isRequestIdEnabled` with configurable `requestIdHeader` adds a request ID to each request for end-to-end tracing.
- **Automatic Tagging**: Standard OpenMCF freeform tags applied to the load balancer (resource kind, resource ID, organization, environment, and user-defined labels from metadata).
- **Infra-Chart Composability**: Exports 2 stack outputs (`load_balancer_id`, `ip_addresses`) for downstream `StringValueOrRef` references. Consumes `compartmentId` from OciCompartment, `subnetId` from OciSubnet, and `networkSecurityGroupId` from OciSecurityGroup.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations producing identical outputs.

## How OCI Application Load Balancers Differ from Other Providers

Understanding these differences is essential when coming from AWS, Azure, or GCP:

| Aspect | OCI Application LB | AWS ALB | Azure Application Gateway | GCP HTTPS LB |
|--------|-------------------|---------|--------------------------|---------------|
| **Shape model** | Flexible bandwidth (10–8000 Mbps) or deprecated fixed shapes | Fixed instance type, scales automatically | SKU tiers (Standard_v2, WAF_v2) with autoscaling | Globally distributed, no shape selection |
| **Backend model** | Backend sets with named groups, per-backend IP/port | Target groups with instance/IP/Lambda targets | Backend pools with HTTP settings | Backend services with instance groups or NEGs |
| **Health checking** | Per-backend-set, HTTP/TCP with response body regex | Per-target-group, HTTP/HTTPS/gRPC | Per-backend-pool HTTP probe | Per-backend-service HTTP/HTTPS/gRPC/HTTP2 |
| **SSL certificates** | Uploaded PEM certs or OCI Certificate Service OCIDs | ACM-managed certificates | Key Vault or uploaded PFX | Google-managed or self-managed certs |
| **Session persistence** | LB-managed cookie or app-managed cookie per backend set | Stickiness via ALB-generated cookie or app cookie | Cookie-based affinity per backend pool | Session affinity via generated cookie |
| **Virtual hostnames** | Named hostname resources, referenced by listeners | Host-header rules in listener rules | Multi-site listeners | URL map host rules |
| **Rule engine** | 11 action types (redirect, header add/remove/extend, access control, connection limits) | Listener rules with conditions and actions | URL path map, rewrite rules, WAF | URL map path matchers, header actions |
| **Proxy protocol** | Backend TCP proxy protocol v1/v2 per listener | Proxy protocol v2 via target group attribute | Not natively supported | PROXY protocol via backend service |
| **IP version** | IPV4 or IPV6 per load balancer | Dual-stack (IPV4 + IPV6) | IPV4 only (Standard_v2) | IPV4 and IPV6 via separate forwarding rules |
| **Delete protection** | `isDeleteProtectionEnabled` flag | Deletion protection attribute | Resource lock (Azure-level, not LB-specific) | No built-in (use IAM constraints) |

Key distinctions for OCI newcomers:

- **Backend Sets Are Named Groups.** Unlike AWS target groups (which are standalone resources), OCI backend sets exist only within a load balancer. Listeners reference backend sets by name. This is why they are bundled into the same component.
- **Certificates Are Load-Balancer-Scoped.** PEM certificates are uploaded to the specific load balancer instance and referenced by name. Alternatively, use `certificateIds` to reference certificates managed by the OCI Certificate Service (preferred for production).
- **Hostnames Are Named Resources.** Virtual hostnames are separate named resources within the load balancer, not inline properties of a listener. A listener references hostnames by name via `hostnameNames`.
- **Rule Sets Use a Flat Action Enum.** All 11 rule types share a single `RuleSetItem` message with an `action` enum. The applicable fields depend on the action — different from AWS's typed rule actions or GCP's URL map matchers.
- **Flexible Shape Is the Standard.** The `"flexible"` shape with configurable min/max bandwidth is the recommended choice. Fixed shapes (`"100Mbps"`, `"400Mbps"`, `"8000Mbps"`) are deprecated.

## Critical Constraints

- **Subnet Changes Force Recreation**: Changing `subnetIds` or `isPrivate` after creation destroys and recreates the load balancer. Plan subnet placement carefully before deployment.
- **One Load Balancer Per Resource**: Each OciApplicationLoadBalancer manifest creates exactly one load balancer. Multiple load balancers require separate manifests.
- **path_route_set Is Excluded**: Oracle has deprecated `path_route_set` in favor of routing policies. Use `routingPolicyName` on a listener to reference an externally managed routing policy.
- **Routing Policies Are External**: Routing policies are not bundled into this component. They can be referenced by name from a listener via `routingPolicyName`, but creation and management must happen outside this resource.
- **SSL Certificates Are Sensitive**: The `privateKey` and `passphrase` fields in the `Certificate` message contain sensitive data. Store manifests containing these values securely and consider using OCI Certificate Service (`certificateIds`) instead of uploaded PEM certificates.
- **State Lifecycle Omitted**: The OCI load balancer `lifecycle_state` field is intentionally omitted. OpenMCF resources are always deployed to their active state. Delete the resource to decommission the load balancer.
- **Default Backend Set Required on All Listeners**: Even listeners with redirect rule sets that never reach the backend must specify a `defaultBackendSetName`. The field is always required.

## Use Cases

- **Web Application HTTPS Termination**: The most common scenario — an HTTPS listener with an uploaded or managed certificate terminating SSL, backed by HTTP backends running on a private subnet. A companion HTTP listener with a redirect rule set sends all HTTP traffic to HTTPS.
- **Internal API Gateway**: A private load balancer (`isPrivate: true`) distributing gRPC or HTTP/2 traffic across backend API servers within a VCN. No public IP assigned, accessible only from within the VCN or via VCN peering.
- **Multi-Domain Hosting**: A single load balancer serving multiple domains using virtual hostnames. Each domain maps to a dedicated backend set, consolidating infrastructure and public IPs. Combined with HTTPS, each domain can share or use its own certificate.
- **Weighted Canary Deployments**: Backend weights enable traffic splitting between stable and canary versions. Assign weight 9 to stable backends and weight 1 to the canary backend, routing approximately 10% of traffic to the new version.
- **Backend Draining for Maintenance**: Set `drain: true` on a backend to stop new connections while allowing in-flight requests to complete. Combined with `backup: true` backends, this enables zero-downtime maintenance windows.
- **Security Header Injection**: Rule sets with `add_http_response_header` inject headers like `Strict-Transport-Security`, `X-Content-Type-Options`, and `X-Frame-Options` on every response without modifying backend applications.
- **IP-Based Rate Limiting**: The `ip_based_max_connections` rule action limits concurrent connections per client IP address, providing basic DDoS mitigation and protecting backends from connection flooding.

## Production Features

This resource provides complete support for production-grade OCI Application Load Balancer deployments, including:

- **HTTPS with Managed Certificates**: Upload PEM certificates directly or reference OCI Certificate Service certificates via OCID for automatic renewal and lifecycle management.
- **End-to-End Encryption**: SSL termination on the listener (client → LB) and SSL re-encryption on the backend set (LB → backend) for environments requiring encrypted traffic on all segments.
- **Cookie-Based Session Persistence**: LB-managed cookies with HttpOnly, Secure, domain, path, and max-age controls. Application-managed cookies for frameworks that maintain their own session identifiers.
- **Health Check Monitoring**: HTTP health checks with URL path, expected status code, response body regex, and configurable intervals. TCP health checks for non-HTTP backends. Unhealthy backends are removed from rotation automatically.
- **Advanced Rule Sets**: HTTP-to-HTTPS redirects, security header injection, HTTP method access control, header size limits, and per-IP connection throttling — all configured declaratively without modifying backend code.
- **High Availability**: Regional load balancers with subnets in two availability domains provide cross-AD redundancy. Reserved public IPs enable IP address persistence across load balancer replacements.
- **Operational Controls**: Delete protection prevents accidental deletion. Request ID injection enables end-to-end request tracing across load balancer, backends, and logging systems.
- **Freeform Tagging**: Standard OpenMCF labels applied as OCI freeform tags for resource management, cost tracking, and compliance.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations producing identical outputs.
- **Infra-Chart Composability**: Designed to compose with OciCompartment (upstream dependency), OciSubnet, OciSecurityGroup, and downstream DNS/certificate components via `StringValueOrRef`.
