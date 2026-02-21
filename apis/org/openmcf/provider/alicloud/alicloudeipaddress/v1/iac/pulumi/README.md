# Pulumi Module to Deploy AlicloudEipAddress

This module provisions an Alibaba Cloud Elastic IP Address (EIP) using the
`ecs.EipAddress` Pulumi resource. The EIP is a standalone public IPv4 address
that can be associated with NAT gateways, load balancers, VPN gateways, and
ECS instances.

## CLI Usage (OpenMCF Pulumi)

```shell
# Preview
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh
openmcf pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Destroy
openmcf pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

**Note**: Credentials are provided via stack input (CLI), not in the manifest `spec`.

For more examples, see [`examples.md`](./examples.md) and [`hack/manifest.yaml`](../hack/manifest.yaml).
