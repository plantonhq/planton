# Overview

The **GCP Cloud Function API Resource** provides a consistent and standardized interface for deploying and managing Google Cloud Functions (Gen 2) within our infrastructure. This resource simplifies the process of running serverless code on Google Cloud Platform (GCP), allowing users to build and deploy functions without the overhead of managing servers or runtime environments.

Gen 2 Cloud Functions are built on Cloud Run and Eventarc, providing superior performance, scalability, and event handling capabilities compared to Gen 1.

## Purpose

We developed this API resource to streamline the deployment and management of serverless functions using GCP Cloud Functions Gen 2. By offering a unified interface, it reduces the complexity involved in setting up event-driven or HTTP-triggered functions, enabling users to:

- **Easily Deploy Cloud Functions**: Quickly create and deploy functions in specified GCP projects and regions.
- **Simplify Configuration**: Abstract the complexities of setting up GCP Cloud Functions, including environment settings, secrets, VPC connectivity, and permissions.
- **Integrate Seamlessly**: Utilize existing GCP credentials and integrate with other GCP services.
- **Focus on Code**: Allow developers to concentrate on writing code rather than managing infrastructure.
- **Production-Ready**: Support for all essential production features including VPC, secrets, scaling, and authentication.

## Key Features

### Core Capabilities

- **Consistent Interface**: Aligns with our existing APIs for deploying open-source software, microservices, and cloud infrastructure.
- **Simplified Deployment**: Automates the provisioning of Cloud Functions, including setting up necessary permissions and environment variables.
- **Scalability**: Leverages GCP's serverless infrastructure (Cloud Run) to automatically scale functions based on demand with support for up to 1,000 concurrent requests per instance.
- **Flexible Configuration**: Supports comprehensive configuration including runtime, entry point, source location, triggers, and service settings.

### Runtime & Build Configuration

- **Multiple Runtime Support**: Python (3.10-3.13), Node.js (20, 22), Go (1.21-1.23), Java (17, 21), .NET (6, 8), Ruby (3.2-3.3), PHP (8.2-8.3)
- **Source Code from GCS**: Deploy from ZIP files stored in Google Cloud Storage buckets
- **Build Environment Variables**: Configure buildpack behavior with custom build-time environment variables

### Trigger Types

- **HTTP Triggers**: Create HTTPS endpoints for webhooks, REST APIs, and microservices
- **Event-Driven Triggers**: Respond to 125+ event sources via Eventarc including:
  - Cloud Pub/Sub messages
  - Cloud Storage object changes (create, delete, archive)
  - Cloud Firestore document changes
  - Custom events from third-party services

### Security & Networking

- **VPC Connectivity**: Access private resources (Cloud SQL, Memorystore, internal APIs) via Serverless VPC Access connectors
- **Secret Manager Integration**: Securely inject secrets as environment variables without exposing them in code
- **Service Account Identity**: Run functions with dedicated, least-privilege service accounts
- **Ingress Controls**: Configure private functions (internal-only) or public functions with authentication
- **Authentication**: Control who can invoke functions (public, private, or specific identities)

### Performance & Scaling

- **Memory Allocation**: Configure from 128MB to 32GB (with proportional CPU allocation)
- **Timeout Configuration**: Set execution timeouts up to 60 minutes for HTTP functions
- **Instance Concurrency**: Handle up to 1,000 concurrent requests per instance
- **Min/Max Instances**: Eliminate cold starts with warm instances or control costs with max instances
- **Auto-scaling**: Automatically scale from zero to thousands of instances based on demand

### Observability

- **Cloud Logging**: Automatic stdout/stderr capture
- **Cloud Monitoring**: Built-in metrics (invocation count, execution time, memory usage, instance count)
- **Error Reporting**: Unhandled exceptions automatically grouped and surfaced
- **Cloud Trace**: Distributed tracing for identifying latency bottlenecks

## Use Cases

### Production Workloads

- **REST APIs**: Build production-grade HTTP APIs with authentication, secrets, and VPC connectivity
- **Webhook Handlers**: Process webhooks from Stripe, Twilio, GitHub, and other services
- **Event Processing**: Respond to Cloud Storage uploads (image thumbnailing, video transcoding), Pub/Sub messages (async job processing), or Firestore changes (data validation)
- **Serverless Backends**: Mobile app backends built on Firebase with server-side logic
- **Integration Glue**: Connect Google Cloud services—trigger workflows when files land in Storage, publish to Pub/Sub when Firestore documents change

### Development & Testing

- **Rapid Prototyping**: Quickly deploy and iterate on function logic without managing infrastructure
- **Development Environments**: Create isolated dev/staging environments with different configurations
- **Testing**: Test event-driven workflows and HTTP endpoints

## Implementation Status

This component is **production-ready** with comprehensive support for:

✅ **Complete Protobuf API**: All essential fields for build config, service config, triggers, and outputs  
✅ **Full Pulumi Support**: Complete Pulumi module with all files (main.go, locals.go, resources.go, function.go, outputs.go)  
✅ **Full Terraform Support**: Complete Terraform module (main.tf, locals.tf, outputs.tf, provider.tf, variables.tf)  
✅ **Comprehensive Tests**: 12 passing unit tests covering valid and invalid configurations  
✅ **Complete Documentation**: Research doc, README, examples, and supporting docs  
✅ **Production Examples**: HTTP functions, event-driven functions, private functions, functions with VPC and secrets
