# GcpDataprocCluster - Pulumi Implementation

## Building

```bash
cd iac/pulumi
make build
```

## Testing

```bash
# Preview changes
make preview

# Apply changes
make up

# Destroy resources
make destroy
```

## Debug

Set the stack input using the hack manifest:

```bash
pulumi config set --path 'target' "$(cat iac/hack/manifest.yaml | yq -o json)"
```
