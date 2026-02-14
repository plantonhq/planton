---
title: "Securing Applications with Cloudflare Zero Trust Access: A Production Guide"
description: "Securing Applications with Cloudflare Zero Trust Access: A Production Guide deployment documentation"
icon: "package"
order: 100
componentName: "cloudflarezerotrustaccessapplication"
---

# Securing Applications with Cloudflare Zero Trust Access: A Production Guide

## Introduction

For decades, the industry gospel was clear: to secure internal applications, put them behind a firewall and have users VPN in. The network perimeter was the security boundary. If you were "inside," you were trusted. If you were outside, you were blocked. This castle-and-moat approach felt intuitive—until reality set in.

The problem? VPNs are painful to use, slow to connect, and create a binary trust model. Once authenticated, a user (or worse, a compromised device) has broad network access. A single stolen credential becomes a golden ticket to every internal resource. Remote work, cloud migration, and zero-trust security principles have exposed this model as fundamentally broken.

**Cloudflare Zero Trust Access** (formerly Cloudflare Access) represents a paradigm shift: instead of granting network-level access, it enforces identity-based authentication **at the application layer, per request**. Every request to a protected application is evaluated individually against policies that check user identity, group membership, device posture, and context like geolocation or MFA status. Access is never assumed; it's continuously verified.

Think of it as replacing a VPN tunnel (which says "you're in the building, go anywhere") with a bouncer at every single door who checks your ID, verifies you're on the guest list, and confirms you're sober before letting you in. And that check happens **at Cloudflare's edge**, across their global network, so it's fast and reliable—not a bottleneck.

Cloudflare Access integrates seamlessly with enterprise identity providers (Okta, Azure AD, Google Workspace, GitHub) and can protect not just web applications but also SSH servers, RDP sessions, and internal APIs. It logs every authentication attempt. It supports MFA enforcement, device posture checks, and fine-grained policies per app. And critically, it works without requiring users to install VPN software or maintain complex network configurations.

This guide explains how Cloudflare Zero Trust Access works, surveys the deployment methods from manual dashboard clicks to production-grade Infrastructure-as-Code, and shows how **OpenMCF** distills the complexity into a clean, protobuf-defined API focused on the 20% of configuration that covers 80% of real-world use cases.

---

## The Zero Trust Access Model: What's Different?

Traditional VPN access works like this:
1. User connects to VPN
2. VPN authenticates user once
3. User has network access to everything behind the VPN until they disconnect
4. Security relies on the network perimeter

Cloudflare Zero Trust Access works like this:
1. User requests a protected application (e.g., `admin.example.com`)
2. Cloudflare intercepts the request at the edge
3. User is redirected to authenticate via identity provider (Okta, Google, etc.)
4. Cloudflare evaluates the access policy: Does this user's email domain match? Are they in the right group? Did they use MFA? Is their device managed?
5. If all conditions are met, Cloudflare issues a session token and proxies traffic to the origin
6. Every new session re-verifies identity; every request is logged

Key differences:
- **Application-scoped**, not network-scoped (you grant access to specific apps, not entire networks)
- **Identity-aware**: policies are based on who you are, which group you're in, and device context—not just IP address
- **Continuous evaluation**: even within a session, conditions like device posture can be re-checked
- **No VPN client needed**: users access apps via normal browser; Cloudflare handles the rest
- **Centralized visibility**: every authentication attempt is logged, creating an audit trail

This is the "BeyondCorp" model Google popularized internally and that zero-trust frameworks advocate. Cloudflare Access makes it accessible to any organization, without requiring Google-scale infrastructure.

---

## The Deployment Spectrum: From Manual to Production

Not all approaches to configuring Cloudflare Access are created equal. Here's the progression from what to avoid to what works at scale:

### Level 0: The Dashboard (Fine for Learning, Risky for Production)

**What it is:** Using Cloudflare's Zero Trust web dashboard to manually add applications, define policies, and configure identity providers through the GUI.

