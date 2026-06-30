# CronJob Kubernetes Pulumi Module

## Key Features

- **Standardized API Resource Model**  
  Uses a concise, consistent definition to manage scheduled tasks on Kubernetes. By describing key parameters (schedule,
  concurrency policy, backoff limits, container details, etc.), you ensure a uniform workflow for running CronJobs
  across all environments.

- **Automated Kubernetes Resource Creation**  
  Automatically creates Namespaces, CronJobs, and (if needed) image pull secrets for private registries, based on the
  provided specifications. Simplifies repetitive tasks and configuration sprawl.

- **Configurable Schedules and Policies**  
  Enables fine-grained control over concurrency (e.g., Forbid, Replace), automatic retries (backoff limits), and
  scheduling (standard cron expressions like “0 0 * * *”). Supports optional features such as `suspend` and
  `startingDeadlineSeconds` to control missed or delayed jobs.

- **Container & Resource Management**  
  Provides robust container configurations: you can specify the image, CPU/memory resource requests and limits,
  environment variables, and secrets. Ensures your scheduled job pods have clear resource boundaries and consistent
  environment setups.

- **Secrets and Environment Variables**  
  Inject sensitive values from external secret stores (or raw definitions) into your CronJob. By mapping secrets to
  environment variables, you keep credentials out of code and container images.

- **Optional Image Pull Secrets**  
  Seamlessly integrates Docker registry credentials if your images are stored in private repositories, creating
  `kubernetes.io/dockerconfigjson` secrets and referencing them in your CronJob.

- **Flexible Namespace Management**  
  Control namespace creation with the `create_namespace` boolean flag. Set to `true` to automatically create and manage
  a dedicated namespace with appropriate labels. Set to `false` to reference an existing namespace, enabling shared
  namespace patterns for multi-tenant scenarios or GitOps workflows where namespace lifecycle is managed separately.

- **Output Exports**  
  Provides essential outputs (e.g., namespace name). You can further integrate these into other stages of your pipeline
  if needed.

## Usage

Refer to the [example](example.md) for usage instructions and sample YAML definitions. In short:

1. Define a YAML file containing a `CronJobKubernetes` resource.
2. Use:
   ```bash
   planton pulumi up --stack-input <your-cronjob-file.yaml>
   ```

to deploy your CronJob to the specified Kubernetes cluster.

## Getting Started

1. **Prepare the Specification**  
   Create a YAML that follows the `CronJobKubernetes` API model, including your desired container image, schedule,
   concurrency settings, and any secrets or environment variables.

2. **Run the CLI**  
   Execute `planton pulumi up --stack-input <cronjob-spec.yaml>` (or whatever command you typically use) to apply the
   resource on your cluster.

3. **Observe Your Jobs**  
   Your Kubernetes CronJob will now schedule pods according to the defined cron expression. Monitor job statuses, logs,
   and pod behavior using standard Kubernetes tooling (`kubectl`, or relevant pipeline logs).

## Module Structure

1. **Initialization**  
   Reads the `CronJobKubernetesStackInput` fields (cluster credentials, Docker config, specification metadata, etc.) and
   sets up local references and labels.

2. **Provider Setup**  
   Creates a Pulumi Kubernetes Provider using your specified cluster credentials.

3. **Namespace Creation or Reference**  
   Based on the `create_namespace` flag, either creates a new dedicated Kubernetes Namespace with proper labels or
   references an existing namespace. This groups all relevant CronJob resources appropriately.

4. **Secret Management**  
   Generates a “main” Kubernetes Secret to store environment secrets for your CronJob container. Keeps sensitive data
   out of plain text or container images.

5. **Image Pull Secret (Optional)**  
   If Docker credentials are provided, creates a secret of type `kubernetes.io/dockerconfigjson` and attaches it to the
   CronJob’s Pod specification.

6. **CronJob Configuration**  
   Defines a Kubernetes CronJob with your specified schedule, concurrency policy, backoff limit, container image,
   resource requests/limits, environment variables, and secrets.

- `schedule` – The cron syntax (e.g., `"0 0 * * *"`).
- `suspend` – Pauses job creation when set to true.
- `startingDeadlineSeconds` – Skips past-due jobs older than this threshold.
- `concurrencyPolicy`, `backoffLimit`, and `restartPolicy` – Fine-tune concurrency, retry behavior, and container
  restarts.

7. **Output Exports**  
   Exports details such as the namespace for reference in subsequent automation or pipeline steps.

## Benefits

- **Streamlined Scheduling**  
  Easily configure one-off or recurring tasks—like backups, data sync, or routine maintenance—without hand-writing
  resource manifests.

- **Security and Compliance**  
  Safely handle secrets, ensuring sensitive information remains encrypted at rest in Kubernetes. Optional image pull
  secrets ensure you can pull from private registries securely.

- **Consistency Across Environments**  
  A single resource definition drives scheduled tasks across dev, staging, and production. Eliminates guesswork and
  custom YAML files by leveraging a standard specification.

- **Scalability and Reliability**  
  Resource limits, concurrency controls, and built-in Kubernetes scheduling logic help keep your cluster stable, even
  when multiple CronJobs run simultaneously.

- **Integration-Friendly**  
  Output values, labeling strategies, and environment variable injection allow easy integration with existing CI/CD
  pipelines, monitoring setups, or orchestrations.

## Contributing

Contributions are welcome! Please open issues or pull requests in the main repository.

## License

This module is provided under the [MIT License](LICENSE). Feel free to fork, extend, or adapt it to your operational
needs.
