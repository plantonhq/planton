# Cloudflare Zero Trust Access Application Examples

This guide provides concrete, copy-and-paste examples for common Zero Trust Access Application deployment scenarios using OpenMCF.

## Table of Contents

- [Minimal Configuration](#minimal-configuration)
- [Company-Wide Internal Tool](#company-wide-internal-tool)
- [Team-Specific Dashboard](#team-specific-dashboard)
- [Staging Environment Access](#staging-environment-access)
- [Production Admin Console](#production-admin-console)
- [Development Environment](#development-environment)
- [Multiple Applications with Different Policies](#multiple-applications-with-different-policies)
- [Multi-Environment Setup](#multi-environment-setup)
- [Email-Based Access](#email-based-access)
- [Google Workspace Group-Based Access](#google-workspace-group-based-access)
- [Mixed Email and Group Access](#mixed-email-and-group-access)
- [Complete CI/CD Workflow](#complete-cicd-workflow)

---

## Minimal Configuration

The simplest possible Zero Trust Access Application with only required fields.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: minimal-app
spec:
  applicationName: "Minimal App"
  zoneId:
    value: "abc123def456..."
    # Or reference a CloudflareDnsZone resource:
    # zoneId:
    #   valueFrom:
    #     kind: CloudflareDnsZone
    #     name: "my-zone"
    #     fieldPath: "status.outputs.zone_id"
  hostname: "app.example.com"
```

**Deploy:**

```bash
planton apply -f minimal-app.yaml
```

**Use Case:** Quick experimentation or proof-of-concept. Not recommended for production (no access restrictions configured).

---

## Company-Wide Internal Tool

Internal tool accessible to all company employees with email domain matching.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: company-wiki
  labels:
    app: wiki
    audience: company-wide
spec:
  applicationName: "Company Wiki"
  zoneId:
    value: "abc123def456..."
  hostname: "wiki.example.com"
  policyType: ALLOW
  allowedEmails:
    - "@example.com"
  sessionDurationMinutes: 480  # 8 hours
  requireMfa: false
```

**Deploy:**

```bash
planton apply -f company-wiki.yaml
```

**Use Case:** Internal documentation, company wiki, or knowledge base accessible to all employees. MFA not required as content is low-sensitivity.

**Access Pattern**: Any user with an `@example.com` email address can access after authenticating with their identity provider.

---

## Team-Specific Dashboard

Dashboard accessible only to engineering team members via Google Workspace groups.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: eng-dashboard
  labels:
    app: dashboard
    team: engineering
spec:
  applicationName: "Engineering Dashboard"
  zoneId:
    value: "abc123def456..."
  hostname: "eng-dash.example.com"
  policyType: ALLOW
  allowedGoogleGroups:
    - "engineering@example.com"
  sessionDurationMinutes: 720  # 12 hours
  requireMfa: true
```

**Deploy:**

```bash
planton apply -f eng-dashboard.yaml
```

**Prerequisites:**
1. Google Workspace configured as identity provider in Cloudflare Zero Trust dashboard
2. Google Workspace group `engineering@example.com` exists and contains relevant users

**Use Case:** Internal tools, dashboards, or monitoring systems scoped to a specific team. MFA enforced for security.

---

## Staging Environment Access

Staging environment accessible to engineering and QA teams.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: staging-admin
  labels:
    environment: staging
    app: admin-console
spec:
  applicationName: "Staging Admin Console"
  zoneId:
    value: "abc123def456..."
  hostname: "staging-admin.example.com"
  policyType: ALLOW
  allowedGoogleGroups:
    - "engineering@example.com"
    - "qa@example.com"
  sessionDurationMinutes: 480  # 8 hours
  requireMfa: true
```

**Deploy:**

```bash
planton apply -f staging-admin.yaml
```

**Use Case:** Staging or preview environments where multiple teams need access for testing and validation. MFA required as staging often contains production-like data.

---

## Production Admin Console

High-security production console with strict access control and short session duration.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: prod-console
  labels:
    environment: production
    app: admin-console
    tier: critical
spec:
  applicationName: "Production Admin Console"
  zoneId:
    value: "abc123def456..."
  hostname: "prod.example.com"
  policyType: ALLOW
  allowedGoogleGroups:
    - "prod-admins@example.com"
  sessionDurationMinutes: 120  # 2 hours
  requireMfa: true
```

**Deploy:**

```bash
planton apply -f prod-console.yaml
```

**Security Highlights:**
- **Group-based access**: Only members of `prod-admins@example.com` group can access
- **MFA enforced**: Multi-factor authentication required
- **Short session**: 2-hour session duration forces re-authentication frequently

**Use Case:** Production database consoles, admin panels, financial dashboards, or any high-security application with access to production data.

---

## Development Environment

Development environment accessible to all engineers without strict session limits.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: dev-console
  labels:
    environment: development
    app: console
spec:
  applicationName: "Development Console"
  zoneId:
    value: "abc123def456..."
  hostname: "dev.example.com"
  policyType: ALLOW
  allowedEmails:
    - "@example.com"
  sessionDurationMinutes: 1440  # 24 hours
  requireMfa: false
```

**Deploy:**

```bash
planton apply -f dev-console.yaml
```

**Use Case:** Development sandboxes where convenience is prioritized over strict security. MFA not required to reduce friction for developers.

---

## Multiple Applications with Different Policies

Deploy multiple applications with varying security policies in one workflow.

### Directory Structure

```
access-apps/
├── company-wiki.yaml
├── eng-dashboard.yaml
├── staging-admin.yaml
└── prod-console.yaml
```

### Deployment Script

**File:** `deploy-all.sh`

```bash
#!/bin/bash
set -e

echo "Deploying Zero Trust Access Applications..."

planton apply -f access-apps/company-wiki.yaml
echo "✅ Company Wiki deployed"

planton apply -f access-apps/eng-dashboard.yaml
echo "✅ Engineering Dashboard deployed"

planton apply -f access-apps/staging-admin.yaml
echo "✅ Staging Admin deployed"

planton apply -f access-apps/prod-console.yaml
echo "✅ Production Console deployed"

echo "🎉 All applications deployed successfully!"
```

**Usage:**

```bash
chmod +x deploy-all.sh
./deploy-all.sh
```

---

## Multi-Environment Setup

Complete multi-environment setup with development, staging, and production applications.

### Development Application

**File:** `dev-app.yaml`

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: my-app-dev
  labels:
    environment: development
    app: my-application
spec:
  applicationName: "My App - Development"
  zoneId:
    value: "abc123def456..."
  hostname: "dev.my-app.example.com"
  policyType: ALLOW
  allowedEmails:
    - "@example.com"
  sessionDurationMinutes: 1440  # 24 hours
  requireMfa: false
```

### Staging Application

**File:** `staging-app.yaml`

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: my-app-staging
  labels:
    environment: staging
    app: my-application
spec:
  applicationName: "My App - Staging"
  zoneId:
    value: "abc123def456..."
  hostname: "staging.my-app.example.com"
  policyType: ALLOW
  allowedGoogleGroups:
    - "engineering@example.com"
    - "qa@example.com"
  sessionDurationMinutes: 480  # 8 hours
  requireMfa: true
```

### Production Application

**File:** `prod-app.yaml`

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: my-app-prod
  labels:
    environment: production
    app: my-application
spec:
  applicationName: "My App - Production"
  zoneId:
    value: "abc123def456..."
  hostname: "my-app.example.com"
  policyType: ALLOW
  allowedGoogleGroups:
    - "prod-users@example.com"
  sessionDurationMinutes: 240  # 4 hours
  requireMfa: true
```

### Environment-Specific Deployment

**File:** `deploy-env.sh`

```bash
#!/bin/bash
set -e

ENVIRONMENT=$1

if [ "$ENVIRONMENT" = "dev" ]; then
  planton apply -f dev-app.yaml
  echo "✅ Development environment deployed"
elif [ "$ENVIRONMENT" = "staging" ]; then
  planton apply -f staging-app.yaml
  echo "✅ Staging environment deployed"
elif [ "$ENVIRONMENT" = "prod" ]; then
  planton apply -f prod-app.yaml
  echo "✅ Production environment deployed"
else
  echo "Usage: ./deploy-env.sh [dev|staging|prod]"
  exit 1
fi
```

**Usage:**

```bash
chmod +x deploy-env.sh
./deploy-env.sh dev
./deploy-env.sh staging
./deploy-env.sh prod
```

---

## Email-Based Access

Allow specific email addresses or domain-wide access.

### Specific Email Addresses

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: contractor-portal
spec:
  applicationName: "Contractor Portal"
  zoneId:
    value: "abc123def456..."
  hostname: "contractors.example.com"
  policyType: ALLOW
  allowedEmails:
    - "john.doe@contractor.com"
    - "jane.smith@vendor.com"
    - "@trustedpartner.com"
  sessionDurationMinutes: 480
  requireMfa: true
```

**Use Case:** External contractors or vendors who need temporary access. Mix individual emails with domain-wide access for trusted partners.

### Domain-Wide Access

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: internal-docs
spec:
  applicationName: "Internal Documentation"
  zoneId:
    value: "abc123def456..."
  hostname: "docs.example.com"
  policyType: ALLOW
  allowedEmails:
    - "@example.com"
  sessionDurationMinutes: 720
  requireMfa: false
```

**Use Case:** Company-wide resources accessible to anyone with a company email domain.

---

## Google Workspace Group-Based Access

Use Google Workspace groups for role-based access control.

### Single Group

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: finance-dashboard
spec:
  applicationName: "Finance Dashboard"
  zoneId:
    value: "abc123def456..."
  hostname: "finance.example.com"
  policyType: ALLOW
  allowedGoogleGroups:
    - "finance@example.com"
  sessionDurationMinutes: 480
  requireMfa: true
```

### Multiple Groups

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: shared-analytics
spec:
  applicationName: "Shared Analytics Platform"
  zoneId:
    value: "abc123def456..."
  hostname: "analytics.example.com"
  policyType: ALLOW
  allowedGoogleGroups:
    - "engineering@example.com"
    - "product@example.com"
    - "marketing@example.com"
  sessionDurationMinutes: 720
  requireMfa: true
```

**Use Case:** Cross-functional tools accessible to multiple departments.

---

## Mixed Email and Group Access

Combine email-based and group-based access for flexible policies.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: mixed-access-app
spec:
  applicationName: "Mixed Access Application"
  zoneId:
    value: "abc123def456..."
  hostname: "mixed.example.com"
  policyType: ALLOW
  allowedEmails:
    - "ceo@example.com"
    - "cto@example.com"
  allowedGoogleGroups:
    - "leadership@example.com"
    - "engineering@example.com"
  sessionDurationMinutes: 480
  requireMfa: true
```

**Access Pattern**: Any user matching **either** the email list **or** the group list can access.

**Use Case**: Executive dashboards or strategic tools where specific individuals and specific teams need access.

---

## Complete CI/CD Workflow

GitHub Actions workflow for automated deployment of Zero Trust Access Applications.

**File:** `.github/workflows/deploy-access-apps.yml`

```yaml
name: Deploy Cloudflare Zero Trust Access Applications

on:
  push:
    branches:
      - main
      - develop
  pull_request:
    branches:
      - main

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Install OpenMCF CLI
        run: |
          curl -fsSL https://get.openmcf.org | bash
          planton version

      - name: Determine environment
        id: env
        run: |
          if [[ "${{ github.ref }}" == "refs/heads/main" ]]; then
            echo "environment=prod" >> $GITHUB_OUTPUT
            echo "manifest=access-apps/prod-app.yaml" >> $GITHUB_OUTPUT
          elif [[ "${{ github.ref }}" == "refs/heads/develop" ]]; then
            echo "environment=staging" >> $GITHUB_OUTPUT
            echo "manifest=access-apps/staging-app.yaml" >> $GITHUB_OUTPUT
          else
            echo "environment=dev" >> $GITHUB_OUTPUT
            echo "manifest=access-apps/dev-app.yaml" >> $GITHUB_OUTPUT
          fi

      - name: Deploy Access Application
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
        run: |
          planton apply -f ${{ steps.env.outputs.manifest }}

      - name: Verify Deployment
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
        run: |
          echo "Deployment to ${{ steps.env.outputs.environment }} completed successfully"
          planton output application_id
          planton output public_hostname
```

**Required GitHub Secrets:**

- `CLOUDFLARE_API_TOKEN`: Cloudflare API token with Zero Trust permissions

---

## Session Duration Examples

Different session durations for different security requirements:

### Ultra-Short (1 hour) - Maximum Security

```yaml
spec:
  applicationName: "Financial Admin Console"
  hostname: "finance-admin.example.com"
  sessionDurationMinutes: 60
  requireMfa: true
```

### Short (2-4 hours) - High Security

```yaml
spec:
  applicationName: "Production Database Console"
  hostname: "db-console.example.com"
  sessionDurationMinutes: 120
  requireMfa: true
```

### Medium (8 hours) - Standard Workday

```yaml
spec:
  applicationName: "Internal Dashboard"
  hostname: "dash.example.com"
  sessionDurationMinutes: 480
  requireMfa: true
```

### Long (24 hours) - Developer Convenience

```yaml
spec:
  applicationName: "Development Tools"
  hostname: "dev-tools.example.com"
  sessionDurationMinutes: 1440
  requireMfa: false
```

---

## Access Policy Types

### Allow Policy (Most Common)

Grant access to users matching the rules:

```yaml
spec:
  applicationName: "Standard Application"
  hostname: "app.example.com"
  policyType: ALLOW
  allowedEmails:
    - "@example.com"
```

### Block Policy (Rare Use Case)

Explicitly deny access to users matching the rules:

```yaml
spec:
  applicationName: "Restricted Application"
  hostname: "restricted.example.com"
  policyType: BLOCK
  allowedEmails:
    - "blocked-user@example.com"
```

**Use Case**: Temporarily block specific users or domains while keeping default allow policy.

---

## Validation

After deployment, verify your Access Application:

```bash
# Check outputs
planton output application_id
planton output public_hostname
planton output policy_id

# Test access
curl https://your-app.example.com
# Should redirect to identity provider login
```

---

## Monitoring and Auditing

### View Access Logs in Cloudflare Dashboard

1. Navigate to **Zero Trust** → **Logs** → **Access**
2. Filter by application name or hostname
3. Review authentication attempts, user emails, IPs, countries, MFA status

### Alert on Suspicious Activity

Configure alerts for:
- Repeated failed login attempts
- Logins from unexpected countries
- Disabled users attempting access
- Non-MFA logins to MFA-required applications

---

## Common Patterns

### Per-Team Application

Each team gets their own application with team-specific access:

```yaml
# Engineering Team
---
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: eng-tools
spec:
  applicationName: "Engineering Tools"
  hostname: "eng.example.com"
  allowedGoogleGroups:
    - "engineering@example.com"

# Marketing Team
---
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: marketing-tools
spec:
  applicationName: "Marketing Tools"
  hostname: "marketing.example.com"
  allowedGoogleGroups:
    - "marketing@example.com"
```

### Progressive Security

Start with lenient access, tighten over time:

**Phase 1 - Initial Rollout**:
```yaml
spec:
  allowedEmails:
    - "@example.com"
  sessionDurationMinutes: 1440
  requireMfa: false
```

**Phase 2 - Add MFA**:
```yaml
spec:
  allowedEmails:
    - "@example.com"
  sessionDurationMinutes: 1440
  requireMfa: true
```

**Phase 3 - Restrict to Groups**:
```yaml
spec:
  allowedGoogleGroups:
    - "approved-users@example.com"
  sessionDurationMinutes: 480
  requireMfa: true
```

---

## Next Steps

- **Explore Architecture**: Read [docs/README.md](docs/README.md) for in-depth architectural guidance
- **User Guide**: See [README.md](README.md) for general usage instructions
- **Pulumi Docs**: Check [iac/pulumi/README.md](iac/pulumi/README.md) for Pulumi-specific details
- **Terraform Docs**: Check [iac/tf/README.md](iac/tf/README.md) for Terraform-specific details

---

## Support

For questions or issues:
- **OpenMCF**: [openmcf.org](https://openmcf.org)
- **Cloudflare Zero Trust Docs**: [developers.cloudflare.com/cloudflare-one/applications](https://developers.cloudflare.com/cloudflare-one/applications)

