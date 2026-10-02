// Package manifestprojection renders a Kubernetes manifest projection kind --
// a kind whose registry entry carries kubernetes_manifest_projection, so its
// spec mirrors one custom resource field for field -- into the object both IaC
// engines apply.
//
// A projection kind's spec holds two things: the custom resource's own spec,
// and a small envelope that describes the object instead (Envelope). The
// envelope is found by annotation, never by field name:
//
//   - the namespace: the field whose foreign key defaults to
//     KubernetesNamespace, routed to metadata.namespace;
//   - the object's own labels and annotations: top-level string maps marked
//     (dev.planton.shared.options.kubernetes_object_metadata), routed to
//     metadata.labels and metadata.annotations.
//
// Everything else is the custom resource's spec, projected by
// pkg/iac/specprojection with the custom resource's own JSON keys. Planton's
// identity labels (IdentityLabels) are stamped on every object and win over a
// manifest's labels, so an object can never be relabelled as another resource.
//
// One implementation serves both engines. The Terraform module generator
// (pkg/iac/tofu/generators) reads EnvelopeOf and IdentityLabelKeys to write the
// HCL that does the same split, and the Pulumi helper
// (pkg/iac/pulumi/pulumimodule/provider/kubernetes/manifestcr) applies Render's
// Object directly. The package imports neither engine and never the kind
// registry's message map, so a compiled Pulumi module stays small.
package manifestprojection
