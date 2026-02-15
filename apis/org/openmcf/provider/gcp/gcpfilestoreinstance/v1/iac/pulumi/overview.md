# GcpFilestoreInstance Pulumi Module Architecture

## Module Structure

```
iac/pulumi/
├── main.go           # Entrypoint: loads stack input, delegates to module
├── Pulumi.yaml       # Pulumi project configuration
├── debug.sh          # Build and debug helper
└── module/
    ├── main.go                 # Resources(): orchestrates resource creation
    ├── locals.go               # Locals struct + GCP label initialization
    ├── filestore_instance.go   # filestore.NewInstance() with all configuration
    └── outputs.go              # Output key constants
```

## Data Flow

1. `main.go` loads `GcpFilestoreInstanceStackInput` from Pulumi config
2. `module.Resources()` initializes locals (labels, provider config)
3. `pulumigoogleprovider.Get()` creates the GCP provider from service account key
4. `filestoreInstance()` creates the Filestore instance with:
   - File share (singular) with optional NFS export options
   - Network attachment (singular, MODE_IPV4 hardcoded)
   - Optional CMEK, deletion protection, performance config, protocol
5. Outputs exported: instance ID, name, IP addresses, file share name, create time

## Key Implementation Details

### File Share (Singular)

Filestore supports exactly one file share per instance. The Pulumi SDK uses `InstanceFileSharesArgs` (not an array type). NFS export options are mapped from the spec's repeated message to `InstanceFileSharesNfsExportOptionArray`.

### Network (Array with One Element)

The Pulumi SDK requires `InstanceNetworkArray` even though only one network is supported. We pass a single-element array. `modes` is hardcoded to `["MODE_IPV4"]`.

### IP Address Extraction

IP addresses are a computed field nested inside the networks array. We use `ApplyT` on `Networks` to extract `networks[0].IpAddresses` as a `StringArrayOutput`.

### Performance Config

Optional sub-message with mutually exclusive `FixedIops` and `IopsPerTb` blocks. Only set on the Pulumi resource when the spec provides it.

## GCP Labels

Framework labels applied automatically:
- `openmcf-resource: "true"`
- `openmcf-resource-name: <instance_name>`
- `openmcf-resource-kind: gcpfilestoreinstance`
- `openmcf-organization: <org>` (if set)
- `openmcf-environment: <env>` (if set)
- `openmcf-resource-id: <id>` (if set)
