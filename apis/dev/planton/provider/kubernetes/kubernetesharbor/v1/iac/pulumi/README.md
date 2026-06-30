# Harbor Kubernetes - Pulumi Module

This Pulumi module deploys Harbor cloud-native registry on Kubernetes.

## Prerequisites

- Pulumi CLI installed
- Go 1.21 or higher
- Kubernetes cluster with a configured `kubeconfig`
- Harbor Helm chart repository access

## Usage

### Local Development

1. Initialize Go dependencies:
```bash
make deps
```

2. Create a stack input file (e.g., `stack-input.yaml`):
```yaml
target:
  apiVersion: kubernetes.planton.dev/v1
  kind: HarborKubernetes
  metadata:
    name: my-harbor
  spec:
    database:
      isExternal: false
      managedDatabase:
        container:
          replicas: 1
          isPersistenceEnabled: true
          diskSize: 20Gi
    cache:
      isExternal: false
      managedCache:
        container:
          replicas: 1
          isPersistenceEnabled: true
          diskSize: 8Gi
    storage:
      type: filesystem
      filesystem:
        diskSize: 100Gi
providerConfig:
  kubernetesProviderConfig:
    kubeConfigPath: ~/.kube/config
```

3. Run Pulumi:
```bash
pulumi stack init dev
pulumi config set --path harbor-kubernetes-stack-input --plaintext file://./stack-input.yaml
pulumi up
```

## Module Structure

- `main.go`: Entry point for the Pulumi program
- `module/main.go`: Main orchestration logic
- `module/locals.go`: Local variables and initialization
- `module/outputs.go`: Output constants
- `module/harbor.go`: Harbor Helm chart deployment
- `module/ingress_core.go`: Ingress for Harbor Core/Portal
- `module/ingress_notary.go`: Ingress for Notary service

## Outputs

The module exports the following outputs:
- `namespace`: Kubernetes namespace
- `core_service`: Harbor Core service name
- `portal_service`: Harbor Portal service name
- `registry_service`: Harbor Registry service name
- `external_hostname`: External hostname (if ingress enabled)
- `admin_username`: Harbor admin username
- `admin_password_secret`: Kubernetes secret containing admin password

## Debugging

To debug the Pulumi program:

1. Uncomment the `binary` option in `Pulumi.yaml`
2. Install delve: `go install github.com/go-delve/delve/cmd/dlv@latest`
3. Run `pulumi up` - it will start a debug server on port 2345
4. Attach your IDE debugger to localhost:2345

## References

- [Harbor Documentation](https://goharbor.io/docs/)
- [Harbor Helm Chart](https://github.com/goharbor/harbor-helm)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/docs/intro/cloud-providers/kubernetes/)

