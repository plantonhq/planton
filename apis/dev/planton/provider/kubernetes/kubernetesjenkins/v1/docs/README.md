# Deploying Jenkins on Kubernetes: From Anti-Patterns to Production

## Introduction

For years, the conventional wisdom was clear: Jenkins, a stateful application designed in the pre-container era, and Kubernetes, a platform optimized for ephemeral workloads, were an awkward pairing at best. Many teams assumed that deploying Jenkins to Kubernetes meant fighting the platform rather than embracing it.

That conventional wisdom is outdated.

The maturity of Kubernetes primitives (StatefulSets, CSI drivers, VolumeSnapshots), the evolution of the official Helm chart, and the game-changing Jenkins Configuration as Code (JCasC) plugin have transformed Jenkins-on-Kubernetes from a risky experiment into a production-ready pattern. The key isn't whether to run Jenkins on Kubernetes—it's understanding which deployment methods are production-viable and which are catastrophic mistakes waiting to happen.

This document explains the deployment landscape for Jenkins on Kubernetes, from the anti-patterns that cause data loss to the modern GitOps workflows that power resilient CI/CD systems. It details why Planton standardizes on the official `jenkinsci/jenkins` Helm chart as the foundation for its `JenkinsKubernetes` API and how JCasC enables fully declarative, version-controlled Jenkins configuration.

## The Deployment Maturity Spectrum

Understanding how to run Jenkins on Kubernetes requires seeing the landscape not as a list of options, but as an evolutionary progression from naive patterns to production-ready solutions.

### Level 0: The Data Loss Anti-Pattern

**The Pattern:** Deploy Jenkins using a standard Kubernetes `Deployment` resource with an `emptyDir` volume for `$JENKINS_HOME`.

**Why It Fails:** This is the most common and most catastrophic mistake. A `Deployment` is designed for stateless applications, treating its Pods as ephemeral and interchangeable. When a Pod is deleted or rescheduled (which happens routinely during node maintenance, scaling, or failures), an `emptyDir` volume and all its data are permanently lost.

For Jenkins, `$JENKINS_HOME` contains everything: all job definitions, build history, plugins, credentials, and configuration. Losing this directory means losing your entire CI/CD system.

**Verdict:** Never use this pattern, not even for development. The data loss risk is unacceptable.

### Level 1: StatefulSets—The Foundation

**The Pattern:** Deploy Jenkins using a Kubernetes `StatefulSet` with a `PersistentVolumeClaim` (PVC) for `$JENKINS_HOME`.

**What It Solves:** A `StatefulSet` is purpose-built for stateful applications and provides two critical guarantees:
1. **Stable, Unique Identity:** Pods are created with predictable names (e.g., `jenkins-0`)
2. **Stable, Persistent Storage:** Each Pod identity is bound to a specific PVC. When `jenkins-0` is rescheduled, Kubernetes ensures it always reattaches to its unique volume.

This solves the Kubernetes-level persistence problem. Your Jenkins data survives Pod restarts, node failures, and cluster upgrades.

**What It Doesn't Solve:** A StatefulSet is "dumb" to the application's internal logic. It provides persistent storage, but it knows nothing about:
- Plugin installation and version management
- Jenkins configuration (security realms, credentials, job definitions)
- Application-aware backups
- Ingress configuration for external access
- Resource sizing for production loads

A raw StatefulSet is the foundation, not a complete solution. You still need to manually configure dozens of Kubernetes resources (Services, ConfigMaps, Secrets, Ingress) and manage the complex lifecycle of Jenkins itself.

**Verdict:** The correct primitive, but far too low-level for production use.

### Level 2: The Helm Chart—The De Facto Standard

**The Pattern:** Deploy Jenkins using the official `jenkinsci/jenkins` Helm chart from `https://charts.jenkins.io`.

**What It Adds:** Helm is a package manager for Kubernetes. The official chart bundles all the complexity of a production Jenkins deployment into a single, configurable package:
- The StatefulSet for the Jenkins controller
- The PersistentVolumeClaim for `$JENKINS_HOME`
- Services for networking
- ConfigMaps for injecting Jenkins Configuration as Code (JCasC)
- Optional Ingress resources for external access
- RBAC roles for Kubernetes API integration (dynamic agents)

Instead of managing dozens of individual YAML files, you configure a single `values.yaml` file and run `helm install`. This is the 80/20 solution: it handles the majority of complexity while remaining flexible enough for customization.

