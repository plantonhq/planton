# KubernetesGkeComputeClass Pulumi Module

This module creates one cluster-scoped GKE `ComputeClass` from a `KubernetesGkeComputeClass` manifest. It applies the manifest's projection through `manifestcr.Apply`, the same object the generated Terraform module builds: the spec under its upstream keys, the spec's `name` as the object's name, and the spec's `labels` and `annotations` under Planton's identity labels. No field is mapped by hand, so a field the spec gains reaches the object with no module change.

The module does not wait on the class. GKE's cluster autoscaler reads it the next time a pod that selects it is pending, and reports any problem in the class's `Health` condition.

## Usage

```bash
planton pulumi up --manifest compute-class.yaml
```

## Outputs

| Output | Description |
|--------|-------------|
| `compute_class_name` | Name of the created ComputeClass (the spec's `name`) |
