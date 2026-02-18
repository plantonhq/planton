# AWS IAM User Deployment: The Case for Minimalism

## The Identity Paradox

The best IAM user is the one you never create.

That statement might sound strange in documentation about IAM user deployment, but it captures AWS's modern identity strategy perfectly. For years, IAM users were the default way to grant access to AWS – create a user, attach some policies, generate access keys, and you're done. Simple, repeatable, and seemingly harmless.

But here's what we've learned: **every long-lived credential is a security liability waiting to happen**. Access keys don't expire. They get committed to Git repositories, embedded in CI/CD configurations, copied to laptops, and forgotten in old Kubernetes secrets. When they leak (not if, but when), attackers can use them indefinitely unless you catch it.

AWS now encourages a fundamentally different approach: **roles for workloads, federation for humans, and temporary credentials wherever possible**. IAM Identity Center (formerly AWS SSO) for employee access. OIDC federation for CI/CD pipelines. Instance roles for EC2. Service account roles for Kubernetes. The modern AWS identity landscape is one of ephemeral, scoped credentials that rotate automatically.

Yet IAM users haven't disappeared – and won't. Legacy applications still need them. Third-party integrations sometimes require them. Some CI/CD systems can't use OIDC yet. Cross-account access patterns occasionally demand them. When you truly need a long-lived service account with programmatic access keys, IAM users remain the mechanism.

This document explores the deployment landscape for AWS IAM users: from manual console clicks to sophisticated Infrastructure as Code, from anti-patterns that lead to breaches to production-ready patterns that minimize risk. We'll examine what OpenMCF supports and why – focusing on the essential 20% of configuration that covers 80% of real-world needs, while embedding security best practices by default.

## The Deployment Spectrum: From Manual to Modern

### Level 0: The Console Click (Manual Management)

The AWS Management Console provides a straightforward interface for creating IAM users. You fill in a username, click checkboxes for "Programmatic access" or "AWS Console access," attach some policies from a dropdown, and optionally configure MFA or tags. For creating one or two users in a sandbox account, this works fine.

**The problems emerge at scale:**

- **No reproducibility**: Every manual action is a potential deviation from your security standards. Did you remember to enable MFA? Did you attach the right policies? Did you choose "require password reset on first login"?

