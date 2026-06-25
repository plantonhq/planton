package module

import (
	"github.com/pkg/errors"
	cloudflarelistitemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflarelistitem/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// listItem writes a single entry into the target Cloudflare List. Exactly one of
// the item shapes (ip / asn / hostname / redirect) is set, matching the list kind.
func listItem(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.ListItem, error) {
	spec := locals.CloudflareListItem.Spec

	listId := ""
	if spec.ListId != nil {
		listId = spec.ListId.GetValue()
	}

	args := &cloudflare.ListItemArgs{
		AccountId: pulumi.String(spec.AccountId),
		ListId:    pulumi.String(listId),
	}
	if spec.Comment != "" {
		args.Comment = pulumi.StringPtr(spec.Comment)
	}

	switch v := spec.Item.(type) {
	case *cloudflarelistitemv1.CloudflareListItemSpec_Ip:
		args.Ip = pulumi.StringPtr(v.Ip)
	case *cloudflarelistitemv1.CloudflareListItemSpec_Asn:
		args.Asn = pulumi.IntPtr(int(v.Asn))
	case *cloudflarelistitemv1.CloudflareListItemSpec_Hostname:
		h := v.Hostname
		hostnameArgs := &cloudflare.ListItemHostnameArgs{
			UrlHostname: pulumi.String(h.UrlHostname),
		}
		if h.ExcludeExactHostname != nil {
			hostnameArgs.ExcludeExactHostname = pulumi.BoolPtr(h.GetExcludeExactHostname())
		}
		args.Hostname = hostnameArgs
	case *cloudflarelistitemv1.CloudflareListItemSpec_Redirect:
		r := v.Redirect
		redirectArgs := &cloudflare.ListItemRedirectArgs{
			SourceUrl: pulumi.String(r.SourceUrl),
			TargetUrl: pulumi.String(r.TargetUrl),
		}
		if r.StatusCode != 0 {
			redirectArgs.StatusCode = pulumi.IntPtr(int(r.StatusCode))
		}
		// Omit false booleans for byte-for-byte parity with the Terraform module,
		// which sends null (provider default false) when unset.
		if r.IncludeSubdomains {
			redirectArgs.IncludeSubdomains = pulumi.BoolPtr(true)
		}
		if r.PreservePathSuffix {
			redirectArgs.PreservePathSuffix = pulumi.BoolPtr(true)
		}
		if r.PreserveQueryString {
			redirectArgs.PreserveQueryString = pulumi.BoolPtr(true)
		}
		if r.SubpathMatching {
			redirectArgs.SubpathMatching = pulumi.BoolPtr(true)
		}
		args.Redirect = redirectArgs
	}

	created, err := cloudflare.NewListItem(
		ctx,
		"list-item",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare list item")
	}

	ctx.Export(OpItemId, created.ID())
	ctx.Export(OpListId, pulumi.String(listId))

	return created, nil
}
