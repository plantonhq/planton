package runner

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
)

// requireLaneEngine refuses a lane whose engine the kind does not run on. The
// framework's "terraform" engine runs TerraformBinary (tofu by default), so the check is made
// against the binary that would actually run, not the engine's name.
func requireLaneEngine(kindDir, engine string) error {
	var p provisioner.ProvisionerType
	switch engine {
	case "pulumi":
		p = provisioner.ProvisionerTypePulumi
	case "terraform":
		binary, err := provisioner.FromString(TerraformBinary())
		if err != nil {
			return errors.Wrap(err, "PLANTON_E2E_TF_BINARY")
		}
		p = binary
	default:
		return errors.Errorf("unknown E2E engine %q: lanes run pulumi or terraform", engine)
	}
	return provisioner.RequireForKindName(kindDir, p)
}
