# AWS AWS EKS Cluster Pulumi Module

## Introduction

The AWS AWS EKS Cluster Pulumi Module provides a standardized and efficient way to define and deploy Amazon Elastic Kubernetes Service (EKS) clusters on AWS using a Kubernetes-like API resource model. By leveraging our unified APIs, developers can specify their EKS cluster configurations in simple YAML files, which the module then uses to create and manage AWS EKS resources through Pulumi. This approach abstracts the complexity of AWS interactions and streamlines the deployment process, enabling consistent infrastructure management across multi-cloud environments.

## Key Features

- **Kubernetes-Like API Resource Model**: Utilizes a familiar structure with `apiVersion`, `kind`, `metadata`, `spec`, and `status`, making it intuitive for developers accustomed to Kubernetes to define AWS EKS resources.

- **Unified API Structure**: Ensures consistency across different resources and cloud providers by adhering to a standardized API resource model.

- **Pulumi Integration**: Employs Pulumi for infrastructure provisioning, enabling the use of real programming languages and providing robust state management and automation capabilities.

- **Comprehensive AWS EKS Cluster Configuration**: Supports detailed specification of EKS cluster attributes, including AWS region, VPC integration, and worker node management mode.

- **Region Specification**: Allows the cluster to be deployed in any valid AWS region by specifying the `region` field in the `spec`. This provides flexibility in choosing the geographical location of the cluster.

- **VPC Integration**: Offers the option to deploy the EKS cluster into an existing VPC by specifying the `vpcId`, or to create a new VPC if not specified. This ensures seamless integration with existing network infrastructures.

- **Credential Management**: Securely handles AWS credentials via the `awsProviderConfigId` field, ensuring authenticated and authorized resource deployments without exposing sensitive information.

- **Status Reporting**: Captures and stores outputs such as the cluster endpoint, certificate authority data, and VPC ID in `status.outputs`. This facilitates easy reference and integration with other systems, such as Kubernetes clients or additional automation tools.

- **Scalability and High Availability**: Enables the creation of highly available clusters by deploying worker nodes across multiple availability zones.

## Architecture

The module operates by accepting an AWS AWS EKS Cluster API resource definition as input. It interprets the resource definition and uses Pulumi to interact with AWS, creating the specified EKS resources. The main components involved are:

- **API Resource Definition**: A YAML file that includes all necessary information to define an EKS cluster, following the standard API structure. Developers specify the cluster's desired state in this file, including the AWS region, VPC ID, and worker node management mode.

- **Pulumi Module**: Written in Go, the module reads the API resource and uses Pulumi's AWS SDK to provision EKS resources based on the provided specifications. It abstracts the complexity of resource creation, update, and deletion.

- **AWS Provider Initialization**: The module initializes the AWS provider within Pulumi using the credentials specified by `awsProviderConfigId`. This ensures that all AWS resource operations are authenticated and authorized.

- **Resource Creation**: Provisions the EKS cluster and associated resources as defined in the `spec`, including VPC, subnets, security groups, IAM roles, and worker nodes. If a VPC ID is provided, the cluster is deployed into that VPC; otherwise, a new VPC is created.

- **Worker Node Management**: Depending on the `workersManagementMode` specified, the module sets up worker nodes using different strategies:

  - **Self-Managed Nodes**: Nodes are managed by the user, providing full control over the EC2 instances.

  - **Managed Node Groups**: AWS manages the worker nodes, simplifying the provisioning and lifecycle management.

  - **Fargate Profiles**: Serverless compute for containers, allowing you to run pods without managing servers.

- **Status Outputs**: Outputs from the Pulumi deployment, such as the cluster endpoint, certificate authority data, and VPC ID, are captured and stored in `status.outputs`. This information is crucial for connecting to the cluster and deploying applications.


## Inputs

- **vpcId**: The VPC ID where the EKS cluster will be deployed.
- **subnetIds**: A list of subnet IDs for the EKS cluster.
- **clusterRoleArn**: (Optional) The ARN of the IAM role for the EKS cluster.
- **nodeGroupRoleArn**: (Optional) The ARN of the IAM role for the node group.
- **securityGroupIds**: A list of security group IDs for the EKS cluster.
- **clusterName**: (Optional) The name of the EKS cluster. Defaults to `aws-eks-cluster`.
- **tags**: (Optional) Tags to apply to resources.

## Outputs

- **clusterName**: The name of the EKS cluster.
- **clusterEndpoint**: The endpoint of the EKS cluster.
- **clusterArn**: The ARN of the EKS cluster.

## Usage

- Refer to `example.md` for an example of how to use this module.

## Limitations