**Critical Caveat:** The chart is "production-capable," not "production-ready by default." Its default configuration is designed for a minimal test deployment. Default resource requests (CPU: 200m, Memory: 256Mi in some versions) are dangerously insufficient for a real controller and will lead to immediate Out-Of-Memory (OOM) failures. Production deployments require significant tuning of `values.yaml` to configure proper resources, persistence, ingress, and JCasC settings.

**Verdict:** The correct deployment method for production. Requires expertise to tune properly, which is where Planton's API abstraction provides value.

### Level 3: GitOps Delivery—The Production Pattern

**The Pattern:** Store the Jenkins Helm chart and a custom `values.yaml` in a Git repository. Use a GitOps tool (ArgoCD or Flux) to continuously reconcile the cluster state with the Git-declared state.

**What It Adds:** GitOps is not a deployment method itself—it's a **delivery pattern**. In a GitOps workflow:
1. The `jenkinsci/jenkins` Helm chart and your `values.yaml` are stored in Git
2. An ArgoCD `Application` or Flux `HelmRelease` resource points to this repository
3. ArgoCD/Flux detects any changes (e.g., updated plugins, resource limits, or JCasC configuration)
4. The GitOps agent automatically performs a `helm upgrade` to apply the new state

This combines the best of all layers:
- **GitOps** provides auditable, version-controlled delivery
- **Helm** provides application packaging and lifecycle management
- **StatefulSet** provides the stateful storage foundation

This is the gold standard for production systems. It eliminates manual `helm install` commands, provides full change history via Git commits, and enables safe, reviewable updates via Pull Requests.

**Verdict:** The production best practice. Planton's API is designed to work seamlessly in GitOps workflows.

### Level 4: The Operator—A Specialized Alternative

**The Pattern:** Deploy Jenkins using the `jenkinsci/kubernetes-operator`, which manages Jenkins via a Custom Resource Definition (CRD).

**What It Adds:** A Kubernetes Operator is a custom controller that encodes operational knowledge into software. Instead of running `helm install`, you apply a Custom Resource (e.g., `kind: Jenkins`). The Operator watches for this resource and manages the application's full lifecycle, promising automation for backups, safe upgrades, and plugin management.

**The Trade-Off:** While architecturally elegant, the Jenkins Operator has significantly less adoption than the Helm chart. Community feedback indicates that many teams prefer the flexibility and lower-level control of Helm, finding the Operator too opinionated in its management of configuration and dependencies. For most use cases, the Helm chart provides a better balance of abstraction and control.

**Verdict:** A viable alternative for teams that want higher-level lifecycle automation, but not the recommended default for most users.

## The Official Helm Chart: A Deep Dive

The `jenkinsci/jenkins` Helm chart is the cornerstone of modern Jenkins deployments on Kubernetes. Understanding its capabilities, defaults, and limitations is essential.

### Production Readiness: Template vs. Solution

The chart is widely used in production by enterprises worldwide, but this comes with an important caveat: **it is a template that requires expert-level tuning, not a turnkey solution.**

The default `values.yaml` is configured for a minimal test deployment:
- Resource requests are often as low as CPU: 200m, Memory: 256Mi
- No ingress is configured
- No JCasC customization is provided
- No production-grade plugin list is included

A 2025 analysis of over 100 open-source Helm charts found that most lack critical reliability features like PodDisruptionBudgets (PDBs), HorizontalPodAutoscalers (HPAs), and sensible resource limits. The Jenkins chart is no exception. It provides the mechanism for a production deployment, but the configuration is your responsibility.

This is precisely where Planton's abstraction adds value: it provides opinionated, production-viable defaults while still exposing the flexibility needed for customization.

### Core Features

The chart's primary value lies in its seamless integration of the Jenkins-on-Kubernetes ecosystem:

#### 1. Jenkins Configuration as Code (JCasC)

This is the chart's most important feature. The `controller.JCasC.configScripts` key allows you to embed JCasC YAML directly into the Helm `values.yaml` file. The chart automatically creates a Kubernetes ConfigMap from this configuration and mounts it into the Jenkins controller Pod, where the JCasC plugin consumes it on startup.

This enables **fully declarative, version-controlled configuration**. Every aspect of Jenkins—security realms, credentials, plugin settings, job definitions—can be defined as code, stored in Git, and reviewed via Pull Requests.

