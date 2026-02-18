# Examples

## Basic - Single Config File

```hcl
module "config" {
  source = "."
  metadata = {
    name = "app-config"
    id   = "s3objs-001"
    org  = "acme"
    env  = "dev"
    labels = {}
    annotations = {}
    tags = []
  }
  spec = {
    region  = "us-east-1"
    bucket  = "my-dev-bucket"
    objects = [{
      key          = "config/app.json"
      content      = "{\"debug\": true}"
      content_type = "application/json"
    }]
  }
}
```

## Multiple Objects with Tags

```hcl
module "website" {
  source = "."
  metadata = {
    name = "website-assets"
    id   = "s3objs-002"
    org  = "acme"
    env  = "production"
    labels = {}
    annotations = {}
    tags = []
  }
  spec = {
    region  = "us-west-2"
    bucket  = "my-website-bucket"
    tags    = { environment = "production" }
    objects = [
      {
        key           = "index.html"
        content       = "<html><body>Hello</body></html>"
        content_type  = "text/html"
        cache_control = "max-age=300"
      },
      {
        key          = "style.css"
        content      = "body { margin: 0; }"
        content_type = "text/css"
        cache_control = "max-age=86400"
        tags         = { type = "static-asset" }
      }
    ]
  }
}
```
