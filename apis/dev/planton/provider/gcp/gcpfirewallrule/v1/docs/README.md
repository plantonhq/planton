# GCP Firewall Rules: From Ad-Hoc Security to Codified Network Policy

## Introduction

Every packet that enters or leaves a virtual machine in Google Cloud Platform passes through a firewall evaluation. GCP firewall rules are the gatekeepers of your VPC network—they decide, on a per-packet basis, whether traffic is allowed or denied based on protocol, port, direction, source, and destination. Get them right, and your network is a well-ordered system with clear boundaries. Get them wrong—or worse, manage them ad hoc—and you're one misconfigured rule away from either a security breach or a mysterious connectivity outage at 3 AM.

Unlike traditional hardware firewalls that sit at network boundaries, GCP firewall rules are distributed and stateful. They're evaluated at the VM level, not at a chokepoint, which means they scale automatically with your infrastructure. Every VPC network comes with two implied rules: a default-deny for all ingress traffic and a default-allow for all egress traffic (both at the lowest priority of 65535). Everything you build on top of these defaults shapes the security posture of your entire environment.

The challenge isn't understanding what firewall rules *do*—that's straightforward. The challenge is managing them at scale. When you have dozens of microservices, multiple environments, and a team of engineers who each have their own idea of what "temporary" means, firewall rules become the single most common source of both security gaps and connectivity debugging sessions. This document traces the evolution of how teams manage GCP firewall rules, compares the available approaches, and explains how Planton distills the essential configuration into a clean, composable API.

## Evolution and Historical Context

### The Network Security Landscape in GCP

GCP's approach to network security has evolved significantly since the platform's early days. Initially, firewall rules were simple constructs attached to a network—you specified a protocol, a port, and a source range, and that was about it. Over time, Google added layers of sophistication:

- **Network tags** arrived early, allowing you to target rules at specific groups of VMs without knowing their IP addresses in advance. Tag a VM as `web-server`, and any rule targeting that tag applies automatically.
- **Service account-based targeting** came later, providing a more robust identity-based approach. Unlike tags (which any project editor can modify on a VM), service accounts are IAM-managed, making them harder to accidentally or maliciously change.
- **Priority-based evaluation** allows fine-grained ordering. Rules are evaluated from lowest priority number (highest precedence) to highest, with the first matching rule winning. This enables layered policies: a high-priority deny rule can override more permissive lower-priority allows.
- **Firewall logging** was added for compliance and debugging, letting you capture metadata about connections that match (or are denied by) specific rules.
- **Hierarchical firewall policies** (organization and folder-level policies) were introduced for enterprise governance, allowing central security teams to enforce baseline rules that individual project owners cannot override.

Despite these advances in the firewall *engine*, the management story has lagged. Most teams still manage firewall rules through a mix of console clicks, `gcloud` commands pasted from Slack, and Terraform resources scattered across repositories. The rules themselves are powerful; it's the *lifecycle management* that needs work.

### Why Firewall Rules Deserve First-Class IaC Treatment

Firewall rules occupy a unique position in the infrastructure stack:

1. **They're security-critical**: A single overly permissive rule can expose internal services to the internet. A missing rule can break a deployment.
2. **They're invisible until they're not**: Unlike a misconfigured VM (which fails obviously), a bad firewall rule might silently allow traffic for months before anyone notices.
3. **They accumulate**: Teams add rules during incidents, during feature launches, during migrations. They rarely remove them. Over time, the rule set becomes a sedimentary record of every operational event the team has experienced.
4. **They interact**: Rules don't exist in isolation. Priority ordering, tag overlap, and the interplay between allow and deny rules create a complex evaluation matrix. Understanding the *effective* policy for a given VM requires analyzing all matching rules together.

For all these reasons, firewall rules are one of the *first* resources teams should move to infrastructure-as-code—not one of the last.

## Deployment Methods Landscape

### Level 0: Manual Console Provisioning

The GCP Console provides a visual interface for creating firewall rules under **VPC network → Firewall**. You fill in a form: name, network, direction, action, targets, source/destination filters, protocols and ports, priority, and optionally logging. Click "Create," and the rule takes effect within seconds.

**Why teams start here**:
- Immediate visual feedback. You can see the rule in the list right away.
- No tooling setup. No Terraform state, no CLI authentication—just a browser.
- Good for learning. The form layout teaches you the anatomy of a firewall rule.

