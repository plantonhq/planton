# OciDevopsProject Examples

## Minimal Project

A DevOps project with direct OCID values for quick setup:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDevopsProject
metadata:
  name: my-project
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciDevopsProject.my-project
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  notificationTopicId:
    value: "ocid1.onstopic.oc1..example"
```

## Project with Compartment Reference

A project using `valueFrom` to reference an OciCompartment, enabling infra-chart composability:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDevopsProject
metadata:
  name: platform-cicd
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDevopsProject.platform-cicd
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: cicd-compartment
      fieldPath: status.outputs.compartmentId
  notificationTopicId:
    value: "ocid1.onstopic.oc1..example"
  description: "Platform team CI/CD pipelines for production workloads"
```

## Production Project with Description

A fully described production project for a backend services team:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDevopsProject
metadata:
  name: backend-services
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: backend
    pulumi.openmcf.org/stack.name: prod.OciDevopsProject.backend-services
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  notificationTopicId:
    value: "ocid1.onstopic.oc1..example"
  description: "Backend microservices build and deployment pipelines"
```

## Common Operations

### Update the notification topic

Change `notificationTopicId` to a different ONS topic OCID and re-apply. The project updates in place without recreation.

### Move to a different compartment

Update `compartmentId` to a new compartment OCID and re-apply. OCI supports in-place compartment moves for DevOps projects.

### Use the namespace output

After deployment, use the `namespace` stack output in container registry paths:

```
{region}.ocir.io/{namespace}/{repo-name}:{tag}
```

## Best Practices

1. **One project per team or service domain** — keeps pipeline scope manageable and aligns notification routing with team ownership.
2. **Use `valueFrom` references** for `compartmentId` — avoids hardcoding OCIDs and maintains dependency ordering in infra charts.
3. **Create the ONS topic first** — the notification topic must exist before the DevOps project can reference it.
4. **Add a description** — helps teams identify the project's purpose in the OCI Console and API responses.
