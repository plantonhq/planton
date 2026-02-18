# Create using CLI

Create a YAML file using one of the examples shown below. After the YAML is created, use the command below to apply.

```shell
planton apply -f <yaml-path>
```

# Basic Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: my-aws-secrets
spec:
  region: us-west-2
  secretNames:
    - database-password
    - api-key
```

# Example with Environment Info

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: my-aws-secrets
spec:
  region: us-west-2
  environmentInfo:
    envId: production
  secretNames:
    - database-password
    - api-key
    - encryption-key
```

# Example with Stack Job Settings

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: my-aws-secrets
spec:
  region: us-west-2
  stackUpdateSettings:
    jobTimeout: 3600
  secretNames:
    - database-password
    - api-key
```

# Example with No Secrets Specified

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: my-aws-secrets
spec:
  region: us-west-2
```

# Example Using Created Secrets in a Microservice

The following example shows how to reference secrets created by `AwsSecretsManager` in a `MicroserviceKubernetes` deployment. This assumes that the secrets have been created and are managed by Planton Cloud's `AwsSecretsManager` deployment module.

```yaml
apiVersion: aws.openmcf.org/v1
kind: MicroserviceKubernetes
metadata:
  name: todo-list-api
spec:
  version: main
  container:
    app:
      env:
        secrets:
          # The value before the dot is the ID of the AwsSecretsManager resource on Planton Cloud
          # The value after the dot is the name of the secret as specified in the AwsSecretsManager spec
          DATABASE_PASSWORD: ${my-aws-secrets.database-password}
        variables:
          DATABASE_NAME: todo
      image:
        repo: nginx
        tag: latest
      ports:
        - appProtocol: http
          containerPort: 8080
          isIngressPort: true
          name: rest-api
          networkProtocol: TCP
          servicePort: 80
      resources:
        requests:
          cpu: 100m
          memory: 100Mi
        limits:
          cpu: 2000m
          memory: 2Gi
```

# Notes

These examples demonstrate how to structure your YAML configuration files to create and manage secrets in AWS Secrets Manager using the `AwsSecretsManager` API resource. You can specify one or more secret names under `secretNames` in the `spec` section. If `secretNames` is omitted or left empty, no secrets will be created. Additionally, you can integrate these secrets into other resources like `MicroserviceKubernetes` by referencing them in the configuration.

Remember to replace placeholders like `my-aws-secrets` and secret names with your actual resource IDs and desired secret names.