**What it solves:** Immediate visibility and experimentation. The dashboard is polished and intuitive—ideal for understanding how Access policies work. You can click "Add application," enter a hostname, choose an identity provider, set an email domain rule, and enable MFA in minutes.

**What it doesn't solve:** Drift, auditability, repeatability. Manual changes don't get version-controlled. If multiple people have dashboard access, you risk conflicting changes or accidental deletions. Reproducing the exact same config in another environment (dev vs. staging vs. prod) becomes error-prone. And critically, **manual config doesn't scale**—managing dozens of applications across multiple environments becomes a brittle, time-consuming chore.

**Verdict:** Use the dashboard to learn and prototype. For production, lock it to read-only mode (a Cloudflare feature designed specifically to prevent dashboard edits when using IaC) and manage everything via API or Terraform.

---

### Level 1: REST API (Flexible but Requires Custom Tooling)

**What it is:** Calling Cloudflare's Zero Trust API endpoints directly to create Access applications, policies, groups, and identity provider integrations.

**Example:** `POST /accounts/:account/access/apps` with JSON payload defining application name, domain, session duration, and policies.

**What it solves:** Programmatic control. You can script Access configuration, integrate it into CI/CD pipelines, or build custom automation. The API is comprehensive—every feature in the dashboard is available via REST endpoints. Cloudflare even provides detailed API documentation and client libraries (like `cloudflare-go`).

**What it doesn't solve:** State management, idempotency, abstraction. You're responsible for tracking what exists, handling updates vs. creates, sequencing operations (e.g., create identity provider before referencing it in a policy), and implementing error handling. Essentially, you're building your own IaC layer on top of HTTP calls.

**Verdict:** Useful if you're integrating Cloudflare Access into a larger platform or building custom tooling. But for most teams, higher-level IaC tools (Terraform, Pulumi) abstract the API complexity and provide state management out of the box.

---

### Level 2: Infrastructure-as-Code (Production Standard)

**What it is:** Using Terraform or Pulumi with Cloudflare's official provider to declaratively define Access applications and policies as code.

**Terraform example:**

```hcl
resource "cloudflare_zero_trust_access_application" "internal_dashboard" {
  account_id       = var.cloudflare_account_id
  name             = "Internal Dashboard"
  domain           = "dash.example.com"
  type             = "self_hosted"
  session_duration = "8h"
}

resource "cloudflare_access_policy" "allow_employees" {
  application_id = cloudflare_zero_trust_access_application.internal_dashboard.id
  name           = "Allow employees with MFA"
  decision       = "allow"
  precedence     = 1

  include {
    email_domain = ["example.com"]
  }

  require {
    auth_method = ["mfa"]
  }
}
```

**Pulumi example (Python):**

```python
import pulumi_cloudflare as cloudflare

app = cloudflare.ZeroTrustAccessApplication("internal-dashboard",
    account_id=account_id,
    name="Internal Dashboard",
    domain="dash.example.com",
    type="self_hosted",
    session_duration="8h")

policy = cloudflare.AccessPolicy("allow-employees",
    application_id=app.id,
    name="Allow employees with MFA",
    decision="allow",
    precedence=1,
    includes=[{"email_domain": ["example.com"]}],
    requires=[{"auth_method": ["mfa"]}])
```

**What it solves:** Everything. You get:
- **Declarative configuration**: Define the desired state; the tool figures out how to get there
- **State tracking**: Terraform/Pulumi track what exists and detect drift (if someone changes a policy in the dashboard, the next plan will flag it)
- **Version control**: Access policies become code—diffs, pull requests, code review, rollback
- **Multi-environment support**: Reuse the same config with different variables for dev/staging/prod
- **Idempotency**: Run the same config multiple times with predictable results
- **Comprehensive resource coverage**: Applications, policies, groups, identity providers, service tokens—all manageable

**What it doesn't solve:** The underlying Cloudflare Access model (session-based, not real-time revocation beyond session expiry; device posture requires WARP client; some advanced features like WARP-specific routing may require additional setup). But it makes managing Access policies predictable, auditable, and scalable.

