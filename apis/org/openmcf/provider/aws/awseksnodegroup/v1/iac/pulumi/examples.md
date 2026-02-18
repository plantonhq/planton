```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: example
spec:
  region: us-west-2
```

CLI:

```bash
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```


