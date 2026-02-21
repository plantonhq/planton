package module

import (
	"fmt"

	"github.com/pkg/errors"
	ocifilesystemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocifilesystem/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/filestorage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func exports(ctx *pulumi.Context, locals *Locals, provider *oci.Provider,
	fs *filestorage.FileSystem, mt *filestorage.MountTarget) error {

	for _, exp := range locals.OciFileSystem.Spec.Exports {
		exportArgs := &filestorage.ExportArgs{
			ExportSetId:  mt.ExportSetId,
			FileSystemId: fs.ID(),
			Path:         pulumi.String(exp.Path),
		}

		if len(exp.ExportOptions) > 0 {
			exportArgs.ExportOptions = buildExportOptions(exp.ExportOptions)
		}

		resourceName := fmt.Sprintf("%s-export%s", locals.DisplayName, exp.Path)
		_, err := filestorage.NewExport(ctx, resourceName, exportArgs,
			pulumiOciOpt(provider),
			pulumi.DependsOn([]pulumi.Resource{fs, mt}),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create export for path %s", exp.Path)
		}
	}

	return nil
}

func buildExportOptions(options []*ocifilesystemv1.OciFileSystemSpec_ExportOption) filestorage.ExportExportOptionArray {
	result := make(filestorage.ExportExportOptionArray, len(options))
	for i, opt := range options {
		args := filestorage.ExportExportOptionArgs{
			Source: pulumi.String(opt.Source),
		}

		if v, ok := accessMap[opt.Access]; ok {
			args.Access = pulumi.StringPtr(v)
		}

		if v, ok := identitySquashMap[opt.IdentitySquash]; ok {
			args.IdentitySquash = pulumi.StringPtr(v)
		}

		if opt.RequirePrivilegedSourcePort {
			args.RequirePrivilegedSourcePort = pulumi.BoolPtr(true)
		}

		if opt.IsAnonymousAccessAllowed {
			args.IsAnonymousAccessAllowed = pulumi.BoolPtr(true)
		}

		if opt.AnonymousUid != 0 {
			args.AnonymousUid = pulumi.StringPtr(fmt.Sprintf("%d", opt.AnonymousUid))
		}

		if opt.AnonymousGid != 0 {
			args.AnonymousGid = pulumi.StringPtr(fmt.Sprintf("%d", opt.AnonymousGid))
		}

		result[i] = args
	}
	return result
}
