# GcpVertexAiIndex Terraform Module

This directory contains the Terraform module for provisioning a GCP Vertex AI Vector Search index.

## Usage

```hcl
module "vertex_ai_index" {
  source = "./path/to/module"

  metadata = {
    name = "my-index"
  }

  spec = {
    location     = "us-central1"
    display_name = "My Vector Index"
    config = {
      dimensions                  = 768
      approximate_neighbors_count = 150
      tree_ah_config              = {}
    }
  }
}
```

`spec.project_id` is optional: when empty, the index lands in the provider's
default project.

## Inputs

| Name | Type | Required | Description |
|------|------|----------|-------------|
| metadata | object | yes | Planton resource metadata |
| spec | object | yes | GcpVertexAiIndex specification |

Credentials are never module inputs: the provider block is empty and the
runner injects `GOOGLE_CREDENTIALS` (or the ambient ADC chain applies) —
the catalog-wide contract.

## Outputs

| Name | Description |
|------|-------------|
| index_id | Fully qualified index resource path (the deployed index's composition key) |
| index_name | The GCP-assigned numeric index ID |
| metadata_schema_uri | Schema URI for index-type-specific metadata |
| create_time | RFC3339 creation timestamp |
| update_time | RFC3339 last-update timestamp |

## Search Geometry

`spec.config` is required and entirely immutable (ForceNew): `dimensions`,
the algorithm arm (`tree_ah_config` XOR `brute_force_config`), `shard_size`,
`distance_measure_type`, and `feature_norm_type` shape the physical data
layout. Changing any of them replaces the index.

## Data Loading Quirks

`contents_delta_uri` is write-only upstream: GCP never reports it back, so an
out-of-band data load shows as a one-field diff on the next plan. A change to
the field travels in its own single-field update (no other index field can
ride along in the same apply).

## Destroy Behavior

`spec.deletion_policy` is the client-side destroy lever: empty/`DELETE`
deletes the index and its vectors, `PREVENT` makes destroy fail, `ABANDON`
drops it from state but leaves it standing (and billing). CMEK
(`spec.kms_key_name`) renders as the `encryption_spec` block.

## Provider Requirements

- `hashicorp/google` ~> 8.3

Note: index create/update/delete timeouts are 180 minutes — large batch
builds are genuinely slow.