**Verdict:** This is the production standard. Terraform has broader adoption and maturity. Pulumi offers programming language flexibility (TypeScript, Python, Go) if you prefer imperative logic or already use Pulumi for multi-cloud. Both are production-ready and officially supported by Cloudflare.

**Cloudflare's official stance:** The documentation explicitly recommends using API or Terraform for managing Zero Trust config at scale, even offering a `cf-terraforming` tool to import existing dashboard configurations into Terraform code.

---

## Terraform vs. Pulumi: Which to Choose?

Both tools cover Cloudflare Access comprehensively. The choice comes down to preference and ecosystem fit.

### Terraform: The Proven Standard

**Maturity:** The Cloudflare Terraform provider is actively maintained by Cloudflare, with frequent updates (v5.x as of 2025). It covers all Access resources: applications, policies, groups, service tokens, identity providers, even app launcher settings.

**Configuration Model:** Declarative HCL. You define resources and Terraform manages the dependency graph.

**Strengths:**
- Larger community and more examples available
- HCL is purpose-built for infrastructure (less boilerplate than general-purpose languages)
- Strong state management and plan preview
- Cloudflare's docs often reference Terraform examples directly

**Considerations:**
- Recent provider updates renamed resources (e.g., `cloudflare_access_application` → `cloudflare_zero_trust_access_application`). Upgrading requires migration.
- Edge cases like policy rule logic (AND vs. OR) can be tricky to express in HCL, requiring careful reading of docs.

### Pulumi: The Programmable Alternative

**Maturity:** Pulumi's Cloudflare provider is built on the Terraform provider (via Pulumi's bridge), so it inherits the same resource coverage and stability.

**Configuration Model:** Imperative code in TypeScript, Python, Go, or .NET.

**Strengths:**
- Full programming language features (loops, conditionals, functions)
- Easier to express complex logic (e.g., dynamically generate policies based on a list of environments)
- Native integration with application code (if your stack is already TypeScript/Python)

**Considerations:**
- Smaller community for Cloudflare-specific examples
- More boilerplate code compared to Terraform's concise HCL
- Breaking changes in Terraform provider (like resource renames) flow through to Pulumi

### Recommendation for OpenMCF

Since OpenMCF uses **Pulumi** under the hood (as evidenced by the `iac/pulumi/` directory in the repo), the choice is straightforward: Pulumi provides the runtime, and the OpenMCF API abstracts away the complexity. Users define a simple protobuf spec; Planton generates the Pulumi code to create the Access application, policies, and necessary configurations.

For teams working directly with Cloudflare (outside OpenMCF), **Terraform** is the safer bet due to broader adoption and more examples. For teams already invested in Pulumi or who need programmatic flexibility, Pulumi is equally capable.

---

## Production Best Practices

### Identity Provider Integration

Cloudflare Access is only as strong as your identity provider. In production:
- **Use an enterprise IdP**: Okta, Azure AD, Google Workspace, OneLogin—something that supports MFA, group management, and strong authentication
- **Enable group sync**: Leverage IdP groups (e.g., Okta groups, Azure AD groups, Google Workspace groups) in Access policies rather than hard-coding individual emails
- **Support multiple IdPs if needed**: Cloudflare allows concurrent IdPs (e.g., Azure AD for employees, GitHub for contractors)
- **Configure emergency access**: Enable one-time PIN or a fallback IdP in case your primary IdP is down

### Policy Design

Follow the principle of **least privilege**:
- **Email domain rules** (e.g., `@example.com`) are fine for broad internal apps (company wiki, dashboards)
- **Group-based rules** (e.g., "Engineering" Okta group) are better for scoped access (staging environments, admin consoles)
- **Require MFA** on all sensitive applications (production databases, financial tools, admin panels)
- **Layer multiple conditions** for critical resources: group membership **AND** MFA **AND** device posture **AND** geolocation

