# Civo Firewall Deployment Methods

## Introduction

For years, conventional wisdom held that cloud firewalls were just "security groups with a different name"—simple wrappers around basic IP filtering that you clicked into existence in a web console. This worked fine when infrastructure was small and changes were rare. But as Kubernetes-native clouds like Civo emerged and infrastructure-as-code became the standard, that manual approach started to show its limits. The challenge wasn't whether cloud firewalls could protect your workloads (they can), but whether you could **manage them at scale** alongside the rest of your infrastructure without creating drift, inconsistency, or security gaps.

Civo's firewall architecture reflects this modern reality. As a Kubernetes-native cloud built on K3s, Civo treats firewalls as first-class infrastructure resources that integrate cleanly with networks, instances, and clusters. They're stateful (return traffic is automatically allowed), default-deny for security, and scoped to virtual networks for proper segmentation. Unlike legacy security groups that evolved from manual VM management, Civo firewalls were designed for an IaC-first world where your firewall rules should live in Git alongside your application deployments.

This document explores the deployment methods available for Civo firewalls—from manual dashboard clicks to full infrastructure-as-code automation. We'll examine why some approaches create technical debt while others enable production-grade security at scale, compare the mature IaC options (Terraform and Pulumi), and explain Planton's choice to build on Pulumi with a protobuf-based abstraction layer.

## The Maturity Spectrum: From Manual to Declarative

Deploying Civo firewalls follows a clear evolution from quick-and-dirty manual methods to production-ready declarative infrastructure. Understanding this spectrum helps you choose the right approach for your needs and avoid the pitfalls of under-engineering (manual drift) or over-engineering (premature abstraction).

### Level 0: Manual Dashboard Management (The Anti-Pattern for Production)

**What it is:** Using the Civo web console to create firewalls, add rules, and attach them to instances through point-and-click operations.

**Why it exists:** The Civo dashboard provides an intuitive interface for exploration and learning. When you're new to Civo or testing a single instance, clicking through the Networking → Firewalls section to create a firewall and add a few rules is the fastest path to working infrastructure. It's excellent for understanding how Civo's firewall model works—seeing the relationship between networks, firewalls, and rules visually.

**The fatal flaw:** Manual changes leave no audit trail and create invisible drift. When someone clicks "Add Rule" to temporarily open port 8080 for debugging and forgets to close it, that rule lives on indefinitely. When a second engineer creates a similar firewall with slightly different rules for another instance, you now have undocumented divergence. After six months and three team members, no one knows which firewall rule came from where or whether it's still needed. The "Default" firewall (which allows all traffic) lurks as a trap—many instances inadvertently use it because someone forgot to select a custom firewall during creation.

**The lockout risk:** A common disaster scenario: you're editing firewall rules and accidentally remove SSH access before saving. Suddenly you're locked out of your instance. While Civo provides a web console for recovery, this kind of self-inflicted outage is embarrassing and entirely preventable with IaC review workflows.

**Verdict:** Manual dashboard management is acceptable for throwaway dev instances and initial exploration. For anything that will live longer than a week or that involves multiple team members, it's an anti-pattern that accumulates technical debt.

### Level 1: CLI Scripting (Better, But Not Quite There)

**What it is:** Using the `civo` command-line tool to script firewall operations. Instead of clicking, you run commands like `civo firewall create web-fw` and `civo firewall rule create web-fw --startport=443 --protocol=tcp --direction=ingress`.

**Why it's better:** The CLI creates a reproducible sequence of operations. You can wrap these commands in shell scripts, commit them to Git, and run them from CI/CD pipelines. The Civo CLI is well-designed, wraps the REST API cleanly, and even provides conveniences like `--create-firewall` flags when launching Kubernetes clusters that automatically open the right ports (6443 for the API, 80/443 for ingress).

**Where it falls short:** Scripts are imperative, not declarative. If you run the script twice, you might create duplicate firewalls or get errors trying to recreate existing resources. Handling idempotency ("only create if it doesn't exist") requires parsing `civo firewall list` output and writing conditional logic. State management is manual—you have to track which firewalls exist and compare them to your desired state yourself. When rules change, you must explicitly delete old rules and create new ones; there's no automatic diffing.