#### 2. Plugin Management

The chart supports two methods for plugin installation:

**Method 1: controller.installPlugins**
A simple list of plugin IDs (e.g., `kubernetes:1.31.0`). This is easy for development and staging, but it's slow (plugins are downloaded on every startup) and not immutable.

**Method 2: controller.image / controller.tag**
Specify a custom-built Docker image with plugins "baked in." This is the **production best practice**. Building a custom image that includes all required plugins creates an immutable, tested artifact. Startup is dramatically faster, more reliable, and more secure because you're not downloading plugins from the internet on every restart.

#### 3. Persistence

The chart exposes the `persistence` keys to create a PersistentVolumeClaim for `$JENKINS_HOME`. However, the chart only provisions the volume—it does not back it up. Backup is an external concern that you must handle separately (see Production Operations below).

#### 4. Ingress

The chart can optionally create an Ingress resource to expose Jenkins externally. This is configured via the `ingress.enabled`, `ingress.hostName`, and `ingress.tls` keys, providing a production-ready HTTPS endpoint.

### Licensing: 100% Open Source

The entire stack is open source with no commercial licensing concerns:
- **Jenkins Core:** MIT License
- **Helm Chart:** Apache-2.0 License
- **Plugins:** Plugins hosted by the Jenkins project are required to use OSI-approved open-source licenses (typically MIT)

This eliminates any licensing risk for teams building on this foundation.

## The 80/20 Configuration: What Actually Matters

A comprehensive analysis of production Jenkins deployments reveals that while the Helm chart exposes 100+ configuration options, only a small subset is essential for a production-viable system. Understanding this "80/20" split is crucial for API design.

### The Essential 20%: Must-Have Configuration

These fields are **non-negotiable** for any production deployment:

#### 1. Controller Resources

**Configuration:** `controller.resources.requests` and `controller.resources.limits`

**Why Critical:** Resource exhaustion, specifically Out-Of-Memory (OOM) errors, is the #1 cause of Jenkins instability on Kubernetes. The chart's default resources are unusable in production. A production Jenkins controller requires:
- CPU: 2-4 vCPU minimum
- Memory: 8-16 GiB minimum, with requests equal to limits for Guaranteed QoS

#### 2. Persistence

**Configuration:** `persistence.enabled`, `persistence.size`, `persistence.storageClass`

**Why Critical:**
- `persistence.enabled: true` is mandatory—Jenkins cannot function in production without persistent storage
- `persistence.size` must account for build logs and artifacts (typically 50-250 GiB)
- `persistence.storageClass` must point to a high-IOPS, snapshot-capable storage backend (e.g., `gp3-csi`, `pd-ssd`) to ensure reliability and enable disaster recovery

#### 3. Ingress

**Configuration:** `ingress.enabled`, `ingress.hostName`, `ingress.tls`

**Why Critical:** Production Jenkins requires a stable, secure FQDN with HTTPS. The alternative—NodePort or LoadBalancer services—exposes Jenkins on random or fixed ports, which is unsuitable for enterprise environments.

#### 4. JCasC Configuration

**Configuration:** `controller.JCasC.defaultConfig`, `controller.JCasC.configScripts`

**Why Critical:** This is the "brain" of the deployment. All Jenkins configuration should be defined declaratively via JCasC. This field is the "escape hatch" that provides infinite flexibility while keeping the top-level API simple.

#### 5. Plugins

**Configuration:** `controller.installPlugins` or `controller.image`

**Why Critical:**
- For development/staging: Use `controller.installPlugins` with a list of plugin IDs for ease of iteration
- For production: Use `controller.image` with a custom-built, immutable image that includes all plugins

### The Remaining 80%: Advanced and Specialized

These features represent higher operational maturity but are not required for initial production deployment:
- Custom RBAC roles scoped to specific namespaces
- Sidecar containers for automatic JCasC reloading
- Dynamic agent configuration via the Kubernetes plugin
- Custom init containers or sidecars

## Production Operations: Day 2 Challenges

Deploying Jenkins is "Day 1." Operating it reliably is "Day 2," where the most significant challenges emerge.

### Backup and Disaster Recovery

**The Anti-Pattern:** Using a Jenkins plugin like `thinBackup` to run backups inside the application. If Jenkins is down or broken, you can't run or restore the backup.

