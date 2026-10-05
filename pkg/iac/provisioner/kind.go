package provisioner

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/proto"
)

// A kind may declare the engines it runs on (kind_meta.provisioners): a provider with no Pulumi
// provider, or kinds proven on OpenTofu alone. This file is the one place that turns that
// declaration into a decision, so every surface that picks or runs an engine -- the CLI's
// resolution, the multi-manifest preflight, the engine entry points, module tooling -- resolves
// and refuses in the same words.

// ModuleFamily names the module directory this engine runs, the way module lookups key it:
// pulumi for iac/pulumi, and terraform for iac/tf, the one HCL module OpenTofu and Terraform
// both run. It is not the engine: a lookup handed tofu would find nothing.
func (p ProvisionerType) ModuleFamily() shared.IacProvisioner {
	switch p {
	case ProvisionerTypePulumi:
		return shared.IacProvisioner_pulumi
	case ProvisionerTypeTofu, ProvisionerTypeTerraform:
		return shared.IacProvisioner_terraform
	default:
		return shared.IacProvisioner_iac_provisioner_unspecified
	}
}

// DisplayName is the engine's product name, as refusals and prompts spell it.
func (p ProvisionerType) DisplayName() string {
	switch p {
	case ProvisionerTypePulumi:
		return "Pulumi"
	case ProvisionerTypeTofu:
		return "OpenTofu"
	case ProvisionerTypeTerraform:
		return "Terraform"
	default:
		return "an unspecified engine"
	}
}

func fromProto(p shared.IacProvisioner) ProvisionerType {
	switch p {
	case shared.IacProvisioner_pulumi:
		return ProvisionerTypePulumi
	case shared.IacProvisioner_tofu:
		return ProvisionerTypeTofu
	case shared.IacProvisioner_terraform:
		return ProvisionerTypeTerraform
	default:
		return ProvisionerTypeUnspecified
	}
}

// everyEngine is the order an undeclared kind offers its engines in, Pulumi first as the CLI
// has always defaulted.
var everyEngine = []ProvisionerType{ProvisionerTypePulumi, ProvisionerTypeTofu, ProvisionerTypeTerraform}

// Allowed returns the engines a kind runs on, in declared order; every engine when the kind
// declares none (or is not a registered kind -- only a declaration narrows).
func Allowed(kind catalogkind.CatalogKind) ([]ProvisionerType, error) {
	if kind == catalogkind.CatalogKind_unspecified {
		return everyEngine, nil
	}
	declared, err := catalogkindreflect.Provisioners(kind)
	if err != nil {
		return nil, err
	}
	if len(declared) == 0 {
		return everyEngine, nil
	}
	allowed := make([]ProvisionerType, 0, len(declared))
	for _, p := range declared {
		allowed = append(allowed, fromProto(p))
	}
	return allowed, nil
}

// Require refuses an engine the kind does not run on, before anything runs. A kind that declares
// nothing accepts every engine.
func Require(kind catalogkind.CatalogKind, p ProvisionerType) error {
	allowed, err := Allowed(kind)
	if err != nil {
		return err
	}
	for _, a := range allowed {
		if a == p {
			return nil
		}
	}
	names := make([]string, 0, len(allowed))
	values := make([]string, 0, len(allowed))
	for _, a := range allowed {
		names = append(names, a.DisplayName())
		values = append(values, a.String())
	}
	return errors.Errorf("%s runs on %s only, so %s cannot deploy it: set planton.dev/provisioner to %s, or leave it unset",
		kind, strings.Join(names, " or "), p.DisplayName(), strings.Join(values, " or "))
}

// RequireForKindName is Require for callers that hold the kind's name (a manifest's kind field,
// a kind folder name); a name that resolves to no registered kind is not narrowed.
func RequireForKindName(kindName string, p ProvisionerType) error {
	return Require(catalogkindreflect.KindFromString(kindName), p)
}

// RequireForManifest is Require for callers that hold the manifest.
func RequireForManifest(manifest proto.Message, p ProvisionerType) error {
	return Require(kindOf(manifest), p)
}

// ForManifest resolves the engine a manifest runs on: its planton.dev/provisioner annotation when
// set (refused if the kind does not run on it); else the kind's sole declared engine, used without
// asking; else ProvisionerTypeUnspecified, and the caller asks among AllowedForManifest.
func ForManifest(manifest proto.Message) (ProvisionerType, error) {
	annotated, err := ExtractFromManifest(manifest)
	if err != nil {
		return ProvisionerTypeUnspecified, err
	}
	kind := kindOf(manifest)
	if annotated != ProvisionerTypeUnspecified {
		if err := Require(kind, annotated); err != nil {
			return ProvisionerTypeUnspecified, err
		}
		return annotated, nil
	}
	allowed, err := Allowed(kind)
	if err != nil {
		return ProvisionerTypeUnspecified, err
	}
	if len(allowed) == 1 {
		return allowed[0], nil
	}
	return ProvisionerTypeUnspecified, nil
}

// AllowedForManifest is Allowed for the manifest's kind: the choices a prompt may offer.
func AllowedForManifest(manifest proto.Message) ([]ProvisionerType, error) {
	return Allowed(kindOf(manifest))
}

// SoleEngineNote is the line the CLI prints when it chose the engine from the kind's declaration
// rather than the manifest, or "" when the manifest named it.
func SoleEngineNote(manifest proto.Message, resolved ProvisionerType) string {
	if annotated, _ := ExtractFromManifest(manifest); annotated != ProvisionerTypeUnspecified {
		return ""
	}
	return fmt.Sprintf("%s runs on %s only", kindOf(manifest), resolved.DisplayName())
}

func kindOf(manifest proto.Message) catalogkind.CatalogKind {
	if manifest == nil {
		return catalogkind.CatalogKind_unspecified
	}
	name, err := catalogkindreflect.ExtractKindFromProto(manifest)
	if err != nil {
		return catalogkind.CatalogKind_unspecified
	}
	return catalogkindreflect.KindFromString(name)
}
