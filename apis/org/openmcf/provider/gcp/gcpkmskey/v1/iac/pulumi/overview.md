# GcpKmsKey — Pulumi Architecture Overview

## Execution Flow

```
StackInput (GcpKmsKeyStackInput)
  │
  ├── target: GcpKmsKey (api.proto envelope)
  │     ├── metadata: CloudResourceMetadata
  │     └── spec: GcpKmsKeySpec
  │           ├── key_ring_id (StringValueOrRef → GcpKmsKeyRing)
  │           ├── key_name
  │           ├── purpose
  │           ├── rotation_period
  │           ├── destroy_scheduled_duration
  │           ├── version_template { algorithm, protection_level }
  │           └── skip_initial_version_creation
  │
  └── provider_config: GcpProviderConfig

  ↓ module.Resources()

  1. initializeLocals() → Locals { GcpLabels, spec ref }
  2. pulumigoogleprovider.Get() → gcp.Provider
  3. kmsKey() → kms.NewCryptoKey
       ├── Maps spec fields to CryptoKeyArgs
       ├── Applies framework GcpLabels
       ├── Conditionally sets optional fields
       └── Exports key_id (.ID()) and key_name (.Name)
```

## Resource Mapping

| Spec Field | Pulumi Property | Notes |
|------------|-----------------|-------|
| `key_ring_id` | `KeyRing` | Fully qualified path from GcpKmsKeyRing |
| `key_name` | `Name` | GCP resource name |
| `purpose` | `Purpose` | Optional, defaults to ENCRYPT_DECRYPT |
| `rotation_period` | `RotationPeriod` | Optional, only for symmetric keys |
| `destroy_scheduled_duration` | `DestroyScheduledDuration` | Optional, defaults to 30 days |
| `version_template.algorithm` | `VersionTemplate.Algorithm` | Required within template |
| `version_template.protection_level` | `VersionTemplate.ProtectionLevel` | SOFTWARE or HSM |
| `skip_initial_version_creation` | `SkipInitialVersionCreation` | Optional, creation-time only |
| (framework) | `Labels` | Computed from metadata |