**The Best Practice:** Perform backups at the **storage layer**, external to Jenkins:

1. **Storage Provisioning:** The `persistence.storageClass` in your Helm `values.yaml` must point to a CSI driver that supports the Kubernetes `VolumeSnapshot` API
2. **Snapshot Configuration:** Configure a `VolumeSnapshotClass` to define snapshot parameters. For disaster recovery (DR), configure replication of snapshots to a remote location (e.g., an S3 bucket in another region)
3. **Automation:** Use a Kubernetes-native tool (Velero, Portworx Backup) to automate creation of `VolumeSnapshot` resources on a schedule (e.g., nightly)
4. **Restoration:** In disaster scenarios, provision a new PersistentVolume from the snapshot and point the Jenkins StatefulSet at the restored volume

This pattern ensures that backups are independent of the application state and can be restored even if Jenkins is completely unavailable.

### High Availability: The HA Myth

**The Myth:** "I can set `replicas: 2` in the StatefulSet to make Jenkins highly available."

**The Reality:** This is **architecturally impossible** and will break the application. Open-source Jenkins is not designed for active-active operation.

**The Proof:** Community reports confirm that scaling Jenkins to 2 replicas immediately causes `HTTP ERROR 403 No valid crumb was included in the request` errors. The "crumb" is a CSRF token generated by each Jenkins instance. When a user's browser, holding a crumb from `jenkins-0`, has its next request load-balanced to `jenkins-1`, the second instance rejects the request because it doesn't recognize the token. The two "masters" are completely independent and unsynchronized.

**The Real HA Pattern (Active-Passive Fast-Failover):** True HA for Jenkins is a storage-layer solution:
1. Deploy Jenkins with `replicas: 1`
2. Use a replicated storage solution (e.g., LINSTOR) that synchronously replicates the PVC data across multiple Availability Zones (AZs)
3. If the node or AZ hosting the Jenkins Pod fails, Kubernetes reschedules the Pod to a healthy node in another AZ
4. The storage controller detaches the old volume and attaches the synchronously replicated volume in the new AZ
5. Jenkins restarts in 1-3 minutes with no data loss

This provides true HA with minimal downtime, without the impossible active-active pattern.

### Dynamic Agents: The Kubernetes Advantage

The most compelling reason to run Jenkins on Kubernetes is the ability to use **dynamic, ephemeral agents**.

**Static Agents (Anti-Pattern):** A pool of persistent agent Pods, running as Deployments, waiting for work. This is inefficient (agents sit idle), fragile (builds conflict over shared tools), and expensive.

**Dynamic Agents (Best Practice):** The Jenkins `kubernetes` plugin enables the controller to use the Kubernetes API to create a dedicated, ephemeral agent Pod for every pipeline job. The Pod is created, runs the single job, and is then destroyed.

This provides:
- **Elasticity:** No idle resources. Resources are consumed only during builds
- **Scalability:** Build capacity scales to the limits of the cluster, not a fixed agent pool
- **Isolation:** Each build runs in a fresh container, eliminating "works on my machine" issues and preventing build-to-build contamination
- **Resource Optimization:** Kubernetes efficiently packs builds onto cluster nodes

This pattern is configured entirely via JCasC under the `clouds.kubernetes` key, making it fully declarative and version-controlled.

### Security Posture

Securing Jenkins on Kubernetes involves three layers:

#### 1. RBAC (Least Privilege)

The Jenkins controller needs Kubernetes API permissions to manage agent Pods. The default `ClusterRole` created by the Helm chart is often too permissive.

**Best Practice:**
- Set `rbac.create: false` in the Helm chart
- Manually create a namespace-scoped `Role` (not a `ClusterRole`) that grants pod-management permissions only in specific namespaces where agents run
- Bind this `Role` to the Jenkins `ServiceAccount` via a `RoleBinding`

#### 2. Secrets Management

Jenkins credentials (SSH keys, API tokens) must never be stored in plain text in JCasC YAML within a Git repository.

**Best Practice:**
- Use JCasC's string substitution mechanism: `password: ${JENKINS_GITHUB_TOKEN}`
- Populate these variables from Kubernetes Secrets mounted as **files** into the controller Pod
- **Critical:** Do NOT use environment variables. The JCasC documentation explicitly warns this is a "very bad idea for sensitive data," as environment variables are leaked in the Jenkins UI (`/systemInfo`) and in logs. File-based substitution is the only secure method.

