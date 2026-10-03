# S3 Definition Pipeline

This preset reads the pipeline definition from an S3 object — for
definitions too large to inline, or published to S3 by the pipeline's
own build tooling — with a parallelism cap on executions.

## When to Use

- Definitions generated and uploaded by CI (the SageMaker Python SDK's
  `pipeline.definition()` output)
- Large step graphs that would swamp the manifest inline

## What You Get

- A pipeline whose definition is fetched from
  `s3://my-pipeline-definitions/pipelines/nightly-training.json` at
  create
- A default cap of 4 steps executed in parallel across this pipeline's
  executions

## Customize

- Know the blind spot: AWS's describe API never returns the S3
  location, so drift on the S3 object is invisible to refresh — pin
  `versionId` to a specific object version and treat definition changes
  as manifest changes
- Reference the bucket from an AwsS3Bucket kind (`valueFrom`)
  instead of the literal name when it is managed in the same
  environment
- Drop `parallelismMaxSteps` to remove the pipeline-level cap
