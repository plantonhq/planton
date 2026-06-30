# Pulumi Implementation - GCP Cert Manager Cert

This directory contains the Pulumi implementation (in Go) for the **GcpCertManagerCert** resource.

## Overview

The Pulumi implementation provides a programmatic way to deploy GCP SSL/TLS certificates using Go. It leverages the Pulumi GCP provider to create Certificate Manager certificates or Load Balancer certificates with automatic DNS validation.

## Directory Structure

```
pulumi/
├── main.go              # Entry point for Pulumi program
├── module/              # Core implementation
│   ├── main.go          # Resources function and provider setup
│   ├── cert_manager_cert.go  # Certificate creation logic
│   ├── locals.go        # Local variables and initialization
│   └── outputs.go       # Output constants
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Helper commands
├── debug.sh             # Debug script
├── README.md            # This file
└── examples.md          # Usage examples
```

## Module Components

### main.go (Entry Point)

The entry point loads the stack input and calls the module's `Resources` function:

```go
func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &gcpcertv1.GcpCertManagerCertStackInput{}
        if err := stackinput.Load(stackInput); err != nil {
            return err
        }
        return module.Resources(ctx, stackInput)
    })
}
```

### module/main.go (Provider Setup)

Sets up the GCP provider with credentials and calls the certificate creation logic:

- Configures GCP provider from stack input
- Decodes base64-encoded service account key
- Calls `certManagerCert` function

### module/cert_manager_cert.go (Core Logic)

Implements certificate creation for both types:

- **Certificate Manager (MANAGED)**:
  - Creates DNS authorizations for each domain
  - Creates DNS validation records in Cloud DNS
  - Provisions Certificate Manager certificate
  - Waits for validation

- **Load Balancer Certificates**:
  - Creates Google-managed SSL certificate
  - Handles domain list
  - Exports outputs

### module/locals.go (Local Variables)

Initializes local variables and GCP labels from metadata:

```go
type Locals struct {
    GcpCertManagerCert *gcpcertv1.GcpCertManagerCert
    GcpLabels          map[string]string
}
```

### module/outputs.go (Outputs)

Defines output constants for exported values:

- `certificate-id`
- `certificate-name`
- `certificate-domain-name`
- `certificate-status`

## Prerequisites

- Go 1.21 or later
- Pulumi CLI installed
- GCP account with appropriate permissions
- Cloud DNS managed zone for validation

## Required GCP Permissions

The service account needs the following IAM roles:

- `roles/certificatemanager.editor` (for Certificate Manager certificates)
- `roles/compute.loadBalancerAdmin` (for Load Balancer certificates)
- `roles/dns.admin` (for DNS validation records)

## Usage

### Direct Pulumi Usage

1. **Set up environment**:
   ```bash
   export PULUMI_CONFIG_PASSPHRASE="your-passphrase"
   pulumi stack init org/project/stack
   ```

2. **Create stack input**:
   Create a JSON file with the stack input structure.

3. **Run Pulumi**:
   ```bash
   pulumi up
   ```

### Using Planton CLI (Recommended)

```bash
planton pulumi up --manifest cert.yaml --stack org/project/stack
```

## Development

### Building

```bash
go build -o /tmp/gcp-cert-manager-cert .
```

### Debugging

Use the debug script:

```bash
./debug.sh
```

Or run Pulumi commands directly:

```bash
pulumi preview  # Preview changes
pulumi up       # Apply changes
pulumi destroy  # Destroy resources
```

### Using Makefile

```bash
make install   # Install dependencies
make preview   # Preview changes
make up        # Apply changes
make destroy   # Destroy resources
```

## Outputs

After successful deployment, the following outputs are available:

- `certificate-id`: The GCP certificate ID
- `certificate-name`: The certificate resource name
- `certificate-domain-name`: Primary domain name
- `certificate-status`: Certificate status (PROVISIONING/ACTIVE)

## Error Handling

The implementation includes comprehensive error handling:

- Provider initialization errors
- Certificate creation failures
- DNS record creation issues
- Validation failures

All errors are wrapped with context for easier debugging.

## Testing Locally

1. Create a test manifest
2. Set up GCP credentials
3. Run `pulumi preview` to see planned changes
4. Run `pulumi up` to create resources
5. Verify in GCP Console
6. Run `pulumi destroy` to clean up

## Integration with Planton

This Pulumi module is designed to work seamlessly with the Planton CLI, which:

- Converts YAML manifests to stack inputs
- Manages Pulumi state
- Handles credentials
- Provides a unified interface across providers

## Support

For issues or questions:

- Check the main README.md
- Review examples.md for usage patterns
- Open an issue on GitHub

