# AWS Bedrock Model Access

Accepts the marketplace agreement that entitles an AWS account, in one region, to invoke a specific Bedrock foundation model — the declarative form of the console's "Model access" page, one component instance per model. The modules look the model's public offer token up at deploy time, so the manifest names only the model, and the optional account use-case form (the once-per-account prerequisite for Anthropic models) folds into the same component. Most offers carry no subscription charge; the cost driver is per-token invocation once the model is used. Destroying the component cancels the agreement — access revocation, not cleanup.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Foundation Model Agreement** — acceptance of the model's public offer, with the short-lived offer token resolved fresh at deploy time; create waits until the agreement reaches AVAILABLE (commonly seconds to a few minutes)
- **Use-Case Form** — created only when `useCaseForm` is set: the account's use-case-for-model-access record. This is an account-global, write-once object — the module puts it (a re-put of identical content is a no-op; differing content fails loudly), and deleting the component never removes it because AWS provides no delete.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Bedrock marketplace permissions (`bedrock:CreateFoundationModelAgreement`, `bedrock:ListFoundationModelAgreementOffers`, `bedrock:PutUseCaseForModelAccess` and their read/delete siblings, plus the `aws-marketplace:Subscribe` family). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **The model available in the target region** — agreements are regional; the same model in a second region is a second instance.
- **For Anthropic models** — the account's use-case form on file, supplied through this component's `useCaseForm` or recorded once in the console (only for Anthropic; other vendors' models need no form).

## Deploy

### Console

Open the deployment store, find **AWS Bedrock Model Access**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region and model identifier. Start from the **Marketplace Model** preset in the [Presets](#presets) tab for the plain usage-based-offer shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockModelAccess
metadata:
  name: command-r-access
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  modelId: cohere.command-r-v1:0
```

```shell
planton apply -f model-access.yaml
```

This accepts Cohere Command R's public offer in us-west-2 — the account can invoke the model there once the agreement reaches AVAILABLE. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring model access. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Many models need no agreement at all** — auto-enabled models (Amazon first-party, Mistral, Meta) reject the offers API with "Agreement not supported for this model"; deploying this component for one fails. Probe with `aws bedrock list-foundation-model-agreement-offers --model-id <id>` before adding an instance: this kind is for the models that DO carry marketplace offers (Cohere, Anthropic, and other third-party vendors).

**The use-case form has exactly one owner** — AWS keeps ONE form per account, write-once: a re-put of different content errors loudly and there is no delete. Keep the form in exactly one instance (the Anthropic-access one) and omit it everywhere else; two instances with differing forms will fight.

**Destroy is a revocation event** — removing the component cancels the agreement, and anything still invoking the model in that region starts failing immediately. Treat deletion like removing a production credential.

**One instance per model per region, sequenced first** — agents, provisioned-throughput purchases, and inference profiles consume the model only after the grant exists. The `model_id` output exists precisely so charts can order model-consuming components behind this one.

**The manifest stays timeless** — offer tokens are short-lived credentials; the modules mint one fresh at each deploy and ignore token drift afterwards, so the accepted agreement is keyed by the model, and a re-apply never replaces (and therefore never revokes) the agreement.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies — `modelId` names an AWS-catalog foundation model, not a platform-managed resource, and the offer token is resolved by the modules at deploy time.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `model_id` | The foundation model identifier the agreement covers (echoes `modelId`) | Chart ordering: wiring it into a downstream component makes the InfraPipeline deploy the access grant before agents, throughput purchases, or profiles that use the model |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Marketplace model enablement** — one instance for a third-party model with a plain usage-based public offer, deployed ahead of everything that invokes it. Start from the **Marketplace Model** preset.

**Anthropic access with the use-case form** — one instance carries both the account's use-case form and the Anthropic model agreement; every other Anthropic instance in the account omits the form and relies on the one on file. Start from the **Anthropic Model with Use-Case Form** preset.

**Access as a chart layer** — in an InfraChart that deploys an agent or a throughput purchase, put the access instance in the same chart and consume its `model_id` so the pipeline proves the entitlement exists before anything invokes the model.

## Works With

- [**AWS Bedrock Agent**](/infra-catalog/aws-bedrock-agent) — agents on marketplace models need this agreement in place before they can prepare
- [**AWS Bedrock Provisioned Throughput**](/infra-catalog/aws-bedrock-provisioned-throughput) — capacity purchases for a model require the account to have access first
- [**AWS Bedrock Inference Profile**](/infra-catalog/aws-bedrock-inference-profile) — application profiles over a marketplace model presume the agreement exists
- [**AWS Bedrock Custom Model**](/infra-catalog/aws-bedrock-custom-model) — customization jobs on a marketplace base model require access to it