**Why teams leave**:
- **No history**: The console doesn't tell you *who* created a rule, *when*, or *why*. There's no commit message, no PR review, no audit trail beyond Cloud Audit Logs (which require separate configuration to be useful).
- **No review process**: Anyone with `compute.firewalls.create` permission can add a rule. There's no pull request, no approval gate, no "are you sure you want to allow 0.0.0.0/0 on port 22?"
- **Drift is invisible**: If someone modifies a rule in the console, there's no mechanism to detect that the actual state has diverged from the intended state.
- **Reproduction is manual**: Need the same rule set in staging? You're clicking through the form again, hoping you remember every setting.

**Verdict**: Use the console for learning and one-off debugging. For anything that matters—and firewall rules always matter—move to code.

### Level 1: CLI Automation with gcloud

The `gcloud` CLI provides the `gcloud compute firewall-rules` command group for managing rules programmatically:

```bash
# Allow HTTP/HTTPS from the internet
gcloud compute firewall-rules create allow-web-ingress \
  --network=my-vpc \
  --direction=INGRESS \
  --action=ALLOW \
  --rules=tcp:80,tcp:443 \
  --source-ranges=0.0.0.0/0 \
  --target-tags=web-server \
  --priority=1000 \
  --description="Allow HTTP and HTTPS from the internet to web servers"

# Deny all egress (restrictive baseline)
gcloud compute firewall-rules create deny-all-egress \
  --network=my-vpc \
  --direction=EGRESS \
  --action=DENY \
  --rules=all \
  --destination-ranges=0.0.0.0/0 \
  --priority=65534 \
  --description="Default deny all outbound traffic"

# Allow SSH from Google IAP ranges only
gcloud compute firewall-rules create allow-ssh-iap \
  --network=my-vpc \
  --direction=INGRESS \
  --action=ALLOW \
  --rules=tcp:22 \
  --source-ranges=35.235.240.0/20 \
  --target-tags=allow-ssh \
  --priority=1000 \
  --description="Allow SSH via Identity-Aware Proxy only"
```

**Advantages over the console**:
- **Scriptable**: Wrap commands in bash scripts, run them in CI/CD pipelines, version-control the scripts.
- **Parameterizable**: Use variables for project IDs, network names, and CIDR ranges.
- **Bulk operations**: Create dozens of rules in a loop.

**Limitations**:
- **Imperative, not declarative**: The script says "create this rule." If it already exists, you get an error. If you want to update it, you need a separate `update` command. If the rule was deleted externally, the script has no idea.
- **No state management**: There's no built-in mechanism to know what rules your script has previously created, which ones have been modified, or which should be deleted.
- **No drift detection**: Your script doesn't know if someone added priority-500 rules via the console that override everything your script created.
- **Ordering complexity**: When rules depend on networks or other resources, you need to manage creation order yourself.

**Verdict**: CLI scripts are a solid step up from the console and work well for simple automation. But as the number of rules grows, the lack of state management becomes a real problem.

### Level 2: Infrastructure as Code with Terraform

Terraform (and its open-source fork OpenTofu) treats firewall rules as declarative resources with full lifecycle management:

```hcl
resource "google_compute_firewall" "allow_web" {
  name    = "allow-web-ingress"
  network = google_compute_network.vpc.self_link
  project = var.project_id

  direction = "INGRESS"
  priority  = 1000

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["web-server"]

  description = "Allow HTTP and HTTPS from the internet to web servers"
}

resource "google_compute_firewall" "deny_all_egress" {
  name    = "deny-all-egress"
  network = google_compute_network.vpc.self_link
  project = var.project_id

  direction = "EGRESS"
  priority  = 65534

  deny {
    protocol = "all"
  }

  destination_ranges = ["0.0.0.0/0"]

  description = "Default deny all outbound traffic"
}

resource "google_compute_firewall" "allow_ssh_iap" {
  name    = "allow-ssh-iap"
  network = google_compute_network.vpc.self_link
  project = var.project_id

  direction = "INGRESS"
  priority  = 1000

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["35.235.240.0/20"]
  target_tags   = ["allow-ssh"]

  description = "Allow SSH via Identity-Aware Proxy only"
}
```

