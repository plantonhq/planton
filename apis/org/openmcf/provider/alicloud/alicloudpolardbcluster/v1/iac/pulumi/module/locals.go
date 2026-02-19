package module

import (
	"strings"

	alicloudpolardbclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudpolardbcluster/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudPolardbCluster *alicloudpolardbclusterv1.AlicloudPolardbCluster
	Tags                   map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudpolardbclusterv1.AlicloudPolardbClusterStackInput) *Locals {
	locals := &Locals{}
	locals.AlicloudPolardbCluster = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AlicloudPolardbCluster.String()),
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

func clusterDescription(locals *Locals) string {
	if locals.AlicloudPolardbCluster.Spec.Description != "" {
		return locals.AlicloudPolardbCluster.Spec.Description
	}
	return locals.AlicloudPolardbCluster.Metadata.Name
}

func payType(spec *alicloudpolardbclusterv1.AlicloudPolardbClusterSpec) string {
	if spec.PayType != nil && *spec.PayType != "" {
		return *spec.PayType
	}
	return "PostPaid"
}

func dbNodeCount(spec *alicloudpolardbclusterv1.AlicloudPolardbClusterSpec) int {
	if spec.DbNodeCount != nil {
		return int(*spec.DbNodeCount)
	}
	return 2
}

func accountType(acct *alicloudpolardbclusterv1.AlicloudPolardbAccount) string {
	if acct.AccountType != nil && *acct.AccountType != "" {
		return *acct.AccountType
	}
	return "Normal"
}

func accountPrivilege(priv *alicloudpolardbclusterv1.AlicloudPolardbAccountPrivilege) string {
	if priv.AccountPrivilege != nil && *priv.AccountPrivilege != "" {
		return *priv.AccountPrivilege
	}
	return "ReadOnly"
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}

func optionalStringPtr(s *string) pulumi.StringPtrInput {
	if s == nil || *s == "" {
		return nil
	}
	return pulumi.String(*s)
}

func optionalIntPtr(i *int32) pulumi.IntPtrInput {
	if i == nil {
		return nil
	}
	return pulumi.Int(int(*i))
}
