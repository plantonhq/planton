package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrustaccessgroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflarezerotrustaccessgroup/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// group provisions the Cloudflare Zero Trust Access group and exports its ID.
func group(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.ZeroTrustAccessGroup, error) {
	spec := locals.CloudflareZeroTrustAccessGroup.Spec

	args := &cloudflare.ZeroTrustAccessGroupArgs{
		Name:      pulumi.String(spec.Name),
		Includes:  groupIncludes(ctx, spec.Include),
		IsDefault: pulumi.BoolPtr(spec.IsDefault),
	}
	if spec.AccountId != "" {
		args.AccountId = pulumi.StringPtr(spec.AccountId)
	}
	if spec.ZoneId != nil && spec.ZoneId.GetValue() != "" {
		args.ZoneId = pulumi.StringPtr(spec.ZoneId.GetValue())
	}
	if len(spec.Exclude) > 0 {
		args.Excludes = groupExcludes(ctx, spec.Exclude)
	}
	if len(spec.Require) > 0 {
		args.Requires = groupRequires(ctx, spec.Require)
	}

	created, err := cloudflare.NewZeroTrustAccessGroup(
		ctx,
		"group",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare zero trust access group")
	}

	ctx.Export(OpGroupId, created.ID())

	return created, nil
}

// riskLevels converts the proto user-risk enum slice to the provider's strings.
func riskLevels(levels []cloudflarezerotrustaccessgroupv1.AccessRuleUserRiskScore_Level) pulumi.StringArray {
	out := make(pulumi.StringArray, 0, len(levels))
	for _, l := range levels {
		out = append(out, pulumi.String(l.String()))
	}
	return out
}

