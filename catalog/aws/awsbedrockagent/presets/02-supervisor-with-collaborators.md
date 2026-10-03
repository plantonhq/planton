# Supervisor with Collaborators

This preset creates a multi-agent supervisor: it answers general
questions itself (with product-docs retrieval), delegates billing
questions to a specialist agent through its `live` alias, and carries
30-day session-summary memory.

## When to Use

- Splitting a broad assistant into focused specialist agents while
  keeping one entry point
- Teams composing agents as chart catalog kinds: the collaborator reference
  reads the specialist's `alias_arns` output, so the chart orders the
  deployments

## What You Get

- A SUPERVISOR-mode agent with conversation-history relay to its
  collaborator
- A knowledge-base association for retrieval-augmented answers
- A `live` alias snapshotting the assembled supervisor

## Customize

- Switch `agentCollaboration` to `SUPERVISOR_ROUTER` to route each
  request to exactly one collaborator without a planning step
- Add more `collaborators` entries — each names another agent's alias and
  its delegation instruction
