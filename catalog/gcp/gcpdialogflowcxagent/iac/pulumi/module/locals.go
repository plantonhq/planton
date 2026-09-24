package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpdialogflowcxagentv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdialogflowcxagent/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// startFlowId is the id of the start flow every agent is created with; a
// version or an environment entry with an empty flow_id addresses it.
const startFlowId = "00000000-0000-0000-0000-000000000000"

// defaultPlaybook is the start playbook path Google accepts before the
// agent's id exists: "-" for project, location, and agent -- the form
// Google's own provider tests send. Google allows only the default
// playbook as a start playbook.
const defaultPlaybook = "projects/-/locations/-/agents/-/playbooks/00000000-0000-0000-0000-000000000000"

type Locals struct {
	GcpProviderConfig    *gcpprovider.GcpProviderConfig
	GcpDialogflowCxAgent *gcpdialogflowcxagentv1alpha1.GcpDialogflowCxAgent

	// DisplayName is spec.display_name when set, otherwise metadata.name --
	// Google requires one, and the Terraform module applies the same
	// fallback in locals.tf.
	DisplayName string
}

// initializeLocals derives the defaulted display name. Dialogflow CX
// resources carry no labels, so there is no attribution label set.
func initializeLocals(_ *pulumi.Context, stackInput *gcpdialogflowcxagentv1alpha1.GcpDialogflowCxAgentStackInput) *Locals {
	locals := &Locals{}
	locals.GcpDialogflowCxAgent = stackInput.Target

	locals.DisplayName = locals.GcpDialogflowCxAgent.Spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = locals.GcpDialogflowCxAgent.Metadata.Name
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}

// flowIdOrStart resolves an empty flow_id to the start flow -- the same
// rule the Terraform module's locals.tf applies.
func flowIdOrStart(flowId string) string {
	if flowId == "" {
		return startFlowId
	}
	return flowId
}

// deletionPolicy returns the spec's destroy stance for fanning to every
// folded child, or nil so the provider default stays in charge.
func deletionPolicy(locals *Locals) pulumi.StringPtrInput {
	if policy := locals.GcpDialogflowCxAgent.Spec.DeletionPolicy; policy != "" {
		return pulumi.String(policy)
	}
	return nil
}

// optionalString sends a string only when set -- the null-for-empty rule
// the Terraform module applies.
func optionalString(value string) pulumi.StringPtrInput {
	if value == "" {
		return nil
	}
	return pulumi.String(value)
}

// optionalTrue sends a boolean only when true (Google's default is false).
func optionalTrue(value bool) pulumi.BoolPtrInput {
	if !value {
		return nil
	}
	return pulumi.Bool(true)
}

// optionalSecret sends a credential only when set, marked secret so Pulumi
// encrypts it in state.
func optionalSecret(value string) pulumi.StringPtrInput {
	if value == "" {
		return nil
	}
	return pulumi.ToSecret(pulumi.String(value)).(pulumi.StringOutput)
}

// optionalStringArray sends a list only when it has entries.
func optionalStringArray(values []string) pulumi.StringArrayInput {
	if len(values) == 0 {
		return nil
	}
	return pulumi.ToStringArray(values)
}

// optionalStringMap sends a map only when it has entries.
func optionalStringMap(values map[string]string) pulumi.StringMapInput {
	if len(values) == 0 {
		return nil
	}
	return pulumi.ToStringMap(values)
}