**Error handling headaches:** Scripting with CLI commands works until something fails halfway through. If your script creates a firewall successfully but fails adding rule #3 of 10, you're left with partial state. Resuming requires careful checks. This gets exponentially worse with complex multi-resource deployments (networks, firewalls, instances) where creation order matters.

**Verdict:** CLI scripting is a solid step up from manual clicking—suitable for one-off automation tasks and small deployments where simplicity beats sophistication. But for production infrastructure that evolves over time, it lacks the declarative power and state management of true IaC.

### Level 2: Infrastructure-as-Code Foundations (Terraform/Pulumi)

**What it is:** Defining firewalls declaratively using Terraform (HCL configuration files) or Pulumi (actual programming languages like Go, Python, or TypeScript). Both tools use Civo's official providers to translate your desired state into API calls, track state, and compute diffs when you make changes.

**Why it's a paradigm shift:** This is where infrastructure management becomes truly repeatable and reviewable. You describe **what** you want (a firewall named "web-fw" with these rules, attached to these instances), and the tool figures out **how** to make it happen. If the firewall doesn't exist, it creates it. If a rule changed, it updates it. If you removed a rule from your code, it deletes the rule from Civo. State files track what was deployed, so `terraform plan` or `pulumi preview` shows you exactly what will change before you apply it.

**Terraform approach:** Terraform uses its mature HCL language and a plugin architecture. The official Civo provider (`civo/civo`) covers all firewall resources. You declare a `civo_firewall` resource and one or more `civo_firewall_rule` resources (each rule is a separate HCL block). This explicit structure makes dependencies clear and enables Terraform's graph-based orchestration. Version control and peer review become natural—firewall changes go through pull requests just like application code.

