# AWS Cloud Map Namespace

Deploys an AWS Cloud Map namespace — the service-discovery registry ECS services and custom applications look each other up in — with its services and their statically registered instances managed in-line. The namespace type shapes the whole surface: HTTP namespaces are API-only (consumers call DiscoverInstances, no DNS records exist), PRIVATE_DNS namespaces create a private hosted zone visible inside one VPC, and PUBLIC_DNS namespaces create an internet-resolvable zone. The namespace's name is its domain — service `api` in namespace `corp.internal` answers at `api.corp.internal` — and both the name and the type are fixed for life.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cloud Map Namespace** — exactly one of the three provider resources per `type` (HTTP, private DNS, or public DNS), named after `metadata.name`. DNS namespaces also get a Route 53 hosted zone created by Cloud Map itself — private zones associated to `vpcId`, public zones live on the internet
- **Cloud Map Services** — one per `services` entry, keyed by name: the DNS records instances publish (A/AAAA/SRV/CNAME with TTL, multivalue or weighted routing), Route 53 health checks (public namespaces only), or the custom-health marker
- **Instance Registrations** — one per `instances` entry under each service: an IP and port, a CNAME to any endpoint, a Route 53 alias to a load balancer, or an EC2 instance — plus custom attributes for API-side discovery. Registration is an AWS upsert keyed by `instanceId`, so re-applying the same id updates in place
- **Route 53 Health Checks** — created by Cloud Map (not the module) for instances under a health-checked service; they are linked resources only Cloud Map can delete, and they linger briefly in the account after teardown until AWS's own cleanup sweeps them

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Cloud Map (`servicediscovery`) permissions, plus Route 53 permissions when the namespace type creates a hosted zone. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Only for PRIVATE_DNS: the VPC the namespace's zone should be visible in, with DNS resolution enabled (`vpcId`).
- Only for PUBLIC_DNS: a name Route 53 will accept — it refuses `example.com` and its subdomains as reserved (`.test` names are accepted, and you never need to own or delegate the domain just to create the namespace).

## Deploy

### Console

Open the deployment store, find **AWS Cloud Map Namespace**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region and type, the VPC for private namespaces, and the services with their records, health checking, and static registrations. Start from the **Private DNS Namespace** preset or the **HTTP (API-Only) Namespace** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudMapNamespace
metadata:
  name: corp.internal
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  type: PRIVATE_DNS
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: app-vpc
      fieldPath: status.outputs.vpc_id
  description: Service discovery for the corp platform
  services:
    - name: api
      dnsConfig:
        records:
          - type: A
            ttl: 10
      healthCheckCustomConfig: {}
    - name: db
      dnsConfig:
        records:
          - type: CNAME
            ttl: 30
        routingPolicy: WEIGHTED
      instances:
        - instanceId: primary
          cname: mydb.cluster-abc.us-east-1.rds.amazonaws.com
```

```shell
planton apply -f cloud-map-namespace.yaml
```

This creates the `corp.internal` private zone in the app VPC with an `api` service for runtime platforms to register into and a `db` service resolving to the database endpoint by CNAME. An Infra Job tracks the provisioning in real time.

### InfraChart

When a private namespace deploys alongside its VPC in one chart, wire the VPC reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  type: PRIVATE_DNS
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: app-vpc
      fieldPath: status.outputs.vpc_id
  services:
    - name: api
      dnsConfig:
        records:
          - type: A
            ttl: 10
      healthCheckCustomConfig: {}
```

The InfraPipeline resolves the dependency graph, deploys the VPC first, then creates the namespace's private zone inside it.

## Key Configuration

These are the most important decisions when configuring a Cloud Map namespace. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The type decides everything, once** — `type` is fixed for life and shapes what the rest of the spec may say: HTTP services carry no `dnsConfig` (discovery is API-only), Route 53 health checks work only under PUBLIC_DNS (the checkers live on the public internet), and only PRIVATE_DNS takes a `vpcId`. Moving between models later means a new namespace and re-registering everything in it.

