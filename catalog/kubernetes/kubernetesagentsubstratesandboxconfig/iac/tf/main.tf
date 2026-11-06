# Applies the SandboxConfig custom resource through kubectl_manifest (alekc/kubectl):
# no plan-time cluster dependency (plannable before the CRDs exist), applied
# server-side. No wait, deliberately: the CR is configuration its controller
# consumes; applying it server-side-validated is the whole contract. Pulumi
# equivalent: the typed CR without await annotations.
resource "kubectl_manifest" "sandbox_config" {
  yaml_body = yamlencode({
    apiVersion = "ate.dev/v1"
    kind       = "SandboxConfig"
    metadata = {
      name        = var.spec.name
      labels      = local.labels
      annotations = local.annotations
    }
    spec = local.manifest_spec
  })

  server_side_apply = true
}
