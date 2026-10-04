# AwsCloudMapNamespace

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsCloudMapNamespaceSpec defines one AWS Cloud Map namespace - the
service-discovery registry ECS services and custom applications look
each other up in - with its services and their statically registered
instances managed in-line.

The namespace type shapes the whole surface:

HTTP namespaces are API-only - consumers discover instances by
calling DiscoverInstances; no DNS records exist, so services here
carry no dns_config.

PRIVATE_DNS namespaces create a private hosted zone visible inside
one VPC; services publish A/AAAA/SRV/CNAME records there.

PUBLIC_DNS namespaces create a public hosted zone; services publish
records resolvable from the internet.

Services fold here because a service belongs to exactly one
namespace for life. Statically registered instances fold under
their service - the declarative registration surface for things
that do not register themselves (an RDS endpoint, an on-prem
address). Never declare instances on services that runtime
platforms (ECS service discovery) register into: the runtime owns
those registrations.

## Example

```yaml
# Canonical AwsCloudMapNamespace example (hack/dev manifest and refgen
# Example source): a private DNS namespace with one A-record service
# carrying a statically registered instance, and one CNAME service
# aliasing a database endpoint.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudMapNamespace
metadata:
  name: corp.internal
  id: corp-internal
  org: test-org
  env: dev
spec:
  region: us-west-2
  type: PRIVATE_DNS
  vpcId:
    value: vpc-0123456789abcdef0
  description: Service discovery for the corp platform
  services:
    - name: api
      dnsConfig:
        records:
          - type: A
            ttl: 10
      healthCheckCustomConfig: {}
      instances:
        - instanceId: static-api-1
          ip: 10.0.1.10
          port: 8080
    - name: db
      dnsConfig:
        records:
          - type: CNAME
            ttl: 30
        # CNAME requires the explicit WEIGHTED policy - AWS rejects
        # CNAME under the default MULTIVALUE.
        routingPolicy: WEIGHTED
      instances:
        - instanceId: primary
          cname: mydb.cluster-abc.us-west-2.rds.amazonaws.com
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.type` | `string` |  |  |  |
| `spec.vpcId` | `string \| valueFrom` |  |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.description` | `string` |  |  |  |
| `spec.services` | `[]AwsCloudMapService` |  |  |  |
| `spec.services[].name` | `string` | yes |  |  |
| `spec.services[].description` | `string` |  |  |  |
| `spec.services[].dnsConfig` | `AwsCloudMapServiceDnsConfig` |  |  |  |
| `spec.services[].dnsConfig.records` | `[]AwsCloudMapServiceDnsRecord` | yes |  |  |
| `spec.services[].dnsConfig.records[].type` | `string` |  |  |  |
| `spec.services[].dnsConfig.records[].ttl` | `int64` |  |  |  |
| `spec.services[].dnsConfig.routingPolicy` | `string` |  |  |  |
| `spec.services[].healthCheckConfig` | `AwsCloudMapServiceHealthCheckConfig` |  |  |  |
| `spec.services[].healthCheckConfig.type` | `string` |  |  |  |
| `spec.services[].healthCheckConfig.resourcePath` | `string` |  |  |  |
| `spec.services[].healthCheckConfig.failureThreshold` | `int64` |  |  |  |
| `spec.services[].healthCheckCustomConfig` | `AwsCloudMapServiceHealthCheckCustomConfig` |  |  |  |
| `spec.services[].forceDestroy` | `bool` |  |  |  |
| `spec.services[].instances` | `[]AwsCloudMapInstance` |  |  |  |
| `spec.services[].instances[].instanceId` | `string` | yes |  |  |
| `spec.services[].instances[].ip` | `string` |  |  |  |
| `spec.services[].instances[].ipv6` | `string` |  |  |  |
| `spec.services[].instances[].port` | `int64` |  |  |  |
| `spec.services[].instances[].cname` | `string` |  |  |  |
| `spec.services[].instances[].aliasDnsName` | `string \| valueFrom` |  |  | AwsAlb (`status.outputs.load_balancer_dns_name`) |
| `spec.services[].instances[].ec2InstanceId` | `string \| valueFrom` |  |  | AwsEc2Instance (`status.outputs.instance_id`) |
| `spec.services[].instances[].customAttributes` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the namespace lives in. Example: "us-west-2".

- rule: {"string":{"minLen":"1"}}

### spec.type

`string`

The namespace's discovery model. Fixed for life. A PUBLIC_DNS
namespace creates a real Route 53 public hosted zone for
metadata.name, and Route 53 REFUSES reserved documentation
domains there - example.com and its subdomains fail
CANNOT_CREATE_HOSTED_ZONE "reserved by AWS" (live-proven); .test
names are accepted, and you never need to own or delegate the
domain just to create the namespace.

- rule: {"string":{"in":["HTTP","PRIVATE_DNS","PUBLIC_DNS"]}}

### spec.vpcId

`string | valueFrom`

The VPC a PRIVATE_DNS namespace's hosted zone is visible in.
Reference an AwsVpc vpc_id output or pass a literal vpc-... id.
Fixed for life. (The provider never reads this back from AWS -
imports must supply it.)

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.description

`string`

What this namespace is for. Updatable on DNS namespaces; on an
HTTP namespace the provider has no update path, so changing it
REPLACES the namespace (and everything registered in it).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"1024"}}

### spec.services

`[]AwsCloudMapService`

The services registered in this namespace, keyed by name.

- rule: health_check_config and health_check_custom_config are mutually exclusive
- rule: instance ids must be unique within the service
- rule: an instance's alias_dns_name cannot combine with ip, ipv6, port, cname, or ec2_instance_id
- rule: an instance's ec2_instance_id and ip are mutually exclusive - AWS derives the IP from the instance
- rule: instances with an ip under a health-checked service must set port - Route 53 probes ip:port

### spec.services[].name

`string` · required

Service name - the for_each key on both engines, the key in the
service_ids output, and the DNS label records publish under
(e.g. service "api" in namespace "corp.local" resolves at
"api.corp.local").

- rule: {"string":{"minLen":"1","maxLen":"127"}}

### spec.services[].description

`string`

What this service is.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"1024"}}

### spec.services[].dnsConfig

`AwsCloudMapServiceDnsConfig`

The DNS records instances of this service publish (DNS
namespaces only).

- rule: a CNAME record requires routing_policy WEIGHTED - AWS rejects CNAME under the default MULTIVALUE policy

### spec.services[].dnsConfig.records

`[]AwsCloudMapServiceDnsRecord` · required

The record types instances publish. Fixed per record entry at the
provider (changing a type replaces the service's records).

- rule: {"repeated":{"minItems":"1"}}

### spec.services[].dnsConfig.records[].type

`string`

The record type. A/AAAA publish instance IPs, SRV publishes
host+port, CNAME publishes an alias to the instance's cname
attribute.

- rule: {"string":{"in":["A","AAAA","SRV","CNAME"]}}

### spec.services[].dnsConfig.records[].ttl

`int64`

The record's TTL in seconds.

- rule: ttl must be between 1 and 2147483647 seconds

### spec.services[].dnsConfig.routingPolicy

`string`

How queries pick among healthy instances. MULTIVALUE answers with
up to eight healthy records; WEIGHTED answers with one, picked by
weight. Unset means AWS's default (MULTIVALUE). Fixed for life.
CNAME records REQUIRE the explicit WEIGHTED (see the message rule).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["MULTIVALUE","WEIGHTED"]}}

### spec.services[].healthCheckConfig

`AwsCloudMapServiceHealthCheckConfig`

Route 53 health checks against instance endpoints (PUBLIC_DNS
namespaces only - Route 53 health checkers live on the public
internet).

- rule: resource_path applies only to HTTP and HTTPS health checks

### spec.services[].healthCheckConfig.type

`string`

How the checker probes instances. Unset means AWS's default
(HTTP). Fixed for life.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["HTTP","HTTPS","TCP"]}}

### spec.services[].healthCheckConfig.resourcePath

`string`

The path HTTP/HTTPS probes request (e.g. "/health"). Unset means
"/".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

### spec.services[].healthCheckConfig.failureThreshold

`int64`

Consecutive probe results (1-10) before Route 53 flips an
instance's health state. Unset means AWS's default (3).

- rule: failure_threshold must be between 1 and 10

### spec.services[].healthCheckCustomConfig

`AwsCloudMapServiceHealthCheckCustomConfig`

A custom heartbeat: your workload reports instance health via
UpdateInstanceCustomHealthStatus instead of Route 53 probing it.

### spec.services[].forceDestroy

`bool`

On destroy, deregister EVERYTHING registered in the service first
- including instances runtime platforms (ECS) registered that
this manifest never declared. Leave false unless this service's
registrations are fully owned here.

### spec.services[].instances

`[]AwsCloudMapInstance`

Statically registered instances, keyed by instance_id - the
declarative registration surface for endpoints that do not
register themselves.

- rule: set at most one of ip and cname - an instance resolves to an address or an alias, not both

### spec.services[].instances[].instanceId

`string` · required

The registration's identity within the service - the for_each key
on both engines and half of the import ID
("{service_id}/{instance_id}"). Re-registering the same id
updates the instance in place (AWS upserts).

- rule: {"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z_/:.@-]+$"}}

### spec.services[].instances[].ip

`string`

The instance's IPv4 address (publishes A records; SRV targets
resolve to it). When the service carries health_check_config,
Route 53 REJECTS reserved/documentation/private-range addresses
at registration (RegisterInstance 400 "IPv4 address ... is
forbidden for healthChecks" - TEST-NET included; health-checked
endpoints must be publicly routable).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"ipv4":true}}

### spec.services[].instances[].ipv6

`string`

The instance's IPv6 address (publishes AAAA records).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"ipv6":true}}

### spec.services[].instances[].port

`int64`

The port the instance serves on (SRV records and health checks).

- rule: port must be between 1 and 65535

### spec.services[].instances[].cname

`string`

The domain name CNAME records point at (services whose dns_config
publishes CNAME) - e.g. an RDS endpoint address. Reference
another kind's endpoint output or pass a literal hostname.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

### spec.services[].instances[].aliasDnsName

`string | valueFrom`

Publish a Route 53 ALIAS to an ELB DNS name instead of address
records. Cannot combine with any other attribute. Reference an
AwsAlb load_balancer_dns_name output or pass a literal DNS name.

- references: AwsAlb (`status.outputs.load_balancer_dns_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsAlb, name: <that resource's name>, fieldPath: status.outputs.load_balancer_dns_name}} -- a bare string does not parse

