# Terraform Module - Not Implemented

⚠️ **The Terraform implementation for KubernetesStrimziKafkaOperator is currently not available.**

## Recommended Approach

Use the **Pulumi module** for deployment:

```bash
cd ../pulumi
pulumi up
```

See [../pulumi/README.md](../pulumi/README.md) for complete instructions.

## Why Pulumi Only?

The Pulumi implementation provides proper CRD installation, Helm chart management, and type-safe configuration. A Terraform implementation may be added in future releases based on demand.

## Support

- **Pulumi users**: See [../pulumi/README.md](../pulumi/README.md)
- **General questions**: Refer to [../../README.md](../../README.md)

