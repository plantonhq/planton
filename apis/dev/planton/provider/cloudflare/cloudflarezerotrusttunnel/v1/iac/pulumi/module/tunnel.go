package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrusttunnelv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarezerotrusttunnel/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// tunnel provisions a Cloudflare Tunnel and, when remotely managed with ingress rules,
// its configuration. The connector run token is read from the token data source and
// exported (sensitive) so a downstream cloudflared runner can authenticate.
func tunnel(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareZeroTrustTunnel.Spec

	// config_src defaults to "cloudflare" (remote management); only "local" opts out.
	configSrc := "cloudflare"
	if spec.GetConfigSrc() == cloudflarezerotrusttunnelv1.CloudflareZeroTrustTunnelConfigSource_local {
		configSrc = "local"
	}

	tunnelArgs := &cloudflare.ZeroTrustTunnelCloudflaredArgs{
		AccountId: pulumi.String(spec.AccountId),
		Name:      pulumi.String(spec.Name),
		ConfigSrc: pulumi.String(configSrc),
	}
	if spec.TunnelSecret != "" {
		tunnelArgs.TunnelSecret = pulumi.String(spec.TunnelSecret)
	}

	createdTunnel, err := cloudflare.NewZeroTrustTunnelCloudflared(
		ctx,
		"tunnel",
		tunnelArgs,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudflare tunnel")
	}

	// Remote ingress configuration is provisioned as its own resource, so editing it
	// never recreates the tunnel. Only applies to a cloudflare-managed tunnel with rules.
	if configSrc == "cloudflare" && len(spec.Ingress) > 0 {
		configArgs := &cloudflare.ZeroTrustTunnelCloudflaredConfigArgs{
			AccountId: pulumi.String(spec.AccountId),
			TunnelId:  createdTunnel.ID().ToStringOutput(),
			Source:    pulumi.String("cloudflare"),
			Config: &cloudflare.ZeroTrustTunnelCloudflaredConfigConfigArgs{
				Ingresses:     buildIngresses(spec.Ingress),
				OriginRequest: buildConfigOriginRequest(spec.OriginRequest),
			},
		}

		if _, err := cloudflare.NewZeroTrustTunnelCloudflaredConfig(
			ctx,
			"tunnel-config",
			configArgs,
			pulumi.Provider(cloudflareProvider),
			pulumi.Parent(createdTunnel),
		); err != nil {
			return errors.Wrap(err, "failed to configure cloudflare tunnel ingress")
		}
	}

	// The connector run token is exposed by a data source, not the resource itself.
	tokenResult := cloudflare.GetZeroTrustTunnelCloudflaredTokenOutput(
		ctx,
		cloudflare.GetZeroTrustTunnelCloudflaredTokenOutputArgs{
			AccountId: pulumi.String(spec.AccountId),
			TunnelId:  createdTunnel.ID().ToStringOutput(),
		},
		pulumi.Provider(cloudflareProvider),
	)

	tunnelCname := createdTunnel.ID().ApplyT(func(id string) string {
		return id + ".cfargotunnel.com"
	}).(pulumi.StringOutput)

	ctx.Export(OpTunnelId, createdTunnel.ID())
	ctx.Export(OpTunnelCname, tunnelCname)
	ctx.Export(OpTunnelToken, pulumi.ToSecret(tokenResult.Token()))
	ctx.Export(OpTunnelStatus, createdTunnel.Status)
	ctx.Export(OpAccountTag, createdTunnel.AccountTag)
	ctx.Export(OpCreatedOn, createdTunnel.CreatedAt)

	return nil
}

// buildIngresses maps the spec ingress rules to provider ingress args, preserving order
// (the final rule is the catch-all).
func buildIngresses(rules []*cloudflarezerotrusttunnelv1.CloudflareZeroTrustTunnelIngressRule) cloudflare.ZeroTrustTunnelCloudflaredConfigConfigIngressArray {
	out := make(cloudflare.ZeroTrustTunnelCloudflaredConfigConfigIngressArray, 0, len(rules))
	for _, r := range rules {
		ingress := cloudflare.ZeroTrustTunnelCloudflaredConfigConfigIngressArgs{
			Service: pulumi.String(r.Service),
		}
		if r.Hostname != "" {
			ingress.Hostname = pulumi.String(r.Hostname)
		}
		if r.Path != "" {
			ingress.Path = pulumi.String(r.Path)
		}
		if r.OriginRequest != nil {
			ingress.OriginRequest = buildIngressOriginRequest(r.OriginRequest)
		}
		out = append(out, ingress)
	}
	return out
}

