# OpenMCF

> **Deploy anywhere with one workflow.** Write declarative YAML once, deploy to AWS, GCP, Azure, or Kubernetes with the same CLI and consistent experience.

<p align="center">
  <img src="site/public/icon.png" alt="openmcf-logo"> <img src="site/public/text-logo.svg" alt="openmcf-text-logo">
</p>

## What is OpenMCF?

An open-source multi-cloud deployment framework that brings Kubernetes-style consistency to infrastructure deployments everywhere. No vendor lock-in, no artificial abstractions—just provider-specific configurations with a unified structure and workflow.

<!-- AI-AGENT NOTE: The deployment component count below MUST be verified against the authoritative source:
     apis/org/openmcf/shared/cloudresourcekind/cloud_resource_kind.proto
     Count only non-test, non-unspecified enum values in the CloudResourceKind enum. -->

**360+ deployment components** across **17 cloud providers** — AWS, GCP, Azure, Kubernetes, OCI, Alibaba Cloud, Hetzner Cloud, DigitalOcean, Cloudflare, Civo, Scaleway, OpenStack, Confluent, Snowflake, Auth0, MongoDB Atlas, and OpenFGA.

**[Documentation](https://openmcf.org)** · **[Component Catalog](https://openmcf.org/docs/catalog)** · **[Getting Started](https://openmcf.org/docs/getting-started)**

---

## Why OpenMCF?

- **One structure, any cloud** — Kubernetes Resource Model (apiVersion/kind/metadata/spec) for all deployments
- **Validate before deploy** — Protocol Buffer validations catch errors in seconds, not minutes
- **Zero abstraction** — Provider-specific configs preserve cloud capabilities; consistent experience across all
- **Choose your IaC** — Built-in Pulumi and Terraform/OpenTofu modules with feature parity
- **Build on top** — Auto-generated SDKs in Go, Python, TypeScript, Java from Protocol Buffer definitions

---

## Quick Start

### 1. Install the CLI

```bash
brew install plantonhq/tap/openmcf
```

### 2. Create a YAML Manifest

Example: Deploy PostgreSQL to Kubernetes using the [KubernetesPostgres](https://buf.build/openmcf/openmcf/file/main:org/openmcf/provider/kubernetes/kubernetespostgres/v1/spec.proto) deployment component.

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: my-first-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: local
    pulumi.openmcf.org/project: getting-started
    pulumi.openmcf.org/stack.name: dev
spec:
  namespace:
    value: my-first-postgres
  createNamespace: true
  container:
    replicas: 1
    resources:
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 500m
        memory: 512Mi
    diskSize: 1Gi
```

You can create similar manifests for [AWS VPC](https://github.com/plantonhq/openmcf/tree/main/apis/org/openmcf/provider/aws/awsvpc/v1), [GKE Cluster](https://github.com/plantonhq/openmcf/tree/main/apis/org/openmcf/provider/gcp/gcpgkecluster/v1), [Kafka on Kubernetes](https://github.com/plantonhq/openmcf/tree/main/apis/org/openmcf/provider/kubernetes/kuberneteskafka/v1), and [many more](https://github.com/plantonhq/openmcf/tree/main/apis/org/openmcf/provider).

### 3. Deploy

```bash
# Unified command (auto-detects provisioner from manifest labels)
openmcf apply -f postgres.yaml

# Or use IaC-specific commands directly
openmcf pulumi up -f postgres.yaml
openmcf tofu apply -f postgres.yaml
```

---

## CLI Overview

```
openmcf
├── apply / destroy / plan      Unified commands (provisioner auto-detected)
├── pulumi                      Pulumi-specific commands (init, preview, up, destroy, refresh)
├── tofu                        OpenTofu commands (init, plan, apply, destroy, refresh)
├── terraform                   Terraform commands (init, plan, apply, destroy, refresh)
├── validate                    Validate manifest against protobuf schema
├── pull / checkout             Module management
├── config                      CLI configuration
└── version / upgrade           Version management
```

See the full [CLI Reference](https://openmcf.org/docs/cli/cli-reference) for all commands, flags, and options.

---

## Learn More

- **[Getting Started Guide](https://openmcf.org/docs/getting-started)** — Your first deployment in 10 minutes
- **[Component Catalog](https://openmcf.org/docs/catalog)** — Browse 360+ deployment components across 17 providers
- **[Architecture](https://openmcf.org/docs/concepts/architecture)** — How Protocol Buffers, IaC modules, and CLI work together
- **[Planton Cloud](https://planton.cloud)** — Commercial SaaS platform with UI, CI/CD, and team features

---

## Contributing

Visit [CONTRIBUTING.md](CONTRIBUTING.md) for information on building OpenMCF from source or contributing improvements.

Also, refer to the [Contributor Guide](https://openmcf.org/docs/contributing) for detailed information about becoming a contributor to OpenMCF.

## License

OpenMCF is released under the [Apache 2.0 license](LICENSE). You are free to use, modify, and distribute this software in accordance with the license terms.

## Acknowledgments

- **Brian Grant & Kubernetes API team** for their foundational work on the Kubernetes Resource Model.
- The **[Protobuf Team](https://protobuf.dev/)** for laying the foundation for a powerful language neutral contract definition language.
- The **[Buf](https://github.com/bufbuild/buf) Team** for their Protobuf tooling—including BSR Docs, BSR SDKs, and ProtoValidate — which collectively democratized protobuf adoption and made this project possible.
- The **[Pulumi](https://github.com/pulumi/pulumi)** team for providing a powerful infrastructure as code platform that enables multi-language support.
- The **[spf13/cobra](https://github.com/spf13/cobra)** team for making building command line tools a bliss.