- **Advanced Features**: Certain advanced features of EKS, such as custom networking configurations, advanced IAM roles, or specific add-ons, that are not specified in the current API resource definition may not be supported. Future updates may include additional capabilities based on user needs.

- **Region and VPC Changes**: Updating the `region` or `vpcId` fields in the `spec` may result in the recreation of the EKS cluster, as these are critical properties that affect the cluster's deployment.

## Local Development
### Note: 
* This process will help to iterate locally by making changes and testing without needing to push updates externally.
* Set the [AWS credentials](https://docs.aws.amazon.com/cli/v1/userguide/cli-configure-files.html) using aws cli.
* Setup [Pulumi](https://www.pulumi.com/docs/iac/concepts/state-and-backends/)

### To update the spec of the EKS cluster:
* **Navigate to the Protobuf File**:- To update the EKS cluster specification, navigate to `project-platon/api/project/platon/provider/aws/aws-eks-cluster/vi/spec.proto`.
* **Modify the Protobuf Definition**:
  - Example: Add a new field `repeated string subnets = 4;` to include an additional subnet.
   ```go
   message EKSClusterSpec {
       // Existing fields...
       repeated string subnets = 4;
   }
   ```
### Build APIs Locally and release a local cli version
* Build the APIs and update the local cli
  - From the root of the `planton` repository, run:
    ```bash
    make build-APIs & make local
    ```
  - This command invokes the protobuf compiler, generating updated Golang stubs based on your local changes and cli utilizes the locally generated stubs, reflecting your recent protobuf modifications.

### Point the `aws-eks-cluster-pulumi-module` to utilize the local changes this helps us to iterate locally.
* Update the `go.mod` file in the `aws-eks-cluster-pulumi-module` directory to point to the local `planton` repository.
  ```go
  replace github.com/plantonhq/planton => ../planton
  ```
* Run the following command to update the dependencies:
  ```bash
    make build
    ```

### Test the features locally
* Sample Manifest
```yaml
apiVersion: aws.planton.dev/v1
kind: AwsEksCluster
metadata:
  name: planton-test-aws-eks-cluster
spec:
  region: us-east-1
  securityGroups:
  - "sg-502a052d"
  subnets:
  - "subnet-******"
  - "subnet-******"
  - "subnet-******"
  roleArn: "arn:aws:iam::******:role/EKSClusterRole" # optional
  nodeRoleArn: "arn:aws:iam::******:role/EKSNodeRole" # optional
  instanceType: "t3.medium"
  desiredSize: 2 # Number of worker nodes
  maxSize: 10 # Maximum number of worker nodes
  minSize: 2 # Minimum number of worker nodes
  tags: # optional
    Name: "planton-test-aws-eks-cluster"
    Environment: "test"
    Owner: "planton"
    Team: "planton"

```
* Run command
```bash
planton pulumi up --manifest <manifest_path>/eks.yaml  --stack <stack_path> --module-dir <path>/aws-eks-cluster-pulumi-module
```
* Output
```bash
Enter your passphrase to unlock config/secrets
    (set PULUMI_CONFIG_PASSPHRASE or PULUMI_CONFIG_PASSPHRASE_FILE to remember):
Enter your passphrase to unlock config/secrets
Previewing update (eks):
     Type                     Name                 Plan
 +   pulumi:pulumi:Stack      handson-eks          create
 +   ├─ pulumi:providers:aws  classic-provider     create
 +   ├─ aws:eks:Cluster       awsEksCluster           create
 +   └─ aws:eks:NodeGroup     eksManagedNodeGroup  create

Outputs:
    awsEksClusterName  : "planton-test-aws-eks-cluster"
    eksNodeGroupName: "planton-test-aws-eks-cluster-node-group"

Resources:
    + 4 to create

Do you want to perform this update? yes
Updating (eks):
     Type                     Name                 Status
 +   pulumi:pulumi:Stack      handson-eks          created (553s)
 +   ├─ pulumi:providers:aws  classic-provider     created (0.00s)
 +   ├─ aws:eks:Cluster       awsEksCluster           created (413s)
 +   └─ aws:eks:NodeGroup     eksManagedNodeGroup  created (137s)

Outputs:
    awsEksClusterName  : "planton-test-aws-eks-cluster"
    eksNodeGroupName: "planton-test-aws-eks-cluster-node-group"

Resources:
    + 4 created

Duration: 9m16s
```
* Set aws aws eks cluster config locally
```bash
aws eks update-kubeconfig --name <cluster_name> --region us-east-1
```

## Contributing

We welcome contributions to enhance the functionality of this module. Please submit pull requests or open issues to help improve the module and its documentation.
