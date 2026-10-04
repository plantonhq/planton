# AwsSsmDocument

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsSsmDocumentSpec defines the desired configuration for one
customer-owned AWS Systems Manager document: a reusable definition
of the actions Systems Manager performs (run a command, drive an
automation, open a session, ...), authored as JSON or YAML content.

The document's name is metadata.name (3-128 characters of letters,
digits, underscores, hyphens, and periods - hyphenated names fit)
and changing it forces replacement. Updating the CONTENT creates a
new document version and promotes it to the default version;
documents at schema version 1.x can only be updated when the
content itself changes (an AWS rule for legacy command documents).

Associations (State Manager bindings of a document to targets on a
schedule) are their own AwsSsmAssociation component - an association
binds ANY document, AWS-managed or customer-owned, so it is not this
document's satellite.

## Example

```yaml
# Canonical AwsSsmDocument example (hack/dev manifest and refgen
# Example source): a JSON Command document with a version label,
# target-type restriction, an attachment, and account sharing. Literal
# values stand in so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSsmDocument
metadata:
  name: install-app-agent
  id: install-app-agent
  org: test-org
  env: dev
spec:
  region: us-west-2
  documentType: Command
  documentFormat: JSON
  content: |
    {
      "schemaVersion": "2.2",
      "description": "Install the app agent on managed nodes",
      "parameters": {
        "AgentVersion": {
          "type": "String",
          "description": "Agent version to install",
          "default": "latest"
        }
      },
      "mainSteps": [
        {
          "action": "aws:runShellScript",
          "name": "installAgent",
          "inputs": {
            "runCommand": ["curl -fsSL https://example.com/agent-{{AgentVersion}}.sh | bash"]
          }
        }
      ]
    }
  targetType: /AWS::EC2::Instance
  versionName: release-1.0.0
  attachmentSources:
    - key: S3FileUrl
      name: agent-config.json
      values:
        - https://s3.amazonaws.com/example-artifacts/agent-config.json
  shareWithAccountIds:
    - "123456789012"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.content` | `string` | yes |  |  |
| `spec.documentType` | `string` |  |  |  |
| `spec.documentFormat` | `string` |  |  |  |
| `spec.targetType` | `string` | yes |  |  |
| `spec.versionName` | `string` | yes |  |  |
| `spec.attachmentSources` | `[]AwsSsmDocumentAttachmentSource` |  |  |  |
| `spec.attachmentSources[].key` | `string` |  |  |  |
| `spec.attachmentSources[].name` | `string` | yes |  |  |
| `spec.attachmentSources[].values` | `[]string` | yes |  |  |
| `spec.shareWithAccountIds` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the document lives in.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.content

`string` · required

The document body, in the format document_format declares. For
Command documents this is the schema-versioned action definition
("schemaVersion", "mainSteps", ...); Automation documents define
their steps and parameters the same way.

- rule: {"string":{"minLen":"1"}}

### spec.documentType

`string`

What kind of document this is. Command documents run on managed
nodes, Automation documents drive AWS-API workflows, Session
documents configure Session Manager, Policy documents gather
inventory - the rest are specialty types AWS accepts on the same
API.

- rule: {"string":{"in":["Command","Policy","Automation","Session","Package","ApplicationConfiguration","ApplicationConfigurationSchema","DeploymentStrategy","ChangeCalendar","Automation.ChangeTemplate","ProblemAnalysis","ProblemAnalysisTemplate","CloudFormation","ConformancePackTemplate","QuickSetup","ManualApprovalPolicy","AutoApprovalPolicy"]}}

### spec.documentFormat

`string`

The format the content is written in. Unset = JSON (the provider
default).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["JSON","YAML","TEXT"]}}

### spec.targetType

`string` · required

Restricts the AWS resource types the document can run against
(e.g. "/AWS::EC2::Instance"; "/" alone means all types).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"1","maxLen":"200","pattern":"^\\/[\\w\\.\\-\\:\\/]*$"}}

### spec.versionName

`string` · required

An artifact-style version label for the document version created
by this content (e.g. "release-1.0.0"). Immutable per version -
reusing a label on changed content is rejected by AWS.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"3","maxLen":"128","pattern":"^[0-9A-Za-z_.-]{3,128}$"}}

### spec.attachmentSources

`[]AwsSsmDocumentAttachmentSource`

Attachments for document types that carry artifacts (e.g. Package
documents' installers), each naming where the artifact lives. AWS
never reads attachment metadata back, so a freshly imported
document shows none - the first apply re-asserts them.

- rule: {"repeated":{"maxItems":"20"}}

### spec.attachmentSources[].key

`string`

How the values locate the artifact: a public URL, an S3 file URL,
or a reference to another document's attachment.

- rule: {"string":{"in":["SourceUrl","S3FileUrl","AttachmentReference"]}}

### spec.attachmentSources[].name

`string` · required

The attachment's file name inside the document. Unset = derived
from the location.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"3","maxLen":"128","pattern":"^[0-9A-Za-z_.-]{3,128}$"}}

### spec.attachmentSources[].values

`[]string` · required

The location values for the key (e.g. the S3 URL). One entry for
SourceUrl/S3FileUrl; AttachmentReference entries name
"documentName/attachmentName" pairs.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1","maxLen":"1024"}}}}

### spec.shareWithAccountIds

`[]string`

AWS account IDs this document is shared with, or the single entry
"All" to share it publicly. AWS applies changes in batches of 20
behind the scenes.

- rule: {"repeated":{"unique":true,"items":{"string":{"pattern":"^(\\d{12}|All)$"}}}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSsmDocument, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.document_name` | `string` | The document's name (also the provider's import ID, and what associations reference). |
| `status.outputs.document_arn` | `string` | The document's ARN. |
| `status.outputs.default_version` | `string` | The default document version (the one associations resolve when they pin "$DEFAULT"; updates promote the new version here). |
| `status.outputs.latest_version` | `string` | The latest document version. |
| `status.outputs.document_hash` | `string` | The Sha256 digest of the document content. |
| `status.outputs.status` | `string` | The document's lifecycle status ("Creating", "Active", ...). |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsSsmAssociation | `spec.documentName` | `status.outputs.document_name` |

## See Also

- [Overview](../README.md)
