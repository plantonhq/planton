# KubernetesAgentSandboxTemplate Pulumi Module

This Pulumi module creates one namespaced agent-sandbox `SandboxTemplate` on a target cluster.

It builds the object from the manifest's projection (`pkg/iac/pulumi/pulumimodule/provider/kubernetes/manifestcr`), the same projection the generated Terraform module applies. The blueprint (the pod template, the volume claim templates, `service`), the network policy and the claim policies reach the custom resource under their upstream keys, `namespace` becomes `metadata.namespace`, and `labels` and `annotations` become the object's own metadata under Planton's identity labels. The module maps no field by hand, so every field the spec gains reaches the object with no change here.

## Prerequisites

- The agent-sandbox controller and its extension definitions on the cluster (see the `KubernetesAgentSandboxController` kind, extensions on). The object applies with only the definition present; the controller is what gives it its network policy.
- The target namespace (see `KubernetesNamespace`).
- Go toolchain and the Pulumi CLI, and access to the target cluster.

## Usage

### With the Planton CLI

```bash
planton pulumi up --manifest ../../e2e/manifest.yaml
```

### Direct Pulumi usage

The entrypoint loads the `KubernetesAgentSandboxTemplateIacInput` from the `IAC_INPUT_YAML_FILE` environment variable (path to the IaC input, with the manifest under `target`) or `IAC_INPUT_YAML` (inline YAML content). The CLI builds that input from a manifest and runs Pulumi:

```bash
planton pulumi up --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>
```

## Outputs

| Output | Description |
|--------|-------------|
| `sandbox_template_name` | Name of the created SandboxTemplate (equals `metadata.name`), which warm pools reference |
| `namespace` | Namespace the SandboxTemplate was created in |

## Module Structure

```
pulumi/
├── main.go          # Pulumi entrypoint (loads the IaC input)
├── Pulumi.yaml      # Pulumi project configuration
├── README.md        # This file
└── module/
    ├── main.go      # Applies the projected custom resource
    └── outputs.go   # Output name constants
```
