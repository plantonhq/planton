# Base Images

This directory contains Docker base images that provide pre-configured environments for running infrastructure-as-code programs.

## Available Images

| Image                       | Description                                                    | Registry                                            |
| --------------------------- | -------------------------------------------------------------- | --------------------------------------------------- |
| [iac-runner](./iac-runner/) | Base image with Go, Pulumi, OpenTofu, and pre-warmed Go caches | `ghcr.io/plantonhq/planton/base-images/iac-runner` |

## Why Base Images?

Building Pulumi Go programs requires:

1. Downloading Go module dependencies (~500MB-1GB)
2. Compiling Pulumi provider SDKs (~30-120 seconds)

By pre-warming these caches during image build, runtime compilation drops from minutes to seconds.

## Versioning

Base images are tagged with the same version as planton releases. When you tag a new release (e.g., `v1.2.3`), the corresponding base image is built and published.

## Usage

```dockerfile
FROM ghcr.io/plantonhq/planton/base-images/iac-runner:v1.2.3

# Your application
COPY . /app
RUN cd /app && go build .
```
