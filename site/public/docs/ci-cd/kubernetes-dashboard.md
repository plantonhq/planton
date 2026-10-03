---
title: "Kubernetes Dashboard"
description: "View pods, stream logs, exec into containers, and browse resources for your deployed services directly from the web console"
icon: kubernetes
order: 70
tags:
  - Service Hub
  - Kubernetes
  - Pods
  - Debugging
---

# Kubernetes Dashboard

After deploying a service to Kubernetes, the Kubernetes tab on your infra component shows the live state of your deployment — pods, their status, logs, and a shell for debugging. No kubectl installation required, no kubeconfig to manage, no cluster credentials to distribute.

## What You See

Open any Kubernetes infra component in the web console and click the Kubernetes tab. The dashboard is scoped to your deployment's namespace — you see only the resources that belong to your service, not the entire cluster.

<!-- SCREENSHOT: Kubernetes dashboard overview
  Page: /resource/infra-hub/infra-component/kubernetes/{type}/{id}/kubernetes-resources
  Action: Show the Kubernetes tab with resource graph and pod list visible
  Focus: Full page showing the resource graph and pod list
  Alt: Kubernetes dashboard showing the resource dependency graph and pod list for a deployed service
-->

### Pod List

The pod list displays each pod with:

- **Name** — The pod identifier (clickable for details)
- **Ready** — How many containers are ready out of total
- **Status** — Current phase (Running, Pending, Failed, etc.)
- **Restarts** — Total restart count across containers

Each pod has an actions menu with three operations: stream logs, exec into a container, and delete.

### Resource Graph

The resource graph visualizes all Kubernetes resources created for your deployment as a directed acyclic graph. Nodes represent resources (Deployments, ReplicaSets, Pods, Services, ConfigMaps, Secrets) and edges show relationships between them. Click any node to inspect, edit, or delete the resource.

<!-- SCREENSHOT: Resource graph
  Page: /resource/infra-hub/infra-component/kubernetes/{type}/{id}/kubernetes-resources
  Action: Show the DAG canvas with Deployment -> ReplicaSet -> Pod relationships visible
  Focus: The graph visualization panel
  Alt: Directed acyclic graph showing Kubernetes resource relationships for a deployed service
-->

## Streaming Logs

Click "Stream Logs" on any pod to open the log viewer. Logs stream in real time and display in a scrollable viewer.

The filter panel lets you narrow logs by:

- **Container** — Select a specific container (if the pod has multiple)
- **Time range** — How far back to look (e.g., 5m, 30m, 4h)
- **Tail lines** — Number of previous lines to fetch
- **Content** — Search for specific text in log lines

Use the play/pause control to freeze the log stream while reading, then resume to catch up. Copy the entire log buffer to your clipboard with one click.

## Getting a Shell

Click "Exec into Container" on any pod to open a browser-based terminal at the bottom of the screen.

1. Select the container (if the pod has multiple)
2. Type the shell name when prompted (e.g., `bash` or `sh`)
3. Run commands directly in the container

The terminal supports standard input, scrolling, and can be expanded to full height. Type `exit` or close the drawer to end the session.

This is useful for quick debugging — checking environment variables, verifying file mounts, testing internal connectivity, or inspecting application state — without leaving the browser.

## Inspecting and Editing Resources

Click any resource in the graph or list to view its details:

- **Describe** — Equivalent to `kubectl describe`, showing the resource's current state, events, and conditions
- **YAML** — The full resource definition in YAML format

Click "Edit" to modify the YAML directly in the browser and apply changes immediately. Click "Delete" to remove a resource (with confirmation).

Keep in mind that manual edits are temporary — the next deployment through Service Hub may overwrite your changes. For permanent configuration changes, update your service configuration and redeploy.

## Full Operations Reference

The Kubernetes dashboard in Service Hub uses [Operations](/docs/operations) under the hood. For the complete operations reference — including CLI commands, admin-mode access to any namespace, advanced filtering, and the full command set — see [Operations > Kubernetes Operations](/docs/operations/kubernetes-operations).

## Related Documentation

- [Operations > Kubernetes Operations](/docs/operations/kubernetes-operations) — Full CLI reference and advanced operations
- [Operations Overview](/docs/operations) — How Cloud Ops works, dual access modes, supported providers
- [What is a Service](/docs/ci-cd/what-is-a-service) — How services bridge Git repositories and deployments
- [Deployment Targets](/docs/ci-cd/deployment-targets) — Where services can be deployed