**Pulumi approach:** Pulumi provides the same declarative benefits but lets you write infrastructure code in general-purpose languages (TypeScript, Python, Go, C#). The Civo Pulumi provider is code-generated from the Terraform provider schema, ensuring feature parity. Instead of learning HCL syntax, you use familiar language constructs—loops to generate rules, functions to abstract patterns, type checking to catch errors before deployment. Pulumi's state management mirrors Terraform's: it tracks deployed resources and computes minimal change sets.

**Shared strengths:**
- **Declarative state:** You commit your desired firewall configuration to Git. The IaC tool ensures reality matches your code.
- **Change preview:** Before applying, you see a diff of what will change (rules added, removed, or modified). No surprises.
- **Atomic operations:** Changes succeed or roll back. No more partial deployments from failed scripts.
- **Multi-cloud compatibility:** Both Terraform and Pulumi manage resources across AWS, GCP, Azure, Civo, and dozens of other providers from a single workflow.

**Where they differ:**
- **Language:** Terraform's HCL is domain-specific and simple but sometimes limiting. Pulumi's use of real programming languages offers more expressiveness but can be overkill for simple cases.
- **Ecosystem:** Terraform has a longer history and massive module library. Pulumi is newer but growing fast and often more comfortable for developers already working in TypeScript or Python.
- **State backends:** Both support remote state (Terraform Cloud, Pulumi Service, S3, etc.). Civo's Object Store can even serve as a Terraform backend for full infrastructure self-hosting.

**Verdict:** Infrastructure-as-Code with Terraform or Pulumi is the **production-ready baseline**. This is where most teams should land for anything beyond experimentation. The choice between Terraform and Pulumi comes down to team preference (HCL vs. programming languages) and existing toolchain investment.

### Level 3: Abstraction Layers and Multi-Cloud APIs (Planton's Approach)

**What it is:** Wrapping lower-level IaC tools (like Pulumi) with higher-level abstractions that hide provider-specific details behind a unified API. In Planton's case, this means defining firewalls using protobuf messages that express **intent** (e.g., "create a firewall for this network with these inbound rules"), and letting a Pulumi module translate that intent into the specific Civo API calls.

**Why it matters:** Direct use of Terraform/Pulumi providers is powerful, but each cloud has its own quirks. AWS security groups use separate ingress/egress rules with different syntax than Civo. GCP firewall rules work differently still. If you're deploying to multiple clouds or want to offer a simplified experience (like an internal platform team providing "firewall as a service" to app developers), abstracting these differences behind a common schema reduces cognitive load.

**How Planton does it:** 
- **Protobuf API:** Firewalls are defined in structured protobuf messages (`CivoFirewallSpec`). This provides strong typing, validation, and cross-language support.
- **Unified model:** The spec captures the 80/20 of firewall needs—name, network, inbound rules, outbound rules, tags—without exposing every provider-specific knob.
- **Pulumi modules:** Under the hood, a Go-based Pulumi module reads the protobuf spec and creates the actual `civo.Firewall` and `civo.FirewallRule` resources. This keeps the deployment engine (Pulumi) but hides its complexity behind a consistent interface.
- **Foreign key references:** Network IDs can reference other Planton resources (like `CivoVpc`) by name, and the system resolves them automatically. You don't manually copy-paste UUIDs.

**Benefits:**
- **Multi-cloud portability:** The same protobuf schema could be adapted to provision AWS security groups, Azure NSGs, or GCP firewall rules, with provider-specific modules handling translation.
- **Simplified developer experience:** Application teams can request a firewall by filling out a YAML spec matching the protobuf schema, without learning Civo-specific Terraform syntax.
- **Versioned schema:** Protobuf provides built-in versioning and backward compatibility, making API evolution safer.
- **Validation and defaults:** Protobuf validation rules enforce constraints (e.g., protocol must be `tcp`, `udp`, or `icmp`) at the spec level, catching errors before deployment.

**Trade-offs:**
- **Abstraction overhead:** You introduce a layer between intent and implementation. For users who want full control of every Civo-specific flag, this can feel limiting.
- **Maintenance burden:** Keeping the abstraction in sync with provider updates requires discipline. When Civo adds a new firewall feature, someone must update the protobuf schema and the Pulumi module.

**Verdict:** Abstraction layers like Planton's are ideal for **platform engineering teams** building multi-cloud internal developer platforms. If you're managing infrastructure for a single cloud and a small team, the direct Terraform/Pulumi approach is simpler. But for organizations with cross-cloud ambitions or wanting to offer infrastructure as a self-service platform, a unified API backed by proven IaC engines provides the right balance of flexibility and standardization.

## Comparing Production IaC: Terraform vs. Pulumi

When you've moved past manual and scripting approaches, the choice narrows to two mature IaC tools: **Terraform** and **Pulumi**. Both support Civo firewalls fully, offer declarative state management, and integrate with CI/CD pipelines. The decision comes down to workflow preferences and ecosystem alignment.

### Terraform: Battle-Tested, HCL-Native

**Strengths:**
- **Maturity:** Terraform has been the IaC standard since ~2014. The Civo provider is official, actively maintained, and covers all firewall features (rules, networks, instances).
- **Simplicity:** HCL's declarative syntax is straightforward. A `civo_firewall` resource plus several `civo_firewall_rule` blocks is easy to read and understand even for those new to Terraform.
- **Ecosystem:** Massive library of modules, extensive documentation, and broad community support. Most cloud practitioners already know Terraform.
- **Plan/Apply workflow:** `terraform plan` shows exactly what will change. This is critical for safely evolving production firewalls—you can review rule changes in pull requests before applying.
- **State backends:** Supports numerous remote state backends (S3, Terraform Cloud, Civo Object Store, etc.) for team collaboration.

**Limitations:**
- **Language constraints:** HCL is purpose-built but sometimes awkward for complex logic. Looping over dynamic data or conditional resource creation can feel clunky compared to a full programming language.
- **Separate resources per rule:** Each firewall rule is a distinct Terraform resource. If you have 10 rules, that's 10 resource blocks. This is explicit (good for clarity) but verbose.
- **No in-place rule updates:** Modifying a rule (e.g., changing a port) often requires Terraform to delete and recreate it. This is usually fine (happens quickly), but be aware that there's a brief moment where the old rule is gone before the new one appears.

**Best for:** Teams already invested in Terraform, projects prioritizing simplicity and wide ecosystem support, or those who prefer a domain-specific language over general-purpose code.

### Pulumi: Code-Native, Developer-Friendly

**Strengths:**
- **Real programming languages:** Write infrastructure in TypeScript, Python, Go, or C#. Use loops, functions, classes, and libraries you already know. This is powerful for generating rules programmatically or building reusable abstractions.
- **Type safety:** Pulumi's SDKs provide strong typing. Your IDE autocompletes resource properties, and TypeScript/Go catch errors at compile time before deployment.
- **Developer ergonomics:** If your team consists of software engineers more comfortable with code than HCL, Pulumi feels natural. You can use familiar testing frameworks (Jest, pytest) to test infrastructure code.
- **Feature parity with Terraform:** Pulumi's Civo provider is derived from the Terraform provider (using Pulumi's bridging technology), so it supports the same resources and has the same version cadence.
- **Unified stack model:** Pulumi's concept of "stacks" (dev, staging, prod) with per-stack configuration aligns well with multi-environment deployments.

