# Terraform Module to Deploy AWS DynamoDB table

```shell
planton tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=planton/gcp-stacks/test-gcp-dns-zone.tfstate"
```

```shell
planton tofu plan --manifest hack/manifest.yaml
```

```shell
planton tofu apply --manifest hack/manifest.yaml --auto-approve
```

```shell
planton tofu destroy --manifest hack/manifest.yaml --auto-approve
```
