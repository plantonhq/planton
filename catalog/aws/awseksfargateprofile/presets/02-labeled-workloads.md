# Label-Scoped Fargate

This preset runs only opted-in pods on Fargate -- pods that carry the
selector's labels -- while the rest of the namespace keeps running on
node groups. The mixed-compute pattern for shared namespaces.

## When to Use

- Shared namespaces where only certain workloads (bursty jobs,
  isolation-sensitive services) should be serverless
- Gradual Fargate adoption: workloads opt in by adding a label, no
  namespace reshuffle
- Cost experiments: run the same service on both compute styles and
  compare

## Key Configuration Choices

- **`labels` on the selector** -- AND semantics: a pod must match the
  namespace and carry every listed label (up to 5 pairs); values accept
  `*`/`?` wildcards
- **Label choice is the API** -- `compute: fargate` is a readable
  convention; whatever you pick becomes the opt-in switch workloads use
- **Profiles are immutable** -- adding a label pair later replaces the
  profile; run an overlapping profile through the change if matched
  pods must keep a scheduling target

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<profile-resource-name>` | Name for the profile (AWS limit 63 chars) | Your naming convention |
| `<aws-region>` | AWS region code (e.g., `us-west-2`) | Your deployment region |
| `<cluster-resource-name>` | Name of the AwsEksCluster resource | Your cluster manifest's `metadata.name` |
| `<pod-execution-role-resource-name>` | Name of the AwsIamRole for Fargate pod execution | Your role manifest's `metadata.name` |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of two private AwsSubnet resources in different AZs | Your subnet manifests' `metadata.name` |
| `<namespace>` | The shared namespace | Your cluster's namespace layout |

## Common Additions

- More selectors (up to 5) for additional namespace/label combinations

## Related Presets

- **01-namespace** -- whole-namespace serverless placement