### spec.services[].instances[].ec2InstanceId

`string | valueFrom`

Register an EC2 instance by id - AWS derives the IPv4 attribute
from it. Reference an AwsEc2Instance instance_id output or pass a
literal i-... id.

- references: AwsEc2Instance (`status.outputs.instance_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEc2Instance, name: <that resource's name>, fieldPath: status.outputs.instance_id}} -- a bare string does not parse

### spec.services[].instances[].customAttributes

`map<string, string>`

Any additional AWS instance attributes (e.g.
AWS_INIT_HEALTH_STATUS) or custom key-value metadata returned by
DiscoverInstances. Keys and values follow AWS's attribute charset
rules.

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"255","pattern":"^[0-9A-Za-z!-~]+$"}},"values":{"string":{"maxLen":"1024"}}}}

## Validation Rules

- `spec.private_dns_requires_vpc`: type PRIVATE_DNS requires vpc_id - the VPC the private hosted zone is visible in
- `spec.vpc_only_private_dns`: vpc_id applies only when type is PRIVATE_DNS
- `spec.http_services_carry_no_dns_config`: services in an HTTP namespace cannot set dns_config - discovery is API-only
- `spec.health_check_config_public_only`: health_check_config (Route 53 health checks) applies only in a PUBLIC_DNS namespace - use health_check_custom_config elsewhere
- `spec.service_names_unique`: service names must be unique within the namespace

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCloudMapNamespace, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace_id` | `string` | The namespace's id (ns-...) - the provider's import ID (a PRIVATE_DNS namespace imports as "{namespace_id}:{vpc_id}"). |
| `status.outputs.namespace_arn` | `string` | The namespace's ARN. |
| `status.outputs.hosted_zone_id` | `string` | The Route 53 hosted zone Cloud Map created for a DNS namespace - what record-level tooling and zone delegations reference. |
| `status.outputs.http_name` | `string` | The name an HTTP namespace's DiscoverInstances calls use. |
| `status.outputs.service_ids` | `map<string, string>` | AWS-generated service IDs (srv-...) keyed by service name - what instance registrations and imports reference. |
| `status.outputs.service_arns` | `map<string, string>` | Service ARNs keyed by service name - what ECS service registries wire as the registry_arn. |
| `status.outputs.instance_service_ids` | `map<string, string>` | Each registration's owning service ID keyed by "{service_name}//{instance_id}" - the first half of the instance's composite import ID. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.services[].instances[].aliasDnsName` | AwsAlb | `status.outputs.load_balancer_dns_name` |
| `spec.services[].instances[].ec2InstanceId` | AwsEc2Instance | `status.outputs.instance_id` |

## See Also

- [Overview](../README.md)
