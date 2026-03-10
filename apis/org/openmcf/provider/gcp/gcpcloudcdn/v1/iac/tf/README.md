# Terraform Module to Deploy AWS DynamoDB table

```shell
openmcf tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=openmcf/gcp-stacks/test-gcp-artifact-registry.tfstate"
```

```shell
openmcf tofu plan --manifest hack/manifest.yaml
```

```shell
openmcf tofu apply --manifest hack/manifest.yaml --auto-approve
```

```shell
openmcf tofu destroy --manifest hack/manifest.yaml --auto-approve
```
