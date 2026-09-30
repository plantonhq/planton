// Deterministic Terraform-module generation for Kubernetes-CRD-projection kinds.
// Domain: Planton infra-hub / IaC. These kinds project their spec onto a single
// Kubernetes custom resource, so the whole iac/tf module is a pure function of
// the schema and is generated rather than hand-written.

package generators

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/refannotations"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GenerateManifestModule returns the complete iac/tf module (filename -> HCL
// content) for a Kubernetes-CRD-projection kind. The module is a thin
// kubectl_manifest (alekc/kubectl) passthrough: variable "spec" is typed `any`
// and handed verbatim to the CR, because the proto->tfvars converter
// (ProtoToManifestTFVars) already emits the manifest-shaped, camelCase,
// null-pruned spec (with StringValueOrRef foreign keys resolved to literal
// strings). This removes the snake->camel / null-prune / oneOf locals.tf that
// every CRD module previously hand-wrote.
//
// kubectl_manifest rather than the hashicorp provider's kubernetes_manifest,
// deliberately: kubernetes_manifest fetches the CR's OpenAPI type from the live
// cluster at PLAN time, so a CR could never be planned before its CRDs are
// installed — which breaks single-run infra charts (CRD installer + CRs
// together) and makes offline plan proofs impossible. kubectl_manifest renders
// from yaml_body with no plan-time cluster dependency, applies server-side, and
// supports import (apiVersion//kind//name[//namespace]).
//
// The kind is passed explicitly (rather than read from msg) because callers
// generate from an empty message instance whose `kind` field is unset; msg is
// used only for its descriptors. Returns an error if the kind is not annotated
// with a kubernetes_manifest_projection, or if a stack output cannot be mapped
// to a source field (which would otherwise silently produce a wrong module).
func GenerateManifestModule(kind cloudresourcekind.CloudResourceKind, msg proto.Message) (map[string]string, error) {
	proj := manifestProjection(kind)
	if proj == nil {
		return nil, errors.New("kind has no kubernetes_manifest_projection in CloudResourceKindMeta; " +
			"generate-module only applies to CRD-projection kinds")
	}
	apiVersion, crdKind := proj.GetApiVersion(), proj.GetKind()
	if apiVersion == "" || crdKind == "" {
		return nil, errors.Errorf("kubernetes_manifest_projection must set both api_version and kind (got %q/%q)", apiVersion, crdKind)
	}

	md := msg.ProtoReflect().Descriptor()
	specMsg := messageOfField(md, "spec")
	if specMsg == nil {
		return nil, errors.New("manifest message has no message-typed spec field")
	}

	// A namespace foreign key means the CR is namespaced; its value maps to
	// metadata.namespace, not into the CR spec. Cluster-scoped kinds (e.g.
	// GatewayClass) have no such field.
	nsJSONName, namespaced := namespaceForeignKeyJSONName(specMsg)
	label := kebabFromPascal(crdKind)

	files := map[string]string{
		"provider.tf":  manifestProviderTF(kind.String()),
		"backend.tf":   manifestBackendTF(),
		"variables.tf": manifestVariablesTF(apiVersion, crdKind),
		"locals.tf":    manifestLocalsTF(kind.String(), nsJSONName, namespaced),
		"main.tf":      manifestMainTF(strings.ReplaceAll(label, "-", "_"), apiVersion, crdKind, nsJSONName, namespaced),
	}

	outputs, err := manifestOutputsTF(md, specMsg, crdKind, nsJSONName, namespaced)
	if err != nil {
		return nil, err
	}
	if outputs != "" {
		files["outputs.tf"] = outputs
	}
	return files, nil
}

func manifestProjection(kind cloudresourcekind.CloudResourceKind) *cloudresourcekind.KubernetesManifestProjection {
	meta, err := crkreflect.KindMeta(kind)
	if err != nil {
		return nil
	}
	return meta.GetKubernetesManifestProjection()
}

func messageOfField(md protoreflect.MessageDescriptor, name string) protoreflect.MessageDescriptor {
	fd := md.Fields().ByName(protoreflect.Name(name))
	if fd == nil || fd.Kind() != protoreflect.MessageKind || fd.IsMap() {
		return nil
	}
	return fd.Message()
}

// namespaceForeignKeyJSONName returns the JSON key of the spec field that is the
// KubernetesNamespace foreign key, and whether one exists. It is detected via
// the (foreignkey.default_kind) option rather than the field name, so the rule
// stays correct even if a future kind names the field differently.
func namespaceForeignKeyJSONName(specMsg protoreflect.MessageDescriptor) (string, bool) {
	fields := specMsg.Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if refannotations.Of(fd).DefaultKind == cloudresourcekind.CloudResourceKind_KubernetesNamespace {
			return fd.JSONName(), true
		}
	}
	return "", false
}

