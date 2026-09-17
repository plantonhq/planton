---
title: "Getting Started with Planton"
date: 2025-07-20
author:
  - name: Planton Team
    title: Core Development Team
    bio: "Building the future of DevOps with Planton. We're passionate about making deployment simple and reliable."
    profilePicture: https://api.dicebear.com/7.x/avataaars/svg?seed=PlantonTeam
    github: https://github.com/planton-ai
    website: https://planton.ai
  - name: Sarah Chen
    title: Developer Experience Engineer
    bio: "Focused on making developers' lives easier through intuitive tools and clear documentation."
    profilePicture: https://i.pravatar.cc/150?img=5
    twitter: https://twitter.com/sarahchen_dev
    github: https://github.com/sarahchen
featuredImage: >-
  https://images.unsplash.com/photo-1451187580459-43490279c0fa?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=2072&q=80
featuredImageType: full
tags:
  - Tutorial
  - Getting Started
  - Cloud
excerpt: "Learn how to get started with Planton and deploy your first application in minutes."
---

# Getting Started with Planton

Welcome to Planton! This guide will walk you through the essential steps to get up and running with our platform.

## What is Planton?

Planton turns your own cloud account into a self-service platform. AI designs the infrastructure, verifies the cost and permissions before anything is created, and publishes it as templates your whole team can deploy. Your services then ship onto that infrastructure straight from Git. Whether you're a developer, DevOps engineer, or operations team member, Planton provides the tools you need to ship faster and more reliably.

## Prerequisites

Before you begin, make sure you have:

- A GitHub, GitLab, or Bitbucket account
- Docker installed on your local machine (optional)
- Basic knowledge of Git and containerization

## Step 1: Sign Up

1. Visit [planton.ai](https://planton.ai)
2. Click "Sign Up" and choose your Git provider
3. Authorize Planton to access your repositories
4. Complete your profile setup

## Step 2: Create Your First Project

Once you're signed in:

1. Click "New Project" in the dashboard
2. Choose a repository from your Git account
3. Select your preferred deployment region
4. Configure your project settings

## Step 3: Configure Your Application

Planton supports multiple deployment strategies:

### Docker Deployment
```dockerfile
# planton.yaml
version: '1.0'
services:
  app:
    image: your-app:latest
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
```

### Kubernetes Deployment
```yaml
# planton.yaml
version: '1.0'
kubernetes:
  deployment:
    replicas: 3
    resources:
      requests:
        memory: "256Mi"
        cpu: "250m"
      limits:
        memory: "512Mi"
        cpu: "500m"
```

## Step 4: Deploy

1. Push your configuration to your repository
2. Planton will automatically detect changes
3. Monitor the deployment in real-time
4. Your app will be available at the provided URL

## Best Practices

- **Environment Variables**: Use environment variables for configuration
- **Health Checks**: Implement proper health check endpoints
- **Monitoring**: Set up alerts and monitoring from the start
- **Backups**: Configure automated backups for your data

## Next Steps

Now that you have your first application running:

- Explore the [documentation](/docs)
- Join our [Discord community](https://discord.gg/planton)
- Check out our [tutorials](/tutorials)
- Learn about [advanced features](/product)

## Need Help?

If you run into any issues:

- Read the [documentation](/docs)
- Reach out to our support team
- Ask questions in our community forums

---

*Ready to deploy? [Get started now](https://planton.ai)!* 