# AwsElasticIp Pulumi Module — Architecture Overview

## Resource Graph

This is one of the simplest modules in OpenMCF. It creates a single AWS resource:

```
AwsElasticIp (ec2.Eip)
  └── domain: "vpc" (hardcoded)
  └── tags: merged from metadata
  └── [optional] public_ipv4_pool
  └── [optional] address
  └── [optional] network_border_group
```

## Data Flow

```
StackInput (manifest YAML)
  → initializeLocals() → Locals{tags, resource ref}
  → eip() → ec2.NewEip → EipResult{allocationId, publicIp, arn, publicDns}
  → ctx.Export() → Stack Outputs
```

## Design Decisions

1. **Single resource, no sub-resources.** The EIP is a single `aws_eip` / `ec2.Eip`. No subnet groups, parameter groups, or auxiliary resources.

2. **Domain hardcoded to "vpc".** EC2-Classic is retired. There is no valid reason to expose `domain` as a user-facing field.

3. **No association.** This module allocates the IP; it does not bind it to an instance or ENI. Association is the consumer's concern (NLB uses `allocation_id` directly; instances use `AwsEipAssociation`).

4. **All spec fields are optional.** The empty-spec case (allocate from Amazon's default pool) is the primary use case.