// groupIncludes maps the access rules onto the provider's include array.
func groupIncludes(ctx *pulumi.Context, rules []*cloudflarezerotrustaccessgroupv1.CloudflareAccessRule) cloudflare.ZeroTrustAccessGroupIncludeArray {
	out := cloudflare.ZeroTrustAccessGroupIncludeArray{}
	for _, r := range rules {
		e := &cloudflare.ZeroTrustAccessGroupIncludeArgs{}
		switch v := r.Rule.(type) {
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Email:
			e.Email = &cloudflare.ZeroTrustAccessGroupIncludeEmailArgs{Email: pulumi.String(v.Email.Email)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_EmailDomain:
			e.EmailDomain = &cloudflare.ZeroTrustAccessGroupIncludeEmailDomainArgs{Domain: pulumi.String(v.EmailDomain.Domain)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_EmailList:
			e.EmailList = &cloudflare.ZeroTrustAccessGroupIncludeEmailListArgs{Id: pulumi.String(v.EmailList.Id.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Everyone:
			e.Everyone = &cloudflare.ZeroTrustAccessGroupIncludeEveryoneArgs{}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Ip:
			e.Ip = &cloudflare.ZeroTrustAccessGroupIncludeIpArgs{Ip: pulumi.String(v.Ip.Ip)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_IpList:
			e.IpList = &cloudflare.ZeroTrustAccessGroupIncludeIpListArgs{Id: pulumi.String(v.IpList.Id.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Certificate:
			e.Certificate = &cloudflare.ZeroTrustAccessGroupIncludeCertificateArgs{}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Group:
			e.Group = &cloudflare.ZeroTrustAccessGroupIncludeGroupArgs{Id: pulumi.String(v.Group.Id.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_AzureAd:
			e.AzureAd = &cloudflare.ZeroTrustAccessGroupIncludeAzureAdArgs{Id: pulumi.String(v.AzureAd.Id), IdentityProviderId: pulumi.String(v.AzureAd.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_GithubOrganization:
			ga := &cloudflare.ZeroTrustAccessGroupIncludeGithubOrganizationArgs{IdentityProviderId: pulumi.String(v.GithubOrganization.IdentityProviderId.GetValue()), Name: pulumi.String(v.GithubOrganization.Name)}
			if v.GithubOrganization.Team != "" {
				ga.Team = pulumi.StringPtr(v.GithubOrganization.Team)
			}
			e.GithubOrganization = ga
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Gsuite:
			e.Gsuite = &cloudflare.ZeroTrustAccessGroupIncludeGsuiteArgs{Email: pulumi.String(v.Gsuite.Email), IdentityProviderId: pulumi.String(v.Gsuite.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Okta:
			e.Okta = &cloudflare.ZeroTrustAccessGroupIncludeOktaArgs{Name: pulumi.String(v.Okta.Name), IdentityProviderId: pulumi.String(v.Okta.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Saml:
			e.Saml = &cloudflare.ZeroTrustAccessGroupIncludeSamlArgs{AttributeName: pulumi.String(v.Saml.AttributeName), AttributeValue: pulumi.String(v.Saml.AttributeValue), IdentityProviderId: pulumi.String(v.Saml.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Oidc:
			e.Oidc = &cloudflare.ZeroTrustAccessGroupIncludeOidcArgs{ClaimName: pulumi.String(v.Oidc.ClaimName), ClaimValue: pulumi.String(v.Oidc.ClaimValue), IdentityProviderId: pulumi.String(v.Oidc.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_AuthContext:
			e.AuthContext = &cloudflare.ZeroTrustAccessGroupIncludeAuthContextArgs{Id: pulumi.String(v.AuthContext.Id), AcId: pulumi.String(v.AuthContext.AcId), IdentityProviderId: pulumi.String(v.AuthContext.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_AuthMethod:
			e.AuthMethod = &cloudflare.ZeroTrustAccessGroupIncludeAuthMethodArgs{AuthMethod: pulumi.String(v.AuthMethod.AuthMethod)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_CommonName:
			e.CommonName = &cloudflare.ZeroTrustAccessGroupIncludeCommonNameArgs{CommonName: pulumi.String(v.CommonName.CommonName)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Geo:
			e.Geo = &cloudflare.ZeroTrustAccessGroupIncludeGeoArgs{CountryCode: pulumi.String(v.Geo.CountryCode)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_DevicePosture:
			e.DevicePosture = &cloudflare.ZeroTrustAccessGroupIncludeDevicePostureArgs{IntegrationUid: pulumi.String(v.DevicePosture.IntegrationUid.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_ExternalEvaluation:
			e.ExternalEvaluation = &cloudflare.ZeroTrustAccessGroupIncludeExternalEvaluationArgs{EvaluateUrl: pulumi.String(v.ExternalEvaluation.EvaluateUrl), KeysUrl: pulumi.String(v.ExternalEvaluation.KeysUrl)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_LoginMethod:
			e.LoginMethod = &cloudflare.ZeroTrustAccessGroupIncludeLoginMethodArgs{Id: pulumi.String(v.LoginMethod.Id.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_ServiceToken:
			e.ServiceToken = &cloudflare.ZeroTrustAccessGroupIncludeServiceTokenArgs{TokenId: pulumi.String(v.ServiceToken.TokenId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_AnyValidServiceToken:
			e.AnyValidServiceToken = &cloudflare.ZeroTrustAccessGroupIncludeAnyValidServiceTokenArgs{}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_LinkedAppToken:
			e.LinkedAppToken = &cloudflare.ZeroTrustAccessGroupIncludeLinkedAppTokenArgs{AppUid: pulumi.String(v.LinkedAppToken.AppUid.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_UserRiskScore:
			e.UserRiskScore = &cloudflare.ZeroTrustAccessGroupIncludeUserRiskScoreArgs{UserRiskScores: riskLevels(v.UserRiskScore.UserRiskScore)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_CloudflareAccountMember:
			ctx.Log.Warn("cloudflare_account_member access rule is not supported by the Pulumi Cloudflare SDK (v6.17.0); skipping this rule. Use the Terraform engine to provision it. See the Pulumi module README.", nil)
			continue
		default:
			continue
		}
		out = append(out, e)
	}
	return out
}

// groupExcludes maps the access rules onto the provider's exclude array.
func groupExcludes(ctx *pulumi.Context, rules []*cloudflarezerotrustaccessgroupv1.CloudflareAccessRule) cloudflare.ZeroTrustAccessGroupExcludeArray {
	out := cloudflare.ZeroTrustAccessGroupExcludeArray{}
	for _, r := range rules {
		e := &cloudflare.ZeroTrustAccessGroupExcludeArgs{}
		switch v := r.Rule.(type) {
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Email:
			e.Email = &cloudflare.ZeroTrustAccessGroupExcludeEmailArgs{Email: pulumi.String(v.Email.Email)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_EmailDomain:
			e.EmailDomain = &cloudflare.ZeroTrustAccessGroupExcludeEmailDomainArgs{Domain: pulumi.String(v.EmailDomain.Domain)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_EmailList:
			e.EmailList = &cloudflare.ZeroTrustAccessGroupExcludeEmailListArgs{Id: pulumi.String(v.EmailList.Id.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Everyone:
			e.Everyone = &cloudflare.ZeroTrustAccessGroupExcludeEveryoneArgs{}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Ip:
			e.Ip = &cloudflare.ZeroTrustAccessGroupExcludeIpArgs{Ip: pulumi.String(v.Ip.Ip)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_IpList:
			e.IpList = &cloudflare.ZeroTrustAccessGroupExcludeIpListArgs{Id: pulumi.String(v.IpList.Id.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Certificate:
			e.Certificate = &cloudflare.ZeroTrustAccessGroupExcludeCertificateArgs{}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Group:
			e.Group = &cloudflare.ZeroTrustAccessGroupExcludeGroupArgs{Id: pulumi.String(v.Group.Id.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_AzureAd:
			e.AzureAd = &cloudflare.ZeroTrustAccessGroupExcludeAzureAdArgs{Id: pulumi.String(v.AzureAd.Id), IdentityProviderId: pulumi.String(v.AzureAd.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_GithubOrganization:
			ga := &cloudflare.ZeroTrustAccessGroupExcludeGithubOrganizationArgs{IdentityProviderId: pulumi.String(v.GithubOrganization.IdentityProviderId.GetValue()), Name: pulumi.String(v.GithubOrganization.Name)}
			if v.GithubOrganization.Team != "" {
				ga.Team = pulumi.StringPtr(v.GithubOrganization.Team)
			}
			e.GithubOrganization = ga
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Gsuite:
			e.Gsuite = &cloudflare.ZeroTrustAccessGroupExcludeGsuiteArgs{Email: pulumi.String(v.Gsuite.Email), IdentityProviderId: pulumi.String(v.Gsuite.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Okta:
			e.Okta = &cloudflare.ZeroTrustAccessGroupExcludeOktaArgs{Name: pulumi.String(v.Okta.Name), IdentityProviderId: pulumi.String(v.Okta.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Saml:
			e.Saml = &cloudflare.ZeroTrustAccessGroupExcludeSamlArgs{AttributeName: pulumi.String(v.Saml.AttributeName), AttributeValue: pulumi.String(v.Saml.AttributeValue), IdentityProviderId: pulumi.String(v.Saml.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Oidc:
			e.Oidc = &cloudflare.ZeroTrustAccessGroupExcludeOidcArgs{ClaimName: pulumi.String(v.Oidc.ClaimName), ClaimValue: pulumi.String(v.Oidc.ClaimValue), IdentityProviderId: pulumi.String(v.Oidc.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_AuthContext:
			e.AuthContext = &cloudflare.ZeroTrustAccessGroupExcludeAuthContextArgs{Id: pulumi.String(v.AuthContext.Id), AcId: pulumi.String(v.AuthContext.AcId), IdentityProviderId: pulumi.String(v.AuthContext.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_AuthMethod:
			e.AuthMethod = &cloudflare.ZeroTrustAccessGroupExcludeAuthMethodArgs{AuthMethod: pulumi.String(v.AuthMethod.AuthMethod)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_CommonName:
			e.CommonName = &cloudflare.ZeroTrustAccessGroupExcludeCommonNameArgs{CommonName: pulumi.String(v.CommonName.CommonName)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Geo:
			e.Geo = &cloudflare.ZeroTrustAccessGroupExcludeGeoArgs{CountryCode: pulumi.String(v.Geo.CountryCode)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_DevicePosture:
			e.DevicePosture = &cloudflare.ZeroTrustAccessGroupExcludeDevicePostureArgs{IntegrationUid: pulumi.String(v.DevicePosture.IntegrationUid.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_ExternalEvaluation:
			e.ExternalEvaluation = &cloudflare.ZeroTrustAccessGroupExcludeExternalEvaluationArgs{EvaluateUrl: pulumi.String(v.ExternalEvaluation.EvaluateUrl), KeysUrl: pulumi.String(v.ExternalEvaluation.KeysUrl)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_LoginMethod:
			e.LoginMethod = &cloudflare.ZeroTrustAccessGroupExcludeLoginMethodArgs{Id: pulumi.String(v.LoginMethod.Id.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_ServiceToken:
			e.ServiceToken = &cloudflare.ZeroTrustAccessGroupExcludeServiceTokenArgs{TokenId: pulumi.String(v.ServiceToken.TokenId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_AnyValidServiceToken:
			e.AnyValidServiceToken = &cloudflare.ZeroTrustAccessGroupExcludeAnyValidServiceTokenArgs{}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_LinkedAppToken:
			e.LinkedAppToken = &cloudflare.ZeroTrustAccessGroupExcludeLinkedAppTokenArgs{AppUid: pulumi.String(v.LinkedAppToken.AppUid.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_UserRiskScore:
			e.UserRiskScore = &cloudflare.ZeroTrustAccessGroupExcludeUserRiskScoreArgs{UserRiskScores: riskLevels(v.UserRiskScore.UserRiskScore)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_CloudflareAccountMember:
			ctx.Log.Warn("cloudflare_account_member access rule is not supported by the Pulumi Cloudflare SDK (v6.17.0); skipping this rule. Use the Terraform engine to provision it. See the Pulumi module README.", nil)
			continue
		default:
			continue
		}
		out = append(out, e)
	}
	return out
}

// groupRequires maps the access rules onto the provider's require array.
func groupRequires(ctx *pulumi.Context, rules []*cloudflarezerotrustaccessgroupv1.CloudflareAccessRule) cloudflare.ZeroTrustAccessGroupRequireArray {
	out := cloudflare.ZeroTrustAccessGroupRequireArray{}
	for _, r := range rules {
		e := &cloudflare.ZeroTrustAccessGroupRequireArgs{}
		switch v := r.Rule.(type) {
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Email:
			e.Email = &cloudflare.ZeroTrustAccessGroupRequireEmailArgs{Email: pulumi.String(v.Email.Email)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_EmailDomain:
			e.EmailDomain = &cloudflare.ZeroTrustAccessGroupRequireEmailDomainArgs{Domain: pulumi.String(v.EmailDomain.Domain)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_EmailList:
			e.EmailList = &cloudflare.ZeroTrustAccessGroupRequireEmailListArgs{Id: pulumi.String(v.EmailList.Id.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Everyone:
			e.Everyone = &cloudflare.ZeroTrustAccessGroupRequireEveryoneArgs{}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Ip:
			e.Ip = &cloudflare.ZeroTrustAccessGroupRequireIpArgs{Ip: pulumi.String(v.Ip.Ip)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_IpList:
			e.IpList = &cloudflare.ZeroTrustAccessGroupRequireIpListArgs{Id: pulumi.String(v.IpList.Id.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Certificate:
			e.Certificate = &cloudflare.ZeroTrustAccessGroupRequireCertificateArgs{}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Group:
			e.Group = &cloudflare.ZeroTrustAccessGroupRequireGroupArgs{Id: pulumi.String(v.Group.Id.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_AzureAd:
			e.AzureAd = &cloudflare.ZeroTrustAccessGroupRequireAzureAdArgs{Id: pulumi.String(v.AzureAd.Id), IdentityProviderId: pulumi.String(v.AzureAd.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_GithubOrganization:
			ga := &cloudflare.ZeroTrustAccessGroupRequireGithubOrganizationArgs{IdentityProviderId: pulumi.String(v.GithubOrganization.IdentityProviderId.GetValue()), Name: pulumi.String(v.GithubOrganization.Name)}
			if v.GithubOrganization.Team != "" {
				ga.Team = pulumi.StringPtr(v.GithubOrganization.Team)
			}
			e.GithubOrganization = ga
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Gsuite:
			e.Gsuite = &cloudflare.ZeroTrustAccessGroupRequireGsuiteArgs{Email: pulumi.String(v.Gsuite.Email), IdentityProviderId: pulumi.String(v.Gsuite.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Okta:
			e.Okta = &cloudflare.ZeroTrustAccessGroupRequireOktaArgs{Name: pulumi.String(v.Okta.Name), IdentityProviderId: pulumi.String(v.Okta.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Saml:
			e.Saml = &cloudflare.ZeroTrustAccessGroupRequireSamlArgs{AttributeName: pulumi.String(v.Saml.AttributeName), AttributeValue: pulumi.String(v.Saml.AttributeValue), IdentityProviderId: pulumi.String(v.Saml.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Oidc:
			e.Oidc = &cloudflare.ZeroTrustAccessGroupRequireOidcArgs{ClaimName: pulumi.String(v.Oidc.ClaimName), ClaimValue: pulumi.String(v.Oidc.ClaimValue), IdentityProviderId: pulumi.String(v.Oidc.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_AuthContext:
			e.AuthContext = &cloudflare.ZeroTrustAccessGroupRequireAuthContextArgs{Id: pulumi.String(v.AuthContext.Id), AcId: pulumi.String(v.AuthContext.AcId), IdentityProviderId: pulumi.String(v.AuthContext.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_AuthMethod:
			e.AuthMethod = &cloudflare.ZeroTrustAccessGroupRequireAuthMethodArgs{AuthMethod: pulumi.String(v.AuthMethod.AuthMethod)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_CommonName:
			e.CommonName = &cloudflare.ZeroTrustAccessGroupRequireCommonNameArgs{CommonName: pulumi.String(v.CommonName.CommonName)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_Geo:
			e.Geo = &cloudflare.ZeroTrustAccessGroupRequireGeoArgs{CountryCode: pulumi.String(v.Geo.CountryCode)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_DevicePosture:
			e.DevicePosture = &cloudflare.ZeroTrustAccessGroupRequireDevicePostureArgs{IntegrationUid: pulumi.String(v.DevicePosture.IntegrationUid.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_ExternalEvaluation:
			e.ExternalEvaluation = &cloudflare.ZeroTrustAccessGroupRequireExternalEvaluationArgs{EvaluateUrl: pulumi.String(v.ExternalEvaluation.EvaluateUrl), KeysUrl: pulumi.String(v.ExternalEvaluation.KeysUrl)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_LoginMethod:
			e.LoginMethod = &cloudflare.ZeroTrustAccessGroupRequireLoginMethodArgs{Id: pulumi.String(v.LoginMethod.Id.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_ServiceToken:
			e.ServiceToken = &cloudflare.ZeroTrustAccessGroupRequireServiceTokenArgs{TokenId: pulumi.String(v.ServiceToken.TokenId.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_AnyValidServiceToken:
			e.AnyValidServiceToken = &cloudflare.ZeroTrustAccessGroupRequireAnyValidServiceTokenArgs{}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_LinkedAppToken:
			e.LinkedAppToken = &cloudflare.ZeroTrustAccessGroupRequireLinkedAppTokenArgs{AppUid: pulumi.String(v.LinkedAppToken.AppUid.GetValue())}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_UserRiskScore:
			e.UserRiskScore = &cloudflare.ZeroTrustAccessGroupRequireUserRiskScoreArgs{UserRiskScores: riskLevels(v.UserRiskScore.UserRiskScore)}
		case *cloudflarezerotrustaccessgroupv1.CloudflareAccessRule_CloudflareAccountMember:
			ctx.Log.Warn("cloudflare_account_member access rule is not supported by the Pulumi Cloudflare SDK (v6.17.0); skipping this rule. Use the Terraform engine to provision it. See the Pulumi module README.", nil)
			continue
		default:
			continue
		}
		out = append(out, e)
	}
	return out
}