**Why Terraform works well for firewall rules**:
- **Declarative**: You describe the desired state. Terraform calculates the diff and applies only the necessary changes.
- **Plan/apply cycle**: Before modifying any rule, `terraform plan` shows exactly what will change. This makes firewall rule changes reviewable—critical for security-sensitive resources.
- **Drift detection**: Running `terraform plan` against live infrastructure reveals any rules that have been modified outside of Terraform.
- **Cross-resource references**: The `network` field can reference a VPC resource directly (`google_compute_network.vpc.self_link`), ensuring the rule is always attached to the correct network without hardcoding self-links.
- **State tracking**: Terraform tracks every rule it manages, so it knows when to create, update, or delete.

**Considerations**:
- **One `allow` or `deny` block per rule resource**: In Terraform, each `google_compute_firewall` resource maps to exactly one GCP firewall rule. If you need both allow and deny rules, you create separate resources.
- **HCL verbosity for many rules**: Each rule is a separate `resource` block. In environments with 50+ rules, the HCL files can get long. Modules and `for_each` help, but add their own complexity.
- **State file management**: The state file must be stored remotely and locked for team use. This is solved infrastructure (GCS backend + state locking), but it's an operational consideration.

**Verdict**: Terraform is the production standard for firewall rule management. Its plan/apply cycle provides the review gate that security-sensitive resources demand.

### Level 3: Infrastructure as Code with Pulumi

Pulumi brings general-purpose programming languages to infrastructure management. A firewall rule in Pulumi (Go, matching the Planton implementation language):

```go
firewallRule, err := compute.NewFirewall(ctx, "allow-web-ingress", &compute.FirewallArgs{
    Name:      pulumi.String("allow-web-ingress"),
    Network:   vpc.SelfLink,
    Project:   pulumi.String(projectID),
    Direction: pulumi.String("INGRESS"),
    Priority:  pulumi.IntPtr(1000),
    Allows: compute.FirewallAllowArray{
        &compute.FirewallAllowArgs{
            Protocol: pulumi.String("tcp"),
            Ports:    pulumi.StringArray{pulumi.String("80"), pulumi.String("443")},
        },
    },
    SourceRanges: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
    TargetTags:   pulumi.StringArray{pulumi.String("web-server")},
    Description:  pulumi.StringPtr("Allow HTTP and HTTPS from the internet to web servers"),
})
```

**Strengths for firewall rules**:
- **Type safety**: The `compute.FirewallArgs` struct catches invalid field names and types at compile time. You can't accidentally set `Direction` to an integer.
- **Programmatic rule generation**: Loops and functions make it natural to generate rule sets from configuration maps or external data sources.
- **Same preview workflow**: `pulumi preview` shows planned changes, equivalent to `terraform plan`.
- **Shared abstractions**: Wrap common patterns (web server rules, IAP SSH rules, deny-all baselines) in Go functions or npm packages and reuse them across projects.

**Considerations**:
- **Language-specific complexity**: The Pulumi GCP SDK uses `pulumi.String()`, `pulumi.StringPtr()`, and typed arrays extensively. This is more verbose than Terraform's HCL for simple cases, but the type safety pays off at scale.
- **Smaller community for GCP-specific patterns**: While Pulumi's GCP support is comprehensive, the library of community-shared firewall rule modules is smaller than Terraform's.

**Verdict**: Pulumi is an excellent choice for teams already using Go, TypeScript, or Python for infrastructure. Its type safety and programmatic flexibility make it particularly well-suited for generating complex rule sets.

## Comparative Analysis

| Criterion | Console (Manual) | gcloud CLI | Terraform/OpenTofu | Pulumi |
|---|---|---|---|---|
| **Repeatability** | None | Script-dependent | Full (declarative) | Full (declarative) |
| **Drift Detection** | None | None | On-demand (`plan`) | On-demand (`preview`) |
| **Review Gate** | None | Script review only | Plan output in PR | Preview output in PR |
| **State Tracking** | None | None | State file | State backend |
| **Cross-Resource Refs** | Manual copy-paste | Variable substitution | Native references | Native references |
| **Multi-Cloud** | N/A | GCP only | 3000+ providers | 150+ providers |
| **Learning Curve** | Low | Low-Medium | Medium (HCL) | Low-Medium (if you know the language) |
| **Rule Set Complexity** | Unmanageable at scale | Fragile at scale | Manageable with modules | Manageable with functions |
| **Audit Trail** | Cloud Audit Logs only | Script version control | Full VCS + plan logs | Full VCS + preview logs |
| **Best For** | Learning, one-off debug | Simple automation, CI/CD | Production rule management | Teams using general-purpose languages |