- **The secret key dilemma**: When you create access keys through the console, AWS shows the secret key exactly once. If you don't download it immediately, it's gone forever (and you'll need to create a new key). This often leads to "let me just paste this into Slack real quick" moments that violate security policy.

- **Configuration drift**: Manual changes aren't tracked in version control. Six months later, when someone asks "why does this user have S3 admin access?", there's no audit trail explaining the decision.

- **Scale nightmare**: Managing permissions for dozens or hundreds of IAM users through the console is like performing surgery with a sledgehammer.

**Verdict**: Acceptable for initial learning or emergency break-glass scenarios. Unacceptable for production infrastructure management.

### Level 1: Scripting and CLI (Automation Without State)

The AWS CLI and SDKs (like Boto3 for Python) allow scripting IAM user creation. You can write a shell script that calls `aws iam create-user`, `aws iam attach-user-policy`, and `aws iam create-access-key` to provision users consistently.

This is better than clicking: your script can enforce standards (always tag users, always create with specific policy patterns), and you can version control the scripts. The access key secret can be programmatically stored in AWS Secrets Manager or a vault immediately after creation, avoiding the "paste into Slack" problem.

**The missing piece is state management**: CLI scripts are imperative, not declarative. They tell AWS what to do, but they don't track what *should* exist versus what *does* exist. If someone manually attaches an extra policy to a user, your script won't detect or correct that drift. You also need to handle idempotency yourself (checking if a user exists before trying to create it, etc.).

**Verdict**: Useful for one-off automation tasks or integration into custom workflows. Insufficient for managing production IAM at scale where drift detection and declarative state matter.

### Level 2: Configuration Management (Ansible, etc.)

Ansible provides IAM management modules (`amazon.aws.iam_user`, `amazon.aws.iam_access_key`, `amazon.aws.iam_user_policy_attachment`) that bring declarative configuration to IAM. You define your desired state in a playbook, and Ansible ensures it matches reality when the playbook runs.

Ansible's idempotency helps: running the same playbook twice won't duplicate users. It's agentless and integrates well with broader server configuration workflows. If you're already using Ansible to configure your infrastructure, adding IAM user management to your playbooks creates a unified automation story.

**The limitations:**

- **No persistent state file**: Ansible checks current state on each run but doesn't maintain a persistent record between runs. Drift detection requires re-running the playbook (often in check mode).

- **Manual orchestration**: Ansible runs when you tell it to. It won't automatically detect if someone manually changed a user in the console between playbook executions.

- **Limited secret handling**: While Ansible can create access keys and output them, you need to explicitly integrate with vaults or secret managers. There's no built-in encryption of sensitive outputs.

**Verdict**: Good for organizations already standardized on Ansible for infrastructure configuration. Works well when IAM user management is part of broader configuration tasks (e.g., provisioning a server and its service account together). Not ideal as a standalone IaC solution for AWS infrastructure.

### Level 3: Infrastructure as Code (The Production Standard)

Modern IAM user management happens through declarative Infrastructure as Code tools that maintain state, detect drift, and integrate secrets management. Four tools dominate this space:

**Terraform/OpenTofu**

Terraform defines IAM users as resources in HCL (HashiCorp Configuration Language). You declare `aws_iam_user` resources, attach policies via `aws_iam_user_policy_attachment` or `aws_iam_policy`, and create access keys with `aws_iam_access_key`. Terraform's state file tracks what exists, enabling powerful drift detection: run `terraform plan` and you'll see if someone manually modified a user outside of Terraform.

The ecosystem is massive: thousands of modules, deep community knowledge, and support for every AWS service. OpenTofu, the open-source fork of Terraform, provides identical functionality without HashiCorp's licensing concerns.

The Achilles' heel is **secret management**: Terraform state files can contain sensitive data (like access key IDs and potentially secrets) in plaintext unless you use encrypted backends (S3 with encryption, Terraform Cloud, etc.). Terraform marks certain outputs as sensitive to hide them from console logs, but the state file itself requires external protection. This isn't insurmountable – use encrypted remote state, integrate with AWS Secrets Manager via additional resources, and treat state files as sensitive – but it requires explicit care.

**Pulumi**

Pulumi brings IaC to general-purpose programming languages (TypeScript, Python, Go, Java, .NET). Instead of learning HCL, you write code that declares infrastructure. For IAM users, you might instantiate `aws.iam.User` objects, attach policies, and create access keys – all with the full power of a programming language (loops, conditionals, functions, imported libraries).

Pulumi's standout feature is **native secrets encryption**: mark a value as secret (like an access key), and Pulumi automatically encrypts it in state files and logs using your chosen encryption key (passphrase, AWS KMS, Azure Key Vault, etc.). You don't need to remember to encrypt state – it's the default behavior. Pulumi can also easily integrate with secret stores programmatically, since you're writing real code.

State management mirrors Terraform: Pulumi tracks resources and detects drift. The tradeoff is complexity: simple configurations might feel verbose as code, and teams need programming language knowledge rather than just learning a DSL.

**AWS CloudFormation**

CloudFormation is AWS's native IaC service. You write YAML or JSON templates defining `AWS::IAM::User`, `AWS::IAM::AccessKey`, and related resources. CloudFormation deploys these as stacks, managing the lifecycle entirely within AWS.

The advantages are tight AWS integration and zero additional tooling: CloudFormation is built into AWS, state is managed by AWS (no remote backend to configure), and drift detection is a first-class feature. You can run drift reports to see if resources have changed outside of CloudFormation.

The disadvantages are AWS lock-in (CloudFormation only works for AWS, so multi-cloud organizations need separate tools) and verbosity (YAML templates can be tedious for complex setups). Secret handling is minimal: if you create an access key in CloudFormation, the secret is available only at creation time via stack outputs (which you must handle carefully with `NoEcho` to avoid logging it). There's no built-in vault integration.

**AWS CDK (Cloud Development Kit)**

CDK is CloudFormation with a programming layer on top. You write TypeScript, Python, Java, or C# code that defines infrastructure using high-level constructs. When you run `cdk synth`, it generates a CloudFormation template that gets deployed.

CDK combines the power of code (like Pulumi) with CloudFormation's native AWS integration. It's excellent for complex scenarios where you want abstractions, reusable constructs, and unit-testable infrastructure code. The high-level constructs can simplify common patterns (e.g., "create an IAM user and give it access to this specific S3 bucket" might be a few lines).

The tradeoff: CDK is AWS-only, and it adds a layer of complexity (you need Node.js/Python runtime, the CDK CLI, and understanding of how constructs translate to CloudFormation). Secret handling inherits CloudFormation's limitations unless you add custom logic.

## Comparative Analysis: Choosing Your IaC Tool

When deploying IAM users at scale, the choice of IaC tool matters. Here's how the production-ready options compare across key criteria:

| Criterion | Terraform/OpenTofu | Pulumi | CloudFormation | AWS CDK |
|-----------|-------------------|---------|----------------|---------|
| **Configuration Language** | HCL (domain-specific) | TypeScript, Python, Go, etc. | YAML/JSON | TypeScript, Python, Java, etc. |
| **State Management** | Local or remote state file | Local or remote (Pulumi Service or self-hosted) | AWS-managed (no separate file) | AWS-managed (via CloudFormation) |
| **Drift Detection** | `terraform plan` compares state to reality | `pulumi preview --refresh` detects drift | Native drift detection via console/CLI | Inherits CloudFormation drift detection |
| **Secrets Handling** | State can contain plaintext secrets unless encrypted backend used; requires external vault integration | Automatic encryption of secrets in state/logs; easy vault integration via code | No built-in secret encryption; manual vault integration needed | Same as CloudFormation (no built-in); custom Lambda/code needed |
| **Multi-Cloud Support** | Excellent (AWS, Azure, GCP, 100+ providers) | Excellent (AWS, Azure, GCP, Kubernetes, etc.) | AWS only | AWS only |
| **Community & Ecosystem** | Massive community, thousands of modules | Growing rapidly, strong language ecosystem integration | AWS official support, comprehensive docs | AWS official, growing construct library |
| **Learning Curve** | Medium (learn HCL syntax and Terraform patterns) | Medium (need programming language knowledge) | Low-Medium (YAML is simple, but templates verbose) | Medium-High (need programming + CDK constructs) |
| **CI/CD Integration** | Excellent via terraform CLI | Excellent via pulumi CLI | Good via AWS CLI or CloudFormation change sets | Good via cdk CLI and CloudFormation |
| **Licensing** | OpenTofu: Open Source (MPL 2.0); Terraform: BSL | Apache 2.0 (fully open source) | Proprietary but free to use | Apache 2.0 |

**When to choose Terraform/OpenTofu:**

- Multi-cloud or multi-service deployments
- Team prefers declarative DSL over programming languages
- Strong community modules and patterns are important
- Need to avoid vendor-specific licensing (use OpenTofu)

**When to choose Pulumi:**

- Team has strong programming language skills
- Need complex logic, loops, or conditional resource creation
- Security team requires automatic secret encryption in state
- Want to share code/libraries between infrastructure and application logic

**When to choose CloudFormation:**

- AWS-only infrastructure
- Want zero external tooling dependencies
- Need AWS-native features (StackSets for multi-account, etc.)
- Prefer AWS official support and documentation

**When to choose AWS CDK:**

- AWS-only, but want programming language power
- Love high-level abstractions and reusable constructs
- Already using TypeScript/Python and want infrastructure as code
- Want to unit test infrastructure definitions

For most teams, **Terraform/OpenTofu or Pulumi** emerge as the strongest choices due to multi-cloud flexibility and mature ecosystems. For AWS-centric organizations, **CloudFormation or CDK** offer deep integration and official support. All four can manage IAM users in production – the choice is about team preference, existing skills, and broader infrastructure needs.

## The OpenMCF Choice: Pulumi with Minimalist Configuration

OpenMCF uses **Pulumi** as its underlying deployment engine, wrapped in a protobuf-based API that presents a Kubernetes-style declarative interface. For AWS IAM users, this means you interact with a simple, focused API while Pulumi handles the complexity underneath.

**Why Pulumi?**

1. **Multi-cloud by default**: OpenMCF is a multi-cloud framework. Using Pulumi allows deploying IAM users on AWS, service principals on Azure, and service accounts on GCP with the same underlying engine.

2. **Automatic secret encryption**: IAM user access keys are sensitive. Pulumi's native secret encryption means keys are never stored in plaintext in state files or logs – a critical security requirement.

3. **Programming power when needed**: While OpenMCF abstracts most complexity, the Pulumi foundation allows advanced users to extend or customize deployments with full programming language capabilities.

4. **Open source with no licensing concerns**: Pulumi is Apache 2.0 licensed, aligning with OpenMCF's open-source philosophy.

**The 80/20 Configuration Philosophy**

Research into production IAM user deployments reveals a pattern: **most IAM users need just 4-5 core configuration fields**. The vast majority of use cases are:

- **CI/CD service accounts**: Automated pipelines that need programmatic access to deploy infrastructure or push artifacts
- **Third-party integrations**: External services that require AWS API access
- **Legacy application service accounts**: Apps running outside AWS that haven't migrated to roles or federation

These use cases share common needs: a username, one or more policies (usually managed policies like `AmazonS3ReadOnlyAccess` or custom policies), programmatic access keys, and tags for tracking. They rarely need console access, SSH keys for CodeCommit, permissions boundaries, or multiple MFA devices.

OpenMCF's IAM User API reflects this 80/20 reality:

```protobuf
message AwsIamUserSpec {
  string user_name = 1;                              // Required: The IAM username
  repeated string managed_policy_arns = 2;           // Attach AWS-managed or customer-managed policies
  map<string, google.protobuf.Struct> inline_policies = 3;  // Define custom inline policies
  bool disable_access_keys = 4;                      // Default: create access keys; set true to disable
}
```

**What's included:**

- **user_name**: The IAM username, validated against AWS naming rules (1-64 characters, alphanumeric plus `+=,.@_-`)
- **managed_policy_arns**: A list of managed policy ARNs to attach (e.g., `arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess`)
- **inline_policies**: A map of policy names to policy documents (as JSON structs) for cases requiring tightly-scoped custom permissions
- **disable_access_keys**: By default, OpenMCF creates an access key for the user (since programmatic access is the primary use case). Set this to `true` if you only need the IAM user identity without keys.

**What's intentionally omitted:**

- **Console login / password**: Modern best practice is to use AWS IAM Identity Center or SAML federation for human users. IAM users should be service accounts. If you truly need console access, you can still create a login profile manually or via lower-level Pulumi code.
- **SSH keys for CodeCommit**: Niche use case; most teams use HTTPS or IAM Identity Center for Git access.
- **Permissions boundaries**: Advanced feature needed only in delegation scenarios (less than 1% of deployments). Can be added if genuinely required.
- **MFA device associations**: Typically configured by the end user, not deployed via IaC.

This minimalism serves two purposes: **simplicity** (developers don't need to learn 30 fields when 4 cover their needs) and **security by default** (by focusing on service accounts with programmatic access and explicitly discouraging console passwords, we guide users toward modern patterns).

**Secret Handling**

When OpenMCF creates an IAM user with access keys, the secret access key is automatically encrypted by Pulumi and made available as a secure output. You can retrieve it once for distribution (e.g., storing in a secret manager or CI/CD secret store), and it's never exposed in plaintext logs or unencrypted state.

The recommended pattern:

1. Create IAM user via OpenMCF
2. Retrieve the encrypted secret access key from outputs
3. Store it in AWS Secrets Manager, HashiCorp Vault, or your CI/CD system's secret store
4. Configure your application/pipeline to fetch credentials from that secret store
5. Implement 90-day key rotation by updating the secret and creating a new key

This flow ensures keys are never hardcoded, never committed to Git, and have a clear lifecycle management path.

## Real-World Patterns and Examples

To ground this in practice, here are three common IAM user patterns and how OpenMCF handles them:

### Pattern 1: CI/CD Pipeline User (GitHub Actions ECR/ECS Deployment)

**Scenario**: A GitHub Actions workflow needs to build a Docker image, push to Amazon ECR, and deploy to Amazon ECS.

**Traditional approach**: Create an IAM user manually, attach `AmazonEC2ContainerRegistryFullAccess` and `AmazonECS_FullAccess` managed policies, generate access keys, and store in GitHub Secrets.

**OpenMCF approach**:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamUser
metadata:
  name: cicd-deployer
spec:
  region: us-east-1
  userName: github-actions-deployer
  managedPolicyArns:
    - arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryFullAccess
    - arn:aws:iam::aws:policy/AmazonECS_FullAccess
  # Access key will be created by default
  # Encrypted secret available in outputs
```

**Better approach (when possible)**: Instead of a static IAM user, use GitHub's OIDC federation with AWS to assume an IAM role. This eliminates long-lived credentials entirely. OpenMCF encourages this pattern but supports IAM users for workflows not yet on OIDC.

### Pattern 2: Least-Privilege Application Service Account

**Scenario**: An external payment processing service needs to read from one specific DynamoDB table and invoke one specific Lambda function – nothing more.

**Traditional approach**: Attach broad managed policies, or spend hours crafting a custom policy JSON.

**OpenMCF approach**:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamUser
metadata:
  name: payment-service-user
spec:
  region: us-east-1
  userName: payment-service-prod
  inlinePolicies:
    PaymentServiceAccess:
      Version: "2012-10-17"
      Statement:
        - Effect: Allow
          Action:
            - dynamodb:GetItem
            - dynamodb:Query
          Resource: arn:aws:dynamodb:us-east-1:123456789012:table/PaymentsTable
        - Effect: Allow
          Action: lambda:InvokeFunction
          Resource: arn:aws:lambda:us-east-1:123456789012:function:ProcessPaymentFunction
```

This creates a user with exactly the permissions needed, demonstrating **least privilege** in practice. The inline policy is scoped to specific resources, limiting blast radius if the credentials are ever compromised.

### Pattern 3: Read-Only Auditor Account

**Scenario**: A compliance tool needs read-only access to AWS resources for auditing and security scanning.

**OpenMCF approach**:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsIamUser
metadata:
  name: compliance-auditor
spec:
  region: us-east-1
  userName: compliance-readonly
  managedPolicyArns:
    - arn:aws:iam::aws:policy/ReadOnlyAccess
  # Could scope this further with a custom inline policy restricting to specific regions/resources
```

Using AWS's `ReadOnlyAccess` managed policy gives broad read access across services. For tighter control, replace with a custom policy listing specific read actions needed.

## Security Best Practices Embedded by Default

OpenMCF's IAM user implementation embeds AWS security best practices:

**Least Privilege by Design**

By requiring explicit policy ARNs or inline policy documents, there's no "default permissive" mode. Users must consciously grant permissions, encouraging thought about what's truly needed.

**No Console Access by Default**

IAM users created through OpenMCF are service accounts: programmatic access only. This eliminates password management, MFA complexity, and the risk of console credential leakage. For human access, AWS IAM Identity Center or federated SSO is the recommended path.

**Automatic Tagging Support**

While not shown in the minimal examples above, OpenMCF allows adding tags to IAM users (e.g., `Environment: Production`, `Owner: Platform-Team`). These tags are crucial for:

- Cost allocation (seeing IAM-related activity in AWS Cost Explorer)
- Compliance reporting (e.g., SOC 2 audits asking "who owns this identity?")
- Automated governance (AWS Config rules checking that all IAM users have required tags)

**Secret Encryption by Default**

Access key secrets are never stored unencrypted. Pulumi's encryption ensures they're protected in state and logs, and OpenMCF outputs make it easy to pipe secrets directly into secure stores.

**Drift Detection and Reconciliation**

Because OpenMCF uses Pulumi, any manual changes to an IAM user (someone attaching an extra policy via the console) will be detected on the next deployment. This prevents configuration drift and maintains the IaC-defined state as the source of truth.

## When NOT to Use IAM Users

Understanding when to avoid IAM users is as important as knowing how to deploy them. Modern AWS best practices encourage these alternatives:

**For AWS workloads: Use IAM Roles**

If your application runs on EC2, ECS, Lambda, or any AWS compute service, use an IAM Role attached to that service. AWS automatically provides temporary credentials that rotate at least every 6 hours. No static keys to manage, no credentials to distribute. This is the gold standard for AWS-native workloads.

**For human users: Use AWS IAM Identity Center**

Employees should access AWS through federated SSO, not individual IAM users. IAM Identity Center integrates with corporate identity providers (Active Directory, Okta, Google Workspace) and provides centralized user management, automatic offboarding, and MFA enforcement. No static passwords or access keys for humans.

**For CI/CD pipelines: Use OIDC Federation**

GitHub Actions, GitLab CI, CircleCI, and other modern CI/CD systems support OpenID Connect federation with AWS. Your pipeline can assume an IAM role temporarily using JWT tokens, eliminating the need to store AWS access keys in CI/CD secrets. This is more secure than static credentials and should be your first choice for new pipelines.

**For cross-account access: Use IAM Roles with Assume Role**

If Account A needs to access resources in Account B, create an IAM role in Account B that Account A's principals can assume. This is more auditable and controllable than distributing IAM user credentials across accounts.

**For end-user applications: Use Amazon Cognito or Web Identity Federation**

If you have a mobile or web app where end users need limited AWS access (e.g., upload to an S3 bucket), use Amazon Cognito Identity Pools or web identity federation to grant temporary, scoped credentials. Never create IAM users for end customers.

**IAM users remain valid for:**

- Legacy applications or third-party tools that only accept static AWS access keys
- Service accounts for external integrations where OIDC/federation isn't supported
- Break-glass emergency access accounts (with strong MFA and tight policies)
- Specific use cases like AWS CodeCommit Git credentials (though IAM Identity Center is preferred)

The trend is clear: **minimize IAM users, maximize roles and federation**. OpenMCF supports IAM users because real-world infrastructure still needs them, but the framework encourages modern patterns through defaults and documentation.

## Lifecycle and Operations

Deploying an IAM user is only the beginning. Production management requires ongoing operations:

**Key Rotation**

Access keys should rotate every 90 days (a common compliance requirement for PCI DSS, SOC 2, and internal security policies). AWS allows up to two active keys per user to facilitate rotation:

1. Generate a second access key (via OpenMCF or Pulumi update)
2. Update all applications/pipelines to use the new key
3. Verify the new key works
4. Deactivate the old key
5. After a grace period, delete the old key

Automate this with scheduled workflows and monitoring. AWS IAM provides "access key last used" timestamps to identify stale keys.

**Credential Auditing**

Regularly review IAM users and their permissions:

- **IAM Credential Report** (CSV export): Lists all users, their access keys, last login, MFA status, and more
- **AWS IAM Access Analyzer**: Identifies overly permissive policies and unused permissions
- **CloudTrail logs**: Audit what actions each IAM user performed
- **AWS Config rules**: Automate checks like "flag IAM users without MFA" or "flag access keys older than 90 days"

OpenMCF deployments are auditable by design: because IAM users are defined in version-controlled YAML, you have a Git history of every permission change.

**Offboarding and Deletion**

When a service is decommissioned or an integration is removed, delete the IAM user:

1. Verify no applications are still using the credentials (check CloudTrail for recent activity)
2. Deactivate access keys to test impact (AWS allows toggling keys to inactive without deletion)
3. Remove the IAM user from OpenMCF configuration and apply
4. Pulumi will handle cleanup: detaching policies, deleting access keys, removing the user

If deletion fails (e.g., user is in groups managed outside OpenMCF), manually clean up dependencies first.

**Monitoring and Alerting**

Set up CloudWatch alarms or Security Hub findings for:

- New IAM user creation (should be rare and reviewed)
- Access keys used from unexpected IP addresses or regions (possible compromise)
- IAM policy changes (especially privilege escalations)
- Failed authentication attempts (brute force or stolen credentials)

AWS GuardDuty can detect compromised IAM credentials based on anomalous API activity patterns.

**Compliance and Governance**

For organizations subject to compliance frameworks (HIPAA, GDPR, SOC 2), IAM user management intersects with several requirements:

- **SOC 2**: Unique identifiers per user, prompt access revocation upon offboarding, periodic access reviews
- **HIPAA**: Least privilege access to PHI, MFA for privileged accounts, audit logging (CloudTrail)
- **GDPR**: Data access minimization (only necessary personnel have access), audit trails (who accessed what data)

OpenMCF supports these through IaC-based access control (every permission change is code-reviewed), automatic tagging (owner/purpose documentation), and integration with AWS's native compliance tools (CloudTrail, Config, GuardDuty).

## Conclusion: Security Through Simplicity

The evolution of AWS IAM users tells a story of growing sophistication: from manual console creation to Infrastructure as Code, from static credentials to federated identity, from permissive policies to least privilege.

OpenMCF's approach embraces this maturity while respecting pragmatic reality. The framework makes it **easy to do the right thing** (create minimally-privileged service accounts with encrypted secrets) and **hard to do the wrong thing** (no default passwords, no overly permissive templates, no plaintext secrets).

By focusing on the essential 20% of configuration (username, policies, access keys) and omitting rarely-used complexity (console passwords, SSH keys, permissions boundaries), OpenMCF presents a clean API that guides users toward secure patterns. The Pulumi foundation provides power when needed, while the protobuf interface hides unnecessary detail.

**The best IAM user deployment is:**

- **Justified**: You've confirmed that roles, federation, or OIDC can't solve the use case
- **Minimal**: Exactly the permissions needed, no more
- **Tracked**: Defined in version-controlled IaC, not created manually
- **Auditable**: Tagged with owner/purpose, monitored with CloudTrail
- **Time-bounded**: Keys rotated regularly, user deleted when no longer needed

When you do need an IAM user, OpenMCF makes it straightforward to create one that meets all these criteria. And when you don't need one, the framework encourages better alternatives.

Because in the modern cloud, the best credential is the one you never have to manage.

