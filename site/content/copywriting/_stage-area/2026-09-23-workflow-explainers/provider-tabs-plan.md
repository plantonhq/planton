# Scope update

The research below is historical. See handoff.md for the approved five-provider implementation, automatic rotation, explicit connector routes, and dark icon treatment.

---

# Customer Architecture Tabs — Research and Implementation Plan

Status: researched proposal, 2026-09-23. No provider-tab or connector code changed in this planning pass.

## Decision

Use four manually selected tabs: AWS, GCP, Azure, Cloudflare. Each presents a different customer application and deployment model. Replace the internal hosted-cluster story as the landing-page default; retain its research as history. Replace the earlier four-resource AWS example with a Bedrock application while preserving its visual clarity.

Keep the current SVG renderer, deterministic playback, dark palette, accessible controls, and separate Remotion export adapter. No new animation framework is necessary. Refactor the infrastructure-specific data/rendering assumptions into explicit architecture stories; do not build a general-purpose graph editor.

Proposed section promise: **Your Stack. Your Cloud. One Deployment Workflow.**

Supporting copy: “Connect managed services, Kubernetes workloads, and edge applications. Planton deploys the resources in dependency order and gives you a visual record of the architecture.” Each tab is a separate example, not a claim that one manifest is portable unchanged across providers.

## Four Customer Stories

### AWS — An AI Assistant for Your Product Docs

Show seven customer-facing cards:

- **Product Documents** — Amazon S3
- **Vector Search** — S3 Vectors
- **Knowledge Base** — Amazon Bedrock
- **Foundation Model** — Bedrock Inference Profile
- **Response Guardrails** — Amazon Bedrock
- **Assistant API** — AWS Lambda
- **HTTPS Endpoint** — API Gateway

The principal verified kinds are AwsS3Bucket, AwsS3VectorBucket, AwsBedrockKnowledgeBase, AwsBedrockInferenceProfile, AwsBedrockGuardrail, AwsLambda, and AwsHttpApiGateway. Supporting execution roles, artifact storage, logs, authorization, and permissions belong in the expanded resource view rather than extra headline boxes.

Deployment narrative: prepare document/vector stores and access; connect the knowledge base; configure the model and guardrail; deploy application code with its configuration; expose the API. Independent resources remain independent. The knowledge base references the S3 source and vector bucket; the HTTP API references the Lambda function. Application environment strings are not automatically typed output references: any drawn ordering must be backed by actual chart configuration/dependency mechanisms. Do not depict model requests as deployment dependencies.

This demonstrates managed AI and serverless application infrastructure without a customer-managed Kubernetes cluster. Deploying the knowledge base does not upload customer documents or run ingestion. The assistant code must invoke the model, retrieval, and guardrail; creating those resources does not automatically connect application behavior.

Choose Lambda plus Bedrock for this first story. Bedrock Agents Classic is closed to new customers as of July 30, 2026. AgentCore Runtime is a catalog candidate, but a current AWS runtime requirement needs reconciliation with the local schema/module before using it in a confidently deployable example. Neither issue needs to expand this website task into catalog changes.

### GCP — Your Own AI Inference Stack on GKE

Show seven cards, with a visible boundary between Google Cloud infrastructure and workloads inside Kubernetes:

- **Cloud Network** — VPC + Subnet
- **GKE Cluster** — Kubernetes + General-Purpose Nodes
- **GPU Node Pool** — GPU Capacity
- **Model Server** — vLLM
- **Vector Search** — Qdrant
- **AI Application** — Your Container
- **HTTPS Endpoint** — Gateway + TLS

Use GcpVpcNetwork, GcpSubnetwork, GcpGkeCluster, GcpGkeNodePool, KubernetesQdrant, KubernetesDeployment, and the actual Gateway/TLS/route components selected during blueprint authoring. General-purpose and GPU pools are separate; the vector database and API should not imply a GPU requirement.