**Key takeaway**: For firewall rules specifically, the plan/preview workflow is non-negotiable in production. Any method that doesn't show you exactly what will change before it changes is a risk you shouldn't accept for security-critical resources.

## The Planton Approach

### Action + Rules Abstraction

GCP's native API splits firewall rule traffic matching into two mutually exclusive modes: `allowed` blocks and `denied` blocks. You can't have both in the same rule. Terraform mirrors this with separate `allow {}` and `deny {}` dynamic blocks, and Pulumi uses separate `Allows` and `Denies` arrays.

Planton takes a cleaner approach by separating the **action** from the **rules**:

```protobuf
// Action: "ALLOW" or "DENY"
string action = 5;

// Protocol and port combinations this rule matches.
repeated GcpFirewallProtocolPort rules = 6;
```

Instead of two separate lists, you declare the action once and the protocol/port rules once. The IaC layer (Pulumi or Terraform) routes the rules to the correct allow/deny block based on the action field. This eliminates the cognitive overhead of remembering which list to populate and makes YAML manifests more readable:

```yaml
spec:
  action: ALLOW
  rules:
    - protocol: tcp
      ports:
        - "80"
        - "443"
```

Compared to Terraform's:

```hcl
allow {
  protocol = "tcp"
  ports    = ["80", "443"]
}
```

The difference is subtle for a single rule, but at scale—when you're reading through dozens of rules in a review—the explicit `action` field makes intent immediately clear without scanning for whether the block is `allow` or `deny`.

### StringValueOrRef for Composition

The `project_id` and `network` fields use Planton's `StringValueOrRef` pattern, which enables two modes:

1. **Direct value**: Provide a literal string when you know the value at authoring time.

```yaml
spec:
  projectId:
    value: "my-prod-project-123"
  network:
    value: "projects/my-prod-project-123/global/networks/prod-vpc"
```

2. **Foreign key reference**: Reference another Planton resource's output, creating a dependency chain.

```yaml
spec:
  projectId:
    ref:
      kind: GcpProject
      name: my-project
  network:
    ref:
      kind: GcpVpc
      name: my-vpc
```

This pattern is critical for firewall rules because they almost always depend on a VPC network. With foreign key references, you don't hardcode self-links—the framework resolves them from the referenced resource's outputs. This enables composable infrastructure where a VPC, its subnets, and its firewall rules can be defined in separate manifests and wired together through references.

### Validation at the Schema Level

Planton's protobuf schema encodes GCP's constraints directly:

- **Direction validation**: The `direction` field only accepts `"INGRESS"` or `"EGRESS"` via CEL expressions.
- **Action validation**: The `action` field only accepts `"ALLOW"` or `"DENY"`.
- **Ingress source requirement**: A cross-field validation ensures INGRESS rules specify at least one of `source_ranges`, `source_tags`, or `source_service_accounts`. Without this, an INGRESS rule would match *all* traffic—a common misconfiguration.
- **Tag/SA mutual exclusion**: GCP does not allow mixing tag-based targeting with service-account-based targeting in the same rule. Planton validates this at the schema level, catching the error before any API call is made.
- **Priority bounds**: Priority must be between 0 and 65535.
- **Protocol port structure**: At least one protocol/port rule is required.

These validations mean that invalid configurations are rejected at authoring time (during `buf lint` or CI validation), not during `terraform apply` or `pulumi up` when you're already in the middle of a deployment.

## 80/20 Scoping: What We Included and What We Excluded

### What's Included (and Why)

Planton's `GcpFirewallRuleSpec` covers the fields that appear in the vast majority of real-world firewall rules:

