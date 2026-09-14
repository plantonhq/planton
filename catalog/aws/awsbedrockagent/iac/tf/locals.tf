locals {
  # The agent name is metadata.name -- the naming basis both engines share
  # so a manifest deploys identically on either. AWS allows up to 100
  # characters of alphanumeric plus - and _ (no spaces or dots).
  agent_name = var.metadata.name

  # Resource-identity tags match the Pulumi module key-for-key.
  aws_tags = {
    "Name"                     = var.metadata.name
    "planton.ai/resource"      = "true"
    "planton.ai/organization"  = var.metadata.org
    "planton.ai/environment"   = var.metadata.env
    "planton.ai/resource-kind" = "AwsBedrockAgent"
    "planton.ai/resource-id"   = var.metadata.id
  }

  # Optional single-entry list attributes render only when declared. Memory
  # also carries its own switch, on by default once the block is declared, so
  # a null switch reads as on and only an explicit false keeps memory off.
  has_guardrail       = var.spec.guardrail != null
  has_memory          = var.spec.memory != null && coalesce(try(var.spec.memory.enabled, null), true)
  has_prompt_override = var.spec.prompt_override != null

  # Satellites keyed by their stable entry names (the for_each keys both
  # engines share and the output map keys).
  action_groups   = { for g in var.spec.action_groups : g.name => g }
  aliases         = { for a in var.spec.aliases : a.name => a }
  collaborators   = { for c in var.spec.collaborators : c.name => c }
  kb_associations = { for k in var.spec.knowledge_base_associations : k.name => k }
}
