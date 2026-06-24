# Pulumi Implementation for Cloudflare Zero Trust Access Application

This directory contains the Pulumi-based Infrastructure-as-Code (IaC) implementation for provisioning Cloudflare Zero Trust Access Applications.

## Overview

The Pulumi implementation creates:
1. **Cloudflare Access Application**: The protected application resource with hostname and session configuration
2. **Cloudflare Access Policy**: The policy defining who can access the application (email-based or group-based rules)

The implementation is written in **Go** and uses the official [Pulumi Cloudflare provider](https://www.pulumi.com/registry/packages/cloudflare/).

## Prerequisites

1. **Pulumi CLI**: Install from [pulumi.com](https://www.pulumi.com/docs/get-started/install/)
2. **Go**: Version 1.21 or higher
3. **Cloudflare Account**: Active Cloudflare account with Zero Trust Access enabled
4. **Cloudflare API Token**: API token with Zero Trust permissions
5. **Cloudflare DNS Zone**: DNS zone for the domain you want to protect
6. **Identity Provider**: Configured identity provider (Google Workspace, Okta, Azure AD, etc.) in Cloudflare Zero Trust dashboard

## Directory Structure

```
iac/pulumi/
├── README.md           # This file
├── main.go             # Pulumi program entrypoint
├── Pulumi.yaml         # Pulumi project definition
├── Makefile            # Build and deployment helpers
├── debug.sh            # Debug script for local testing
└── module/
    ├── main.go         # Module orchestration
    ├── locals.go       # Local variables and configuration
    ├── outputs.go      # Output constants
    └── application.go  # Access Application and Policy creation logic
```

## Configuration

### Environment Variables

Set these environment variables before running Pulumi:

```bash
export CLOUDFLARE_API_TOKEN="your-cloudflare-api-token"
```

### Stack Input

The Pulumi program expects a `CloudflareZeroTrustAccessApplicationStackInput` JSON input containing:

```json
{
  "target": {
    "metadata": {
      "name": "my-access-app"
    },
    "spec": {
      "application_name": "My Application",
      "zone_id": "your-zone-id",
      "hostname": "app.example.com",
      "policy_type": "ALLOW",
      "allowed_emails": ["@example.com"],
      "session_duration_minutes": 480,
      "require_mfa": true,
      "allowed_google_groups": ["engineering@example.com"]
    }
  },
  "provider_config": {
    "api_token": "${CLOUDFLARE_API_TOKEN}"
  }
}
```

## Deployment

### Using OpenMCF CLI (Recommended)

The simplest way to deploy is via OpenMCF CLI:

```bash
# Create a manifest file
cat > access-app.yaml <<EOF
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: my-app
spec:
  application_name: "My Application"
  zone_id: "your-zone-id"
  hostname: "app.example.com"
  policy_type: ALLOW
  allowed_emails:
    - "@example.com"
  session_duration_minutes: 480
  require_mfa: true
EOF

# Deploy
planton apply -f access-app.yaml
```

### Using Pulumi CLI Directly

For advanced use cases or debugging:

```bash
# Navigate to the pulumi directory
cd iac/pulumi

# Initialize Pulumi stack
pulumi stack init dev

# Configure Cloudflare API token
pulumi config set cloudflare:apiToken $CLOUDFLARE_API_TOKEN --secret

# Set stack input (example)
pulumi config set target '{
  "metadata": {"name": "my-app"},
  "spec": {
    "application_name": "My Application",
    "zone_id": "your-zone-id",
    "hostname": "app.example.com",
    "policy_type": "ALLOW",
    "allowed_emails": ["@example.com"],
    "session_duration_minutes": 480,
    "require_mfa": true
  }
}'

# Preview changes
pulumi preview

# Deploy
pulumi up
```

## Debugging

### Local Testing

Use the included `debug.sh` script to test the Pulumi program locally:

```bash
# Set environment variables
export CLOUDFLARE_API_TOKEN="your-token"

# Run debug script
./debug.sh
```

### Enable Verbose Logging

```bash
pulumi up --logtostderr -v=9
```

### View Pulumi Logs

```bash
pulumi logs --follow
```

## Outputs

After successful deployment, the following outputs are available:

| Output Name | Description |
|-------------|-------------|
| `application_id` | The unique ID of the Cloudflare Access Application |
| `public_hostname` | The hostname being protected |
| `policy_id` | The ID of the Access Policy |

Access outputs:

```bash
pulumi stack output application_id
pulumi stack output public_hostname
pulumi stack output policy_id
```

## How It Works

### Application Creation

The `application.go` module creates:

1. **Zone Lookup** (`cloudflare.LookupZone`):
   - Looks up the Cloudflare zone to retrieve the account ID (required for the account-scoped policy and application)

2. **Access Policy** (`cloudflare.NewZeroTrustAccessPolicy`):
   - Creates a standalone, account-scoped policy with decision type (allow or deny)
   - Adds Include rules for allowed emails and Google groups
   - Adds Require rules for MFA enforcement (if enabled)

3. **Access Application** (`cloudflare.NewZeroTrustAccessApplication`):
   - Sets application name and hostname
   - Configures session duration (converted from minutes to duration string, e.g., "480m")
   - Sets type to "self_hosted"
   - References the policy through its `policies` list

### Policy Logic

**Include Rules** (who can access):
- Each allowed email is added as a separate Include block
- Each allowed Google group is added as a separate Include block
- Users matching **any** Include rule are allowed (OR logic)

**Require Rules** (additional requirements):
- If `require_mfa: true`, adds an auth_method requirement for "mfa"
- All Require rules must be satisfied (AND logic)

## Common Issues

### "Zone not found" Error

**Cause**: Invalid or missing `zone_id` in spec.

**Solution**: Verify the zone ID is correct:
```bash
curl -X GET "https://api.cloudflare.com/client/v4/zones" \
  -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN"
```

### "Insufficient permissions" Error

**Cause**: API token doesn't have Zero Trust permissions.

**Solution**: Ensure your API token has the following permissions:
- Account > Zero Trust > Edit
- Zone > DNS > Read

### MFA Not Being Enforced

**Cause**: Identity provider doesn't support MFA prompting, or `require_mfa: false`.

**Solution**:
1. Set `require_mfa: true` in the spec
2. Verify your IdP supports MFA (Google Workspace, Okta, Azure AD all support this)
3. Re-deploy: `planton apply -f access-app.yaml`

### Session Duration Format Error

**Cause**: Session duration isn't properly converted to Cloudflare's format.

**Solution**: The implementation automatically converts `session_duration_minutes` to a duration string (e.g., `480` becomes `"480m"`). Ensure the value is a positive integer.

## Development

### Building

```bash
cd iac/pulumi
go build -o pulumi-main main.go
```

### Testing

Run the Go tests:

```bash
go test ./...
```

### Code Structure

- **main.go**: Entrypoint that loads stack input and calls the module
- **module/main.go**: Orchestrates resource creation
- **module/locals.go**: Initializes local variables from stack input
- **module/application.go**: Contains the core logic for creating Access Application and Policy
- **module/outputs.go**: Defines output constant names

## Best Practices

1. **Use Pulumi Secrets**: Store sensitive values like API tokens as Pulumi secrets:
   ```bash
   pulumi config set cloudflare:apiToken $CLOUDFLARE_API_TOKEN --secret
   ```

2. **Separate Stacks**: Use separate Pulumi stacks for dev/staging/prod:
   ```bash
   pulumi stack init dev
   pulumi stack init staging
   pulumi stack init prod
   ```

3. **Version Control**: Commit `Pulumi.yaml` and stack configuration files to git

4. **State Backend**: Use Pulumi Cloud or self-hosted backend for team collaboration

5. **CI/CD Integration**: Integrate Pulumi with GitHub Actions, GitLab CI, or your CI/CD platform

## Further Reading

- **Component Architecture**: See [../../docs/README.md](../../docs/README.md) for architectural overview
- **User Guide**: See [../../README.md](../../README.md) for usage instructions
- **Examples**: See [../../examples.md](../../examples.md) for common use cases
- **Pulumi Cloudflare Docs**: [pulumi.com/registry/packages/cloudflare](https://www.pulumi.com/registry/packages/cloudflare/)
- **Cloudflare Zero Trust Docs**: [developers.cloudflare.com/cloudflare-one/applications](https://developers.cloudflare.com/cloudflare-one/applications)

## Support

For issues or questions:
- **OpenMCF**: [openmcf.org](https://openmcf.org)
- **Pulumi Support**: [pulumi.com/support](https://www.pulumi.com/support/)
- **Cloudflare Community**: [community.cloudflare.com](https://community.cloudflare.com)

