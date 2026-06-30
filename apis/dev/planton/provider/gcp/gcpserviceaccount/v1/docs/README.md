# Deploying GCP Service Accounts: From Keys to Keyless

## Introduction

For years, the conventional wisdom around Google Cloud service accounts was simple: create an account, download a JSON key, deploy it with your application. This was straightforward, well-documented, and universally supported. It was also, as Google's security team would eventually conclude after analyzing breach patterns, one of the most common attack vectors in cloud environments.

The paradigm has shifted dramatically. What was once standard practice—distributing long-lived service account keys—is now considered an anti-pattern. Modern cloud infrastructure has evolved to eliminate static credentials entirely through mechanisms like Workload Identity Federation, Application Default Credentials, and token-based impersonation. These aren't just "nice to have" security improvements; they represent a fundamental rethinking of how applications authenticate to cloud services.

This document explores the landscape of GCP service account deployment methods, from manual console creation to sophisticated Infrastructure-as-Code automation. More importantly, it explains the evolution from key-based authentication (what not to do) to production-ready keyless solutions (what to do). Understanding this progression is critical for anyone building secure, scalable cloud infrastructure—whether you're managing a handful of service accounts in a single project or orchestrating hundreds across a multi-cloud platform.

## The Maturity Spectrum: How Service Account Deployment Evolved

Service account management has evolved through distinct phases, each solving limitations of the previous approach. Understanding this progression helps explain why modern platforms favor certain patterns over others.

### Level 0: The Anti-Pattern (Manual Key Distribution)

**The approach:** Create a service account in the GCP Console, click "Add Key," download the JSON file, and manually distribute it to systems that need it. Store the key in configuration management, environment variables, or—worse—commit it to a repository.

**Why it was common:** This was the path of least resistance when service accounts were first introduced. The GCP documentation showed this workflow prominently, and client libraries were designed to work seamlessly with JSON key files via the `GOOGLE_APPLICATION_CREDENTIALS` environment variable.

