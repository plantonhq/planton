# HetznerCloudCertificate Pulumi Module — Architecture Overview

## Data Flow

```
manifest.yaml
  └─> HetznerCloudCertificateStackInput (proto)
        ├── target: HetznerCloudCertificate
        │     ├── metadata.name → certificate name
        │     ├── metadata.org, env, id, labels → label computation
        │     └── spec (oneof certificate)
        │           ├── uploaded.certificate (PEM chain)
        │           ├── uploaded.private_key (PEM key, sensitive)
        │           └── managed.domain_names (string list)
        └── provider_config: HetznerCloudProviderConfig
              └── hcloud_token (or HCLOUD_TOKEN env var)
```

## Module Structure

1. **main.go (entrypoint)**: Loads `HetznerCloudCertificateStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML) via `stackinput.LoadStackInput`, then calls `module.Resources`.

2. **module/main.go**: Orchestrates resource creation:
   - Initializes locals from stack input
   - Creates a Hetzner Cloud Pulumi provider via `pulumihcloudprovider.Get`
   - Calls `certificate()` to create the certificate and export outputs

3. **module/locals.go**: Extracts provider config and target resource, then builds the label map:
   - Standard labels are set from metadata (`resource`, `name`, `kind`, `org`, `env`, `id`)
   - User-specified `metadata.labels` are merged in; standard labels take precedence on key conflicts

4. **module/certificate.go**: The routing and resource creation file. Contains three functions:

   **`certificate()`** — the router. Performs a Go type switch on `spec.Certificate` (the proto oneof) to dispatch to the correct creation function:
   - `*HetznerCloudCertificateSpec_Uploaded` → `uploadedCertificate()`
   - `*HetznerCloudCertificateSpec_Managed` → `managedCertificate()`
   - `default` → error (should never occur if proto validation passed)

   **`uploadedCertificate()`** — creates `hcloud.NewUploadedCertificate` with:
   - `Name` from `metadata.name`
   - `Certificate` from `config.Certificate` (PEM chain)
   - `PrivateKey` wrapped with `pulumi.ToSecret` (encrypted in state, masked in output)
   - `Labels` from locals

   **`managedCertificate()`** — creates `hcloud.NewManagedCertificate` with:
   - `Name` from `metadata.name`
   - `DomainNames` converted via `pulumi.ToStringArray`
   - `Labels` from locals

   Both functions export the same five outputs, making the module polymorphic.

5. **module/outputs.go**: Five constants matching `stack_outputs.proto` field names:
   - `OpCertificateId`, `OpType`, `OpFingerprint`, `OpNotValidBefore`, `OpNotValidAfter`

## Resource Graph

```
                    spec.Certificate (oneof)
                           │
              ┌────────────┴────────────┐
              │                         │
    uploaded variant             managed variant
              │                         │
    hcloud.UploadedCertificate  hcloud.ManagedCertificate
    ("certificate")             ("certificate")
              │                         │
              └────────────┬────────────┘
                           │
                    Shared Outputs:
                    ├── certificate_id  ← .ID()
                    ├── type            ← .Type
                    ├── fingerprint     ← .Fingerprint
                    ├── not_valid_before ← .NotValidBefore
                    └── not_valid_after  ← .NotValidAfter
```

## Key Design Points

- **Go type switch on proto oneof**: The `certificate()` function uses `spec.Certificate.(type)` to route to the correct creation path. This is a compile-time-safe pattern — if a new variant were added to the oneof, the Go compiler would not warn, but the `default` case returns an error. This is the idiomatic way to handle proto oneofs in Go.

- **`pulumi.ToSecret` for private key**: The uploaded certificate's private key is wrapped with `pulumi.ToSecret` before being passed to `UploadedCertificateArgs.PrivateKey`. The type assertion `.(pulumi.StringInput)` is needed because `ToSecret` returns `pulumi.Output`, not `pulumi.StringInput`. This ensures the secret is encrypted in the Pulumi state file and never printed to console.

- **No ID conversion**: Unlike components that reference external resources by ID (e.g., HetznerCloudSnapshot converts `serverId` from string to int), this module has no input ID conversion. The certificate IDs are outputs, not inputs. The only inputs are PEM strings and domain name lists.

- **Polymorphic outputs**: Both `uploadedCertificate()` and `managedCertificate()` export the same five outputs using the same constant names. Downstream consumers (like HetznerCloudLoadBalancer using `valueFrom`) can reference `status.outputs.certificate_id` regardless of which certificate type was used.

- **Single resource file**: Both creation paths live in `certificate.go` because they share the same output contract and are logically a single concern (create a certificate). The file contains three small functions rather than being split across two files.

- **Label merge strategy**: Same CG01 pattern as all other components. Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) always take precedence over user-specified labels. Both `hcloud_uploaded_certificate` and `hcloud_managed_certificate` support labels in the Hetzner Cloud API.
