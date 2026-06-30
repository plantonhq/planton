
## Key Features

- **Unified API Model**: The module follows a Kubernetes-style resource model, making it easy for users familiar with Kubernetes to define cloud function specifications.
- **Pulumi Integration**: By leveraging Pulumi, the module can manage Google Cloud resources with state management, making deployments reproducible and trackable.
- **Cloud Function Deployment**: This module supports the creation, configuration, and management of Google Cloud Functions, including defining the function's runtime environment and handling deployment logistics.
- **Google Cloud Provider Setup**: Automatically configures the Google Cloud provider using the credentials specified in the API resource, ensuring a seamless authentication and setup process.
- **Project and Credential Management**: The module supports defining the GCP project in which the Cloud Function will be created and uses the credentials provided in the API resource.
- **Status Management**: Outputs of the deployment process, including function URLs, runtime logs, and other operational details, are captured in `status.outputs`.

## Module Functionality

- **GCP Project Configuration**: The module sets up the target GCP project where the Cloud Function will be created.
- **Cloud Function Creation**: Handles the full lifecycle of a Google Cloud Function, including creation, deployment, and configuration. It also supports updates to existing functions based on changes to the API resource specification.
- **Provider Setup**: Automatically configures the GCP provider using the credentials provided in the `gcp_credential_id`, ensuring that the required permissions and authentication are in place.
- **Stack Outputs**: All outputs from the Pulumi stack, such as Cloud Function URLs or IDs, are automatically captured and added to the `status.outputs` of the resource, making it easy for users to access important details post-deployment.