**Avoid overly broad policies**: An "Allow Everyone" policy with no additional requirements defeats the purpose of Zero Trust. Always add at least an email domain or IdP requirement.

### Session Management

- **Default session duration** is 24 hours, which is reasonable for most internal tools
- **Shorten sessions** for sensitive apps (e.g., 1-4 hours for production consoles, financial dashboards)
- **Balance security and UX**: Re-authenticating every 30 minutes frustrates users; 8-12 hours for a workday is a sweet spot

### MFA Enforcement

- **Enforce MFA at the IdP level** when possible (e.g., Azure Conditional Access, Okta sign-on policies)
- **Use Cloudflare's "Require MFA" policy option** as a backstop to ensure even IdPs with lax MFA settings are forced to prompt
- **Consider hardware keys (WebAuthn)** for highest-security apps

### Device Posture

For organizations deploying the Cloudflare WARP client:
- **Require WARP** for access to production resources
- **Integrate device posture checks** (e.g., "device must be domain-joined," "antivirus must be running")
- **Combine with IdP groups** (e.g., "Engineering group" **AND** "managed device")

Device posture is a powerful zero-trust signal, but it requires WARP client deployment and endpoint management integration—it's not trivial to roll out.

### Logging and Monitoring

- **Enable Access logs**: Cloudflare logs every authentication attempt (success, failure, user, IP, country, MFA status)
- **Push logs to a SIEM**: Use Cloudflare Logpush to send Access logs to Splunk, Datadog, S3, or your logging platform
- **Alert on anomalies**: Repeated failed logins, logins from unexpected countries, disabled users attempting access
- **Audit policy changes**: If using IaC, policy changes are auditable via Git history. If using the dashboard, enable Cloudflare's audit logging (enterprise feature).

### Common Pitfalls to Avoid

**1. Bypass policies without intent:** The "Bypass" action skips all Access checks—useful for public endpoints (like a health check or webhook receiver) but dangerous if applied too broadly. Use sparingly and document clearly.

**2. Mixing AND/OR logic incorrectly:** Cloudflare's policy rules can be tricky. Multiple values in a single "Require" block are AND'd together, which can lock everyone out if misconfigured. Use Access Groups or separate policies for complex logic.

**3. Forgetting to test IdP integration:** Always test that group claims are being passed correctly from your IdP to Cloudflare. A misconfigured SAML assertion can break policies that rely on group membership.

**4. Not planning for IdP downtime:** Have a fallback (e.g., one-time PIN) for emergency access if your IdP goes down.

**5. Orphaned service tokens:** Rotate and revoke Access service tokens regularly. Cloudflare supports token expiration and alerting before expiry—use it.

---

## The 80/20 Configuration: What Most Teams Actually Need

Cloudflare Access has dozens of settings, but most production deployments use a small core set:

### Essential Fields (The 80%)

These fields cover the vast majority of real-world Access applications:

**Application Configuration:**
- **Application Name**: Friendly identifier for dashboards and logs
- **Hostname**: The FQDN to protect (e.g., `admin.example.com`)
- **Zone ID**: Ties the application to a Cloudflare DNS zone
- **Session Duration**: How long a login is valid (default: 24h; tune as needed)

**Access Policy:**
- **Allowed Emails or Email Domains**: Simplest rule—allow `@company.com` or specific addresses
- **Allowed Groups**: IdP group membership (e.g., Google Workspace group, Okta group)
- **Require MFA**: Boolean toggle to enforce multi-factor authentication
- **Policy Type**: Almost always "Allow" (Block and Bypass are edge cases)

**Examples:**

1. **Basic: Company-wide internal tool**
   - Allow: Anyone with `@example.com` email
   - Require: MFA
   - Session: 8 hours

2. **Intermediate: Staging environment for engineering team**
   - Allow: Google Workspace group `engineering@example.com`
   - Require: MFA
   - Session: 12 hours

3. **Advanced: Production admin console**
   - Allow: Okta group "ProdAdmins"
   - Require: MFA **AND** managed device (WARP + device posture check)
   - Session: 2 hours