func manifestProviderTF(kindName string) string {
	// Pin the provider source and version floor so a generated module is
	// self-contained: without required_providers, Terraform would silently
	// fall back to an unconstrained default, and provider upgrades would
	// ripple through every module unpinned.
	return fmt.Sprintf(`# Provider requirements for the %s module.
#
# kubectl (alekc/kubectl) applies the CR: unlike the hashicorp kubernetes
# provider's kubernetes_manifest resource, kubectl_manifest needs no cluster
# connection at plan time, so the CR can be planned before its CRDs exist
# (single-run infra charts, offline plan proofs). This module creates no other
# Kubernetes objects, so kubectl is its only provider.
#
# The provider is configured by the calling workspace/environment (the same
# kubeconfig environment contract).

terraform {
  required_version = ">= 1.0"

  required_providers {
    kubectl = {
      source  = "alekc/kubectl"
      version = ">= 2.0"
    }
  }
}

provider "kubectl" {
}
`, kindName)
}

func manifestBackendTF() string {
	return `terraform {
  backend "local" {}
}
`
}

func manifestModuleHeader(apiVersion, crdKind string) string {
	return fmt.Sprintf(`# Generated by 'planton tofu generate-module'. Do not edit by hand; regenerate instead.
#
# Thin projection of the %s %q custom resource.
#
# The proto->tfvars converter emits the manifest-shaped (camelCase, null-pruned)
# spec with StringValueOrRef foreign keys resolved to literal strings, so this
# module hands it to kubectl_manifest verbatim -- no snake->camel, null-prune,
# or oneOf logic. variable "spec" is typed 'any' because the apiserver plus
# Planton protovalidate are the schema authority; re-encoding the CRD schema as
# HCL types would only reintroduce the pruning this design removes.
`, apiVersion, crdKind)
}

func manifestVariablesTF(apiVersion, crdKind string) string {
	// The spec is typed `any` (pure passthrough — re-encoding the CRD spec as
	// HCL types would only reintroduce the optional/null friction this design
	// eliminates), but metadata is a typed OBJECT with defaulted optional
	// attributes: the module's identity labels read metadata.id/org/env, and a
	// tfvars document carrying only `name` (the common case) would make those
	// attribute reads FAIL on `any` — HCL's && evaluates both operands, so
	// even a null-guarded read errors when the attribute does not exist at
	// all. Optional-with-default gives every read a value.
	var b strings.Builder
	b.WriteString(manifestModuleHeader(apiVersion, crdKind))
	b.WriteString("\n")
	b.WriteString("variable \"metadata\" {\n")
	b.WriteString("  description = \"Cloud resource metadata (name plus the optional Planton identity attributes the module renders as labels).\"\n")
	b.WriteString("  type = object({\n")
	b.WriteString("    name        = string\n")
	b.WriteString("    id          = optional(string, \"\")\n")
	b.WriteString("    org         = optional(string, \"\")\n")
	b.WriteString("    env         = optional(string, \"\")\n")
	b.WriteString("    labels      = optional(map(string), {})\n")
	b.WriteString("    annotations = optional(map(string), {})\n")
	b.WriteString("    tags        = optional(list(string), [])\n")
	b.WriteString("  })\n")
	b.WriteString("}\n\n")
	b.WriteString("variable \"spec\" {\n")
	fmt.Fprintf(&b, "  description = %q\n",
		fmt.Sprintf("Spec for the %s %q custom resource, passed through verbatim to kubectl_manifest. Typed 'any': the apiserver and Planton protovalidate are the schema authority.", apiVersion, crdKind))
	b.WriteString("  type        = any\n")
	b.WriteString("}\n")
	return b.String()
}

