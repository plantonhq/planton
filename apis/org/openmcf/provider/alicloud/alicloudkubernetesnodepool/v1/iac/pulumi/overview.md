# AlicloudKubernetesNodePool Pulumi Module Overview

This module creates an ACK Kubernetes node pool within an existing managed cluster.

## Architecture

```
AlicloudKubernetesCluster (parent)
└── cs.NodePool
    ├── Worker ECS instances (managed by Auto Scaling group)
    ├── System disk (ESSD/SSD/efficiency)
    ├── Data disks (optional)
    ├── Kubernetes labels and taints
    └── Auto-scaling configuration (optional)
```

## Module Files

- `main.go` -- Entrypoint that loads stack input and delegates to the module
- `module/main.go` -- Creates the node pool resource with all field mappings
- `module/locals.go` -- Initializes tags, resolves foreign keys, computes defaults
- `module/outputs.go` -- Defines output constant names