// buildIngressOriginRequest maps origin settings for a single ingress rule.
func buildIngressOriginRequest(or *cloudflarezerotrusttunnelv1.CloudflareZeroTrustTunnelOriginRequest) cloudflare.ZeroTrustTunnelCloudflaredConfigConfigIngressOriginRequestPtrInput {
	args := &cloudflare.ZeroTrustTunnelCloudflaredConfigConfigIngressOriginRequestArgs{}
	if or.CaPool != "" {
		args.CaPool = pulumi.String(or.CaPool)
	}
	if or.ConnectTimeout > 0 {
		args.ConnectTimeout = pulumi.Int(int(or.ConnectTimeout))
	}
	if or.DisableChunkedEncoding {
		args.DisableChunkedEncoding = pulumi.Bool(true)
	}
	if or.Http2Origin {
		args.Http2Origin = pulumi.Bool(true)
	}
	if or.HttpHostHeader != "" {
		args.HttpHostHeader = pulumi.String(or.HttpHostHeader)
	}
	if or.KeepAliveConnections > 0 {
		args.KeepAliveConnections = pulumi.Int(int(or.KeepAliveConnections))
	}
	if or.KeepAliveTimeout > 0 {
		args.KeepAliveTimeout = pulumi.Int(int(or.KeepAliveTimeout))
	}
	if or.MatchSniToHost {
		args.MatchSnItoHost = pulumi.Bool(true)
	}
	if or.NoHappyEyeballs {
		args.NoHappyEyeballs = pulumi.Bool(true)
	}
	if or.NoTlsVerify {
		args.NoTlsVerify = pulumi.Bool(true)
	}
	if or.OriginServerName != "" {
		args.OriginServerName = pulumi.String(or.OriginServerName)
	}
	if or.ProxyType != "" {
		args.ProxyType = pulumi.String(or.ProxyType)
	}
	if or.TcpKeepAlive > 0 {
		args.TcpKeepAlive = pulumi.Int(int(or.TcpKeepAlive))
	}
	if or.TlsTimeout > 0 {
		args.TlsTimeout = pulumi.Int(int(or.TlsTimeout))
	}
	if or.Access != nil {
		args.Access = &cloudflare.ZeroTrustTunnelCloudflaredConfigConfigIngressOriginRequestAccessArgs{
			AudTags:  audTags(or.Access),
			TeamName: pulumi.String(or.Access.TeamName),
			Required: pulumi.Bool(or.Access.Required),
		}
	}
	return args
}

// buildConfigOriginRequest maps the tunnel-level origin defaults.
func buildConfigOriginRequest(or *cloudflarezerotrusttunnelv1.CloudflareZeroTrustTunnelOriginRequest) cloudflare.ZeroTrustTunnelCloudflaredConfigConfigOriginRequestPtrInput {
	if or == nil {
		return nil
	}
	args := &cloudflare.ZeroTrustTunnelCloudflaredConfigConfigOriginRequestArgs{}
	if or.CaPool != "" {
		args.CaPool = pulumi.String(or.CaPool)
	}
	if or.ConnectTimeout > 0 {
		args.ConnectTimeout = pulumi.Int(int(or.ConnectTimeout))
	}
	if or.DisableChunkedEncoding {
		args.DisableChunkedEncoding = pulumi.Bool(true)
	}
	if or.Http2Origin {
		args.Http2Origin = pulumi.Bool(true)
	}
	if or.HttpHostHeader != "" {
		args.HttpHostHeader = pulumi.String(or.HttpHostHeader)
	}
	if or.KeepAliveConnections > 0 {
		args.KeepAliveConnections = pulumi.Int(int(or.KeepAliveConnections))
	}
	if or.KeepAliveTimeout > 0 {
		args.KeepAliveTimeout = pulumi.Int(int(or.KeepAliveTimeout))
	}
	if or.MatchSniToHost {
		args.MatchSnItoHost = pulumi.Bool(true)
	}
	if or.NoHappyEyeballs {
		args.NoHappyEyeballs = pulumi.Bool(true)
	}
	if or.NoTlsVerify {
		args.NoTlsVerify = pulumi.Bool(true)
	}
	if or.OriginServerName != "" {
		args.OriginServerName = pulumi.String(or.OriginServerName)
	}
	if or.ProxyType != "" {
		args.ProxyType = pulumi.String(or.ProxyType)
	}
	if or.TcpKeepAlive > 0 {
		args.TcpKeepAlive = pulumi.Int(int(or.TcpKeepAlive))
	}
	if or.TlsTimeout > 0 {
		args.TlsTimeout = pulumi.Int(int(or.TlsTimeout))
	}
	if or.Access != nil {
		args.Access = &cloudflare.ZeroTrustTunnelCloudflaredConfigConfigOriginRequestAccessArgs{
			AudTags:  audTags(or.Access),
			TeamName: pulumi.String(or.Access.TeamName),
			Required: pulumi.Bool(or.Access.Required),
		}
	}
	return args
}

// audTags flattens the StringValueOrRef audience tags to a plain string array.
func audTags(access *cloudflarezerotrusttunnelv1.CloudflareZeroTrustTunnelAccessConfig) pulumi.StringArray {
	tags := make(pulumi.StringArray, 0, len(access.AudTag))
	for _, t := range access.AudTag {
		tags = append(tags, pulumi.String(t.GetValue()))
	}
	return tags
}
