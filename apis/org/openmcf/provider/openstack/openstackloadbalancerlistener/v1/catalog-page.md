# OpenStack Load Balancer Listener

Deploys an Octavia listener on an OpenStack load balancer, binding a protocol and port combination that accepts incoming traffic and forwards it to a backend pool. Supports HTTP, HTTPS pass-through, TCP, UDP, and TLS-terminated HTTPS with Barbican certificate integration.

## What Gets Created

When you deploy an OpenStackLoadBalancerListener resource, OpenMCF provisions:

- **Octavia Listener** -- a `loadbalancer.Listener` resource bound to the specified load balancer, accepting traffic on the configured protocol and port. When `defaultTlsContainerRef` is provided with the `TERMINATED_HTTPS` protocol, the listener terminates TLS using a certificate stored in Barbican. When `insertHeaders` is set, the listener injects HTTP headers (such as `X-Forwarded-For`) into requests before forwarding them to backends. When `allowedCidrs` is set, only traffic from those CIDR ranges reaches the listener.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **An existing load balancer** (by UUID or via an OpenStackLoadBalancer resource) in ACTIVE provisioning status
- **A Barbican secret container** holding the TLS certificate, private key, and optional intermediates if using the `TERMINATED_HTTPS` protocol

## Quick Start

Create a file `listener.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: http-listener
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackLoadBalancerListener.http-listener
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackloadbalancerlistener/v1/iac/pulumi/module
spec:
  loadbalancerId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  protocol: HTTP
  protocolPort: 80
```

Deploy:

```shell
openmcf apply -f listener.yaml
```

This creates an Octavia listener on the specified load balancer, accepting HTTP traffic on port 80.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `loadbalancerId` | `StringValueOrRef` | UUID of the load balancer to attach this listener to. Can reference an OpenStackLoadBalancer resource via `valueFrom`. ForceNew: changing this recreates the listener. | Required |
| `protocol` | `string` | The protocol the listener accepts. ForceNew: changing this recreates the listener. | Must be one of `HTTP`, `HTTPS`, `TCP`, `UDP`, `TERMINATED_HTTPS` |
| `protocolPort` | `int32` | The port on which the listener accepts traffic. ForceNew: changing this recreates the listener. | Must be between 1 and 65535 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | -- | Human-readable description of the listener. |
| `connectionLimit` | `int32` | -- | Maximum number of connections the listener allows. -1 means unlimited (Octavia default). Leave unset to use the Octavia default. |
| `defaultTlsContainerRef` | `string` | -- | URI of the Barbican TLS secret container for TLS termination. Required when `protocol` is `TERMINATED_HTTPS`. The container must hold the certificate, private key, and optional intermediates. |
| `insertHeaders` | `map<string, string>` | `{}` | Headers to insert into HTTP requests before forwarding to backends. Common use: `{"X-Forwarded-For": "true", "X-Forwarded-Proto": "true"}`. Only applicable to `HTTP` and `TERMINATED_HTTPS` protocols. |
| `allowedCidrs` | `string[]` | `[]` | List of CIDRs allowed to access this listener. When set, only traffic from these CIDRs reaches the listener; all other traffic is dropped. When empty, all traffic is allowed. |
| `adminStateUp` | `bool` | `true` | Administrative state of the listener. When false, the listener stops accepting traffic. |
| `tags` | `string[]` | `[]` | Tags applied to the listener in OpenStack. Must be unique within this resource. |
| `region` | `string` | provider default | Overrides the region from the provider config for this listener. |

## Examples

### Basic HTTP Listener

A minimal listener accepting HTTP traffic on port 80:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: http-listener
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackLoadBalancerListener.http-listener
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackloadbalancerlistener/v1/iac/pulumi/module
spec:
  loadbalancerId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  protocol: HTTP
  protocolPort: 80
  insertHeaders:
    X-Forwarded-For: "true"
    X-Forwarded-Proto: "true"
  tags:
    - web
    - http
```

### TLS-Terminated HTTPS Listener

A listener that terminates TLS at the load balancer using a Barbican certificate. Backends receive decrypted HTTP traffic with forwarded headers indicating the original protocol:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: https-listener
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackLoadBalancerListener.https-listener
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackloadbalancerlistener/v1/iac/pulumi/module
spec:
  loadbalancerId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  protocol: TERMINATED_HTTPS
  protocolPort: 443
  defaultTlsContainerRef: https://barbican.example.com/v1/containers/12345678-abcd-efgh-ijkl-123456789abc
  insertHeaders:
    X-Forwarded-For: "true"
    X-Forwarded-Proto: "true"
    X-Forwarded-Port: "true"
  description: Production HTTPS listener with TLS termination
  tags:
    - production
    - https
```

### Restricted Listener with Connection Limit

A listener limited to specific CIDR ranges and a maximum number of concurrent connections, suitable for internal admin panels or APIs that should not be publicly accessible:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: admin-api-listener
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackLoadBalancerListener.admin-api-listener
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackloadbalancerlistener/v1/iac/pulumi/module
spec:
  loadbalancerId: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  protocol: HTTP
  protocolPort: 8080
  connectionLimit: 5000
  allowedCidrs:
    - 10.0.0.0/8
    - 172.16.0.0/12
  description: Internal admin API with restricted access
  tags:
    - internal
    - admin
```

### Using Foreign Key References

Reference an OpenMCF-managed load balancer instead of hardcoding the UUID:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackLoadBalancerListener
metadata:
  name: app-listener
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackLoadBalancerListener.app-listener
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackloadbalancerlistener/v1/iac/pulumi/module
spec:
  loadbalancerId:
    valueFrom:
      kind: OpenStackLoadBalancer
      name: app-lb
      field: status.outputs.loadbalancer_id
  protocol: TERMINATED_HTTPS
  protocolPort: 443
  defaultTlsContainerRef: https://barbican.example.com/v1/containers/12345678-abcd-efgh-ijkl-123456789abc
  insertHeaders:
    X-Forwarded-For: "true"
    X-Forwarded-Proto: "true"
  allowedCidrs:
    - 10.0.0.0/8
  adminStateUp: true
  tags:
    - production
    - app-tier
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `listener_id` | `string` | UUID of the created Octavia listener. This is the primary output used as a foreign key by pools. |
| `name` | `string` | Name of the listener, derived from `metadata.name` |
| `protocol` | `string` | The protocol the listener accepts (`HTTP`, `HTTPS`, `TCP`, `UDP`, or `TERMINATED_HTTPS`) |
| `protocol_port` | `int32` | The port on which the listener accepts traffic |
| `region` | `string` | OpenStack region where the listener was created |

## Related Components

- [OpenStackLoadBalancer](/docs/catalog/openstack/openstackloadbalancer) -- provides the load balancer that this listener attaches to
- [OpenStackLoadBalancerPool](/docs/catalog/openstack/openstackloadbalancerpool) -- defines the backend pool that receives traffic forwarded by this listener
