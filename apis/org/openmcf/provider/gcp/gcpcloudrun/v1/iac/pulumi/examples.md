Here are a few examples of the `GcpCloudRun` API resource based on the information you provided. These examples demonstrate how developers can configure the `GcpCloudRun` API resource for deploying services on Google Cloud using Planton's standard structure.

### Example 1: Basic Google Cloud Run Deployment

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudRun
metadata:
  name: simple-api
spec:
  gcp_project_id: my-gcp-project
  container:
    app:
      image:
        repo: gcr.io/my-project/my-app
        tag: latest
      ports:
        - appProtocol: http
          containerPort: 8080
          isIngressPort: true
          servicePort: 80
      resources:
        requests:
          cpu: 500m
          memory: 512Mi
        limits:
          cpu: 1000m
          memory: 1Gi
```

### Example 2: Google Cloud Run with Environment Variables

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudRun
metadata:
  name: todo-list-api
spec:
  gcp_project_id: my-gcp-project
  container:
    app:
      env:
        variables:
          DATABASE_NAME: todo
          DATABASE_HOST: localhost
      image:
        repo: gcr.io/my-project/todo-list-api
        tag: v1.0.0
      ports:
        - appProtocol: http
          containerPort: 8080
          isIngressPort: true
          servicePort: 80
      resources:
        requests:
          cpu: 200m
          memory: 256Mi
        limits:
          cpu: 500m
          memory: 512Mi
```

### Example 3: Google Cloud Run with Secrets

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudRun
metadata:
  name: todo-list-api
spec:
  gcp_project_id: my-gcp-project
  container:
    app:
      env:
        secrets:
          DATABASE_PASSWORD: ${gcpsm-my-org-prod-gcp-secrets.database-password}
        variables:
          DATABASE_NAME: todo
      image:
        repo: gcr.io/my-project/todo-list-api
        tag: v1.0.1
      ports:
        - appProtocol: http
          containerPort: 8080
          isIngressPort: true
          servicePort: 80
      resources:
        requests:
          cpu: 300m
          memory: 512Mi
        limits:
          cpu: 1000m
          memory: 1Gi
```

### Example 4: Minimal Google Cloud Run Deployment

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudRun
metadata:
  name: minimal-app
spec:
  gcp_project_id: my-gcp-project
```
