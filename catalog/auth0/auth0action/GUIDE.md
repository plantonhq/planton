# Auth0 Action Guide

## Security
## Platform Security Posture

The certifications below are Auth0's own published claims about their hosted platform (verify current status on Auth0's compliance page). They describe the vendor's service — never this catalog kind, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Auth0's published certifications and security standards:

- SOC 2 Type II (annual audit)
- ISO 27001, ISO 27018 (privacy controls)
- HIPAA BAA available on enterprise plans
- PCI DSS Level 1 Service Provider
- FedRAMP Authorized (moderate baseline)
- CSA STAR Level 2
- GDPR compliant with Data Processing Agreement

## Data Protection

- **Data residency**: US, EU, AU regions
- **Encryption in transit**: TLS 1.2+
- **Encryption at rest**: AES-256
- **Penetration testing**: Annual third-party assessments

## Action-Specific Security Notes

### Sandboxed Runtime

Action code runs in Auth0's sandboxed Node.js runtime environment. Each Action execution is isolated -- Actions cannot access the filesystem, spawn processes, or communicate with other Actions except through the event/API objects provided by Auth0.

### Secrets Management

Actions support injecting secrets via environment variables configured in the Auth0 dashboard or Management API. These secrets are:

- Encrypted at rest in Auth0's configuration store
- Available to Action code via `event.secrets`
- Not logged or exposed in Action execution logs
- Scoped to a single Action (not shared across Actions)

Never hardcode credentials in Action source code. Always use the secrets mechanism.

### Code Execution Context

Actions in the post-login trigger receive the full user profile and authentication context. The Action code has access to:

- User profile data (email, name, metadata)
- Authentication method details (connection, MFA status)
- Client information (application making the request)
- The ability to modify tokens, deny access, or redirect users

### Supply Chain Security

Actions can include npm dependencies. Auth0 resolves and bundles these at deploy time. Pin dependency versions to avoid unintended updates. Review dependencies for known vulnerabilities before deploying to production.

### Action Versioning

Auth0 maintains a version history for each Action. Only the deployed version executes in the authentication pipeline. Draft versions can be tested before deployment.

## Permissions
## Management API Scopes

Auth0 Action resources require the following Management API scopes for CRUD operations. These scopes must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:actions` | List and retrieve Action definitions and their code |
| Create | `create:actions` | Create new Actions with code, triggers, and secrets |
| Update | `update:actions` | Modify Action code, dependencies, and secrets |
| Delete | `delete:actions` | Remove Actions from the tenant |

## Deployment and Trigger Scopes

Managing Action deployments and trigger bindings requires the same core scopes. The following operations are covered by the scopes above:

| Operation | Required Scope | Description |
|-----------|---------------|-------------|
| Deploy Action | `update:actions` | Deploy a draft Action version to the pipeline |
| List versions | `read:actions` | Retrieve Action version history |
| Get trigger bindings | `read:actions` | List Actions bound to a specific trigger |
| Update trigger bindings | `update:actions` | Reorder or change Actions bound to a trigger |

## Minimum Required Scopes

For basic lifecycle management (create, read, update, deploy, delete), the minimum required scopes are:

```
read:actions create:actions update:actions delete:actions
```

## Secrets in Actions

Action secrets (environment variables) are managed through the Action update endpoint. No additional scopes are required beyond `update:actions` to set or modify Action secrets.

## Compliance
## Regulatory Frameworks

The table below records Auth0's own published compliance posture for their hosted platform, as Auth0 states it (verify current status in Auth0's trust documentation). These are vendor facts about the service this component configures — not properties of the component or of your deployment, and nothing here transfers to your application without your own assessment.

Auth0's published framework posture:

| Framework | Status | Notes |
|-----------|--------|-------|
| SOC 2 Type II | Certified | Annual audit cycle |
| ISO 27001:2022 | Certified | Information security management |
| ISO 27018:2019 | Certified | PII protection in public clouds |
| HIPAA | BAA Available | Enterprise plans only |
| PCI DSS Level 1 | Certified | Service provider certification |
| FedRAMP Moderate | Authorized | US government workloads |
| CSA STAR Level 2 | Certified | Cloud security assurance |
| GDPR | Compliant | Data Processing Agreement available |
| CCPA | Compliant | California consumer privacy |
| Privacy Shield | Historical | Deprecated framework, replaced by SCCs |

## Action-Specific Compliance Notes

### Code as Configuration

Actions contain executable code that runs in Auth0's infrastructure. For regulated environments, treat Action code as auditable configuration. Maintain version control for all Action source code outside Auth0 (e.g., in the planton repository) to satisfy change management requirements.

### Data Access in Actions

Post-login Actions receive the authenticated user's profile data. If your Action logic processes PII (name, email, phone), ensure the Action's behavior complies with applicable data protection regulations. Avoid logging PII in Action console output.

### Custom Claims and Data Minimization

Actions that add custom claims to tokens should follow GDPR data minimization principles. Only include claims that the downstream API requires. Token claims are visible to any party that can decode the token (access tokens are JWTs).

### Audit Trail

All Action CRUD operations (create, update, deploy, delete) are recorded in Auth0 tenant logs. Action execution failures are also logged. Log retention depends on plan tier (2 days free, 30 days enterprise).

### Third-Party Dependencies

Actions that use npm packages introduce third-party code into the authentication pipeline. For compliance-sensitive environments, maintain an approved package list and review transitive dependencies.

## Cost
## Pricing Model

Auth0 pricing is based on Monthly Active Users (MAUs), not on the number of resources created. Actions are free API objects with no per-resource cost.

## Free Tier

The Auth0 Free plan includes:

- 25,000 MAUs
- 1 tenant
- Unlimited Actions
- Unlimited Action triggers (post-login, pre-registration, etc.)

## Cost Impact

Creating, updating, or deleting Auth0 Action resources has no direct billing impact. There is no charge per Action definition, per deployment, or per trigger binding.

The only cost driver is the number of monthly active users authenticating through your tenant. Actions execute as part of the authentication pipeline at no additional cost.

## Execution Considerations

While Actions are free to define, they affect authentication latency. Each Action in a trigger's pipeline adds execution time to the authentication flow:

| Factor | Impact |
|--------|--------|
| Action code complexity | Directly increases login latency |
| External API calls in Actions | Adds network round-trip time |
| Number of Actions per trigger | Cumulative latency increase |
| Action timeout | 20 seconds maximum per Action |

## Rate Limits

The Auth0 Management API enforces rate limits on Action operations. Deploying Actions (creating new versions) counts against the Management API rate limit. Action executions during authentication flows are governed by separate per-tenant limits.

## npm Dependencies

Actions can include npm packages. There is no cost for npm dependencies, but each Action deployment must resolve and bundle its dependencies, which affects deployment time.
