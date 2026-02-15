---
title: "Presets"
description: "Ready-to-deploy configuration presets for Artifact Registry Repo"
type: "preset-list"
componentSlug: "artifact-registry-repo"
componentTitle: "Artifact Registry Repo"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-docker-private"
    rank: "01"
    title: "Private Docker Repository"
    excerpt: "This preset creates a private Artifact Registry repository for Docker container images. It is the most common repository type in cloud-native workflows, used for storing application images built by..."
  - slug: "02-npm-private"
    rank: "02"
    title: "Private NPM Repository"
    excerpt: "This preset creates a private Artifact Registry repository for NPM packages. Use this for hosting internal JavaScript/TypeScript libraries that should not be published to the public NPM registry."
---

# Artifact Registry Repo Presets

Ready-to-deploy configuration presets for Artifact Registry Repo. Each preset is a complete manifest you can copy, customize, and deploy.