**Why it fails:** 
- **Security risk:** JSON keys are long-lived credentials (they don't expire unless manually revoked). Once a key leaks—through a committed file, a compromised system, or a data breach—an attacker has unfettered access until someone notices and revokes it. Google's own analysis showed that a significant percentage of security incidents involved leaked service account keys.
- **Operational burden:** Keys require rotation policies, secure storage infrastructure (vaults, secret managers), and distribution mechanisms. Each system that needs the key becomes a potential leak point.
- **No audit context:** When a service account acts using a JSON key, logs show the service account's actions but provide limited context about which system or workload actually used the key.

**Verdict:** This approach should be avoided entirely unless there is absolutely no alternative. Even then, it requires rigorous key rotation, encrypted storage, and monitoring—so much overhead that investing in a keyless solution is almost always more efficient.

### Level 1: Scripted Provisioning (gcloud CLI & Automation Scripts)

**The approach:** Use `gcloud iam service-accounts create` and related commands in shell scripts or configuration management tools (Ansible, Chef, custom Python scripts) to create service accounts and assign IAM roles programmatically.

**What it solves:** Removes manual clicking through the console. Service accounts can be provisioned consistently across multiple projects or environments by running the same script. This is a step toward repeatability.

**Limitations:**
- **No state tracking:** Scripts don't inherently know what was already created. Running the same script twice might fail or create duplicates unless you add idempotency checks.
- **No dependency management:** If a service account must be created before granting it a role on another resource, you have to manually ensure ordering. Complex setups with multiple dependencies become brittle.
- **Still requires handling credentials:** If you need to create a key (even for CI/CD), the script downloads it and you're back to manual secret distribution.

**Verdict:** Useful for small-scale setups or one-off migrations, but doesn't scale to complex infrastructure. The lack of state management makes it error-prone in production environments.

### Level 2: Infrastructure-as-Code (Terraform, Pulumi, OpenTofu)

**The approach:** Define service accounts, IAM bindings, and optionally keys as declarative resources in Terraform (HCL), Pulumi (TypeScript/Python/Go), or OpenTofu (Terraform's open-source fork). The IaC tool maintains a state file tracking what exists and ensures the actual infrastructure matches the declared configuration.

**What it solves:**
- **Declarative state:** The desired configuration lives in version-controlled code. Changes go through code review. The tool detects drift and can reconcile it.
- **Dependency management:** IaC tools automatically understand that an IAM role binding depends on the service account existing. Complex multi-resource setups (service account → role bindings → workload resources) are handled gracefully.
- **Repeatability:** The same configuration can be applied to dev, staging, and production with variable changes. Modules and functions encapsulate common patterns (e.g., "create a logging service account with these standard roles").

**Key differences between tools:**

| Tool | Language | Secret Handling | State Backend | Best For |
|------|----------|----------------|---------------|----------|
| **Terraform** | HCL (declarative) | Historically stored keys in plaintext state; TF 1.10+ supports ephemeral resources to avoid this | GCS, S3, Terraform Cloud | Teams comfortable with declarative DSL; extensive module ecosystem |
| **Pulumi** | TypeScript, Python, Go, etc. | Treats secrets as encrypted outputs by default; never exposes them in plaintext logs | Pulumi Service (encrypted) or self-hosted | Teams that prefer general-purpose languages; need complex logic or custom abstractions |
| **OpenTofu** | HCL (identical to Terraform) | Same as Terraform (evolving with TF parity) | Same backends as Terraform | Organizations requiring open-source licensing; otherwise identical to Terraform |

**Limitations:**
- **Doesn't eliminate keys by itself:** IaC tools can create service account keys, but that just automates the anti-pattern. The real value comes when you use IaC to set up *keyless* authentication (Workload Identity pools, IAM impersonation, attached service accounts) rather than creating keys.
- **Complexity for beginners:** Requires learning Terraform/Pulumi syntax, understanding state backends, and managing backend credentials securely.

**Verdict:** This is the minimum recommended approach for production infrastructure. However, simply using IaC doesn't make your setup secure—you must combine it with keyless authentication patterns (Level 3) to achieve best-in-class security.

### Level 3: Keyless Production Patterns (The Modern Standard)

**The approach:** Design infrastructure so that workloads never use long-lived service account keys. Instead, leverage built-in platform mechanisms to provide short-lived tokens automatically.

**Core mechanisms:**

1. **For workloads running on GCP:**
   - Attach a service account to the resource (Compute Engine VM, Cloud Run service, GKE pod, Cloud Functions).
   - The GCP metadata server automatically provides short-lived OAuth tokens to the workload. Client libraries using Application Default Credentials (ADC) fetch these tokens transparently.
   - **Zero configuration required in application code**—just ensure the attached service account has the necessary IAM roles.

2. **For GKE workloads (Kubernetes):**
   - Use **Workload Identity Federation for GKE**: map a Kubernetes service account to a GCP service account.
   - Pods running with the Kubernetes SA automatically get credentials for the GCP SA without any key files. Tokens are short-lived and bound to the cluster workload context.
   - This replaces the old pattern of mounting JSON keys as Kubernetes Secrets (which was both insecure and operationally complex).

3. **For external systems (AWS, Azure, GitHub Actions, on-premises):**
   - Use **Workload Identity Federation (WIF)** to establish trust between the external identity provider (AWS IAM, Azure AD, GitHub OIDC) and GCP.
   - External workloads exchange their native tokens for GCP tokens, allowing them to impersonate a GCP service account without ever possessing a JSON key.
   - For example, a GitHub Actions workflow can authenticate to GCP using GitHub's OIDC token, deploy infrastructure, and never store a GCP credential.

**Why this is production-ready:**
- **Security by design:** No long-lived secrets to leak, rotate, or store. Tokens are short-lived (typically 1 hour) and scoped to specific workloads.
- **Reduced operational burden:** No key rotation policies, no secret management infrastructure, no distribution logistics.
- **Better audit trails:** Logs capture not just "service account X acted" but contextual information about which workload/pod/VM actually performed the action.
- **Cloud-native alignment:** This is how modern cloud platforms (GCP, AWS with IAM Roles for Service Accounts, Azure with Managed Identities) are designed to work.

**When you might still need a key (rare exceptions):**
- Legacy systems that cannot be updated to use ADC or WIF.
- Third-party SaaS integrations that explicitly require a JSON key (though many now support OIDC federation).
- Local development or testing where setting up WIF is overkill (though developers should use their own user credentials via `gcloud auth application-default login` instead of service account keys).

Even in these cases, treat keys as temporary technical debt: store them in Secret Manager or Vault, rotate them frequently, and have a plan to migrate to keyless auth.

**Verdict:** This is the gold standard. Modern multi-cloud platforms should default to keyless authentication and treat key creation as an exceptional case requiring explicit justification.

## Infrastructure-as-Code Comparison: Terraform vs. Pulumi for Service Accounts

Both Terraform and Pulumi excel at managing GCP service accounts as code, but their approaches differ in ways that matter for complex, multi-cloud platforms.

### Terraform: Declarative Consistency

**Strengths:**
- **Mature ecosystem:** Extensive module library for common patterns (service account + roles, WIF pool setup, etc.). Many organizations already use Terraform, so adding service account management fits naturally.
- **State-based reconciliation:** Terraform's state file is a single source of truth. Running `terraform plan` shows exactly what will change before applying.
- **Ephemeral resources (TF 1.10+):** The Google provider now supports marking `google_service_account_key` as ephemeral, meaning the private key is generated but never written to the state file. This eliminates the historical security concern of keys in plaintext state.

**Example pattern (Terraform):**

```hcl
resource "google_service_account" "app_sa" {
  account_id   = "my-app-service"
  display_name = "My Application Service Account"
  project      = var.project_id
}

resource "google_project_iam_member" "app_sa_logging" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.app_sa.email}"
}

resource "google_project_iam_member" "app_sa_storage" {
  project = var.project_id
  role    = "roles/storage.objectViewer"
  member  = "serviceAccount:${google_service_account.app_sa.email}"
}
```

**Considerations:**
- **Limited logic:** HCL is declarative, which is great for simple scenarios but can be verbose for complex conditionals or loops. Multi-project role grants require looping constructs (`for_each`) that feel more awkward than in a general-purpose language.
- **State security:** The state file must be treated as sensitive (encrypted backend storage, limited access). Even with ephemeral keys, other resources might contain secrets.

### Pulumi: Programmatic Flexibility

**Strengths:**
- **Real programming languages:** Write infrastructure code in TypeScript, Python, Go, etc. Loops, conditionals, functions, and classes are first-class constructs. This makes abstractions natural—you can create a `createServiceAccountWithRoles()` function that encapsulates common patterns.
- **Built-in secret management:** Pulumi treats secrets as a first-class concept. Outputs marked as secrets are automatically encrypted in state (using Pulumi Cloud KMS or your own backend). Service account keys are always encrypted, never exposed in logs.
- **Native async/await:** Pulumi handles resource dependencies through language-native constructs (e.g., `.apply()` in TypeScript). Complex dependency chains feel more natural than Terraform's implicit graph.

**Example pattern (Pulumi TypeScript):**

```typescript
const appSa = new gcp.serviceaccount.Account("appSa", {
  accountId: "my-app-service",
  displayName: "My Application Service Account",
  project: projectId,
});

const roles = ["roles/logging.logWriter", "roles/storage.objectViewer"];

roles.forEach((role, index) => {
  new gcp.projects.IAMMember(`appSaRole${index}`, {
    project: projectId,
    role: role,
    member: appSa.email.apply(email => `serviceAccount:${email}`),
  });
});
```

**Considerations:**
- **Smaller ecosystem:** Fewer pre-built modules than Terraform, though the gap is closing. You may need to write more custom code.
- **Language dependency:** Teams must be comfortable with the chosen language and manage its runtime/dependencies (Node.js for TypeScript, Python interpreter, etc.).

### Which to Choose?

- **Terraform** is ideal if your organization has existing Terraform infrastructure, prefers a purely declarative approach, and values the mature module ecosystem. The recent ephemeral resource improvements address previous security concerns around keys in state.
- **Pulumi** shines in scenarios requiring complex logic, custom abstractions, or when building a platform with its own higher-level API (like Planton). The ability to write infrastructure in the same language as application code can simplify cognitive load for full-stack teams.

For **multi-cloud platforms**, either tool works—the key is to leverage them to implement keyless patterns, not just automate key creation.

## Planton's Approach

Planton adopts a **minimal-configuration, security-first** philosophy for service account management:

**Default behavior:**
- **No keys by default:** The `create_key` field defaults to `false`. Users must explicitly opt into key creation, which discourages the anti-pattern.
- **80/20 configuration:** The API exposes only essential fields: `service_account_id`, `project_id`, `project_iam_roles`, and optionally `org_id` with `org_iam_roles` for advanced scenarios. This covers 80% of use cases without overwhelming users with options.
- **Separation of concerns:** Project-level and organization-level IAM roles are distinct fields, making it clear when granting broad access.

**Why this design:**
- **Encourages best practices:** By making keyless the default and requiring explicit action to create keys, the API nudges users toward secure patterns.
- **Simplicity:** Most service accounts need a handful of roles in a single project. Advanced use cases (org-level roles, cross-project access) are still supported but don't clutter the common path.
- **Protobuf-defined consistency:** Using protobuf to define the API ensures that service accounts are created consistently across different backends (Pulumi, Terraform, or direct API calls) and different clouds.

**Integration with modern auth:**
While the API can create service accounts, the expectation is that workloads will use those accounts via:
- Attached service accounts (for GCP-native resources)
- Workload Identity (for GKE)
- Workload Identity Federation (for external systems)

This is reflected in the API design: there's no field for "distribute this key to system X" because that's not the intended workflow. If a key *is* created (for legacy integration), handling it securely is the user's responsibility—but the API makes that scenario the exception, not the rule.

**Example usage (declarative YAML):**

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpServiceAccount
metadata:
  name: prod-app-logger
spec:
  service_account_id: prod-app-logger
  project_id: my-prod-project
  project_iam_roles:
    - roles/logging.logWriter
    - roles/storage.objectViewer
  create_key: false
```

This creates a service account with just the permissions needed for logging and reading storage objects. No key is created; the account is intended to be attached to a Cloud Run service or GKE workload.

## Deep Dive Topics

This document provides a strategic overview of service account deployment methods and best practices. For implementation details, see:

- **[Workload Identity Federation Setup Guide](./workload-identity-federation.md)** (planned): Step-by-step instructions for configuring WIF for AWS, Azure, and GitHub Actions.
- **[Service Account Security Patterns](./security-patterns.md)** (planned): Detailed security recommendations, including IAM impersonation chains, least-privilege role design, and audit logging strategies.
- **[Multi-Project Service Account Management](./multi-project-patterns.md)** (planned): Patterns for centralized vs. distributed service account management in large organizations.

These guides will cover specific implementation details without overwhelming this overview.

## Conclusion

The evolution from key-based authentication to keyless patterns represents more than a technical improvement—it's a shift in how we think about identity in cloud infrastructure. Long-lived credentials made sense in an era of static servers and manual provisioning, but modern cloud platforms provide better primitives: short-lived tokens, platform-provided identities, and federated trust.

**Planton's service account API embraces this paradigm:** it makes the secure path the easy path. By defaulting to no keys, exposing only essential configuration, and integrating naturally with Workload Identity and federation, it helps teams avoid the anti-patterns that have plagued cloud security for years.

Whether you're migrating legacy workloads or building greenfield infrastructure, the principle is the same: treat service accounts as identities that workloads *assume* through platform mechanisms, not as credentials that you distribute. The result is infrastructure that's more secure, easier to audit, and simpler to operate—because the platform handles the complexity of credential management, leaving you to focus on building.

