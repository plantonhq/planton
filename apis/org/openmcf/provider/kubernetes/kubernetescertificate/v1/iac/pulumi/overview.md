# KubernetesCertificate Pulumi Module Architecture

## Data Flow

```
StackInput (spec + cluster credentials)
   ↓
Locals (namespace, certificate_name, dns_names, secret_name, issuer_ref, duration, private_key)
   ↓
Resources (Certificate CR)
   ↓
Outputs (namespace, certificate_name, secret_name)
```

## File Organization

| File | Responsibility |
|------|---------------|
| `main.go` | Creates the cert-manager Certificate CR using the typed crd2pulumi SDK |
| `locals.go` | Extracts spec fields, maps proto enums to cert-manager CRD values, resolves issuer ref oneof |
| `outputs.go` | Defines output constant names matching stack_outputs.proto |

## Key Design Decisions

- **Typed SDK** -- Uses `certmanagerv1.NewCertificate` (crd2pulumi) for compile-time type safety, matching the pattern used by all 15+ OpenMCF ingress components
- **Issuer ref oneof** -- Proto oneof maps to cert-manager's `issuerRef.kind` ("ClusterIssuer" vs "Issuer") via Go type switch
- **Proto enum mapping** -- Proto uses lowercase names (rsa, pkcs1, always); cert-manager CRD expects PascalCase (RSA, PKCS1, Always). Mapping functions bridge this gap.
- **Nil private key = CRD defaults** -- When the proto `private_key` submessage is absent, no `privateKey` block is emitted, letting cert-manager apply its own defaults (RSA 2048, PKCS1, Always)
- **Optional duration** -- Duration and RenewBefore are only set when `duration_config` is present in the spec