**Limitations:**
- **Newer tool:** Pulumi launched in 2018, so it has a smaller ecosystem and less community content than Terraform. You'll find fewer third-party modules and blog posts.
- **Abstraction complexity:** The code-first approach can lead to over-abstraction. It's easy to build intricate class hierarchies that make simple firewall definitions harder to understand. Discipline is required.
- **State management:** Pulumi's state model is conceptually similar to Terraform's but uses its own format. Migrating between the two isn't trivial (though possible).

**Best for:** Teams with strong software engineering culture, projects where infrastructure logic is complex (e.g., generating dozens of dynamic rules), or those building platform abstractions (like Planton) where programming language features are essential.

### Side-by-Side Example: Web Server Firewall

To make the comparison concrete, here's how you'd define a firewall allowing SSH (from office IP) and HTTP/HTTPS (from anywhere) in both tools.

**Terraform (HCL):**

```hcl
resource "civo_firewall" "web" {
  name       = "web-server-fw"
  network_id = civo_network.main.id
}

resource "civo_firewall_rule" "ssh" {
  firewall_id = civo_firewall.web.id
  protocol    = "tcp"
  start_port  = 22
  end_port    = 22
  cidr        = ["203.0.113.10/32"]
  direction   = "ingress"
  label       = "SSH from office"
  action      = "allow"
}

resource "civo_firewall_rule" "http" {
  firewall_id = civo_firewall.web.id
  protocol    = "tcp"
  start_port  = 80
  end_port    = 80
  cidr        = ["0.0.0.0/0"]
  direction   = "ingress"
  label       = "HTTP"
  action      = "allow"
}

resource "civo_firewall_rule" "https" {
  firewall_id = civo_firewall.web.id
  protocol    = "tcp"
  start_port  = 443
  end_port    = 443
  cidr        = ["0.0.0.0/0"]
  direction   = "ingress"
  label       = "HTTPS"
  action      = "allow"
}
```

**Pulumi (Python):**

```python
import pulumi_civo as civo

web_fw = civo.Firewall("web",
    name="web-server-fw",
    network_id=main_network.id
)

ssh_rule = civo.FirewallRule("ssh",
    firewall_id=web_fw.id,
    protocol="tcp",
    start_port=22,
    end_port=22,
    cidr=["203.0.113.10/32"],
    direction="ingress",
    label="SSH from office",
    action="allow"
)

http_rule = civo.FirewallRule("http",
    firewall_id=web_fw.id,
    protocol="tcp",
    start_port=80,
    end_port=80,
    cidr=["0.0.0.0/0"],
    direction="ingress",
    label="HTTP",
    action="allow"
)

https_rule = civo.FirewallRule("https",
    firewall_id=web_fw.id,
    protocol="tcp",
    start_port=443,
    end_port=443,
    cidr=["0.0.0.0/0"],
    direction="ingress",
    label="HTTPS",
    action="allow"
)
```

Both are similarly verbose for this simple case. Pulumi shines when you want to generate rules from a list:

```python
web_ports = [{"port": 80, "label": "HTTP"}, {"port": 443, "label": "HTTPS"}]
for p in web_ports:
    civo.FirewallRule(f"web-{p['port']}",
        firewall_id=web_fw.id,
        protocol="tcp",
        start_port=p["port"],
        end_port=p["port"],
        cidr=["0.0.0.0/0"],
        direction="ingress",
        label=p["label"],
        action="allow"
    )
```

This kind of programmatic generation is awkward in HCL (requiring `for_each` with maps) but natural in Pulumi.

### Licensing and Open Source Considerations

