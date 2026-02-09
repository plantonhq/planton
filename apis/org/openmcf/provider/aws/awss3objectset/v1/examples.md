# Create using CLI

Create a YAML file using one of the examples shown below. After the YAML file is created, use the command below to apply the configuration:

```shell
planton apply -f <yaml-path>
```

# Basic Example

Upload a single JSON configuration file to an S3 bucket.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3ObjectSet
metadata:
  name: app-config-objects
spec:
  bucket:
    value: my-app-bucket
  awsRegion: us-east-1
  objects:
    - key: config/app.json
      content: |
        {
          "database": "postgres",
          "port": 5432,
          "debug": false
        }
      contentType: application/json
```

# Example with Bucket Reference

Reference an AwsS3Bucket component instead of providing a literal bucket name.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3ObjectSet
metadata:
  name: static-assets
spec:
  bucket:
    valueFrom:
      name: my-s3-bucket
  awsRegion: us-west-2
  objects:
    - key: index.html
      content: |
        <!DOCTYPE html>
        <html><body><h1>Hello World</h1></body></html>
      contentType: text/html
      cacheControl: max-age=3600
```

# Example with Multiple Objects

Upload multiple objects with different content types and settings.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3ObjectSet
metadata:
  name: website-assets
spec:
  bucket:
    value: my-website-bucket
  awsRegion: us-east-1
  tags:
    environment: production
    project: website
  objects:
    - key: index.html
      content: |
        <!DOCTYPE html>
        <html><body><h1>Welcome</h1></body></html>
      contentType: text/html
      cacheControl: max-age=300
    - key: css/style.css
      content: |
        body { font-family: sans-serif; margin: 0; padding: 20px; }
      contentType: text/css
      cacheControl: max-age=86400
    - key: config/settings.json
      content: |
        {"theme": "dark", "language": "en"}
      contentType: application/json
```

# Example with Base64 Binary Content

Upload binary content using base64 encoding.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3ObjectSet
metadata:
  name: binary-assets
spec:
  bucket:
    value: my-assets-bucket
  awsRegion: us-west-2
  objects:
    - key: images/favicon.ico
      contentBase64: AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQ...
      contentType: image/x-icon
      cacheControl: max-age=604800
```

# Example with Per-Object Tags and ACL

Customize tags and access control per object.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3ObjectSet
metadata:
  name: mixed-access-objects
spec:
  bucket:
    value: my-mixed-bucket
  awsRegion: us-east-1
  tags:
    team: platform
  objects:
    - key: public/readme.txt
      content: "This file is publicly readable."
      contentType: text/plain
      acl: public-read
      tags:
        visibility: public
    - key: private/config.yaml
      content: |
        secret: placeholder
      contentType: application/x-yaml
      acl: private
      tags:
        visibility: private
```

---

These examples illustrate various configurations of the `AwsS3ObjectSet` API resource. Replace placeholder values with your actual bucket names, regions, and content.
