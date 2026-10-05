# KubernetesPodMonitor Pulumi Module

This Pulumi module creates one namespaced prometheus-operator `PodMonitor` on a target cluster.

It builds the object from the manifest's projection (`pkg/iac/pulumi/pulumimodule/provider/kubernetes/manifestcr`), the same projection the generated Terraform module applies. Every spec field reaches the custom resource under its upstream key, `namespace` becomes `metadata.namespace`, and `labels` and `annotations` become the object's own metadata under Planton's identity labels. The module maps no field by hand, so every field the spec gains reaches the object with no change here.

## Prerequisites

- The prometheus-operator CRDs on the cluster (see the `KubernetesKubePrometheusStack` kind).
- A Prometheus whose PodMonitor selector matches the object, for the pods to be scraped. The object applies with only the CRDs present.
- The target namespace (see `KubernetesNamespace`).
- Go toolchain and the Pulumi CLI, and access to the target cluster.

## Usage

### With the Planton CLI

```bash
planton pulumi up --manifest ../../e2e/manifest.yaml
```

### Direct Pulumi usage

The entrypoint loads the `KubernetesPodMonitorIacInput` from the `IAC_INPUT_YAML_FILE` environment variable (path to the IaC input, with the manifest under `target`) or `IAC_INPUT_YAML` (inline YAML content). The CLI builds that input from a manifest
and runs Pulumi:

```bash
planton pulumi up --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>
```

## Outputs

| Output | Description |
|--------|-------------|
| `pod_monitor_name` | Name of the created PodMonitor (equals `metadata.name`) |
| `namespace` | Namespace the PodMonitor was created in |

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
