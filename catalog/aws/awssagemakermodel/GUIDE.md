# AwsSagemakerModel — Kind Guide

Authored operational judgment for the SageMaker model kind: the
design decisions behind the spec's shape, and what to know before
running models in production.

## Design decisions

- **The model's AWS name derives from `metadata.name`.** SageMaker's
  model charset (letters, digits, hyphens, ≤ 63) matches component
  names as-is — no separate name field to keep in sync.
- **Exactly one container form.** A model is a single
  `primary_container` or a pipeline of 2–15 `containers` — never both,
  never neither. AWS's CreateModel contract is not schema-enforced
  upstream, so the spec owns it and rejects invalid manifests before
  any API call.
- **`inference_execution_mode` pairs with the pipeline form.** AWS
  accepts the execution mode only on multi-container models; the spec
  enforces the pairing so a stray `Serial` on a single-container model
  fails at manifest time, not at apply.
- **Two artifact forms, at most one per container.** The classic
  compressed `model_data_url` and the uncompressed `model_data_source`
  (which also carries gated-model EULA acceptance) are alternatives —
  AWS rejects both together, and so does the spec. The upstream
  single-member `s3_data_source` wrapper is flattened to one message.
- **Container-level pairing rules live in the spec.** At least one of
  `image` and `model_package_arn`; `multi_model_cache` only in
  `MultiModel` mode; registry credentials only with `Vpc` access mode
  — each an AWS rule surfaced at manifest time.

## Running models in production

- **Treat every change as a replacement.** All fields are create-time
  only (the provider's update is tags-only); any spec change replaces
  the model. That is AWS's intended flow: roll a new model, repoint the
  endpoint, and the old model disappears once nothing references it.
- **Models cost nothing to keep.** Only endpoints and batch jobs
  deploying a model bill — keep prior model versions around as instant
  rollback targets.
- **Scope the execution role tightly.** It needs ECR pull on the image
  and S3 read on the artifacts, nothing more; SageMaker assumes it at
  deploy time, so a missing grant surfaces as an endpoint failure, not
  a model-create failure.
- **AWS's prebuilt framework images live in per-region ACCOUNTS.** The
  scikit-learn/XGBoost library account for us-west-2 is 246618743249;
  us-west-1's is 746614075791 — and using another region's account
  fails `CreateModel` with a misleading 400 claiming the repository
  "does not grant ... to sagemaker.amazonaws.com service principal"
  (live-verified). That error means "wrong account or nonexistent
  repo", not a policy you can fix. Derive the account from the
  sagemaker-python-sdk `image_uri_config` files for your region.
- **Network isolation is absolute.** `enable_network_isolation` blocks
  all inbound and outbound calls from the container — even to AWS
  services. Artifacts and images still load normally; combine with
  `vpc_config` for private serving.
- **Uncompressed artifacts for large models.** `model_data_source`
  with `compression_type: None` on an `S3Prefix` skips the tarball
  extraction step that dominates load time for multi-GB models — and
  it is the only form that accepts a gated model's EULA.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
