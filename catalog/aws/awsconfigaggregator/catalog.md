# AWS Config Aggregator

Creates the cross-account, cross-region rollup of AWS Config data -- one queryable view of resource configurations and rule compliance across an explicit account list or the whole AWS Organization. Aggregation has two sides and this kind models both as arms: the aggregator itself (deployed in the account that collects) and the reciprocal authorization grants (deployed in each source account). The aggregator references no Config recorder -- it works in an account with zero recorders, because the data comes from the source accounts' recorders.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions whichever arms the spec carries (at least one is required):

- **Configuration Aggregator** -- created only when `aggregation` is set; the collector, sourced from an explicit account list or the whole organization, across listed regions or all of them
- **Aggregator Authorizations** -- one per `authorizations` entry; the reciprocal grant a SOURCE account issues, naming the aggregator account and region allowed to collect from it. Grants are keyed by their `{account_id}:{authorized_aws_region}` identity, so reordering the list never churns resources

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with AWS Config permissions; for an organization source, its credentials must belong to the management account or the Config delegated administrator. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Nothing for a same-account rollup. For a cross-account account-list rollup, each source account outside the deployer's own needs this component deployed on its side with the grant arm.
- For an organization source: an IAM role trusting `config.amazonaws.com` with the `AWSConfigRoleForOrganizations` managed policy, referenced by `roleArn`.
- Data appears only from accounts and regions running a Config recorder -- the aggregator itself needs none.

## Deploy

### Console

Open the deployment store, find **AWS Config Aggregator**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the arms: the source shape for a collector, or the grant entries for a source account. Start from the **Organization Rollup** preset in the [Presets](#presets) tab for the org-wide compliance view.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsConfigAggregator
metadata:
  name: org-compliance-rollup
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  aggregation:
    organizationSource:
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: config-org-role
          fieldPath: status.outputs.role_arn
      allRegions: true
```

```shell
planton apply -f aws-config-aggregator.yaml
```

This creates an organization-sourced aggregator collecting every member account across every region into one queryable rollup, with membership self-discovering as accounts join the organization. An Infra Job tracks the provisioning in real time.

### InfraChart

When the aggregator deploys alongside its organization-reader role in one chart, wire the role via ValueFromRef:

```yaml
spec:
  region: us-east-1
  aggregation:
    organizationSource:
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: config-org-role
          fieldPath: status.outputs.role_arn
      allRegions: true
```

The InfraPipeline resolves the dependency graph, creates the role first, then the aggregator that assumes it.

## Key Configuration

These are the most important decisions when configuring Config aggregation. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Which arm, in which account** -- the `aggregation` arm deploys in the collector; the `authorizations` arm deploys in each source account of a cross-account topology, naming the AGGREGATOR's account and region (not its own). A same-account rollup needs only the aggregation arm; an account that both runs its own rollup and grants a sibling's may honestly declare both arms in one instance.

**Organization source versus account list** -- the organization source self-discovers membership (new member accounts join the view automatically), needs no per-account grants, and requires the management or delegated-administrator account plus the `AWSConfigRoleForOrganizations` role. The account list is for topologies outside an organization or deliberately narrower than one -- at the price of one grant deployment per source account. AWS accepts exactly one source shape, and each shape needs `regions` listed or `allRegions: true`.

**Pending authorization is a state, not an error** -- an account-list aggregator naming a source that has not yet granted it shows that source as pending; data flows the moment the source account applies its grant arm. Nothing is misconfigured -- the reciprocal record just has not landed yet.

**The aggregator sees only what recorders record** -- no recorder in a source account or region means no data from it, silently. The rollup is only as complete as the AWS Config Recorder coverage beneath it; pair them deliberately where coverage matters.

**Replacement semantics are asymmetric** -- adding a source block to an existing aggregator replaces the aggregator; editing or removing one updates in place (the provider's own diff rule). A grant's two leaves are both immutable -- changing either replaces that grant.

**Destroy touches the rollup only** -- destroying this component deletes whichever arms it manages; aggregated data disappears from the view, but every source account's own Config data is untouched.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `aggregation.organizationSource.roleArn` | `status.outputs.role_arn` |

### What This Kind Provides

`status.outputs` echoes the identities of whatever the instance manages: `aggregator_name` and `aggregator_arn` (set only when the aggregation arm is configured), and `authorization_arns`, a map keyed `{account_id}:{authorized_aws_region}` with one entry per grant. These are audit and import values rather than composition inputs -- the aggregated view is queried through the Config console and APIs, not referenced by other Infra Components.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Organization rollup** -- one aggregator in the management or delegated-administrator account, organization-sourced across all regions: the single pane of glass for org-wide Config data and rule compliance, with new member accounts joining automatically. Aggregators themselves are free; the recorders in source accounts bill as themselves. Start from the **Organization Rollup** preset.

**Security-account topology with grants** -- an account-list aggregator in a dedicated security account, with each source account deploying the grant arm authorizing it. The deliberate-membership alternative when the sources are not (all) one organization. Start from the **Source Account Grant** preset for the source side.

**Rollup plus coverage** -- an aggregator is only as good as the recorders feeding it: pair the org-wide view with AWS Config Recorder deployments in every account and region that must appear, and treat a permanently thin region in the view as a recorder gap, not an aggregator bug.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the organization-reader role an organization-sourced aggregator assumes, wired via `roleArn`
- [**AWS Config Recorder**](/infra-catalog/aws-config-recorder) -- the source of everything aggregated; accounts and regions without one contribute nothing
- [**AWS Config Rule**](/infra-catalog/aws-config-rule) -- the compliance results the rollup makes queryable across accounts
- [**AWS Organization**](/infra-catalog/aws-organization) -- the membership authority behind an organization source, with `config.amazonaws.com` in its trusted-access list
