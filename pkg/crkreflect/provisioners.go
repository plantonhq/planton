package crkreflect

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// Provisioners returns the IaC engines one kind runs on (kind_meta.provisioners), in declared
// order. An empty result is the common case, not an absence of knowledge: the kind declares no
// restriction and runs on every engine it ships a module for. The declaration is carried as
// IacProvisioner value names because the kind proto cannot import the enum's package; a name that
// is not a real provisioner is an error here and a registry-test failure at build time.
func Provisioners(kind cloudresourcekind.CloudResourceKind) ([]shared.IacProvisioner, error) {
	kindMeta, err := KindMeta(kind)
	if err != nil {
		return nil, errors.Wrap(err, "while getting cloud resource kind meta")
	}
	declared := kindMeta.GetProvisioners()
	if len(declared) == 0 {
		return nil, nil
	}
	provisioners := make([]shared.IacProvisioner, 0, len(declared))
	for _, name := range declared {
		value, ok := shared.IacProvisioner_value[name]
		if !ok || shared.IacProvisioner(value) == shared.IacProvisioner_iac_provisioner_unspecified {
			return nil, errors.Errorf("%s declares provisioner %q, which is not an IacProvisioner (tofu, terraform, pulumi)", kind, name)
		}
		provisioners = append(provisioners, shared.IacProvisioner(value))
	}
	return provisioners, nil
}

// RunsOn reports whether one kind may run on the given engine: always, when the kind declares no
// provisioners; otherwise only when the engine is among those it declares.
func RunsOn(kind cloudresourcekind.CloudResourceKind, provisioner shared.IacProvisioner) (bool, error) {
	declared, err := Provisioners(kind)
	if err != nil {
		return false, err
	}
	if len(declared) == 0 {
		return true, nil
	}
	for _, p := range declared {
		if p == provisioner {
			return true, nil
		}
	}
	return false, nil
}
