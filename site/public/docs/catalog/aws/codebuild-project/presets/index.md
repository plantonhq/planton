---
title: "Presets"
description: "Ready-to-deploy configuration presets for CodeBuild Project"
type: "preset-list"
componentSlug: "codebuild-project"
componentTitle: "CodeBuild Project"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-github-ci-linux"
    rank: "01"
    title: "GitHub CI (Linux)"
    excerpt: "This preset creates a CI-only CodeBuild project that triggers on pushes and pull requests to the main branch of a GitHub repository. Build status is reported back to GitHub as commit status checks...."
  - slug: "02-docker-build-ecr"
    rank: "02"
    title: "Docker Build with ECR Push"
    excerpt: "This preset creates a CodeBuild project optimized for building Docker images and pushing them to Amazon ECR. Privileged mode enables Docker daemon access inside the build container. Local Docker..."
  - slug: "03-codepipeline-stage"
    rank: "03"
    title: "CodePipeline Build Stage"
    excerpt: "This preset creates a CodeBuild project designed to run as a build or test stage within an AWS CodePipeline. Source code and build artifacts are managed by the pipeline — CodeBuild receives input..."
---

# CodeBuild Project Presets

Ready-to-deploy configuration presets for CodeBuild Project. Each preset is a complete manifest you can copy, customize, and deploy.
