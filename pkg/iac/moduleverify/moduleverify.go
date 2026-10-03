//go:build !codegen
// +build !codegen

// Package moduleverify proves that an IaC module directory conforms to the
// contract of a catalog kind — the confidence gate behind customizing
// the module that deploys a catalog kind.
//
// The checks are anchored to the kind's schema (through the same generator
// and transformer machinery the deploy path uses), never to the official
// module's exact text: a customized module is allowed to diverge in
// everything except the surfaces the platform feeds and reads. Severity
// follows what actually breaks a deployment — a mismatch that fails every
// deploy is an error; one that fails only for particular configurations, or
// merely goes unused, is a warning. A module that would print a secret the
// schema declares (a generated credential exported in the clear) is an error
// too: it leaks on every deployment.
package moduleverify

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// Severity classifies a violation by deployment impact.
type Severity string

const (
	// SeverityError marks a contract break that fails deployments.
	SeverityError Severity = "error"
	// SeverityWarning marks a divergence that fails only particular
	// configurations, or leaves part of the contract unused.
	SeverityWarning Severity = "warning"
)

// Violation is one finding: where, how bad, and — in one sentence — what is
// wrong and the way out.
type Violation struct {
	Severity Severity
	// File is the finding's location relative to the module directory
	// ("" for module-level findings).
	File string
	// Summary names the cause and the fix in plain language.
	Summary string
}

// Result reports a verification run.
type Result struct {
	KindName    string
	Provisioner provisioner.ProvisionerType
	ModuleDir   string
	Violations  []Violation
	// Notices records checks that were skipped and why (e.g. no toolchain
	// on PATH) — a skipped check is reported, never silent.
	Notices []string
}

// HasErrors reports whether any violation is deployment-breaking.
func (r *Result) HasErrors() bool {
	for _, v := range r.Violations {
		if v.Severity == SeverityError {
			return true
		}
	}
	return false
}

// Input configures a verification run.
type Input struct {
	// KindName is the catalog kind whose contract the module must
	// honor. Tolerant forms accepted (AwsS3Bucket, awss3bucket, aws-s3-bucket).
	KindName string

	// ModuleDir is the module directory to verify.
	ModuleDir string

	// Provisioner names the module's engine. Unspecified means infer from
	// the directory's shape.
	Provisioner provisioner.ProvisionerType

	// SampleOutputs, when provided, drives a dry-run of the outputs
	// transformation with representative raw outputs.
	SampleOutputs map[string]interface{}

	// SkipToolchainChecks disables the checks that shell out to the
	// engine's toolchain (tofu validate / go build). The static contract
	// checks always run.
	SkipToolchainChecks bool
}

// Verify runs every applicable contract check. A returned error means the
// run itself could not happen (unknown kind, unreadable directory);
// conformance findings are reported through Result.Violations.
func Verify(in Input) (*Result, error) {
	kind := catalogkindreflect.KindFromString(in.KindName)
	if kind == catalogkind.CatalogKind_unspecified {
		return nil, errors.Errorf("unknown catalog kind %q — kind names follow the catalog (e.g. AwsS3Bucket)", in.KindName)
	}
	kindName := catalogkindreflect.ExtractKindNameByKind(kind)

	moduleDir, err := filepath.Abs(in.ModuleDir)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve absolute path for %s", in.ModuleDir)
	}
	if info, err := os.Stat(moduleDir); err != nil || !info.IsDir() {
		return nil, errors.Errorf("the module directory %s does not exist or is not a directory", moduleDir)
	}

	prov := in.Provisioner
	if prov == provisioner.ProvisionerTypeUnspecified {
		prov, err = inferProvisioner(moduleDir)
		if err != nil {
			return nil, err
		}
	}
	// A module for an engine the kind does not run on is not a module of that
	// kind, however well it satisfies the contract otherwise.
	if err := provisioner.Require(kind, prov); err != nil {
		return nil, err
	}

	result := &Result{
		KindName:    kindName,
		Provisioner: prov,
		ModuleDir:   moduleDir,
	}

	if prov == provisioner.ProvisionerTypePulumi {
		verifyPulumi(kind, kindName, moduleDir, in, result)
	} else {
		verifyTofu(kind, kindName, moduleDir, in, result)
	}

	return result, nil
}

// inferProvisioner reads the module's engine off the directory shape, the
// same signals the deploy-path resolvers key on.
func inferProvisioner(moduleDir string) (provisioner.ProvisionerType, error) {
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		return provisioner.ProvisionerTypeUnspecified, errors.Wrapf(err, "cannot read the module directory %s", moduleDir)
	}

	hasTf, hasPulumi := false, false
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		switch {
		case strings.HasSuffix(entry.Name(), ".tf"):
			hasTf = true
		case entry.Name() == "Pulumi.yaml" || strings.HasSuffix(entry.Name(), ".go"):
			hasPulumi = true
		}
	}

	switch {
	case hasTf && hasPulumi:
		return provisioner.ProvisionerTypeUnspecified, errors.Errorf("the directory %s holds both .tf and Go/Pulumi files — name the engine explicitly with --provisioner", moduleDir)
	case hasTf:
		return provisioner.ProvisionerTypeTofu, nil
	case hasPulumi:
		return provisioner.ProvisionerTypePulumi, nil
	default:
		return provisioner.ProvisionerTypeUnspecified, errors.Errorf("the directory %s does not look like an IaC module — an OpenTofu/Terraform module holds .tf files, a Pulumi module holds Pulumi.yaml and Go sources", moduleDir)
	}
}

func (r *Result) addError(file, summary string) {
	r.Violations = append(r.Violations, Violation{Severity: SeverityError, File: file, Summary: summary})
}

func (r *Result) addWarning(file, summary string) {
	r.Violations = append(r.Violations, Violation{Severity: SeverityWarning, File: file, Summary: summary})
}

func (r *Result) addNotice(notice string) {
	r.Notices = append(r.Notices, notice)
}