| Field | Why It's Included |
|---|---|
| `project_id` | Every rule needs a project. StringValueOrRef enables composition. |
| `network` | Every rule needs a network. StringValueOrRef enables VPC references. |
| `rule_name` | GCP requires a unique name per project. Pattern-validated for GCP naming rules. |
| `direction` | Fundamental: INGRESS or EGRESS. Every rule has one. |
| `action` | Fundamental: ALLOW or DENY. Every rule has one. |
| `rules` (protocol/port) | The core of what traffic is matched. Required (at least one entry). |
| `priority` | Essential for rule ordering. Defaults to 1000 (GCP's default). |
| `description` | Operational necessity. Rules without descriptions become mysteries. |
| `source_ranges` | The most common INGRESS filter. CIDR-based source matching. |
| `destination_ranges` | The most common EGRESS filter. CIDR-based destination matching. |
| `source_tags` | Tag-based INGRESS source matching (common in tag-centric environments). |
| `target_tags` | Tag-based rule targeting (applies rule to tagged VMs only). |
| `source_service_accounts` | SA-based INGRESS source matching (more secure than tags). |
| `target_service_accounts` | SA-based rule targeting (more secure than tags). |
| `disabled` | Enables rule suspension without deletion—critical for incident response. |
| `log_config` | Firewall logging for compliance and debugging. |

### What's Excluded (and Why)

| Excluded Field | Why It's Not in the Spec |
|---|---|
| `resource_manager_tags` | Resource Manager tags (the newer, IAM-governed tag system) are a distinct concept from network tags. While powerful for hierarchical firewall policies, they're used by a small minority of organizations and add significant complexity. Can be added in a future version. |
| `enable_logging` (boolean) | GCP's API supports both a boolean `enable_logging` and a structured `log_config`. Planton uses only `log_config` (presence = enabled), which is the modern, recommended approach. |
| Hierarchical firewall policies | Organization-level and folder-level policies are a different resource type entirely (`google_compute_firewall_policy` + `google_compute_firewall_policy_rule`). They're out of scope for a per-project firewall rule component. |
| `deny` + `allow` in the same rule | GCP doesn't support this. Planton's action+rules design makes this impossible by construction. |

The goal is to cover the 80% of use cases with 100% of the schema, rather than covering 100% of use cases with a schema so large that nobody can hold it in their head.

## Production Best Practices

### Principle of Least Privilege

The most important firewall rule principle mirrors IAM: **deny by default, allow by exception**. GCP's implied rules already deny all ingress and allow all egress (both at priority 65535). Build on this:

1. **Create an explicit deny-all-egress rule at priority 65534**: While GCP allows all egress by default, adding an explicit deny-all at a near-lowest priority establishes a clear baseline. Then add specific allow rules at higher priorities for the exact destinations your workloads need.

2. **Never use `0.0.0.0/0` as a source range unless you genuinely need internet access**: The "allow from anywhere" pattern is appropriate for public web servers. For everything else, restrict source ranges to known CIDRs (internal subnets, partner VPNs, Google IAP ranges).

3. **Prefer service account targeting over network tags**: Network tags can be modified by anyone with `compute.instances.setTags` permission. Service accounts are IAM-managed, providing a stronger identity guarantee. Use SA-based targeting for sensitive rules.

### Tag-Based vs. Service Account-Based Targeting

GCP enforces a hard constraint: you cannot mix tag-based and service-account-based targeting in the same rule. This means you need to choose a strategy:

| Approach | Pros | Cons | Best For |
|---|---|---|---|
| **Network Tags** | Simple, flexible, easy to add/remove from VMs | Any project editor can modify tags on a VM, potentially gaining access | Development environments, non-sensitive rules |
| **Service Accounts** | IAM-controlled, harder to spoof | Max 10 SAs per rule, requires SA-aware architecture | Production environments, security-sensitive rules |
| **Neither (network-wide)** | Simplest—rule applies to all VMs in the network | No granularity; all-or-nothing | Baseline deny rules, network-wide allow rules |

**Recommendation**: Use service accounts for production security-critical rules (database access, internal APIs). Use tags for development environments and less sensitive rules where operational flexibility matters more than strict identity control.

### Firewall Logging

Enable logging judiciously:

- **Always log deny rules**: You want to know what traffic is being blocked. This is essential for debugging connectivity issues and detecting probes.
- **Log sensitive allow rules**: Rules that permit access to databases, admin interfaces, or internal APIs should be logged for audit purposes.
- **Skip logging for high-volume, low-risk rules**: Logging every allowed HTTP packet to a public web server generates massive log volume with low security value. Log the deny rules on those targets instead.

Use `INCLUDE_ALL_METADATA` when you need source/destination details for incident investigation. Use `EXCLUDE_ALL_METADATA` when you only need connection counts and basic flow data (cheaper storage, less noise).

### Priority Design

A well-designed priority scheme makes the rule set predictable:

| Priority Range | Purpose | Examples |
|---|---|---|
| 0-999 | Emergency overrides | Incident response: block a specific IP, emergency access |
| 1000-1999 | Standard application rules | Allow HTTP/HTTPS, allow internal APIs |
| 2000-2999 | Infrastructure rules | Allow health checks, IAP SSH, monitoring agents |
| 3000-9999 | Reserved for future use | — |
| 10000-64999 | Low-priority catchalls | Log-only rules, broad allow rules being phased out |
| 65000-65534 | Baseline deny rules | Deny all egress, deny all ingress (explicit) |
| 65535 | GCP implied rules | Cannot be used; reserved by GCP |

**Key principle**: Leave gaps between priority ranges so you can insert rules without renumbering. Starting your standard rules at 1000 (GCP's default) is a sensible choice—it leaves room for higher-priority emergency rules and lower-priority baselines.

### Naming Conventions

Adopt a consistent naming pattern for firewall rules. A common scheme:

```
{direction}-{action}-{what}-{from/to}
```

Examples:
- `ingress-allow-http-from-internet`
- `ingress-allow-ssh-from-iap`
- `egress-deny-all`
- `ingress-allow-pg-from-app-servers`
- `egress-allow-https-to-internet`

This makes rules self-documenting when viewed in a `gcloud compute firewall-rules list` or in the console. Combine with meaningful descriptions for the full picture.

### Rule Lifecycle Management

Firewall rules tend to accumulate over time. Establish processes to prevent rule sprawl:

1. **Tag rules with purpose**: Use the `description` field to document *why* the rule exists, not just *what* it does. "Allow port 8080 for temporary debugging on 2024-03-15" tells you when it's safe to remove it.
2. **Use the `disabled` field for soft-deletion**: Before removing a rule, disable it first and monitor for breakage. If nothing breaks after a week, delete it.
3. **Audit periodically**: Run `terraform plan` or `pulumi preview` regularly, even when you don't intend to make changes. Drift in firewall rules is a security signal.
4. **Review in PRs**: Every firewall rule change should go through code review. The `plan` output should be attached to the PR so reviewers can see exactly what network traffic is being affected.

## Conclusion

GCP firewall rules are deceptively simple resources with outsized security impact. Each rule is just a few fields—direction, action, protocol, ports, source, destination—but the aggregate of all rules defines the security boundary of your entire VPC network.

The progression from console clicks to codified, reviewed, and validated infrastructure follows a clear path:

1. **Stop clicking**: Move every firewall rule to code, whether that's Terraform, Pulumi, or Planton manifests.
2. **Validate early**: Catch misconfigurations at authoring time, not deployment time. Planton's schema-level validations (direction, action, source requirements, tag/SA mutual exclusion) prevent entire categories of errors.
3. **Review changes**: Use plan/preview outputs in pull requests. Firewall rule changes should receive the same scrutiny as application security changes—because that's exactly what they are.
4. **Design for composition**: Use StringValueOrRef to wire firewall rules to VPCs and projects through references rather than hardcoded values. This enables modular, reusable infrastructure patterns.
5. **Follow least privilege**: Default-deny, explicit-allow. Log denies. Prefer service accounts over tags for production. Leave priority gaps for future rules.

Planton's `GcpFirewallRule` component captures these principles in a focused API that covers the 80% of real-world use cases cleanly, with schema-level validation that catches the mistakes that matter most. It doesn't try to replace Terraform or Pulumi—it generates them. What it adds is a typed, validated, composable layer that makes firewall rules as reviewable and predictable as application code.

For further reading:
- [GCP Firewall Rules Overview (Google Cloud Documentation)](https://cloud.google.com/vpc/docs/firewalls)
- [Best Practices and Reference Architectures for VPC Design](https://cloud.google.com/architecture/best-practices-vpc-design)
- [Using Identity-Aware Proxy for TCP Forwarding](https://cloud.google.com/iap/docs/using-tcp-forwarding)
- [Firewall Rules Logging](https://cloud.google.com/vpc/docs/firewall-rules-logging)
- [Hierarchical Firewall Policies](https://cloud.google.com/vpc/docs/firewall-policies)
