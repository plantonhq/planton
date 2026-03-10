# Overview

The `grafana-kubernetes-pulumi-module` automates the deployment and management of Grafana within a Kubernetes environment. It utilizes Planton’s unified API structure to define a declarative `GrafanaKubernetes` resource, which specifies the configuration for deploying Grafana, including container resources, Kubernetes services, and ingress settings. This module takes the input specification and creates all necessary Kubernetes resources, including setting up the namespace, provisioning the Grafana container, and configuring ingress for external access if required.

The module is built using Pulumi’s Go SDK, allowing developers to manage Grafana’s lifecycle within Kubernetes using infrastructure-as-code. Key outputs, such as Kubernetes service names, namespace details, ingress endpoints, and port-forwarding commands (when ingress is disabled), are captured and made available in `status.outputs`. This approach allows developers to deploy Grafana consistently across environments with minimal manual configuration, ensuring scalable and maintainable monitoring infrastructure.

