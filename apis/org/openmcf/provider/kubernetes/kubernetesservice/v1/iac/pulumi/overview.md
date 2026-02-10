# KubernetesService Pulumi Module: Architecture Overview

## High-Level Architecture

```
manifest.yaml
     |
     v
+-------------------+
| OpenMCF CLI       |
| (manifest loader) |
+-------------------+
     |
     v  stack-input (protobuf JSON)
+-------------------+
| main.go           |
| (entrypoint)      |
+-------------------+
     |
     v  KubernetesServiceStackInput
+-------------------+
| module/main.go    |
| (orchestrator)    |
+-------------------+
     |
     +---> module/locals.go     (transform spec -> derived values)
     |
     +---> K8s Provider          (from provider_config credentials)
     |
     +---> module/service.go    (create kubernetes.core/v1.Service)
     |
     +---> module/outputs.go    (export stack outputs)
```

## Data Flow

1. **Input Loading**: The OpenMCF CLI reads the YAML manifest, serializes it as protobuf JSON, and passes it to Pulumi as the `stack-input` config key.

2. **Locals Initialization**: `locals.go` extracts fields from the spec and computes derived values:
   - Merges user labels with standard OpenMCF labels
   - Resolves protobuf enum values to Kubernetes API strings (e.g., `cluster_ip` -> `"ClusterIP"`)
   - Builds the internal DNS name

3. **Provider Setup**: Creates a Pulumi Kubernetes provider from the `provider_config` credentials, supporting GKE, EKS, AKS, and DOKS clusters.

4. **Service Creation**: `service.go` creates a single `kubernetes.core/v1.Service` resource with:
   - Metadata (name, namespace, labels, annotations)
   - Spec (type, ports, selector, clusterIP, externalTrafficPolicy, sessionAffinity, loadBalancerSourceRanges)
   - Conditional configuration based on service type (headless, ExternalName, LoadBalancer)

5. **Output Export**: `outputs.go` exports observable values from the created resource, including the allocated cluster IP and load balancer ingress address.

## Key Design Decisions

- **Single Resource**: Unlike workload components that create multiple resources (namespace, quota, policies), the Service component creates exactly one `kubernetes.core/v1.Service`. This reflects the component's role as a primitive.

- **Enum Resolution in Locals**: Protobuf enum-to-string conversion happens once in `locals.go`, keeping the resource creation code in `service.go` clean and focused on Pulumi API calls.

- **Conditional Configuration**: Service type determines which fields are set on the Pulumi spec. For example, `externalTrafficPolicy` is only set for NodePort and LoadBalancer types, matching Kubernetes API behavior.

- **Target Port Flexibility**: The `target_port` field accepts both numeric strings and named ports. `service.go` parses the value and uses the appropriate Pulumi type (`pulumi.Int` vs `pulumi.String`).
