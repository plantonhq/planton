# Examples

## Basic Usage

Upload a single configuration file to an S3 bucket:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3ObjectSet
metadata:
  name: app-config
spec:
  bucket:
    value: my-app-bucket
  awsRegion: us-east-1
  objects:
    - key: config/app.json
      content: |
        {"database": "postgres", "port": 5432}
      contentType: application/json
```

## Multiple Objects with Tags

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3ObjectSet
metadata:
  name: website-content
spec:
  bucket:
    valueFrom:
      name: my-s3-bucket
  awsRegion: us-west-2
  tags:
    environment: production
  objects:
    - key: index.html
      content: "<html><body><h1>Hello</h1></body></html>"
      contentType: text/html
      cacheControl: max-age=300
    - key: config/settings.json
      content: '{"theme": "dark"}'
      contentType: application/json
```