func manifestLocalsTF(kindName, nsJSONName string, namespaced bool) string {
	var b strings.Builder
	b.WriteString("locals {\n")
	b.WriteString("  # Planton identity labels — the planton.ai/* convention, identical to the\n")
	b.WriteString("  # Pulumi module's label set (twin discipline). Conditional entries use the\n")
	b.WriteString("  # null-prune idiom: heterogeneous conditional merges fail HCL type\n")
	b.WriteString("  # unification when sibling entries infer as different object types.\n")
	b.WriteString("  labels = {\n")
	b.WriteString("    for k, v in {\n")
	b.WriteString("      \"planton.ai/resource\"      = \"true\"\n")
	b.WriteString("      \"planton.ai/resource-name\" = var.metadata.name\n")
	fmt.Fprintf(&b, "      \"planton.ai/resource-kind\" = %q\n", kindName)
	b.WriteString("      \"planton.ai/resource-id\"   = (var.metadata.id != null && var.metadata.id != \"\") ? var.metadata.id : null\n")
	b.WriteString("      \"planton.ai/organization\"  = (var.metadata.org != null && var.metadata.org != \"\") ? var.metadata.org : null\n")
	b.WriteString("      \"planton.ai/environment\"   = (var.metadata.env != null && var.metadata.env != \"\") ? var.metadata.env : null\n")
	b.WriteString("    } : k => v if v != null\n")
	b.WriteString("  }\n\n")
	if namespaced {
		fmt.Fprintf(&b, "  # The CR spec is var.spec minus the Planton %q foreign key, which maps to\n", nsJSONName)
		b.WriteString("  # metadata.namespace rather than into the CR spec. The converter already emits\n")
		b.WriteString("  # camelCase, null-pruned keys with StringValueOrRef foreign keys resolved to\n")
		b.WriteString("  # literal strings, so no other transformation is needed.\n")
		fmt.Fprintf(&b, "  manifest_spec = { for k, v in var.spec : k => v if k != %q }\n", nsJSONName)
	} else {
		b.WriteString("  # Cluster-scoped CR: the converter already emits camelCase, null-pruned keys,\n")
		b.WriteString("  # so the spec is passed through unchanged.\n")
		b.WriteString("  manifest_spec = var.spec\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func manifestMainTF(resourceName, apiVersion, crdKind, nsJSONName string, namespaced bool) string {
	var metadata string
	if namespaced {
		metadata = fmt.Sprintf(`    metadata = {
      name      = var.metadata.name
      namespace = var.spec.%s
      labels    = local.labels
    }`, nsJSONName)
	} else {
		metadata = `    metadata = {
      name   = var.metadata.name
      labels = local.labels
    }`
	}
	return fmt.Sprintf(`# Applies the %s custom resource through kubectl_manifest (alekc/kubectl):
# no plan-time cluster dependency (plannable before the CRDs exist), applied
# server-side. No wait, deliberately: the CR is configuration its controller
# consumes; applying it server-side-validated is the whole contract. Pulumi
# equivalent: the typed CR without await annotations.
resource "kubectl_manifest" %q {
  yaml_body = yamlencode({
    apiVersion = %q
    kind       = %q
%s
    spec = local.manifest_spec
  })

  server_side_apply = true
}
`, crdKind, resourceName, apiVersion, crdKind, metadata)
}

func manifestOutputsTF(md, specMsg protoreflect.MessageDescriptor, crdKind, nsJSONName string, namespaced bool) (string, error) {
	statusMsg := messageOfField(md, "status")
	if statusMsg == nil {
		return "", nil
	}
	outputsMsg := messageOfField(statusMsg, "outputs")
	if outputsMsg == nil {
		return "", nil
	}

	var b strings.Builder
	fields := outputsMsg.Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		name := string(fd.Name())
		specJSONName := specFieldJSONName(specMsg, name)

		var value, desc string
		switch {
		case name == "namespace" && namespaced && specJSONName == "":
			// The resource's namespace identity (resolved spec.namespace), which
			// the converter routes to metadata.namespace, not into the CR spec.
			value = "var.spec." + nsJSONName
			desc = fmt.Sprintf("Namespace the %s was created in.", crdKind)
		case strings.HasSuffix(name, "_name") && specJSONName == "":
			// An identity-name output with no matching spec field is the resource
			// name (e.g. destination_rule_name, route_name) == metadata.name.
			value = "var.metadata.name"
			desc = fmt.Sprintf("Name of the created %s (equals metadata.name).", crdKind)
		case specJSONName != "":
			// Output backed by a top-level spec field (e.g. GatewayClass
			// controller_name). The projection spec is camelCase under var.spec.
			value = "var.spec." + specJSONName
			desc = fmt.Sprintf("%s of the created %s.", humanizeFieldName(name), crdKind)
		default:
			// No safe mapping: fail loudly rather than emit a wrong module.
			return "", errors.Errorf("cannot map stack output %q to a source for %s; "+
				"generate-module needs an explicit rule for it", name, crdKind)
		}
		fmt.Fprintf(&b, "output %q {\n  description = %q\n  value       = %s\n}\n\n", name, desc, value)
	}
	return strings.TrimRight(b.String(), "\n") + "\n", nil
}

// humanizeFieldName turns a snake_case field name into a capitalized phrase,
// e.g. "controller_name" -> "Controller name".
func humanizeFieldName(snake string) string {
	s := strings.ReplaceAll(snake, "_", " ")
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func specFieldJSONName(specMsg protoreflect.MessageDescriptor, snakeName string) string {
	if specMsg == nil {
		return ""
	}
	fd := specMsg.Fields().ByName(protoreflect.Name(snakeName))
	if fd == nil {
		return ""
	}
	return fd.JSONName()
}

// kebabFromPascal converts a PascalCase CRD kind to a kebab-case label,
// handling acronym runs: DestinationRule->destination-rule, HTTPRoute->http-route,
// GatewayClass->gateway-class, Telemetry->telemetry.
var (
	reKebabAcronym    = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
	reKebabLowerUpper = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

func kebabFromPascal(s string) string {
	s = reKebabAcronym.ReplaceAllString(s, "$1-$2")
	s = reKebabLowerUpper.ReplaceAllString(s, "$1-$2")
	return strings.ToLower(s)
}
