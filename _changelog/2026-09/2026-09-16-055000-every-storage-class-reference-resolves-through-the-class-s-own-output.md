# Every storage-class reference resolves through the class's own output

## What changed

- **Fourteen `KubernetesStorageClass` references across nine Kubernetes kinds resolve at `status.outputs.storage_class_name`** instead of `metadata.name`: ClickHouse (data and Keeper volumes), kube-prometheus-stack (three volumes), Kafka (two), Loki (two), NATS, RabbitMQ, Tempo, SigNoz, and the GitHub Actions runner scale set. The other fourteen references in the catalog already resolved there.
- No manifest, preset, or module changes: a preset that writes `value:` is untouched, a `valueFrom` that names its `fieldPath` is untouched, and the two paths resolve to the same string for a class Planton created.

## Why

The StorageClass kind's outputs say it in one line: `storage_class_name` "is the composition handle" -- the value claims put in their own `storageClassName`. A reference that resolves through a resource's metadata name is a claim that the cluster object is named after the Planton resource; a reference that resolves through the output is a claim about what was actually created. The catalog should make the second claim everywhere, and it now does. The ClickHouse coordination enum's bare `unspecified` zero value, found in the same review, stays: renaming a zero value buys a naming convention and risks a stored manifest that spells it.

## How to check

```bash
grep -rn 'KubernetesStorageClass,' catalog/kubernetes/*/v1alpha1/spec.proto -A1 | grep -c 'storage_class_name'   # 28
planton validate-refs --check
```