#### 3. Network Policies

By default, all Pods in a Kubernetes cluster can communicate. Apply a `NetworkPolicy` to the Jenkins controller to lock down traffic:
- **Ingress:** Deny all, except from the Ingress Controller and monitoring (Prometheus)
- **Egress:** Deny all, except to the Kubernetes API server, SCM (GitHub/GitLab), and artifact repositories

## Jenkins Configuration as Code (JCasC): The Engine of Modern Jenkins

JCasC is not an optional add-on—it's a core component of any modern Jenkins deployment. It replaces the traditional "click-ops" method of configuring Jenkins via the UI with a fully declarative, version-controlled approach.

### The JCasC Value Proposition

JCasC brings the "Infrastructure as Code" paradigm to Jenkins configuration, providing:
1. **Reproducibility:** A new, fully-configured Jenkins controller can be spun up in minutes from JCasC YAML
2. **Version Control:** Configuration is stored in Git with full change history, PR-based review, and rollback capability
3. **Automation:** Enables "zero-touch" deployments with no manual setup steps

### The "Rosetta Stone" Workflow

The JCasC syntax for 2,000+ plugins is not documented in a single place. The most effective workflow for discovering the correct syntax is:
1. Temporarily configure a plugin or setting via the Jenkins UI
2. Navigate to **Manage Jenkins > Configuration as Code > View Configuration**
3. Jenkins generates the JCasC YAML for the setting you just configured
4. Copy this generated YAML into your source-controlled `values.yaml`

This workflow is the "Rosetta Stone" that bridges legacy "click-ops" knowledge and modern "as-code" artifacts.

### The End-to-End GitOps Workflow

The complete production pattern combines Helm + JCasC + GitOps:
1. JCasC YAML and Helm `values.yaml` are stored in a Git repository
2. An ArgoCD `Application` points to this repository
3. To change a Jenkins setting, submit a PR to modify the JCasC YAML
4. On merge, ArgoCD triggers a `helm upgrade`, updating the ConfigMap containing JCasC
5. If `controller.sidecarContainers` is configured with a JCasC reload sidecar, the change is applied without a Pod restart

This pattern provides fully auditable, version-controlled, automated Jenkins configuration management.

### Common JCasC Patterns

**Credentials (Critical):**
- Define credential placeholders using the `credentials.system` block
- Use string substitution syntax: `password: ${JENKINS_GITHUB_TOKEN}`
- Populate from Kubernetes Secrets mounted as files (never environment variables)

**Security:**
- Define `securityRealm` (authentication) and `authorizationStrategy` (permissions) at the root of the JCasC YAML
- Example: LDAP authentication with role-based authorization

**Plugin Configuration:**
- Global tool configurations (Git, Maven) and cloud configurations (Kubernetes plugin for dynamic agents) are top-level keys in JCasC YAML

## Jenkins in the Cloud-Native Ecosystem

While this document details how to run Jenkins on Kubernetes, it's equally important to understand when to use Jenkins versus modern alternatives.

### Jenkins's Enduring Strengths

**1. Unmatched Plugin Ecosystem**
With over 2,000 plugins, Jenkins can integrate with virtually any tool, SCM, or on-premise system. This is critical for enterprises with legacy systems (mainframes, custom SSH scripts) that modern CI tools do not support.

**2. Flexibility and Power**
Jenkins Pipeline, particularly Scripted Pipeline (Groovy), is Turing-complete. This enables incredibly complex, dynamic, and conditional pipeline logic that is often impossible to express in the simple YAML of competitors.

**3. Maturity and Enterprise Adoption**
Jenkins is battle-tested over 15+ years and deeply entrenched in enterprise CI/CD workflows. The operational knowledge and existing job libraries represent significant investment.

### Jenkins's Architectural Weaknesses

**1. Legacy Monolithic Architecture**
Jenkins is a heavyweight, Java-based monolith. Its controller is a "constantly running" process that is memory and processor hungry.

**2. Not Kubernetes-Native**
Jenkins integrates with Kubernetes, but it is not Kubernetes-native. Its architecture, based on a single stateful controller, is fundamentally pre-container.

**3. Plugin Dependency Hell**
The plugin ecosystem is both the greatest strength and weakness. An update to one plugin can break the entire instance due to dependency conflicts. The UI is a disjointed collection of plugin UIs.

