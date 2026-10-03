# Team Experiment Tracking

This preset stands up a `Small` tracking server for a single ML team —
experiments, runs, and model tracking with artifacts in your S3
bucket, and automatic model registration left off (the safe default:
turning it on is effectively one-way).

## When to Use

- The first MLflow deployment for a team of up to ~25 users
- Steady, daily tracking load that justifies an always-on server —
  the `Small` size bills hourly from Created onward, traffic or not;
  the verified figure lives in the kind's generated estimate at
  `catalog/_pricing/estimates/awssagemakermlflowserver.yaml`

## What You Get

- A managed MLflow tracking server with its UI and tracking URI
  (`tracking_server_url` in the outputs)
- Model files and run outputs stored in your S3 bucket through the
  server's IAM role
- The latest MLflow version, since no pin is set

## Customize

- Resize to `Medium` (~50 users) or `Large` (~100) in place as the
  team grows
- Point `roleArn` at a composed `AwsIamRole` with `valueFrom` instead
  of a literal ARN
- If tracking is intermittent, consider the serverless
  `AwsSagemakerMlflowApp` instead — no idle charge, billed per use