### Advanced Fields (The 20%)

These are powerful but less commonly needed:
- **Device posture checks** (requires WARP client and endpoint management integration)
- **Custom OIDC claims or SAML attributes** (for highly specific rules)
- **Service authentication** (service tokens or mTLS for machine-to-machine access)
- **Geolocation or IP restrictions** (e.g., block logins from certain countries)
- **Custom deny messages or branding** (aesthetic, not security-critical)

### OpenMCF's Design Philosophy

OpenMCF's `CloudflareZeroTrustAccessApplicationSpec` focuses squarely on the 80%:

```protobuf
message CloudflareZeroTrustAccessApplicationSpec {
  string application_name = 1;
  string zone_id = 2;
  string hostname = 3;
  CloudflareZeroTrustPolicyType policy_type = 4;  // ALLOW or BLOCK
  repeated string allowed_emails = 5;
  int32 session_duration_minutes = 6;
  bool require_mfa = 7;
  repeated string allowed_google_groups = 8;
}
```

This captures the core configuration needed for most internal applications:
- Identity-based access (email or Google Workspace groups)
- MFA enforcement
- Session duration tuning
- Simple policy type (allow/block)

It intentionally omits:
- Device posture (requires additional WARP setup; can be layered in later via raw Terraform/Pulumi)
- Service tokens (different workflow; typically managed separately)
- Custom OIDC claims (rare; power users can extend the spec or use raw IaC)
- Advanced UI settings (app launcher visibility, custom logos—nice-to-haves, not essentials)

This is the 80/20 principle in action: **focus on the 20% of configuration that delivers 80% of the value**, making the API surface small, predictable, and easy to understand.

---

## Cloudflare Tunnel Integration

One of the most powerful integrations is **Cloudflare Tunnel** (formerly Argo Tunnel). This allows you to expose internal applications (running on private IPs or behind firewalls) to the internet securely, without opening inbound ports, by running the `cloudflared` connector.

**How it works:**
1. Run `cloudflared` on a server or in your network
2. It establishes an outbound tunnel to Cloudflare's edge
3. You map a Cloudflare DNS name (e.g., `internal-app.example.com`) to the tunnel
4. Configure a Zero Trust Access policy for that hostname
5. Users access `internal-app.example.com`, Cloudflare enforces the Access policy, and proxies traffic through the tunnel to your internal service

**Why it matters:** You can protect truly internal resources (private IPs, on-prem servers, VPC services) with Zero Trust policies, without exposing them directly to the internet. The tunnel is outbound-only, so no inbound firewall rules are needed.

**Example use case:** Protect an internal Jenkins dashboard running on `10.x.x.x` by creating a Cloudflare Tunnel, mapping it to `jenkins.example.com`, and applying an Access policy that allows only the DevOps team.

**OpenMCF consideration:** Cloudflare Tunnel setup is typically a separate resource (Cloudflare Tunnel configuration, DNS records). Once the tunnel is established, you configure a Zero Trust Access Application on the public hostname, which is what OpenMCF's spec addresses.

---

## App Launcher and User Experience

Cloudflare offers an **Access App Launcher**—a portal at `yourcompany.cloudflareaccess.com` where users can see all applications they have access to in one dashboard. It's a unified landing page for internal tools.

**Key features:**
- **Auto-discovery**: Applications you define with Access automatically appear (if marked visible)
- **Bookmarks**: You can add external SaaS apps as bookmarks for convenience
- **Customization**: Upload logos, set application descriptions, organize tiles

**Best practice:** Enable the App Launcher for user-facing internal tools. It significantly improves discoverability and reduces the "where do I find X?" support burden.

**OpenMCF consideration:** The spec doesn't currently include app launcher visibility flags, but it could be added as an optional boolean (`app_launcher_visible`) if teams find it valuable.

---

## Cost Considerations

Cloudflare Zero Trust Access pricing tiers (as of 2025):

