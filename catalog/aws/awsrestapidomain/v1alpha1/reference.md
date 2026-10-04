# AwsRestApiDomain

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsRestApiDomainSpec defines the desired configuration for an AWS
API Gateway custom domain name (the API Gateway v1 domain surface).

A custom domain fronts deployed REST APIs at your own hostname:
callers hit https://api.example.com/orders instead of the
execute-api endpoint, TLS terminates on your certificate, and
base-path mappings fan the domain's paths out across APIs and
stages. A domain outlives any one API and maps many - which is why
it is its own kind rather than part of AwsRestApiGateway.

The component bundles the domain, its base-path mappings, and - for
PRIVATE domains - the VPC-endpoint access associations. DNS is not
modeled here: point an AwsRoute53DnsRecord alias at the domain's
regional or CloudFront target (both are outputs).

Rule-based routing (an API Gateway v2 surface that also attaches to
v1 domains) stays on the AwsHttpApiDomain kind; this kind
models the v1 routing_mode knob that arbitrates between the two
mechanisms.

## Example

```yaml
# Canonical AwsRestApiDomain example (hack/dev manifest and refgen
# Example source): a REGIONAL custom domain with ACM certificate, one
# root mapping, and (commented in spirit) the PRIVATE access-association
# arm shown as a second mapping. Literal ARNs stand in for composed
# references so the offline `tofu plan` renders every coexisting arm.
# Mutual TLS and uploaded certificates are XOR with ACM and live in
# spec tests; PRIVATE access associations require endpoint type PRIVATE.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRestApiDomain
metadata:
  name: api-example-com
  id: api-example-com
  org: test-org
  env: dev
spec:
  region: us-west-2
  domainName: api.example.com
  certificateArn:
    value: arn:aws:acm:us-west-2:123456789012:certificate/abcd-1234
  endpointConfiguration:
    type: REGIONAL
    ipAddressType: ipv4
  # endpoint_access_mode pairs only with the SecurityPolicy_* family
  # (AWS rejects it with legacy policy names -- the CEL enforces it).
  endpointAccessMode: BASIC
  securityPolicy: SecurityPolicy_TLS13_1_2_2021_06
  routingMode: BASE_PATH_MAPPING_ONLY
  basePathMappings:
    - basePath: orders
      restApiId:
        value: abcdef1234
      stageName:
        value: prod
    - restApiId:
        value: abcdef1234
      stageName:
        value: prod
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.domainName` | `string` | yes |  |  |
| `spec.certificateArn` | `string \| valueFrom` |  |  | AwsCertManagerCert (`status.outputs.cert_arn`) |
| `spec.uploadedCertificate` | `AwsRestApiDomainUploadedCertificate` |  |  |  |
| `spec.uploadedCertificate.name` | `string` | yes |  |  |
| `spec.uploadedCertificate.body` | `string` | yes |  |  |
| `spec.uploadedCertificate.chain` | `string` |  |  |  |
| `spec.uploadedCertificate.privateKey` | `string` (sensitive) | yes |  |  |
| `spec.endpointConfiguration` | `AwsRestApiDomainEndpointConfiguration` |  |  |  |
| `spec.endpointConfiguration.type` | `string` |  |  |  |
| `spec.endpointConfiguration.ipAddressType` | `string` |  |  |  |
| `spec.endpointAccessMode` | `string` |  |  |  |
| `spec.securityPolicy` | `string` |  |  |  |
| `spec.mutualTls` | `AwsRestApiDomainMutualTls` |  |  |  |
| `spec.mutualTls.truststoreUri` | `string` | yes |  |  |
| `spec.mutualTls.truststoreVersion` | `string` |  |  |  |
| `spec.ownershipVerificationCertificateArn` | `string \| valueFrom` |  |  | AwsCertManagerCert (`status.outputs.cert_arn`) |
| `spec.policy` | `object` |  |  |  |
| `spec.routingMode` | `string` |  |  |  |
| `spec.basePathMappings` | `[]AwsRestApiDomainBasePathMapping` |  |  |  |
| `spec.basePathMappings[].basePath` | `string` |  |  |  |
| `spec.basePathMappings[].restApiId` | `string \| valueFrom` | yes |  | AwsRestApiGateway (`status.outputs.rest_api_id`) |
| `spec.basePathMappings[].stageName` | `string \| valueFrom` |  |  | AwsRestApiGateway (`status.outputs.stage_name`) |
| `spec.accessAssociations` | `[]AwsRestApiDomainAccessAssociation` |  |  |  |
| `spec.accessAssociations[].vpcEndpointId` | `string \| valueFrom` | yes |  | AwsVpcEndpoint (`status.outputs.vpc_endpoint_id`) |

