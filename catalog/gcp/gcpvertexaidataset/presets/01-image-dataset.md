# Image Dataset

## Use Case

An image dataset for classification, object detection, or segmentation: register it, import images from Cloud Storage with their labels (or label them with a Vertex AI labeling task), and train AutoML or custom models on it.

## When to Use

- The first computer-vision dataset in a project
- Product, document, or defect images a team labels over time
- Any image training data that needs an owner and labels of its own

## What This Creates

- An empty image dataset in `us-central1` with Google's `image_1.0.0.yaml` schema, displayed as "Product images" and labelled by team and use case

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | The region your training jobs run in. |
| `metadataSchemaUri` | image schema | `text_1.0.0.yaml`, `video_1.0.0.yaml`, `tabular_1.0.0.yaml`, or `time_series_1.0.0.yaml` for other data types (fixed at creation). |
| `displayName` | "Product images" | The name your team sees in the console. |
| `deletionPolicy` | `DELETE` | `PREVENT` once the dataset holds labels you cannot recreate. |