**Free Plan:**
- Up to 50 users
- One identity provider
- Basic Access policies (email, IP, country)
- 24-hour log retention
- Suitable for small teams or testing

**Teams Standard ($7/user/month):**
- Unlimited users
- Multiple identity providers
- Advanced policies (device posture, WARP client integration)
- Longer log retention (30 days)
- Access for Infrastructure (SSH/RDP)
- Suitable for production use

**Enterprise (custom pricing):**
- SCIM provisioning
- Longer log retention (180+ days)
- SLA guarantees
- Advanced integrations (CASB, browser isolation)
- Dedicated support

**80/20 insight:** Most teams start with **Free** for proof-of-concept, then move to **Teams Standard** for production. Enterprise features are rarely needed unless you're a large organization with compliance requirements or thousands of users.

**No per-application pricing:** Unlike some competitors, Cloudflare charges per user, not per application. Protecting 10 apps or 100 apps costs the same—encouraging broad adoption of Zero Trust principles across all internal tools.

---

## Real-World Adoption Patterns

**Startups and SMBs:**
- Replace VPN entirely with Cloudflare Access
- Protect internal dashboards (Grafana, Kibana, admin panels) with email domain rules
- Use Google Workspace or Okta integration for SSO

**Mid-Sized Companies:**
- Layer Access on top of existing infrastructure (databases, CI/CD, internal APIs)
- Use group-based policies for role-based access (engineering vs. finance vs. support)
- Deploy Cloudflare Tunnel to expose on-prem or private cloud resources

**Enterprises:**
- Combine Access with WARP client for device posture enforcement
- Integrate with multiple IdPs (Azure AD for employees, contractor portals for vendors)
- Push logs to SIEM for compliance and threat detection
- Use Access for Infrastructure (SSH/RDP) to eliminate bastion hosts

---

## OpenMCF's Approach

OpenMCF treats Cloudflare Zero Trust Access as a first-class cloud resource, abstracting the Cloudflare API and Pulumi provider complexity into a simple protobuf spec.

**What you define:**
```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustAccessApplication
metadata:
  name: internal-dashboard
spec:
  applicationName: Internal Dashboard
  zoneId: ${cloudflare_zone.example.id}
  hostname: dash.example.com
  policyType: ALLOW
  allowedEmails:
    - "@example.com"
  sessionDurationMinutes: 480
  requireMfa: true
  allowedGoogleGroups:
    - engineering@example.com
```

**What OpenMCF generates:**
- Pulumi code to create the Access application
- Access policy with Include rules (email domain, Google groups)
- Require rule for MFA
- Session duration configuration
- Integration with Cloudflare's API using scoped tokens

**What you get:**
- Declarative, version-controlled Access policy
- Consistent application across environments (dev/staging/prod)
- Abstraction of Cloudflare provider quirks (resource naming, API nuances)
- Integration with OpenMCF's broader multi-cloud orchestration

---

## Conclusion

The migration from VPN-based access to application-layer Zero Trust is one of the most impactful security shifts an organization can make. Cloudflare Zero Trust Access makes this transition accessible—offering the BeyondCorp model without Google-scale infrastructure, the security rigor of enterprise solutions without the complexity tax, and the flexibility to integrate with any identity provider or application type.

For teams starting out, the Cloudflare dashboard is a useful learning tool. For production, Infrastructure-as-Code (Terraform or Pulumi) is the standard, providing the repeatability, auditability, and scale required for serious deployments. And for teams using OpenMCF, the complexity collapses further—into a clean protobuf API that captures the essence of what most teams need: identity-based policies, MFA enforcement, and session management, applied consistently across environments.

Zero Trust Access isn't just a security upgrade; it's a rethinking of how we grant access to resources. Every request authenticated. Every decision logged. No implicit trust. And with Cloudflare's global edge network handling enforcement, it's fast, reliable, and scalable from day one.

If you're still using VPNs to protect internal applications, it's time to rethink the model. And if you're adopting Zero Trust, Cloudflare Access—automated via OpenMCF—is a pragmatic, production-ready path forward.

