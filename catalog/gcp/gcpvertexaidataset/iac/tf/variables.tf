variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "GcpVertexAiDataset specification"
  type = object({
    # The GCP project the dataset lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Vertex AI location (region) the dataset lives in, e.g.
    # "us-central1". Training jobs that read the dataset must run in the same
    # location. Immutable.
    location = string

    # Human-readable name shown in the console -- up to 128 UTF-8
    # characters. Defaults to metadata.name. Mutable in place.
    display_name = optional(string, "")

    # The Cloud Storage YAML file that fixes what kind of data the dataset
    # holds -- one of Google's published schemas under
    # gs://google-cloud-aiplatform/schema/dataset/metadata/:
    #   image_1.0.0.yaml       images (classification, object detection,
    #                          segmentation)
    #   text_1.0.0.yaml        text (classification, entity extraction,
    #                          sentiment)
    #   tabular_1.0.0.yaml     rows from BigQuery or Cloud Storage CSV
    #   video_1.0.0.yaml       video (classification, action recognition,
    #                          object tracking)
    #   time_series_1.0.0.yaml forecasting data
    # e.g. "gs://google-cloud-aiplatform/schema/dataset/metadata/image_1.0.0.yaml".
    # Immutable.
    metadata_schema_uri = string

    # Labels on the dataset. The platform attribution labels are merged in
    # and win on key conflicts.
    labels = optional(map(string), {})

    # Customer-managed encryption key protecting the dataset and everything
    # imported into it: a GcpKmsKey reference or a literal
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
    # in the same region. The Vertex AI Service Agent needs
    # roles/cloudkms.cryptoKeyEncrypterDecrypter on the key. Omit to use
    # Google-managed encryption. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # What happens to the dataset when this resource is destroyed:
    #   "" / "DELETE" -- the dataset is deleted, imported data items and
    #                    annotations included
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the dataset leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
