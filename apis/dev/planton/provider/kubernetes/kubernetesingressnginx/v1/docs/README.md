# NGINX Ingress Controller on Kubernetes: The Path to Production

## Introduction

When Kubernetes first emerged, the conventional wisdom was simple: expose your services directly with `LoadBalancer` or `NodePort`, and call it a day. But as applications grew in complexity, teams faced a crucial question: *How do we route HTTP traffic to dozens—or hundreds—of services without creating dozens of cloud load balancers?*

Enter the **Ingress** concept: a single entry point that routes traffic based on hostnames and paths. And among ingress controllers, **NGINX Ingress Controller** (kubernetes-ingress-nginx) has become the de facto standard for production Kubernetes deployments. It's not just popular—it's proven. With 19,000+ GitHub stars, battle-tested at massive scale, and a vibrant community, it represents the sweet spot between power and practicality.

But here's the challenge: deploying kubernetes-ingress-nginx properly across different clouds (GKE, EKS, AKS) involves navigating a maze of annotations, load balancer configurations, and cloud-specific quirks. Planton abstracts this complexity, providing a unified API to deploy production-grade ingress controllers anywhere.

This document explores:
1. **The deployment landscape** – from raw manifests to managed services
2. **Why the official Helm chart** is the production-ready choice
3. **Cloud integration patterns** that make or break deployments
4. **The 80/20 configuration** that covers real-world needs
5. **Production best practices** for reliability and security

## The Deployment Method Spectrum

### Level 0: Raw YAML Manifests

