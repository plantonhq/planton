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
  # crd2pulumi-generated typed SDK that the Istio components are built against.
  istio_release = "release-1.26"

  # Full URL of the istio/base CRDs-only bundle (CRDs only -- no istiod, no controller).
  manifest_url = "https://raw.githubusercontent.com/istio/istio/${local.istio_release}/manifests/charts/base/files/crd-all.gen.yaml"

  # Resource labels
  labels = {
    "app.kubernetes.io/name"       = "istio-base-crds"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "crds"
    "istio/release"                = local.istio_release
  }
}