## Field Details

### spec.region

`string` · required

The AWS region where the domain will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.domainName

`string` · required

The fully-qualified domain name (e.g. "api.example.com").
Immutable in AWS - changing it replaces the domain.

- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.certificateArn

`string | valueFrom`

ACM certificate covering the domain name. For REGIONAL and PRIVATE
domains the certificate must live in this region; for EDGE domains
it must live in us-east-1 (the CloudFront region). The modules wire
it to the endpoint type automatically.

- references: AwsCertManagerCert (`status.outputs.cert_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCertManagerCert, name: <that resource's name>, fieldPath: status.outputs.cert_arn}} -- a bare string does not parse

### spec.uploadedCertificate

`AwsRestApiDomainUploadedCertificate`

Upload certificate material directly instead of using ACM (legacy
path for certificates that cannot be imported into ACM).

### spec.uploadedCertificate.name

`string` · required

Display name for the uploaded certificate.

- rule: {"string":{"minLen":"1"}}

### spec.uploadedCertificate.body

`string` · required

PEM-encoded certificate body.

- rule: {"string":{"minLen":"1"}}

### spec.uploadedCertificate.chain

`string`

PEM-encoded certificate chain.

### spec.uploadedCertificate.privateKey

`string` · required · sensitive

PEM-encoded private key.

- rule: {"required":true}

### spec.endpointConfiguration

`AwsRestApiDomainEndpointConfiguration`

Endpoint type and addressing. Omitted = a REGIONAL endpoint (the
right default; EDGE routes through CloudFront, PRIVATE is
reachable only through VPC endpoints).

- rule: PRIVATE endpoints require ip_address_type 'dualstack' (or omit it for the AWS default)

### spec.endpointConfiguration.type

`string`

The endpoint type. REGIONAL serves from this region; EDGE
provisions a managed CloudFront distribution; PRIVATE is reachable
only through VPC endpoints (pair with access_associations).

- rule: {"string":{"in":["REGIONAL","EDGE","PRIVATE"]}}

### spec.endpointConfiguration.ipAddressType

`string`

Endpoint addressing: "ipv4" or "dualstack". Omitted = the AWS
default. PRIVATE endpoints require dualstack.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ipv4","dualstack"]}}

### spec.endpointAccessMode

`string`

Restrict how the domain resolves callers: BASIC (default) or
STRICT (rejects requests whose Host header does not match -
hardens against host-header confusion).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BASIC","STRICT"]}}

### spec.securityPolicy

`string`

Minimum TLS version and cipher policy. Omitted = the AWS default
for the endpoint type. The SecurityPolicy_TLS13_* values are the
2025-09 policy family (FIPS/PFS/PQ variants); TLS_1_0 and TLS_1_2
are the legacy names.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["TLS_1_0","TLS_1_2","SecurityPolicy_TLS13_1_3_2025_09","SecurityPolicy_TLS13_1_3_FIPS_2025_09","SecurityPolicy_TLS13_1_2_PFS_PQ_2025_09","SecurityPolicy_TLS13_1_2_FIPS_PQ_2025_09","SecurityPolicy_TLS13_1_2_FIPS_PFS_PQ_2025_09","SecurityPolicy_TLS13_1_2_PQ_2025_09","SecurityPolicy_TLS13_1_2_2021_06","SecurityPolicy_TLS13_2025_EDGE","SecurityPolicy_TLS12_PFS_2025_EDGE","SecurityPolicy_TLS12_2018_EDGE"]}}

### spec.mutualTls

`AwsRestApiDomainMutualTls`

Require clients to present certificates from your truststore
(mutual TLS).

### spec.mutualTls.truststoreUri

`string` · required

S3 URI of the truststore bundle of CA certificates clients must
chain to (e.g. "s3://my-bucket/truststore.pem"). Version the
object - truststore mistakes lock every caller out.

- rule: {"string":{"minLen":"1","pattern":"^s3://"}}

### spec.mutualTls.truststoreVersion

`string`

S3 object version of the truststore to pin (recommended - an
unpinned truststore changes under the domain).

### spec.ownershipVerificationCertificateArn

`string | valueFrom`

ACM certificate proving you own the domain - required by AWS when
mutual TLS is enabled on a domain whose primary certificate cannot
serve as the ownership proof.

- references: AwsCertManagerCert (`status.outputs.cert_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCertManagerCert, name: <that resource's name>, fieldPath: status.outputs.cert_arn}} -- a bare string does not parse

### spec.policy

`object`

Resource policy on the domain (IAM policy document as structured
YAML/JSON) - controls which principals/VPC endpoints may use a
PRIVATE domain.

### spec.routingMode

`string`

How the domain routes requests when both mechanisms exist:
BASE_PATH_MAPPING_ONLY (default), ROUTING_RULE_ONLY, or
ROUTING_RULE_THEN_BASE_PATH_MAPPING. Routing rules themselves are
the AwsHttpApiDomain component's surface; set a ROUTING_RULE mode
here only when that component manages rules on this same domain.
Uppercase canonical spellings only (the provider accepts any case;
the spec pins the canonical form).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BASE_PATH_MAPPING_ONLY","ROUTING_RULE_ONLY","ROUTING_RULE_THEN_BASE_PATH_MAPPING"]}}

### spec.basePathMappings

`[]AwsRestApiDomainBasePathMapping`

Base-path mappings fanning the domain's paths out across REST APIs
and stages ("/orders" -> the orders API's prod stage).

### spec.basePathMappings[].basePath

`string`

The path segment under the domain ("orders" maps
https://{domain}/orders). Empty = the domain root. No slashes -
base paths are single segments.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"128","pattern":"^[^/]+$"}}

### spec.basePathMappings[].restApiId

`string | valueFrom` · required

The REST API requests on this path route to.

- references: AwsRestApiGateway (`status.outputs.rest_api_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRestApiGateway, name: <that resource's name>, fieldPath: status.outputs.rest_api_id}} -- a bare string does not parse

### spec.basePathMappings[].stageName

`string | valueFrom`

The stage requests route to. Omitted = the API's stage is selected
by the request path's next segment (AWS's stage-in-path behavior);
set it to pin the mapping to one stage - the usual choice.

- references: AwsRestApiGateway (`status.outputs.stage_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRestApiGateway, name: <that resource's name>, fieldPath: status.outputs.stage_name}} -- a bare string does not parse

### spec.accessAssociations

`[]AwsRestApiDomainAccessAssociation`

VPC endpoints granted access to a PRIVATE domain.

### spec.accessAssociations[].vpcEndpointId

`string | valueFrom` · required

The interface VPC endpoint (com.amazonaws.{region}.execute-api)
callers reach the private domain through. Every field change
replaces the association.

- references: AwsVpcEndpoint (`status.outputs.vpc_endpoint_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpcEndpoint, name: <that resource's name>, fieldPath: status.outputs.vpc_endpoint_id}} -- a bare string does not parse

## Validation Rules

- `certificate_exactly_one_source`: set exactly one of certificate_arn (ACM) or uploaded_certificate
- `base_paths_unique`: base_path_mappings entries must have unique base_path values
- `access_associations_private_only`: access_associations apply only to PRIVATE endpoint types
- `access_association_endpoints_unique`: access_associations entries must reference unique VPC endpoints
- `policy_private_only`: policy applies only to PRIVATE endpoint types
- `access_mode_requires_modern_security_policy`: endpoint_access_mode requires a security_policy from the SecurityPolicy_* family

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsRestApiDomain, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.domain_name` | `string` | The custom domain name. |
| `status.outputs.domain_name_arn` | `string` | The domain's ARN. |
| `status.outputs.domain_name_id` | `string` | The domain name ID (distinguishes PRIVATE domains sharing one hostname). |
| `status.outputs.regional_domain_name` | `string` | The regional target hostname - point Route 53 alias records here for REGIONAL/PRIVATE domains. |
| `status.outputs.regional_zone_id` | `string` | The Route 53 hosted zone ID of the regional endpoint (alias records need it). |
| `status.outputs.cloudfront_domain_name` | `string` | The CloudFront target hostname - point alias records here for EDGE domains. |
| `status.outputs.cloudfront_zone_id` | `string` | The CloudFront hosted zone ID (a fixed global value; alias records need it). |
| `status.outputs.base_path_mapping_ids` | `map<string, string>` | Base-path mapping IDs keyed by each entry's base_path ("(none)" for the empty base path). |
| `status.outputs.access_association_arns` | `map<string, string>` | Access-association ARNs keyed by each entry's resolved VPC endpoint ID. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.certificateArn` | AwsCertManagerCert | `status.outputs.cert_arn` |
| `spec.ownershipVerificationCertificateArn` | AwsCertManagerCert | `status.outputs.cert_arn` |
| `spec.basePathMappings[].restApiId` | AwsRestApiGateway | `status.outputs.rest_api_id` |
| `spec.basePathMappings[].stageName` | AwsRestApiGateway | `status.outputs.stage_name` |
| `spec.accessAssociations[].vpcEndpointId` | AwsVpcEndpoint | `status.outputs.vpc_endpoint_id` |

## See Also

- [Overview](../README.md)