**4. Complex Configuration**
Without JCasC, managing Jenkins is a manual "click-ops" nightmare that is not reproducible or version-controlled.

### Strategic Comparison: Jenkins vs. Modern Alternatives

**Jenkins vs. GitHub Actions / GitLab CI**

These are fully integrated SCM platforms, not standalone CI tools. Their CI is built-in, YAML-based, and simple to start.

**Verdict:** For simple, SCM-centric CI (build and test on commit), they are better and simpler. They lack Jenkins's plugin flexibility for complex, multi-system, legacy integrations.

**Jenkins vs. Tekton (The Architectural Contrast)**

- **Jenkins (Orchestration):** A central controller is a "Process Manager" that actively runs the pipeline. This is an inherently brittle design—if the controller fails, the build fails.
- **Tekton (Choreography):** Kubernetes-native. A pipeline is a set of Kubernetes Custom Resources (CRDs). Each step is its own Pod, scheduled independently by Kubernetes. It's a "serverless" (no-idle) model using choreography (event-driven reactions) rather than central orchestration. More robust, no single point of failure.

**Verdict:** Tekton is architecturally superior for new, Kubernetes-native workloads. Jenkins is the "bridge to legacy," better for hybrid or existing enterprise workloads.

### When to Use Jenkins (and When Not To)

**Use Jenkins when:**
1. You have large existing investment in Jenkins jobs, shared Groovy libraries, and operational knowledge
2. Your pipelines require complex, programmatic, dynamic logic best expressed in Groovy
3. You must integrate with diverse or legacy ecosystems (on-prem, mainframes, custom SCMs) that only the Jenkins plugin ecosystem supports

**Use a Modern Alternative (Tekton, GitHub Actions) when:**
1. Starting a new, greenfield project
2. Workflows are 100% container-native
3. You value a "serverless" (no-idle) resource model and event-driven architecture over a central orchestrator

## Planton's Approach

Planton's `JenkinsKubernetes` API standardizes on the official `jenkinsci/jenkins` Helm chart as the deployment foundation. This choice is based on the research findings:

1. **The Helm Chart is the De Facto Standard:** It is mature, flexible, widely adopted, and strikes the right balance between abstraction and control compared to the Jenkins Operator.

2. **Focus on the 80/20:** The API abstracts the hundreds of Helm `values.yaml` keys into the essential fields that 80% of users need for production deployments: controller resources, persistence configuration, ingress settings, and JCasC YAML.

3. **JCasC as the Escape Hatch:** The API exposes a `jcasc_yaml` field as the primary mechanism for all advanced configuration (security, dynamic agents, plugin settings). This keeps the top-level API simple while supporting infinite flexibility for power users.

4. **Opinionated Best Practices:** The API enforces production requirements by making fields like `persistence.storage_class` and `controller_resources` mandatory rather than optional. It prevents anti-patterns like active-active HA (enforcing `replicas: 1`).

5. **Two Plugin Management Paths:**
   - For development/staging: Simple `plugins` list for ease of iteration
   - For production: `custom_image` field for immutable, pre-baked images with plugins included

This design provides a "golden path" for users: opinionated defaults that work reliably in production, with the flexibility to customize when needed via JCasC.

## Conclusion

The journey from anti-pattern to production-ready Jenkins on Kubernetes is not about fighting the platform—it's about understanding the maturity spectrum and choosing the right tools at each layer.

The combination of StatefulSets (for stable persistence), the official Helm chart (for packaging), JCasC (for declarative configuration), and GitOps (for delivery) has transformed Jenkins-on-Kubernetes from a risky experiment into a proven pattern. The key insight is that while Jenkins was not designed for Kubernetes, the ecosystem has matured to provide the abstractions necessary for production-grade deployments.

Planton's `JenkinsKubernetes` API codifies this understanding into a simple, opinionated interface that hides the complexity while preserving the flexibility that makes Jenkins powerful. It recognizes that Jenkins remains the bridge between legacy systems and cloud-native platforms—a role that is critical for enterprises modernizing their CI/CD infrastructure.

For teams starting greenfield projects with purely cloud-native workloads, alternatives like Tekton may be architecturally superior. But for the vast majority of enterprises with existing Jenkins investments, legacy system integrations, and complex pipeline requirements, Jenkins-on-Kubernetes—deployed the right way—remains the pragmatic choice.

