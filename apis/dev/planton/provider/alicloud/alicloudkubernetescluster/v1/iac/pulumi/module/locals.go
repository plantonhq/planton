package module

import (
	"strings"

	alicloudkubernetesclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudkubernetescluster/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudKubernetesCluster *alicloudkubernetesclusterv1.AliCloudKubernetesCluster
	Tags                      map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *alicloudkubernetesclusterv1.AliCloudKubernetesClusterStackInput) *Locals {
	locals := &Locals{}
	locals.AliCloudKubernetesCluster = stackInput.Target
	target := stackInput.Target

	locals.Tags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AliCloudKubernetesCluster.String()),
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

func clusterSpec(spec *alicloudkubernetesclusterv1.AliCloudKubernetesClusterSpec) string {
	if spec.ClusterSpec != nil {
		return *spec.ClusterSpec
	}
	return "ack.standard"
}

func proxyMode(spec *alicloudkubernetesclusterv1.AliCloudKubernetesClusterSpec) string {
	if spec.ProxyMode != nil {
		return *spec.ProxyMode
	}
	return "ipvs"
}

func nodeCidrMask(spec *alicloudkubernetesclusterv1.AliCloudKubernetesClusterSpec) int {
	if spec.NodeCidrMask != nil {
		return int(*spec.NodeCidrMask)
	}
	return 24
}

func newNatGateway(spec *alicloudkubernetesclusterv1.AliCloudKubernetesClusterSpec) bool {
	if spec.NewNatGateway != nil {
		return *spec.NewNatGateway
	}
	return true
}

func slbInternetEnabled(spec *alicloudkubernetesclusterv1.AliCloudKubernetesClusterSpec) bool {
	if spec.SlbInternetEnabled != nil {
		return *spec.SlbInternetEnabled
	}
	return true
}

func enableRrsa(spec *alicloudkubernetesclusterv1.AliCloudKubernetesClusterSpec) bool {
	if spec.EnableRrsa != nil {
		return *spec.EnableRrsa
	}
	return false
}

func isEnterpriseSg(spec *alicloudkubernetesclusterv1.AliCloudKubernetesClusterSpec) bool {
	if spec.IsEnterpriseSecurityGroup != nil {
		return *spec.IsEnterpriseSecurityGroup
	}
	return false
}

func deletionProtection(spec *alicloudkubernetesclusterv1.AliCloudKubernetesClusterSpec) bool {
	if spec.DeletionProtection != nil {
		return *spec.DeletionProtection
	}
	return false
}

func controlPlaneLogTtl(logging *alicloudkubernetesclusterv1.AliCloudKubernetesClusterLogging) string {
	if logging.ControlPlaneLogTtl != nil {
		return *logging.ControlPlaneLogTtl
	}
	return "30"
}
