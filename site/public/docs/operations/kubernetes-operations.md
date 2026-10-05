---
title: "Kubernetes Operations"
description: "Manage Kubernetes pods, stream logs, exec into containers, and browse resources through the web console and CLI"
icon: kubernetes
order: 20
tags:
  - Cloud Ops
  - Kubernetes
  - Pods
  - Logs
  - Exec
---

# Kubernetes Operations

Kubernetes has the deepest Cloud Ops integration — pod management, real-time log streaming, interactive container shells, resource browsing with visual graphs, inline YAML editing, and resource deletion. All operations work through both the web console and CLI, in both [developer mode and admin mode](/docs/operations#two-ways-to-access).

## Viewing Pods

Pod viewing is the starting point for most Kubernetes operations. You can find pods by their managing resource (Deployment, StatefulSet, DaemonSet, Job, CronJob) and see their current state at a glance.

### In the Web Console

Navigate to any Kubernetes Infra Component and open the Kubernetes tab. The pod list shows each pod with its status, ready state, restart count, and an actions menu. Click a pod name to view its full details.

<!-- SCREENSHOT: Pod list view
  Page: /resource/infra-hub/infra-component/kubernetes/{type}/{id}/kubernetes-resources/view-pods
  Action: Show pod list with at least 3 pods in Running state
  Focus: The pod table with Name, Ready, Status, Restarts columns and actions menu
  Alt: Pod list table showing three running pods with their status, ready count, restart count, and actions dropdown
-->

From the actions menu on any pod, you can:

- **Stream Logs** — Open the real-time log viewer
- **Exec into Container** — Open a browser-based terminal
- **Delete** — Remove the pod (with confirmation)

### Using the CLI

Find pods for an Infra Component (developer mode):

```bash
planton kubectl get pods -r my-org/prod/KubernetesDeployment/payments-api
```

Find pods using a direct connection (admin mode):

```bash
planton kubectl get pods --connection k8s-prod-cluster -n payments
```

The `-r` flag accepts an Infra Component reference in the format `org/env/Kind/slug`. The `--connection` flag accepts a provider connection slug and requires `-n` to specify the namespace.

## Streaming Logs

Log streaming delivers real-time log output from running pods, with filtering to narrow down to exactly what you need.

### In the Web Console

Click "Stream Logs" on any pod to open the log viewer. Logs begin streaming immediately and appear in a scrollable, read-only editor.

<!-- SCREENSHOT: Log streaming view
  Page: /resource/infra-hub/infra-component/kubernetes/{type}/{id}/kubernetes-resources/{podStreamId}
  Action: Show log stream with active output and filter panel expanded
  Focus: The log viewer with filter panel showing container, since duration, and tail lines options
  Alt: Log streaming view showing real-time pod logs with an expandable filter panel at the top
-->

The log viewer provides:

- **Play/Pause** — Pause log updates to read through output, then resume
- **Filters** — Expand the filter panel to narrow logs by container name, time range, tail lines, or content search
- **Copy** — Copy all displayed logs to the clipboard
- **Clear** — Clear the displayed log buffer
- **Restart** — Stop and restart the log stream (useful after changing filters)

Each log line includes the pod name and container name as a prefix, so you can distinguish output when streaming from pods with multiple containers.

### Using the CLI

Stream logs for pods managed by a deployment (developer mode):

```bash
planton kubectl logs deployment payments-api -r my-org/prod/KubernetesDeployment/payments-api
```

Stream logs with filters (admin mode):

```bash
planton kubectl logs deployment payments-api --connection k8s-prod -n payments \
  --tail-lines 500 \
  --since-duration 30m \
  --container-name-filter app \
  --content-filter "ERROR"
```

The `logs` command takes two positional arguments: the pod manager kind (e.g., `deployment`, `statefulset`, `pod`) and the name. Available flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--tail-lines` | 200 | Number of previous log lines to fetch |
| `--since-duration` | 5m | How far back to look (e.g., `20s`, `2m`, `4h`) |
| `--container-name-filter` | — | Show logs only from containers matching this name |
| `--content-filter` | — | Show only log lines containing this substring |

## Exec into Containers

Container exec gives you an interactive shell inside a running container — directly from the browser or CLI. No kubectl required, no kubeconfig to manage.

### In the Web Console

Click "Exec into Container" on any pod. A terminal drawer opens at the bottom of the screen.

<!-- SCREENSHOT: Exec terminal
  Page: /resource/infra-hub/infra-component/kubernetes/{type}/{id}/kubernetes-resources/view-pods
  Action: Show exec drawer open with a shell session active in a container
  Focus: The terminal drawer showing a shell prompt with command output
  Alt: Browser-based terminal drawer showing an interactive shell session inside a Kubernetes container
-->

The workflow:

1. **Select a container** — If the pod has multiple containers, pick the one you want to shell into
2. **Choose a shell** — Type the shell name when prompted (`bash`, `sh`, or another available shell)
3. **Run commands** — The terminal connects and you get a live shell session

The terminal supports standard input, output display, and scrolling. You can expand the drawer to full height for more space. Type `exit` or click the close button to end the session.

Common use cases:

```bash
# Check environment variables
env | grep DATABASE

# Verify file mounts
ls -la /app/config

# Test internal connectivity
curl http://other-service:8080/health

# Inspect application state
cat /tmp/debug.log
```

### Using the CLI

Exec into a container (developer mode):

```bash
planton kubectl exec my-pod-7d8f9c6b4d-x2k9m app -r my-org/prod/KubernetesDeployment/payments-api
```

Exec with a specific shell (admin mode):

```bash
planton kubectl exec my-pod-7d8f9c6b4d-x2k9m app --connection k8s-prod -n payments --shell bash
```

The `exec` command takes two required positional arguments: the pod name and the container name. You can optionally append a command to run instead of an interactive shell:

```bash
planton kubectl exec my-pod app -r my-org/prod/KubernetesDeployment/payments-api -- cat /etc/config/app.yaml
```

| Flag | Default | Description |
|------|---------|-------------|
| `--shell` | sh | Shell to use for interactive sessions |

## Browsing Resources

Beyond pods, Cloud Ops lets you browse, inspect, edit, and delete any Kubernetes resource in a namespace.

### Listing Resources

In the web console, the Kubernetes tab on an Infra Component shows all resources organized by kind. The namespace graph view displays resources as a directed acyclic graph (DAG), visualizing relationships between Deployments, ReplicaSets, Pods, Services, ConfigMaps, and other resources.

<!-- SCREENSHOT: Namespace graph view
  Page: /resource/infra-hub/infra-component/kubernetes/{type}/{id}/kubernetes-resources
  Action: Show the DAG canvas with multiple resource types visible
  Focus: The graph visualization showing resource relationships
  Alt: Directed acyclic graph showing Kubernetes resources and their relationships for a deployment
-->

Using the CLI:

```bash
# List all resources in the namespace
planton kubectl get all -r my-org/prod/KubernetesDeployment/payments-api

# List resources by kind
planton kubectl get deployments -r my-org/prod/KubernetesDeployment/payments-api

# List namespaces (admin mode only)
planton kubectl get namespaces --connection k8s-prod
```

The `get` command supports multiple output formats:

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output-format` | table | Output format: `table`, `yaml`, or `graph` (graph works with `get all`) |

### Inspecting Resources

View the full details of any resource — equivalent to `kubectl describe` plus the complete YAML definition.

In the web console, click any resource node in the graph or any resource in a list to see its describe output and YAML.

Using the CLI:

```bash
# Describe a resource
planton kubectl describe deployment payments-api -r my-org/prod/KubernetesDeployment/payments-api

# Get a specific resource
planton kubectl get deployment payments-api -r my-org/prod/KubernetesDeployment/payments-api
```

### Editing Resources

Edit a resource's YAML directly. In the web console, click "Edit" on any resource to modify its YAML in the browser and apply changes immediately.

Using the CLI, the `edit` command opens the resource YAML in your `$EDITOR` (or vi), and applies changes when you save and close:

```bash
planton kubectl edit deployment payments-api -r my-org/prod/KubernetesDeployment/payments-api
```

Common edits include adjusting resource limits, updating environment variables, modifying labels or annotations, and changing replica counts.

### Deleting Resources

Delete a resource with confirmation. In the web console, click "Delete" on any resource — a confirmation dialog appears before the deletion proceeds.

Using the CLI:

```bash
planton kubectl delete pod payments-api-7d8f9c6b4d-x2k9m -r my-org/prod/KubernetesDeployment/payments-api
```

## CLI Reference

All `planton kubectl` subcommands share three persistent flags that control the access mode:

| Flag | Description |
|------|-------------|
| `-r, --resource` | Infra Component reference for developer mode: `org/env/Kind/slug` |
| `--connection` | Provider connection slug for admin mode (mutually exclusive with `-r`) |
| `-n, --namespace` | Kubernetes namespace (required in admin mode, ignored in developer mode) |

### Command Summary

| Command | Arguments | Description |
|---------|-----------|-------------|
| `kubectl get` | `all`, `namespaces`, `<kind>`, `<kind> <name>` | List or get Kubernetes resources |
| `kubectl describe` | `<kind> <name>` | Show detailed resource information |
| `kubectl edit` | `<kind> <name>` | Edit resource YAML in your editor |
| `kubectl delete` | `<kind> <name>` | Delete a resource |
| `kubectl logs` | `<pod-manager-kind> <name>` | Stream pod logs |
| `kubectl exec` | `<pod-name> <container-name>` | Exec into a container |

## Best Practices

### Use the Web Console for Quick Debugging

The browser-based terminal and log viewer are the fastest path to understanding what's happening in a running pod. No local tooling setup required — open the Infra Component, click a pod, and start debugging.

### Use the CLI for Automation and Scripting

The CLI supports the same operations with structured output formats (`table`, `yaml`). Use it in scripts, CI pipelines, or when you prefer terminal-based workflows.

### Prefer Developer Mode for Service Work

Developer mode scopes operations to your deployment's namespace automatically. You cannot accidentally inspect or modify resources belonging to other deployments. This is the safer default for day-to-day service work.

### Use Admin Mode for Cluster-Wide Operations

Admin mode gives you access to any namespace accessible through the connection. Use it for cross-cutting operational tasks like inspecting `kube-system`, checking node-level resources, or troubleshooting cluster infrastructure.

### Remember That Manual Edits Are Temporary

Changes made through Cloud Ops (editing a Deployment's replica count, modifying a ConfigMap) take effect immediately but may be overwritten by the next deployment. For permanent changes, update your infrastructure code or service configuration and redeploy.

## Related Documentation

- [Operations Overview](/docs/operations) — What Cloud Ops is, dual access modes, supported providers
- [Resource Browser](/docs/operations/resource-browser) — Multi-cloud listing of provider resources in AWS, GCP, and Azure
- [Runner](/docs/runner) — The secure execution agent that processes Cloud Ops requests
- [Connections > Kubernetes Clusters](/docs/connections/kubernetes-clusters) — How Kubernetes cluster credentials are managed