**Declared instances and runtime registrations do not mix** — ECS service discovery registers and deregisters task instances at runtime. Declare static `instances` only on services this manifest fully owns; a declared instance on an ECS-managed service invites two owners for one registry. And never set `forceDestroy` on such a service — it deregisters EVERYTHING in the service on destroy, including registrations ECS made that this manifest never declared.

**Routing policy and record type are a contract** — MULTIVALUE (the default) answers with up to eight healthy records; WEIGHTED answers with one. A CNAME record answers with exactly one value, so AWS rejects CNAME services under MULTIVALUE — the spec requires the explicit `routingPolicy: WEIGHTED` so the manifest fails at validation instead of at deploy. The policy is fixed for life per service.

**SRV changes the discovery contract for clients** — a service publishing SRV returns hostname-plus-port tuples, and clients must ask for SRV records (most naive resolvers ask A). Use SRV only where consumers understand it — gRPC clients, service meshes; A records plus a fixed port cover the common case with zero client changes.

**Custom health is a heartbeat, not monitoring** — `healthCheckCustomConfig` services mark instances healthy until told otherwise; the workload must push status updates via the Cloud Map API to ever mark one unhealthy. Silence means permanently healthy. Route 53 health checks (`healthCheckConfig`) probe for you, but only public, routable endpoints: registration rejects private-range IPs under a health-checked service, and IP instances there must declare a `port` for the probe.

**The private namespace's VPC is write-only** — AWS never returns which VPC a private namespace binds; the provider carries it in configuration, and imports need the `{namespace_id}:{vpc_id}` composite. If the VPC id is lost, it is recovered from the hosted zone's associations, not from Cloud Map.

**Destroy registrations through the module** — deregistering an already-gone instance errors at the provider with no not-found tolerance. An out-of-band deregistration leaves the next declarative destroy red until state is reconciled.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsVpc** | `vpcId` | `status.outputs.vpc_id` |
| **AwsAlb** | `services[].instances[].aliasDnsName` | `status.outputs.load_balancer_dns_name` |
| **AwsEc2Instance** | `services[].instances[].ec2InstanceId` | `status.outputs.instance_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `service_arns` | Service ARNs keyed by service name | What ECS service registries wire as the registry ARN for task auto-registration |
| `hosted_zone_id` | The Route 53 hosted zone Cloud Map created for a DNS namespace | Record-level tooling and zone delegations |
| `http_name` | The name an HTTP namespace's DiscoverInstances calls use | Application discovery configuration |
| `namespace_id` | The namespace's id (`ns-...`) | Addressing the namespace in AWS tooling; the provider's import ID |

`namespace_arn`, `service_ids`, and `instance_service_ids` are also exported, but they are operational echoes kept for audits and the composite import IDs of services and registrations.

## Common Patterns

**Private DNS for a service platform** — the ECS-style shape: `api.corp.internal` for tasks to register into (wire the service ARN to the ECS service's registry), plus a statically registered CNAME so the database resolves under the same private domain. Start from the **Private DNS Namespace** preset.

**Discovery without DNS** — an HTTP namespace where consumers call DiscoverInstances with the `http_name` output and filter on custom attributes. The shape for worker fleets and anything that should never be resolvable. Start from the **HTTP (API-Only) Namespace** preset.

**One private domain for managed endpoints** — a private namespace whose services are all statically registered CNAMEs to managed endpoints (RDS clusters, ElastiCache, internal ALBs). Applications learn one stable name per dependency, and an endpoint swap is a one-line instance update instead of an application redeploy.

## Works With

- [**AWS VPC**](/infra-catalog/aws-vpc) — where a private namespace's zone is visible, wired via `vpcId`
- [**AWS ECS Service**](/infra-catalog/aws-ecs-service) — registers its tasks into a service here via the `service_arns` output
- [**AWS ALB**](/infra-catalog/aws-alb) — the load balancer an alias registration points at, wired via `aliasDnsName`
- [**AWS EC2 Instance**](/infra-catalog/aws-ec2-instance) — registered by id via `ec2InstanceId`; AWS derives the address
