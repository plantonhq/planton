# Local values and computed configuration

locals {
  # The synthetic resource id mirrors the kind's registry id_prefix so output
  # shapes look exactly like a real kind's.
  resource_id = "tckg-${var.metadata.name}"

  # A deterministic URL derived from inputs: proves output values flow
  # from resolved inputs through the engine into outputs.
  url = "test://${var.metadata.name}"
}
