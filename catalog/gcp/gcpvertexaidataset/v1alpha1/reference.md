# GcpVertexAiDataset

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVertexAiDatasetSpec defines a Vertex AI managed dataset
(`google_vertex_ai_dataset`) -- the registered container Vertex AI
training, AutoML, labeling, and evaluation read their examples from.
The block declares the dataset itself: its data type (through
metadata_schema_uri), its name, its labels, and its encryption. The data
items inside it (images, text snippets, rows, annotations) are imported
through the Vertex AI API, the console, or the SDK and are deliberately
not part of this block.

Google assigns the dataset a numeric id at creation; the stack outputs
carry it and the full resource name for the training jobs and pipelines
that consume the dataset.

Immutable: location, metadata_schema_uri, and the encryption key (a
change replaces the dataset and everything imported into it). The
display name and labels update in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiDataset
metadata:
  name: product-images
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Product images
  # The schema fixes the dataset's data type: image, text, tabular, video,
  # or time series. It cannot change after creation.
  metadataSchemaUri: gs://google-cloud-aiplatform/schema/dataset/metadata/image_1.0.0.yaml
  labels:
    team: vision
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.metadataSchemaUri` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the dataset lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Vertex AI location (region) the dataset lives in, e.g.
"us-central1". Training jobs that read the dataset must run in the same
location. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.displayName

`string`

Human-readable name shown in the console -- up to 128 UTF-8
characters. Defaults to metadata.name. Mutable in place.

- rule: {"string":{"maxLen":"128"}}

### spec.metadataSchemaUri

`string` · required

The Cloud Storage YAML file that fixes what kind of data the dataset
holds -- one of Google's published schemas under
gs://google-cloud-aiplatform/schema/dataset/metadata/:
  image_1.0.0.yaml       images (classification, object detection,
                         segmentation)
  text_1.0.0.yaml        text (classification, entity extraction,
                         sentiment)
  tabular_1.0.0.yaml     rows from BigQuery or Cloud Storage CSV
  video_1.0.0.yaml       video (classification, action recognition,
                         object tracking)
  time_series_1.0.0.yaml forecasting data
e.g. "gs://google-cloud-aiplatform/schema/dataset/metadata/image_1.0.0.yaml".
Immutable.

- rule: {"required":true,"string":{"pattern":"^gs://[^/]+/.+\\.yaml$"}}

### spec.labels

`map<string, string>`

Labels on the dataset. The platform attribution labels are merged in
and win on key conflicts.

### spec.kmsKeyName

`string | valueFrom`

Customer-managed encryption key protecting the dataset and everything
imported into it: a GcpKmsKey reference or a literal
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
in the same region. The Vertex AI Service Agent needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on the key. Omit to use
Google-managed encryption. Immutable.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.deletionPolicy

`string`

What happens to the dataset when this resource is destroyed:
  "" / "DELETE" -- the dataset is deleted, imported data items and
                   annotations included
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the dataset leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiDataset, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project_number}/locations/{location}/datasets/{dataset_id}. |
| `status.outputs.dataset_id` | `string` | The numeric id Google assigned at creation (the last segment of name). |
| `status.outputs.location` | `string` | The location the dataset lives in. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |

## See Also

- [Overview](../README.md)