The simplest path to deploying kubernetes-ingress-nginx is applying the [official manifests](https://github.com/kubernetes/kubernetes-ingress-nginx) via `kubectl apply -f`. The kubernetes-ingress-nginx project provides ready-to-use YAML files that create all necessary components: namespace, RBAC, Deployment, Service, and admission webhooks.

**Why this works for quick starts:**
- Zero additional tools required
- Perfect for testing or air-gapped environments
- Complete control over every Kubernetes object

**Why it's not production-ready:**
- Manual upgrades across all clusters
- No parameterization (must edit YAML for each environment)
- Difficult to track configuration drift
- No systematic way to apply cloud-specific customizations

**Verdict:** Great for learning and experimentation, but production teams need better tooling.

### Level 1: Package Managers (Helm)

The **official kubernetes-ingress-nginx Helm chart** transforms deployment from a copy-paste exercise into a parameterized, version-controlled operation. This chart is maintained alongside the controller itself, ensuring tight version compatibility.

**What makes it production-grade:**
- **Version synchronization**: Chart version 4.14.0 maps to controller v1.14.0, with explicit Kubernetes version compatibility matrices
- **Upgrade management**: Helm tracks releases and enables rollbacks
- **Parameterization**: A single `values.yaml` controls behavior across clouds
- **Automated complexity**: Handles admission webhook certificates, RBAC, and IngressClass creation
- **Battle-tested**: Used by enterprises worldwide with proven stability

**Configuration flexibility examples:**
```yaml
controller:
  replicaCount: 3
  service:
    type: LoadBalancer
    annotations:
      # Cloud-specific annotations go here
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true  # Prometheus integration
```

**Alternative charts:** Bitnami previously maintained an kubernetes-ingress-nginx chart, but even Bitnami now recommends the official kubernetes/kubernetes-ingress-nginx chart. It's the most current and broadly supported option.

**Verdict:** This is the **production standard**. Planton uses the official Helm chart under the hood, augmenting it with cloud-aware configuration generation.

### Level 2: Cloud-Managed Ingress Controllers

Every major cloud provider offers native ingress solutions that bypass running kubernetes-ingress-nginx entirely:

#### Google Cloud (GKE Ingress)

GKE's built-in ingress creates Google Cloud HTTP(S) Load Balancers—a fully managed, globally distributed L7 load balancer.

**Advantages:**
- No ingress controller pods to manage
- Global load balancing with CDN integration
- Google-managed SSL certificates
- Handles massive scale automatically

**Limitations:**
- Limited to GCP load balancer features (no custom Nginx configs)
- Can't do complex rewrites, Lua scripting, or custom response headers
- Tight coupling to GCP (no multi-cloud portability)

#### AWS Load Balancer Controller (ALB)

The AWS Load Balancer Controller watches Ingress objects and provisions Application Load Balancers (ALBs).

**Advantages:**
- Native AWS WAF integration
- OIDC authentication at load balancer level
- ACM (AWS Certificate Manager) integration
- Target group management for pods

**Limitations:**
- "Limited customization compared to NGINX... less flexible configuration options" (per CloudThat analysis)
- No support for advanced routing logic or response manipulation
- Vendor lock-in to AWS

#### Azure Application Gateway Ingress Controller (AGIC)

AGIC programs an Azure Application Gateway based on Ingress resources.

**Advantages:**
- Managed Azure WAF capabilities
- Integration with Azure Private Link
- Autoscaling managed by Azure

**Limitations:**
- Requires Azure-specific identity configuration (Managed Identity)
- Additional cost for Application Gateway resource
- Complex setup compared to NGINX

**When to choose cloud-managed ingress:**
- You're deeply invested in one cloud and need its native features (WAF, auth, CDN)
- You want to minimize operational overhead at the cost of flexibility
- Your routing needs are straightforward (host/path-based only)

**When to choose NGINX:**
- You need **multi-cloud portability** (consistent behavior everywhere)
- You require **advanced features** like custom headers, complex rewrites, rate limiting per annotation, Lua plugins
- You want configuration-as-code that isn't tied to cloud APIs
- You need fine-grained control over TLS termination and proxy behavior

### Level 3: Operator-Based Deployment

Unlike some Kubernetes components, kubernetes-ingress-nginx **does not have an official Operator**. The community relies on Helm for lifecycle management. (Note: NGINX Inc. offers an operator for their commercial NGINX Plus controller, which is a separate product.)

For GitOps workflows, tools like **Argo CD** or **Flux** manage the Helm chart, providing operator-like continuous reconciliation without requiring a custom operator.

**Verdict:** Not a primary deployment method for kubernetes-ingress-nginx, though GitOps tooling provides similar benefits.

## Why the Official Helm Chart is Production-Ready

The kubernetes/kubernetes-ingress-nginx Helm chart isn't just a packaging tool—it's a **comprehensive deployment system** designed for enterprise reliability.

### Enterprise Features Baked In

**High Availability:** 
- Configurable replica count (default should be ≥2 for HA)
- PodDisruptionBudget support to protect during cluster maintenance
- Pod anti-affinity to spread replicas across nodes/zones
- Resource requests/limits for guaranteed capacity

**Security Controls:**
- Validating admission webhook (prevents invalid Ingress configs)
- Automatic webhook certificate management via `kube-webhook-certgen` job
- Non-root container execution (user 101)
- RBAC with minimal ClusterRole permissions

**Observability:**
- Prometheus metrics endpoint (`/metrics` on port 10254)
- ServiceMonitor CRD support for Prometheus Operator
- Configurable access logging with custom formats
- Official Grafana dashboards available

**Service Exposure Flexibility:**
- LoadBalancer (default for cloud environments)
- NodePort (for on-premise or advanced networking)
- HostNetwork (for bare-metal DaemonSet deployments)
- Complete control via annotations for cloud-specific behaviors

### Version Compatibility & Stability

The chart maintainers provide explicit compatibility matrices:
- Chart version → Controller version → Kubernetes version
- Only recent Kubernetes versions are officially tested (though older may work)
- Synchronized releases ensure no surprises during upgrades

While kubernetes-ingress-nginx is community-driven (no formal SLA), it's used by thousands of production clusters and actively maintained. Many cloud vendors and enterprise users contribute to its stability.

## Cloud Provider Integration Patterns

Deploying kubernetes-ingress-nginx across GKE, EKS, and AKS requires understanding each cloud's load balancer provisioning mechanisms. The Helm chart provides the hooks—via Service annotations—but you must know what to set.

### Google Cloud (GKE)

**Default Behavior:**
When you create a `Service` of type `LoadBalancer`, GKE provisions a **TCP Network Load Balancer** (regional, L4) that forwards traffic to node ports.

**Static IP Assignment:**
```yaml
controller:
  service:
    loadBalancerIP: 34.123.45.67  # Pre-reserved static IP
```

Reserve the IP first:
```bash
gcloud compute addresses create ingress-ip --region=us-central1
```

GCP will assign this IP to the load balancer instead of allocating a random one.

**Internal Load Balancer:**
For private ingress (VPC-only access):
```yaml
controller:
  service:
    annotations:
      cloud.google.com/load-balancer-type: "Internal"
```

This creates an Internal TCP/UDP Load Balancer with a private IP from your subnet.

**Firewall Rules:**
GCP's cloud controller automatically creates firewall rules allowing traffic to the node ports. For external LBs, this opens traffic from `0.0.0.0/0` by default.

### Amazon Web Services (EKS)

**Default Behavior:**
EKS provisions a **Classic ELB** by default for Services of type LoadBalancer. Classic ELBs are outdated for ingress use (limited TLS support, no host-based routing at L7).

**Recommended: Network Load Balancer (NLB)**
```yaml
controller:
  service:
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
```

NLBs provide:
- High performance (millions of requests per second)
- Static IPs per Availability Zone
- Better integration with kubernetes-ingress-nginx (pure L4 pass-through)

**Internal Load Balancer:**
```yaml
annotations:
  service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
  service.beta.kubernetes.io/aws-load-balancer-internal: "true"
```

**Subnet Control:**
```yaml
annotations:
  service.beta.kubernetes.io/aws-load-balancer-subnets: "subnet-abc123,subnet-def456"
```

Critical for environments with separate public/private subnets. This ensures the NLB is created only in specified subnets (e.g., public subnets with internet gateway for external ingress).

**Static IP via Elastic IPs:**
```yaml
annotations:
  service.beta.kubernetes.io/aws-load-balancer-eip-allocations: "eipalloc-xxxxx,eipalloc-yyyyy"
```

Pre-allocate Elastic IPs and attach them to the NLB. The number of EIPs must match the number of AZs/subnets.

**Security Groups:**
With NLB in instance mode, traffic goes to node ports. Worker nodes' security groups must allow inbound traffic on NodePort range. AWS cloud controller typically manages this, but verify if you have custom security group configurations.

**Comparison: NLB vs. ALB**
- **Use NLB + NGINX** when you need advanced ingress features (custom headers, complex routing, rate limiting)
- **Use ALB** (via AWS Load Balancer Controller) when you want managed L7 features (OIDC auth, AWS WAF) but accept less routing flexibility

Many production environments use both: ALB for public APIs (leveraging WAF), NGINX for internal or complex routing needs.

### Microsoft Azure (AKS)

**Default Behavior:**
AKS provisions an **Azure Load Balancer (Standard SKU)** for Services of type LoadBalancer, assigning a public IP from Azure's pool.

**Static Public IP:**
```yaml
controller:
  service:
    loadBalancerIP: 52.168.12.34
    annotations:
      service.beta.kubernetes.io/azure-load-balancer-resource-group: "MC_myCluster_rg_westeurope"
```

Steps:
1. Create a Public IP resource in Azure (same region as cluster)
2. Reference its address in `loadBalancerIP`
3. If the IP is in a different resource group than the cluster's managed resource group, specify it via annotation

**Internal Load Balancer:**
```yaml
controller:
  service:
    annotations:
      service.beta.kubernetes.io/azure-load-balancer-internal: "true"
      service.beta.kubernetes.io/azure-load-balancer-internal-subnet: "mySubnet"
```

This creates an internal LB with a private IP in the specified subnet (must be in the cluster's VNet).

**Network Security Groups (NSGs):**
AKS's cloud controller automatically updates NSGs to allow traffic from the load balancer to node ports. Ensure custom NSG rules don't block this traffic.

**Managed Identity:**
AKS clusters use a system-assigned Managed Identity (or service principal) to provision Azure resources like load balancers and IPs. This is configured at cluster creation—no additional setup needed for basic kubernetes-ingress-nginx.

**Alternative: Azure Application Gateway Ingress Controller (AGIC)**
For Azure-native L7, AGIC programs an Application Gateway. However, this requires:
- Dedicated Application Gateway resource (additional cost)
- User-assigned Managed Identity with Contributor rights to the gateway
- More complex configuration

For multi-cloud consistency, stick with kubernetes-ingress-nginx.

## The 80/20 Configuration: What Most Teams Actually Need

The kubernetes-ingress-nginx Helm chart exposes hundreds of configuration options. In practice, **80% of deployments** only customize a handful of settings.

### Essential Configuration Fields

**1. Internal vs. External Scope**
```yaml
# External (public internet access)
controller:
  service:
    type: LoadBalancer
    # No 'internal' annotation = external by default

# Internal (VPC/private access only)
controller:
  service:
    type: LoadBalancer
    annotations:
      # GCP
      cloud.google.com/load-balancer-type: "Internal"
      # AWS
      service.beta.kubernetes.io/aws-load-balancer-internal: "true"
      # Azure
      service.beta.kubernetes.io/azure-load-balancer-internal: "true"
```

**2. Static IP or DNS**
```yaml
controller:
  service:
    loadBalancerIP: "34.123.45.67"  # GCP/Azure: actual IP
    annotations:
      # AWS: use EIP allocations instead
      service.beta.kubernetes.io/aws-load-balancer-eip-allocations: "eipalloc-xxxxx"
```

**3. High Availability (Replicas)**
```yaml
controller:
  replicaCount: 2  # Minimum for HA; 3+ for critical workloads
  affinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        - labelSelector:
            matchExpressions:
              - key: app.kubernetes.io/name
                operator: In
                values:
                  - kubernetes-ingress-nginx
          topologyKey: kubernetes.io/hostname
```

**4. Resource Sizing**
```yaml
controller:
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 1
      memory: 2Gi
```

**5. TLS/HTTPS Configuration**
```yaml
controller:
  extraArgs:
    # Enforce HTTPS globally
    default-ssl-certificate: "default/tls-secret"
  config:
    force-ssl-redirect: "true"
    hsts: "true"
```

**6. Observability**
```yaml
controller:
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true  # For Prometheus Operator
```

### Advanced Features (The Other 20%)

Most teams don't need these initially but should know they exist:

- **ModSecurity WAF**: Compile-in OWASP Core Rule Set for protection against SQL injection, XSS
- **Rate Limiting**: `nginx.ingress.kubernetes.io/limit-rps` annotation per Ingress
- **Custom NGINX Config**: ConfigMap snippets for Lua, custom headers, etc.
- **Canary Deployments**: Traffic splitting via annotations
- **GeoIP, JWT Auth**: Available via modules/annotations

Planton's API focuses on the essential 80%, with escape hatches for advanced use cases.

## Production Best Practices

### High Availability

**Always run multiple replicas** (minimum 2, ideally 3) with pod anti-affinity to ensure replicas land on different nodes. Configure a PodDisruptionBudget to maintain availability during voluntary disruptions (node drains, cluster upgrades).

```yaml
controller:
  replicaCount: 3
  podDisruptionBudget:
    enabled: true
    minAvailable: 1
```

### TLS/SSL Management

**Use cert-manager** for automated certificate provisioning and renewal. cert-manager integrates with Let's Encrypt and can automatically create/renew TLS certificates referenced by Ingress resources.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
    - hosts:
        - example.com
      secretName: example-tls  # cert-manager creates this
```

Enable global HTTPS redirect to avoid serving HTTP:
```yaml
controller:
  config:
    force-ssl-redirect: "true"
```

### Client IP Preservation

To see real client IPs (not node IPs) in application logs, enable `externalTrafficPolicy: Local`:

```yaml
controller:
  service:
    externalTrafficPolicy: Local
```

**Trade-off:** Traffic only goes to nodes running kubernetes-ingress-nginx pods. Ensure adequate replica spread across zones.

### Monitoring & Alerting

**Integrate with Prometheus:**
- Enable the metrics endpoint (on by default)
- Create ServiceMonitor for Prometheus Operator
- Import the [official Grafana dashboard](https://github.com/kubernetes/kubernetes-ingress-nginx/tree/main/deploy/grafana/dashboards)

**Key metrics to watch:**
- Request rate (requests/sec)
- Response latency (P50, P95, P99)
- Error rates (4xx, 5xx)
- Active connections
- Configuration reload count

Set alerts for:
- Sustained 5xx rate increase (backend or ingress issues)
- High pod CPU/memory usage (capacity planning)
- Configuration reload failures

### Security Hardening

**1. Rate Limiting:**
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/limit-rps: "10"
    nginx.ingress.kubernetes.io/limit-burst-multiplier: "2"
```

**2. IP Whitelisting:**
```yaml
annotations:
  nginx.ingress.kubernetes.io/whitelist-source-range: "10.0.0.0/8,192.168.0.0/16"
```

**3. WAF (Optional):**
Enable ModSecurity with OWASP Core Rule Set:
```yaml
controller:
  config:
    enable-modsecurity: "true"
    enable-owasp-modsecurity-crs: "true"
```

Start in detection mode, tune rules, then enable blocking.

**4. Keep Updated:**
Regularly upgrade kubernetes-ingress-nginx to patch security vulnerabilities. The community is responsive, but you must apply updates. Plan quarterly upgrade windows (HA setup allows zero-downtime rolling updates).

### Common Pitfalls to Avoid

**Forgetting IngressClass:**
If your cluster has multiple ingress controllers, Ingress resources without `ingressClassName` may be ignored. Always specify:
```yaml
spec:
  ingressClassName: nginx
```

**Service/Port Mismatches:**
If an Ingress points to a non-existent Service or wrong port name, you'll get 503 errors. Always verify with `kubectl describe ingress`.

**DNS Before TLS:**
Pointing DNS to the ingress LB before certificates are ready causes cert errors. Use cert-manager to provision certs in parallel with DNS setup, or temporarily allow HTTP for ACME challenges.

**Ignoring Capacity:**
Ingress-nginx can become a bottleneck. Load test before production, monitor CPU/memory, and scale proactively.

**Proxy Protocol Confusion (AWS):**
If enabling proxy protocol on NLB, ensure kubernetes-ingress-nginx expects it:
```yaml
controller:
  config:
    use-proxy-protocol: "true"
```

## Comparison: NGINX vs. Alternatives

### When to Choose NGINX

✅ **Multi-cloud portability** (identical behavior on GKE, EKS, AKS, on-premise)  
✅ **Advanced routing** (custom headers, complex rewrites, Lua scripts)  
✅ **Mature ecosystem** (massive community, extensive documentation)  
✅ **Fine-grained control** (hundreds of tunable options)  
✅ **Cost-effective** (consolidate many routes under one controller)

### When to Choose Cloud-Managed Ingress

✅ **Cloud-native simplicity** (no pods to manage)  
✅ **Integrated features** (AWS WAF, GCP CDN, Azure WAF)  
✅ **Extreme scale** (cloud LBs handle massive traffic automatically)  
✅ **Single-cloud commitment** (already invested in cloud-specific tooling)

### Comparison with Other Controllers

| Controller | Best For | Trade-offs |
|------------|----------|------------|
| **NGINX** | Production-grade, multi-cloud | Battle-tested, flexible, massive community |
| **Traefik** | Dynamic environments, auto-TLS | HTTP-only in OSS, less performance at extreme scale |
| **HAProxy** | Maximum performance | Smaller community, complex configuration |
| **Envoy/Contour** | Modern cloud-native, service mesh prep | Newer, fewer modules than NGINX |
| **Istio Gateway** | If already using Istio mesh | Heavy (requires full Istio), powerful traffic policies |

For most teams, **NGINX is the safe, proven default**. It balances power, flexibility, and operational maturity.

## Planton's Approach

Planton abstracts cloud-specific complexity while preserving NGINX's power:

**Unified API:**
- `scope: external | internal` → Generates correct cloud annotations
- `staticIP: <name-or-address>` → Maps to cloud-specific IP mechanisms
- `replicas: 3` → Sets HA configuration with anti-affinity
- `metrics: enabled` → Configures Prometheus integration

**Cloud-Aware Defaults:**
- On AWS: Provisions NLB (not Classic ELB) with proper subnet selection
- On GCP: Handles static IP reservation and firewall rule verification
- On Azure: Manages Public IP resources and NSG configurations

**Production-Ready Out of Box:**
- cert-manager integration for automated TLS
- ServiceMonitor for Prometheus scraping
- PodDisruptionBudget for HA
- Sensible resource requests/limits

**Escape Hatches:**
For advanced needs (custom Lua, ModSecurity tuning), Planton allows injecting custom Helm values or ConfigMap snippets.

## Conclusion

The journey from "just expose services" to production-grade ingress has been a maturation story. Raw manifests gave way to Helm charts. Cloud providers offered managed alternatives, but teams discovered the need for portability and advanced features. NGINX emerged as the battle-tested standard—not because it's the newest or flashiest, but because it's **proven, powerful, and portable**.

Deploying kubernetes-ingress-nginx correctly requires understanding cloud-specific load balancer provisioning, TLS management, and operational best practices. Planton distills this complexity into a clean API, letting you focus on your applications rather than annotation arcana.

**Key Takeaways:**
1. Use the **official kubernetes-ingress-nginx Helm chart** for production
2. **Cloud integration** is about annotations, not black magic—know what to set for your provider
3. **80% of configuration** is straightforward: internal vs. external, static IPs, replica count
4. **Production readiness** means HA, TLS automation, monitoring, and security hardening
5. **NGINX remains the best choice** for teams needing multi-cloud portability and advanced routing

Whether you're running in GKE, EKS, AKS, or bare metal, kubernetes-ingress-nginx provides a consistent, powerful gateway to your applications. And with Planton, it's just a few lines of protobuf away.