Both tools are **open source**:
- **Terraform:** Mozilla Public License 2.0 (MPL-2.0) for the core and the Civo provider. The licensing changed in 2023 when HashiCorp moved to BSL for Terraform core, but the community forked it as **OpenTofu** (still MPL-2.0). The Civo provider plugin works with both Terraform and OpenTofu.
- **Pulumi:** Apache License 2.0 for the open-source engine and providers. The Pulumi Service (hosted state and secrets management) is a commercial offering, but you can self-host state (e.g., in S3 or Civo Object Store) if preferred.

Neither locks you into proprietary tooling for managing Civo firewalls. If you're committed to avoiding vendor lock-in, both are safe choices.

## Planton's Choice: Pulumi + Protobuf Abstraction

Planton uses **Pulumi** as the deployment engine for Civo firewalls, wrapped behind a **protobuf-defined API**. This choice reflects a deliberate balance between flexibility, type safety, and multi-cloud portability.

### Why Pulumi?

1. **Language-native infrastructure:** Planton's IaC modules are written in Go. Pulumi's Go SDK provides idiomatic Go code for infrastructure, enabling compile-time type checking and seamless integration with existing Go tooling.

2. **Programmatic abstraction:** Building a platform that abstracts multiple clouds requires logic—mapping high-level specs to provider-specific resources, resolving foreign key references, applying defaults. Pulumi's programming model makes this logic straightforward without the contortions HCL would require.

3. **State management flexibility:** Pulumi's state backends (local, S3, cloud service) give flexibility for different deployment contexts (local dev, CI/CD pipelines, managed platforms).

4. **Proven Civo support:** Pulumi's Civo provider (derived from the Terraform provider) ensures complete feature coverage with active maintenance and version parity.

### Why Protobuf Abstraction?

Directly exposing Pulumi code to end users would couple them to Civo-specific details and Go expertise. Instead, Planton defines a **protobuf schema** (`CivoFirewallSpec`) that captures the essential firewall configuration:

- **name:** Human-readable firewall identifier
- **network_id:** Reference to the VPC (with foreign key resolution)
- **inbound_rules / outbound_rules:** Lists of protocol, port range, CIDRs, and labels
- **tags:** Future support for tag-based instance auto-assignment

This protobuf message becomes the **contract** between users and the platform. Users (or other systems) generate YAML or JSON matching the schema, and Planton's Pulumi module translates it into Civo API calls. The benefits:

- **Multi-cloud potential:** Tomorrow, a `GcpFirewallSpec` with similar structure could provision GCP firewall rules using a different Pulumi module. The user-facing API stays consistent.
- **Validation and documentation:** Protobuf validation rules enforce constraints (e.g., protocol must match `tcp|udp|icmp`). The schema self-documents what's required.
- **Versioning:** Protobuf's built-in versioning (e.g., `v1`, `v2`) allows evolving the API without breaking existing deployments.
- **Tooling:** Protobuf generates strongly-typed code in multiple languages, enabling client libraries, CLIs, and web UIs to interact with the firewall API safely.

### The 80/20 Design Philosophy

Planton's protobuf spec deliberately covers the **80% of firewall use cases** most users need:
- Creating a firewall in a specified network
- Defining inbound rules (SSH, HTTP/HTTPS, application ports) with protocol, ports, and source CIDRs
- Defining outbound rules (often a single "allow all" rule for internet access)
- Labeling rules for clarity

It intentionally omits the **20% of edge cases** (like explicitly ordered rules, complex deny patterns, or every Civo-specific flag) to keep the API simple and understandable. Users with advanced needs can still drop down to raw Pulumi/Terraform if necessary, but the vast majority get a clean, consistent experience.

### How It Works

When a user submits a `CivoFirewall` resource:

1. **Validation:** Protobuf validators check that required fields (name, network_id) are present and patterns (protocol regex) match.
2. **Stack input:** The spec is serialized into a `CivoFirewallStackInput` protobuf message that includes the firewall spec plus any resolved dependencies.
3. **Pulumi module:** A Go-based Pulumi program (`iac/pulumi/module`) reads the stack input, creates a `civo.Firewall` resource, iterates over `inbound_rules` to create `civo.FirewallRule` resources, and returns outputs (like the firewall ID) in `CivoFirewallStackOutputs`.
4. **State tracking:** Pulumi's state backend records what was created. On subsequent updates, Pulumi computes the minimal diff and applies only the necessary changes.