The catalog has no dedicated vLLM or KServe component. KubernetesDeployment's typed resource surface does not establish a GPU-limit field in this inspection. For vLLM, use the supported KubernetesManifest path with a pinned upstream GPU deployment, explicitly documented as a custom workload. Do not invent a native vLLM component. Namespace, GPU scheduling, model access, storage, readiness, and the cluster connection must be specified in the underlying blueprint. Stable workload names are not exported typed references from a raw manifest.

Deployment narrative: network → cluster → separate CPU/GPU capacity; install the model server on GPU capacity and Qdrant on CPU capacity; configure the application and publish its HTTPS route. Supporting IAM, cluster access, namespaces, certificate controllers, and DNS remain inspectable below.

This demonstrates both cloud provisioning and actual workloads inside Kubernetes. Model weights/licensing, GPU quota, and application code remain required inputs. Provisioning GPU capacity is not evidence that a model has loaded or is serving requests.

### Azure — An Order-Processing App With Background Jobs

Show six cards:

- **App Environment** — Azure Container Apps
- **Orders Database** — Azure Database for PostgreSQL
- **Job Queue** — Azure Service Bus
- **Secrets & Access** — Key Vault + Managed Identity
- **Orders API** — Container App
- **Background Jobs** — Container Apps Jobs

Verified kinds include AzureContainerAppEnvironment, AzureContainerApp, AzureContainerAppJob, AzurePostgresqlFlexibleServer, AzureServiceBusNamespace, AzureServiceBusQueue, AzureKeyVault, AzureKeyVaultSecret, and AzureUserAssignedIdentity. Supporting resource group, image registry, logging, role assignments, and any private-network/DNS resources belong in the blueprint and expanded view.

Deployment narrative: establish environment, database, queue, and identity; configure application access; deploy API and jobs once their own prerequisites exist. The jobs component supports an event trigger using the azure-servicebus scaler. Queue delivery at runtime is distinct from the deployment order. Database migrations, producer/consumer code, and authorization grants must be explicit rather than implied by the existence of a database and queue.

This demonstrates a familiar business application using managed containers, database, and messaging, without asking the customer to operate a Kubernetes cluster.

### Cloudflare — A File-Sharing App at the Edge

Show six cards:

- **File Storage** — R2
- **Application Data** — D1
- **App Configuration** — Workers KV
- **Background Worker** — Cloudflare Workers
- **Job Queue** — Cloudflare Queues
- **Edge API** — Cloudflare Workers

Use CloudflareR2Bucket, CloudflareD1Database, CloudflareKvNamespace, CloudflareWorker, and CloudflareQueue, plus DNS zone/custom-domain configuration when publishing the endpoint. The two Worker cards are separate resource instances with different jobs.

Deployment narrative: create stores; bind them to the background consumer; create the queue with that consumer; bind the queue and stores to the producer/API Worker; publish the configured domain. This is intentionally different from runtime flow (API → queue → consumer). Do not create a producer/consumer reference cycle: the consumer does not bind the queue that depends on it. Avoid R2 notification dependencies in this first composition.

This demonstrates edge compute with durable data and asynchronous work. Use small metadata/notification tasks for background work, not unqualified heavy media processing. Database migrations and Worker application code remain explicit. KV is configuration storage, not a substitute for transactional application data.

## Visual and Interaction Contract

