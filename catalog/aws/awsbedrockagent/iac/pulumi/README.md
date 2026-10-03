# Pulumi Module: AWS Bedrock Agent

Provisions an Amazon Bedrock agent and its folded satellites using Pulumi
(Go).

## Resources Created

- `bedrock.AgentAgent` — The agent (model, role, instructions, guardrail,
  memory, prompt overrides). The provider prepares the agent after every
  change and retries the preparing/OptLock conflict classes itself.
- `bedrock.AgentAgentActionGroup` — One per `spec.action_groups` entry
  (resource name `action-group-<entry name>`), attached to DRAFT.
- `bedrock.AgentAgentCollaborator` — One per `spec.collaborators` entry
  (resource name `collaborator-<entry name>`), attached to DRAFT.
- `bedrock.AgentAgentKnowledgeBaseAssociation` — One per
  `spec.knowledge_base_associations` entry (resource name
  `kb-association-<entry name>`), attached to DRAFT.
- `bedrock.AgentAgentAlias` — One per `spec.aliases` entry (resource name
  `alias-<entry name>`), created only AFTER every other satellite via
  explicit dependencies: an alias without explicit routing snapshots the
  draft into a new numbered version, and the snapshot must capture the
  fully-assembled agent.

## Notable Behavior

- Behavioral parity with the Terraform module is the contract: identical
  send conditions, identical constants (SESSION_SUMMARY memory type,
  RETURN_CONTROL custom control, OVERRIDDEN prompt creation mode),
  identical outputs.
- `prepare_agent` and `skip_resource_in_use_check` are apply-behavior
  knobs, not desired state — both engines keep the provider defaults
  (recorded in the parity manifest).
- Satellite entries iterate name-sorted for deterministic previews.

## Usage

The module is executed by the Planton platform through the entrypoint in
`main.go`, which loads the `AwsBedrockAgentIacInput`.