This architecture keeps the **user-facing API stable and simple** while leveraging Pulumi's **mature orchestration and state management** under the hood.

### Trade-Offs Acknowledged

Abstraction has costs:
- **Indirection:** Users don't directly see the Pulumi code being executed. Debugging requires understanding both the protobuf spec *and* the underlying module.
- **Maintenance:** When Civo adds new firewall features, Planton must update the schema and module to expose them.
- **Learning curve:** Teams must understand the protobuf API in addition to general Civo concepts.

These trade-offs are acceptable for Planton's goals: **providing a consistent, multi-cloud platform** where infrastructure is managed via declarative specs rather than imperative code. For teams using Planton, the abstraction is a feature—it hides complexity. For teams preferring full control, direct Terraform or Pulumi usage remains an option.

## Alternatives Considered (and Why They're Not the Default)

### Crossplane

**What it is:** A Kubernetes-based control plane that provisions cloud resources via Custom Resource Definitions (CRDs). The `crossplane-contrib/provider-civo` exists but is in early stages (v0.1).

**Why not the default:** As of 2025, the Civo Crossplane provider **doesn't support firewall resources**—only Kubernetes clusters and instances. Crossplane is powerful for teams fully bought into Kubernetes-native workflows (GitOps with Flux/ArgoCD), but the Civo provider isn't mature enough yet. For organizations already running Crossplane and wanting to add Civo, it's promising. For Planton's multi-cloud needs, Pulumi offers broader coverage and stability.

**Future potential:** If the Civo Crossplane provider evolves to include firewalls, it could become a viable option for Kubernetes-centric deployments. Watch this space.

### Ansible

**What it is:** Configuration management tool that can call Civo's REST API via the `uri` module or shell out to the `civo` CLI.

**Why not the default:** Ansible excels at configuring servers (installing packages, setting up services), not at declarative infrastructure state management. You can script firewall creation in Ansible playbooks, but achieving true idempotency (detecting drift, computing minimal updates) requires manual effort. Terraform and Pulumi handle this natively. Ansible makes sense if you're already using it for multi-cloud orchestration and want to keep everything in one tool, but it's not purpose-built for infrastructure state like IaC tools are.

### Direct API Calls

**What it is:** Using Civo's REST API (`/v2/firewalls`) directly via curl, SDKs (Go, Python), or custom tooling.

**Why not the default:** The API is fully documented and capable, but you're responsible for state tracking, retry logic, and orchestration. This is reinventing what Terraform/Pulumi already provide. Direct API use makes sense for integrations where existing IaC tools don't fit (e.g., embedding firewall provisioning into a SaaS product), but for general infrastructure management, it's unnecessarily low-level.

## Conclusion: The Shift from Clicking to Declaring

The evolution of Civo firewall deployment mirrors the broader shift in infrastructure management over the past decade. We've moved from treating infrastructure as snowflakes (manually crafted, impossible to reproduce) to treating it as code (version-controlled, peer-reviewed, repeatable). Civo's firewall design—stateful, default-deny, network-scoped—is built for this declarative world. The question isn't *whether* to use infrastructure-as-code, but *which layer of abstraction* fits your team's maturity and goals.

For most teams deploying on Civo, **Terraform or Pulumi** is the right answer. Both are production-ready, offer full firewall support, and integrate with modern CI/CD workflows. The choice between them comes down to preference: HCL's simplicity vs. programming language power.

For platform teams building multi-cloud abstractions or internal developer platforms, **protobuf-based APIs backed by Pulumi** (as Planton demonstrates) provide a unified interface that hides provider quirks while retaining the benefits of mature IaC engines.

Manual dashboard management and CLI scripting remain useful for learning and experimentation, but they don't scale. Production infrastructure demands repeatability, auditability, and change preview—qualities that only declarative IaC tools deliver.

By understanding the maturity spectrum and choosing the right level of abstraction for your context, you ensure that Civo firewalls become a **security foundation** rather than a source of drift and confusion. The infrastructure is too important to leave to ad-hoc clicks.