- Every card title uses authored Title Case, preserving names such as vLLM, PostgreSQL, and Cloudflare. Customer purpose is the primary label; service/product names are secondary.
- Keep six or seven overview cards per story. Supporting resources appear in an expandable, provider-specific inventory and focused dependency view. Do not carry the old “22 resources” count into unrelated stories.
- Four visible tabs, with one active animation. No automatic tab rotation. Remember pause/reduced-motion preferences; stop hidden playback. Reserve a stable desktop height to avoid page movement on tab change.
- Keyboard-accessible tabs with visible focus and correct tab/panel semantics. Without JavaScript, provide all four stories as readable static sections/disclosures. Reduced motion renders completed diagrams plus the narrative.
- Blue moving dots always mean dependency handoffs in the infrastructure stories. The surrounding copy explicitly says deployment order. Add no runtime traffic animation in this iteration.
- An optional “Explore This Stack” disclosure identifies what each card represents, supporting resources, and the meaning of its edges. This is an illustrative architecture until a deployable blueprint has been validated; do not offer a misleading “Deploy This Stack” button.
- Keep the existing living-architecture message, tied to dependencies, deployment status, and resource inspection. No claim of automatically discovering arbitrary cloud resources or continuously detecting drift.
- Labels sit outside connector routes. The current fan-out labels overlap the moving paths; remove that collision in the provider layouts.

## Connector Correction: Delivery and Coding Agents

The current desktop scene uses 170-unit cards at 198-unit intervals. After the arrow gap, only about 23 units remain for a connector whose vertical displacement is 133 units. The diagonal control points therefore make the line look almost straight and attach awkwardly. Changing arrowheads again will not solve the spacing problem.

Recompose the six stages into a compact, clearly numbered, two-row layout with three stages per row. Use a deliberate folded path with an outside return connector; give local handoffs broad lanes and modest vertical offsets. Verify reading order visually rather than assuming numbering compensates for a confusing route. Cards retain readable type and the entire closed component fits the target desktop viewport.

Routes leave the source edge perpendicular to it, follow a continuous cubic curve, and arrive perpendicular to the target edge. The outside return uses rounded, tangent-continuous segments. Arrow tips align with the final tangent and stop at a consistent gap. Preserve thin strokes, small filled heads, and evenly spaced blue packets. Mobile uses an intentionally authored vertical route.

Share geometry/ports and arc-length sampling across browser and video. If the outside connector needs multiple curve segments, extend the existing small geometry helper with a sampled compound route; do not introduce a graph-routing library.

## Code Structure

- Architecture data: four provider story records with separate public copy, nodes, logical relationships, grouped-resource mappings, and evidence notes. Keep layout coordinates separate from relationships.
- Preserve one reusable playback/control shell and the existing pure scene/export contract. Replace the hard-coded infrastructure story lookup and HOSTED_GROUPS coupling with an explicit selected architecture.
- Generalize HostedResourceGraph only into a resource explorer that accepts data. Its current interaction remains useful; its Planton-internal resource list must not leak into other tabs.
- Keep provider-specific compositions intentional and small. Shared nodes, connectors, and controls do not require every architecture to have the same topology.
- Extend IDs/analytics/export selection with stable provider story IDs. Scope SVG marker IDs to story/instance to avoid collisions between active panels, static alternatives, and exports.
- Lazy playback for the selected, visible story; no cloud SDKs or Remotion runtime in the public browser bundle. Source SVG/TS stays in Git. Generated videos stay outside Git unless separately published.

## Implementation Order and Acceptance

1. Write the four resource/dependency records and source-backed copy. Label each edge as a typed output reference, explicit deployment prerequisite, grouping summary, or application configuration relationship. Reject unsupported automatic-wiring claims.
2. Author static desktop/mobile layouts and correct delivery/agent connectors. Inspect all labels, arrowheads, return curves, and card reading order before animation.
3. Add tabs and connect the existing deterministic playback. Keep the user-requested dark presentation and controls.
4. Replace the internal-cluster disclosure with each story's own resource explorer and text inventory. Update machine-readable page content and evidence/handoff documentation.
5. Validate schema-backed blueprint examples offline where feasible. Confirm components exist in the intended released catalog before calling a story deployable. A catalog enum, a module directory, or a successful website build is not an end-to-end cloud deployment test. No cloud resources are provisioned for this website work.
6. Run site build and targeted browser tests for tab keyboard behavior, hidden playback, reduced motion, no-JS content, graph focus, deterministic phases, and viewport fit at 1366×768, 1440×900, and 1920×1080. Inspect mobile at 320/390 px and tablet separately; never shrink text just to force a fit.
7. Export each provider story and the revised sequence scenes through the same renderer. Inspect start, transfer, approval, and completed frames in both browser and MP4; verify export composition is legible at social-media sizes, not merely a scaled desktop screenshot.
8. Refresh the local preview for the ongoing visual review. The earlier authorization to create/merge the eventual PR remains, but this turn requests a plan; no implementation or merge occurs in this planning pass.

