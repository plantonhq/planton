package module

import (
	"encoding/json"
	"strings"

	alicloudsaeapplicationv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudsaeapplication/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudSaeApplication *alicloudsaeapplicationv1.AliCloudSaeApplication
	Tags                   map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudsaeapplicationv1.AliCloudSaeApplicationStackInput) *Locals {
	locals := &Locals{}
	locals.AliCloudSaeApplication = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AliCloudSaeApplication.String()),
	}

	if target.Metadata.Id != "" {
		locals.Tags["resource_id"] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.Tags["organization"] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.Tags["environment"] = target.Metadata.Env
	}

	for k, v := range target.Spec.Tags {
		locals.Tags[k] = v
	}

	return locals
}

// envsToJSON converts a map of environment variables into the JSON array
// format that the SAE API expects: [{"name":"K","value":"V"},...].
// Returns an empty string when the map is nil or empty.
func envsToJSON(envs map[string]string) string {
	if len(envs) == 0 {
		return ""
	}

	type envEntry struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	entries := make([]envEntry, 0, len(envs))
	for k, v := range envs {
		entries = append(entries, envEntry{Name: k, Value: v})
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return ""
	}
	return string(data)
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}

func optionalInt(v *int32) pulumi.IntPtrInput {
	if v == nil {
		return nil
	}
	return pulumi.Int(int(*v))
}

func optionalBool(v *bool) pulumi.BoolPtrInput {
	if v == nil {
		return nil
	}
	return pulumi.Bool(*v)
}

func optionalStringPtr(v *string) pulumi.StringPtrInput {
	if v == nil || *v == "" {
		return nil
	}
	return pulumi.String(*v)
}
