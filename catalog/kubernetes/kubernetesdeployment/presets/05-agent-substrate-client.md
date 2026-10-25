# Agent Substrate Client

This preset deploys a service that creates and calls actors on a KubernetesAgentSubstrate in the same cluster: everything the substrate's outputs say a caller presents, mounted by the kubelet and kept current as it rotates.

## When to Use

- A product's server that creates, suspends and resumes actors through the Agent Substrate API
- Any workload that pushes to actors through the router's TLS endpoint

## Key Configuration Choices

- **Three projected identity files** — a ServiceAccount token for the API's audience (`api.ate-system.svc`), the API's ClusterTrustBundle (referenced from the substrate's `api_trust_bundle_name` output, so a rotated CA reaches the pod on its own — never copy the PEM into a ConfigMap), and a pod certificate from the router's client-certificate signer (`router_client_signer_name`)
- **Endpoints by reference** — the API address, its server name and the router's TLS endpoint come from the substrate's outputs, not literals
- **A labeled namespace** — the substrate's NetworkPolicy admits callers only from namespaces carrying its `api_access_namespace_label` (`ate.dev/api-access=true`); the namespace reference points at a KubernetesNamespace that sets the label (see the KubernetesNamespace preset **04-agent-substrate-client**)
- **The router's allow-list** — the pod certificate carries `spiffe://cluster.local/ns/<namespace>/sa/<service-account>`; name it in the substrate's `router.allowed_client_spiffe_ids` for the TLS endpoint to accept it
- **An owner of the API** — the API authorizes every call; name the service account (`system:serviceaccount:<namespace>:<service-account>`) in the substrate's `apiServer.authorization.bootstrapOwners`, or have an owner grant it an atespace's access policy, or every call is denied

## Engine and Cluster

The trust-bundle and pod-certificate mounts deploy on the **Pulumi engine** (`planton.dev/provisioner: pulumi`); the OpenTofu engine refuses them by name at plan time. The cluster needs the substrate's own prerequisites — the ClusterTrustBundle, ClusterTrustBundleProjection and PodCertificateRequest features — which any cluster running the substrate already has.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-substrate-client-namespace>` | The KubernetesNamespace (labeled for API access) the service runs in | Your `KubernetesNamespace` resource |
| `<your-agent-substrate>` | The substrate this service calls | Your `KubernetesAgentSubstrate` resource |
| `<your-container-registry>/<your-image>` | Container image repository | Your container registry |
| `<your-image-tag>` | Image tag or version | Your CI/CD pipeline output |

## Related Presets

- **03-worker** — a background process with no inbound traffic
- **04-hardened-production** — the restricted-profile security baseline to layer on top
