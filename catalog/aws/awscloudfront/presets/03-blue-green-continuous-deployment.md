# Blue/Green Rollout with Continuous Deployment

This preset stages a CloudFront configuration change on real production traffic before promoting it: the primary distribution owns a continuous-deployment policy that routes a weighted slice of viewers to a staging distribution carrying the candidate configuration.

## When to Use

- Validating risky distribution changes (new cache policies, origin swaps, function associations) on a fraction of live traffic instead of all of it
- Any production CDN where a bad configuration push must be contained to a bounded blast radius

## Key Configuration Choices

- **Two distributions, one policy** -- deploy the candidate configuration as its own `AwsCloudFront` resource with `staging: true` (staging distributions serve no direct traffic and cannot carry aliases), then reference its `domain_name` output from the primary's `continuousDeployment` block; the primary creates and attaches the policy
- **Weighted routing with session stickiness** -- `weight: 0.10` sends 10% of traffic to staging (AWS caps the slice at 15%); stickiness pins each viewer to one side for their session so nobody sees a mixed experience
- **Header routing as the alternative** -- swap `singleWeight` for `singleHeader` (header must carry the `aws-cf-cd-` prefix) to route only opt-in requests, letting a team test in production before any percentage rollout
- **Promotion is a copy, not a cutover** -- when staging looks healthy, apply the staging configuration to the primary and remove the policy; the staging distribution is reusable for the next rollout

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `replace-me-bucket.s3.us-east-1.amazonaws.com` | The S3 bucket holding the content (replace the bucket and region segments) | `AwsS3Bucket` outputs |
| `<staging-distribution-resource-name>` | The staging `AwsCloudFront` resource (deployed with `staging: true`) | Your staging distribution manifest |

## Related Presets

- **01-s3-static-website** -- The base single-distribution shape this preset stages changes for
- **02-custom-domain-cdn** -- Combine with this preset when the primary serves a custom domain
