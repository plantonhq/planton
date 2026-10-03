package eject

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
)

// writeContractNotes generates the notes file that travels with every
// ejected module. The audience is the engineer who finds this directory in
// their own repository months from now: the notes must explain the contract
// the module serves, how to prove a customization still honors it, and how
// the module runs — without assuming any context beyond the directory
// itself.
func writeContractNotes(outputDir, kindName string, in Input, result *Result) (string, error) {
	var content string
	if in.Provisioner == provisioner.ProvisionerTypePulumi {
		content = pulumiNotes(kindName, result.SourceVersion, in.GoModulePath)
	} else {
		content = tofuNotes(kindName, result.SourceVersion)
	}

	notesPath := filepath.Join(outputDir, NotesFileName)
	if err := os.WriteFile(notesPath, []byte(content), 0644); err != nil {
		return "", errors.Wrapf(err, "failed to write %s", notesPath)
	}
	return notesPath, nil
}

func tofuNotes(kindName, sourceVersion string) string {
	return fmt.Sprintf(`# %[1]s — Customized OpenTofu/Terraform Module

This directory is a customizable copy of the official OpenTofu/Terraform
module for the %[1]s infra component, ejected from release %[2]s of the
Planton catalog. Edit it freely — it is yours now. The one thing that must
survive every customization is the contract below.

## The contract

The platform invokes this module with generated variables, and reads its
results back through a typed outputs schema:

- **Inputs.** Two variables, both declared in variables.tf: metadata (the
  resource's name, identifiers, labels, and tags) and spec (the %[1]s
  configuration, mirroring the resource's spec schema field for field). The
  caller renders values for populated fields only, so every attribute that
  can be absent must be declared optional() with a default — a new required
  attribute the schema cannot supply will fail every deployment.
- **Outputs.** Values declared in outputs.tf are matched by name against the
  %[1]s outputs schema. An output the schema does not know is ignored;
  a schema field no output populates stays empty on the deployed resource.

## Prove a customization still conforms

    planton module verify --kind %[1]s --module-dir .

Run it after every meaningful change. It checks the input surface against
the %[1]s schema and validates the outputs contract, and it names exactly
what broke when something does.

## Run it locally

    planton tofu plan --manifest <your-manifest.yaml> --module-dir .
    planton tofu apply --manifest <your-manifest.yaml> --module-dir .

The commands behave exactly like a deployment of the official module, with
this directory substituted in.

## Use it for platform deployments

Push this directory to a git repository your platform organization can
reach, then register it as the module for %[1]s deployments in your
organization's settings. Deployments pick it up from git; nothing about the
resource manifests changes.

## License

This module derives from the Apache-2.0-licensed Planton catalog
(https://github.com/plantonhq/planton). Keep the LICENSE and NOTICE files
alongside any copy you distribute.
`, kindName, sourceVersion)
}

func pulumiNotes(kindName, sourceVersion, goModulePath string) string {
	return fmt.Sprintf(`# %[1]s — Customized Pulumi Module

This directory is a customizable copy of the official Pulumi module for the
%[1]s infra component, ejected from release %[2]s of the Planton catalog.
Edit it freely — it is yours now. The one thing that must survive every
customization is the contract below.

## The contract

- **Entrypoint.** main.go is a Go Pulumi program: it loads a typed
  %[1]sIacInput (the resource manifest plus deployment context) through
  the iac-input loader and hands it to the module package. Keep that
  shape — the input type and its loading are how every deployment feeds
  this module.
- **Runtime.** Pulumi.yaml declares the go runtime. Deployments build this
  module from source with the Go toolchain; other runtimes are not
  supported for customized modules.
- **Outputs.** Values exported on the Pulumi context are matched by name
  against the %[1]s outputs schema. An export the schema does not
  know is ignored; a schema field no export populates stays empty on the
  deployed resource.

## The Go module

This copy declares its own Go module, %[3]s, with the catalog
(github.com/plantonhq/planton) as a dependency — that is where the %[1]s
stub types and the iac-input loader come from. If dependencies have not
been resolved yet, run:

    go mod tidy

To rename the module path later, change the module line in go.mod and
update the matching self-import of the module subpackage in main.go.

## Prove a customization still conforms

    planton module verify --kind %[1]s --module-dir .

Run it after every meaningful change. It checks the module's shape and
entrypoint against the %[1]s contract, and it names exactly what broke when
something does.

## Run it locally

    planton pulumi preview --manifest <your-manifest.yaml> --module-dir .
    planton pulumi update --manifest <your-manifest.yaml> --module-dir .

The commands behave exactly like a deployment of the official module, with
this directory substituted in.

## Use it for platform deployments

Push this directory to a git repository your platform organization can
reach, then register it as the module for %[1]s deployments in your
organization's settings. Deployments pick it up from git; nothing about the
resource manifests changes.

## License

This module derives from the Apache-2.0-licensed Planton catalog
(https://github.com/plantonhq/planton). Keep the LICENSE and NOTICE files
alongside any copy you distribute.
`, kindName, sourceVersion, goModulePath)
}
