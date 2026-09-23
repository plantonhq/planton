# GcpVertexAiDataset Guide

The judgment this guide protects: a dataset's type and home are permanent, and destroying it destroys the labeled examples inside it. Decide the type, the region, and the key before the first import, and guard a labeled dataset like a database.

## The block declares, imports fill

This block registers the dataset; it never imports an example. Images, text, video, and tabular rows arrive through the Vertex AI API, the console, the SDK, or a pipeline, together with their annotations. That split is deliberate: datasets grow and get relabeled daily, while the dataset itself -- its type, region, key, and owner labels -- changes rarely. Declaring the container keeps training data owned and encrypted without putting every example in a manifest.

## The type is the schema

`metadataSchemaUri` names one of Google's five schemas under `gs://google-cloud-aiplatform/schema/dataset/metadata/`: `image_1.0.0.yaml`, `text_1.0.0.yaml`, `tabular_1.0.0.yaml`, `video_1.0.0.yaml`, or `time_series_1.0.0.yaml`. It decides which AutoML objectives, labeling tasks, and import formats the dataset accepts. Google fixes it at creation, so a change is a new dataset -- plan the type with the model you intend to train.

## One region for data, jobs, and models

A training job reads a dataset in its own region and writes its model there. Put the dataset where your training runs (usually where your accelerators and your quota are). The block requires a region, never `global` or a multi-region.

## Identity comes from Google

Google assigns the dataset a numeric id at creation. The `name` output is the full resource name jobs and pipelines take; `dataset_id` is the id alone for console links and SDK calls. The display name defaults to `metadata.name` and can change freely.

## Encryption and destroy

`kmsKeyName` encrypts the dataset and every imported item under your key, and the Vertex AI Service Agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on it. The key is fixed at creation. `deletionPolicy` decides what a destroy does: `DELETE` removes the dataset and its items and annotations, `PREVENT` makes destroy fail, and `ABANDON` leaves it in place. Labeling costs money and time, so a labeled dataset should carry `PREVENT`.