## Research Evidence

Catalog root inspected: `catalog/`. All component reference pages used for schema facts came from this checkout; the installed skill pack was not mixed into these facts. Read the shared commons, provider indexes, relevant generated reference pages, and the guides for knowledge bases, agents, AgentCore, node pools, workloads, Helm, and queues. Confirmed Terraform module presence for the main selected kinds. No live deployment claim is made.

Registry: `shared/cloudresourcekind/cloud_resource_kind.proto` in that checkout.

Primary references within `catalog/`:

- AWS: `aws/awsbedrockknowledgebase/v1alpha1/reference.md`, `aws/awss3vectorbucket/v1alpha1/reference.md`, `aws/awsbedrockguardrail/v1alpha1/reference.md`, `aws/awsbedrockinferenceprofile/v1alpha1/reference.md`, `aws/awslambda/v1alpha1/reference.md`, `aws/awshttpapigateway/v1alpha1/reference.md`.
- GCP/Kubernetes: `gcp/gcpgkecluster/v1alpha1/reference.md`, `gcp/gcpgkenodepool/v1alpha1/reference.md`, `kubernetes/kubernetesqdrant/v1alpha1/reference.md`, `kubernetes/kubernetesdeployment/v1alpha1/reference.md`, `kubernetes/kubernetesmanifest/v1alpha1/reference.md`.
- Azure: `azure/azurecontainerappenvironment/v1alpha1/reference.md`, `azure/azurecontainerapp/v1alpha1/reference.md`, `azure/azurecontainerappjob/v1alpha1/reference.md`, `azure/azureservicebusqueue/v1alpha1/reference.md`, `azure/azurepostgresqlflexibleserver/v1alpha1/reference.md`.
- Cloudflare: `cloudflare/cloudflareworker/v1alpha1/reference.md`, `cloudflare/cloudflarequeue/v1alpha1/reference.md`, `cloudflare/cloudflared1database/v1alpha1/reference.md`, `cloudflare/cloudflarer2bucket/v1alpha1/reference.md`, `cloudflare/cloudflarekvnamespace/v1alpha1/reference.md`.

Current provider documentation checked:

- [Bedrock ingestion is a separate operation](https://docs.aws.amazon.com/bedrock/latest/userguide/kb-data-source-sync-ingest.html).
- [Bedrock Knowledge Bases with S3 Vectors](https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-vectors-bedrock-kb.html).
- [Bedrock Agents Classic maintenance mode](https://docs.aws.amazon.com/bedrock/latest/userguide/agents-classic-maintenance-mode.html).
- [AgentCore runtime requirements](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/runtime-troubleshooting.html): the MMDSv2 requirement does not have a matching field in the inspected local runtime reference/module. Keep this alternative outside the first AWS story pending verification.
- [Serving models on GKE with vLLM](https://docs.cloud.google.com/kubernetes-engine/docs/tutorials/serve-llama-gpus-vllm).
- [Azure Container Apps jobs](https://learn.microsoft.com/en-us/azure/container-apps/jobs).
- [Cloudflare Worker bindings](https://developers.cloudflare.com/workers/runtime-apis/bindings/).

Editorial visitor check: each tab answers “could I build my application here?” before introducing infrastructure vocabulary. Editorial claim check: label purposes do not turn into assertions that customer code, data ingestion, auth, or model preparation are automatically supplied by provisioning. These are author checks, not an independent design review.
