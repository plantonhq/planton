##############################################
# locals.tf
#
# Computed local values for the
# KubernetesIstioBaseCrds module.
##############################################

locals {
  # Istio release the CRDs are fetched from.
  #
  # MUST stay in sync with `istio_release` in pkg/kubernetes/kubernetestypes/Makefile and
  # the Pulumi module's IstioRelease constant, so the installed CRD schema matches the
  # crd2pulumi-generated typed SDK that the Istio kinds are built against.
  # Always an exact release TAG (e.g. "1.30.3"), never a release BRANCH: a branch ref
  # moves as patches land, so the same deployed resource would install different CRD
  # schemas at different times — tag pinning keeps installs reproducible.
  istio_release = "1.30.3"

  # Full URL of the istio/base CRDs-only bundle (CRDs only -- no istiod, no controller).
  manifest_url = "https://raw.githubusercontent.com/istio/istio/${local.istio_release}/manifests/charts/base/files/crd-all.gen.yaml"

  # Planton identity labels — the planton.ai/* convention, identical to the
  # Pulumi module's label set (twin discipline). Conditional entries use the
  # null-prune idiom: heterogeneous conditional merges fail HCL type
  # unification when sibling entries infer as different object types.
  labels = {
    for k, v in {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesIstioBaseCrds"
      "planton.ai/resource-id"   = (var.metadata.id != null && var.metadata.id != "") ? var.metadata.id : null
      "planton.ai/organization"  = (var.metadata.org != null && var.metadata.org != "") ? var.metadata.org : null
      "planton.ai/environment"   = (var.metadata.env != null && var.metadata.env != "") ? var.metadata.env : null
    } : k => v if v != null
  }
}
