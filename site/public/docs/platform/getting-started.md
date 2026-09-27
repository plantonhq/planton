---
title: "Getting Started"
description: "Create your account, set up an organization, connect a cloud provider, and deploy your first resource."
icon: getting-started
order: 30
tags:
  - Getting Started
  - Tutorial
  - Quickstart
---

# Getting Started

This guide walks through the first steps on Planton: creating an account, setting up an organization and environment, connecting a cloud provider, and deploying a cloud resource. By the end, you will have working infrastructure deployed to your own cloud account.

## Prerequisites

- A modern web browser
- An AWS, GCP, or Azure account with permissions to create resources
- Credentials for that account (access keys, service account key, or service principal)

## Step 1: Create Your Account

Navigate to [planton.ai](https://planton.ai) and click **Login** or **Join Beta** in the header. Both lead to the same authentication page.

The authentication page handles both login and registration. Click **Sign up** to switch to registration mode, then choose one of:

- **Email and password** — enter your email and create a password
- **Continue with Google** — authenticate with your Google account

After completing signup, you land on the dashboard.

<!-- SCREENSHOT: Authentication page
  Page: /auth/signup
  Action: Show the signup form with email/password fields and Google option
  Focus: The signup form area
  Alt: Planton signup page showing email and password fields with Continue with Google option
-->

## Step 2: Create Your Organization

An organization is the top-level container for everything in Planton — environments, connections, team members, and billing. It is similar to a Google Cloud organization or an AWS account.

As a new user without an organization, the dashboard prompts you to create one.

1. Click **Create Organization**
2. Enter the organization name (human-readable, e.g., "Acme Corp")
3. Enter the organization slug, e.g., "acme-corp" — this must be unique across the platform. A slug is lowercase letters and digits joined by single hyphens, like my-app-2; an organization's slug also starts with a letter and is 2 to 15 characters
4. Click **Create**

<!-- SCREENSHOT: Create Organization wizard
  Page: /organizations/new/setup
  Action: Show the organization creation form with name and slug fields
  Focus: The form fields
  Alt: Create Organization form with organization name and slug fields
-->

You are now the organization owner. The context selector in the top-left of the header shows your organization name.

## Step 3: Create Your First Environment

Environments are logical groupings that separate your resources — typically corresponding to deployment stages like development, staging, and production. All infrastructure and services deploy into an environment, not directly into an organization.

1. Navigate to **Infra Hub** in the sidebar
2. Click **Create Environment** (or find the environment creation option)
3. Enter a name — start with something like `dev` or `development`
4. Click **Create**

<!-- SCREENSHOT: Create Environment
  Page: /orgs/{org}/environments/create
  Action: Show the environment creation form
  Focus: The environment name field
  Alt: Create Environment form with name field and validation hint
-->

The context selector now shows both your organization and the active environment:

```
Acme Corp / dev
```

**Using the CLI:**

```bash
# Create an environment
planton env create dev --description "Development environment"

# List environments
planton env list

# Set the environment as your default context
planton context set --org acme-corp --env dev
```

## Step 4: Connect a Cloud Provider

Before deploying infrastructure, connect your cloud provider account. This gives Planton the credentials to create and manage resources on your behalf.

1. Click **Connections** in the sidebar
2. Find your cloud provider card (AWS, GCP, Azure, or others) under the **Infrastructure** section
3. Click **Connect** on the provider card
4. Fill in the credentials:
   - **AWS**: Access Key ID and Secret Access Key
   - **GCP**: Upload a service account key JSON file
   - **Azure**: Subscription ID, Tenant ID, Client ID, and Client Secret
5. **Authorize for your environment** — select which environments can use this connection. Check the box next to your `dev` environment.
6. Click **Submit**

<!-- SCREENSHOT: AWS connection form
  Page: /orgs/{org}/connections (AWS connect wizard)
  Action: Show the AWS credential form with environment authorization checkboxes
  Focus: The credential fields and environment authorization section
  Alt: AWS connection form showing access key fields and environment authorization checkboxes
-->

The connection now appears in the **Connected Providers** list and is ready for use.

**Why environment authorization matters:** A connection exists at the organization level, but it must be explicitly authorized for each environment where it can be used. This prevents a production AWS account from being accidentally used during development. See [Environment Mappings](/docs/connections/environment-mappings) for details.

## Step 5: Deploy Your First Cloud Resource

With a cloud provider connected and authorized, you can deploy infrastructure.

1. Open the **Deployment Component Store** — click the store icon in the header (right side)
2. Browse or search for a component (e.g., search "VPC" and filter by your cloud provider)
3. Click on the component to see its details
4. Click **Deploy**
5. Fill in the configuration form with your desired settings
6. Click **Deploy**

<!-- SCREENSHOT: Deployment Component Store
  Page: /platform/deployment-store
  Action: Show the component catalog with provider filter active
  Focus: The component grid with deploy buttons
  Alt: Deployment Component Store showing infrastructure components filterable by cloud provider
-->

A Stack Job is created automatically. Stack Jobs are the execution units that run Pulumi, Terraform, or OpenTofu to provision your infrastructure. You can watch the deployment progress in real-time as each operation (init, refresh, plan, apply) completes.

<!-- SCREENSHOT: Stack Job progress
  Page: /orgs/{org}/cloud-resources/{id} (Stack Jobs tab)
  Action: Show a Stack Job in progress with real-time log output
  Focus: The Stack Job progress panel with operation status
  Alt: Stack Job execution showing real-time progress through init, refresh, plan, and apply stages
-->

Once the Stack Job completes, your cloud resource is live. Navigate to **Infra Hub** in the sidebar and click **Cloud Resources** to see it listed with its current status.

## Step 6: Complete the Onboarding Checklist

The dashboard includes a getting-started checklist that tracks your progress through the initial setup. The eight tasks are organized into four groups:

**Foundation**
- Connect a cloud account
- Create an environment

**Infrastructure**
- Deploy your first cloud resource
- Deploy an Infra Chart stack

**Applications**
- Connect a Git provider (GitHub or GitLab)
- Deploy your first service

**Team and Operations**
- Invite team members
- Set up billing

You can complete these tasks in any order, and dismiss the checklist at any time from the dashboard.

## What to Explore Next

With your first resource deployed, here are the natural next steps:

- **Deploy more resources** — return to the Deployment Component Store and deploy a database, Kubernetes cluster, or storage bucket. See [Infrastructure](/docs/infrastructure).
- **Deploy an Infra Chart** — instead of individual resources, deploy a coordinated set of resources (e.g., VPC + ECS Cluster + ALB) as a single Infra Chart. See [Infra Charts](/docs/infrastructure/infra-charts).
- **Deploy an application** — connect GitHub, create a Service, and push code to trigger an automated build and deployment pipeline. See [CI/CD](/docs/ci-cd).
- **Invite your team** — go to **Settings > Manage Members** to invite colleagues and assign roles. See [Teams and Access](/docs/teams-and-access).
- **Explore the console** — take a [Platform Tour](/docs/platform/platform-tour) to learn what each section of the console does.

## CLI Quick Reference

These are the CLI commands relevant to the steps in this guide:

```bash
# Set your default context (organization and environment)
planton context set --org <org-slug> --env <env-slug>

# View your current context
planton context get

# List environments
planton env list

# Create an environment
planton env create <name> --description "<description>"

# Get environment details
planton env get <name>
```

For CLI installation instructions, see [CLI](/docs/cli).
