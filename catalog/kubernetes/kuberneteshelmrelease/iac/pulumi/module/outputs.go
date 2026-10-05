package module

import (
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output constants define the keys for outputs exported by this
// module, mirroring KubernetesHelmReleaseOutputs. The Terraform module
// exports the identical set from the helm_release resource's metadata.
const (
	// OpNamespace is the namespace the release is installed in.
	OpNamespace = "namespace"
	// OpReleaseName is the Helm release name (`helm list` NAME column).
	OpReleaseName = "release_name"
	// OpVersion is the installed chart version.
	OpVersion = "version"
	// OpAppVersion is the chart's appVersion (the packaged application's
	// upstream version).
	OpAppVersion = "app_version"
	// OpStatus is the release status as Helm records it (e.g. "deployed").
	OpStatus = "status"
	// OpRevision is the release revision number (1 on install, incremented
	// by upgrades/rollbacks).
	OpRevision = "revision"
)

// exportOutputs exports the release's observable handles. Values come from
// the Release resource's status (populated by the provider after install),
// so they reflect what Helm actually recorded — not what the spec asked for.
func exportOutputs(ctx *pulumi.Context, locals *Locals, createdRelease *helmv3.Release) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpVersion, createdRelease.Status.Version().Elem())
	ctx.Export(OpAppVersion, createdRelease.Status.AppVersion().Elem())
	// Status() is a plain StringOutput at the pinned SDK (not a Ptr).
	ctx.Export(OpStatus, createdRelease.Status.Status())
	ctx.Export(OpRevision, createdRelease.Status.Revision().Elem())
}
