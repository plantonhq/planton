package module

import (
	"github.com/pkg/errors"
	cloudflarezerotrustaccesspolicyv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarezerotrustaccesspolicy/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// policy provisions the Cloudflare Zero Trust Access policy and exports its ID.
func policy(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.ZeroTrustAccessPolicy, error) {
	spec := locals.CloudflareZeroTrustAccessPolicy.Spec

	args := &cloudflare.ZeroTrustAccessPolicyArgs{
		AccountId:                    pulumi.String(spec.AccountId),
		Name:                         pulumi.String(spec.Name),
		Decision:                     pulumi.String(spec.Decision.String()),
		Includes:                     policyIncludes(ctx, spec.Include),
		ApprovalRequired:             pulumi.BoolPtr(spec.ApprovalRequired),
		IsolationRequired:            pulumi.BoolPtr(spec.IsolationRequired),
		PurposeJustificationRequired: pulumi.BoolPtr(spec.PurposeJustificationRequired),
	}
	if len(spec.Exclude) > 0 {
		args.Excludes = policyExcludes(ctx, spec.Exclude)
	}
	if len(spec.Require) > 0 {
		args.Requires = policyRequires(ctx, spec.Require)
	}
	if sd := spec.GetSessionDuration(); sd != "" {
		args.SessionDuration = pulumi.StringPtr(sd)
	}
	if spec.PurposeJustificationPrompt != "" {
		args.PurposeJustificationPrompt = pulumi.StringPtr(spec.PurposeJustificationPrompt)
	}

	if len(spec.ApprovalGroups) > 0 {
		var groups cloudflare.ZeroTrustAccessPolicyApprovalGroupArray
		for _, g := range spec.ApprovalGroups {
			ga := &cloudflare.ZeroTrustAccessPolicyApprovalGroupArgs{
				ApprovalsNeeded: pulumi.Float64(float64(g.ApprovalsNeeded)),
			}
			if len(g.EmailAddresses) > 0 {
				ga.EmailAddresses = pulumi.ToStringArray(g.EmailAddresses)
			}
			if g.EmailListUuid != nil && g.EmailListUuid.GetValue() != "" {
				ga.EmailListUuid = pulumi.StringPtr(g.EmailListUuid.GetValue())
			}
			groups = append(groups, ga)
		}
		args.ApprovalGroups = groups
	}

	if cr := spec.ConnectionRules; cr != nil && cr.Rdp != nil {
		rdp := &cloudflare.ZeroTrustAccessPolicyConnectionRulesRdpArgs{}
		if len(cr.Rdp.AllowedClipboardLocalToRemoteFormats) > 0 {
			rdp.AllowedClipboardLocalToRemoteFormats = pulumi.ToStringArray(cr.Rdp.AllowedClipboardLocalToRemoteFormats)
		}
		if len(cr.Rdp.AllowedClipboardRemoteToLocalFormats) > 0 {
			rdp.AllowedClipboardRemoteToLocalFormats = pulumi.ToStringArray(cr.Rdp.AllowedClipboardRemoteToLocalFormats)
		}
		args.ConnectionRules = &cloudflare.ZeroTrustAccessPolicyConnectionRulesArgs{Rdp: rdp}
	}

	if mc := spec.MfaConfig; mc != nil {
		mfa := &cloudflare.ZeroTrustAccessPolicyMfaConfigArgs{
			MfaDisabled: pulumi.BoolPtr(mc.MfaDisabled),
		}
		if len(mc.AllowedAuthenticators) > 0 {
			var auths []string
			for _, a := range mc.AllowedAuthenticators {
				auths = append(auths, a.String())
			}
			mfa.AllowedAuthenticators = pulumi.ToStringArray(auths)
		}
		if mc.SessionDuration != "" {
			mfa.SessionDuration = pulumi.StringPtr(mc.SessionDuration)
		}
		args.MfaConfig = mfa
	}

	created, err := cloudflare.NewZeroTrustAccessPolicy(
		ctx,
		"policy",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare zero trust access policy")
	}

	ctx.Export(OpPolicyId, created.ID())

	return created, nil
}

// riskLevels converts the proto user-risk enum slice to the provider's strings.
func riskLevels(levels []cloudflarezerotrustaccesspolicyv1.AccessRuleUserRiskScore_Level) pulumi.StringArray {
	out := make(pulumi.StringArray, 0, len(levels))
	for _, l := range levels {
		out = append(out, pulumi.String(l.String()))
	}
	return out
}

const cfAccountMemberWarning = "cloudflare_account_member access rule is not supported by the Pulumi Cloudflare SDK (v6.17.0); skipping this rule. Use the Terraform engine to provision it. See the Pulumi module README."

// policyIncludes maps the access rules onto the provider's include array.
func policyIncludes(ctx *pulumi.Context, rules []*cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule) cloudflare.ZeroTrustAccessPolicyIncludeArray {
	out := cloudflare.ZeroTrustAccessPolicyIncludeArray{}
	for _, r := range rules {
		e := &cloudflare.ZeroTrustAccessPolicyIncludeArgs{}
		switch v := r.Rule.(type) {
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Email:
			e.Email = &cloudflare.ZeroTrustAccessPolicyIncludeEmailArgs{Email: pulumi.String(v.Email.Email)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_EmailDomain:
			e.EmailDomain = &cloudflare.ZeroTrustAccessPolicyIncludeEmailDomainArgs{Domain: pulumi.String(v.EmailDomain.Domain)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_EmailList:
			e.EmailList = &cloudflare.ZeroTrustAccessPolicyIncludeEmailListArgs{Id: pulumi.String(v.EmailList.Id.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Everyone:
			e.Everyone = &cloudflare.ZeroTrustAccessPolicyIncludeEveryoneArgs{}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Ip:
			e.Ip = &cloudflare.ZeroTrustAccessPolicyIncludeIpArgs{Ip: pulumi.String(v.Ip.Ip)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_IpList:
			e.IpList = &cloudflare.ZeroTrustAccessPolicyIncludeIpListArgs{Id: pulumi.String(v.IpList.Id.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Certificate:
			e.Certificate = &cloudflare.ZeroTrustAccessPolicyIncludeCertificateArgs{}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Group:
			e.Group = &cloudflare.ZeroTrustAccessPolicyIncludeGroupArgs{Id: pulumi.String(v.Group.Id.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_AzureAd:
			e.AzureAd = &cloudflare.ZeroTrustAccessPolicyIncludeAzureAdArgs{Id: pulumi.String(v.AzureAd.Id), IdentityProviderId: pulumi.String(v.AzureAd.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_GithubOrganization:
			ga := &cloudflare.ZeroTrustAccessPolicyIncludeGithubOrganizationArgs{IdentityProviderId: pulumi.String(v.GithubOrganization.IdentityProviderId.GetValue()), Name: pulumi.String(v.GithubOrganization.Name)}
			if v.GithubOrganization.Team != "" {
				ga.Team = pulumi.StringPtr(v.GithubOrganization.Team)
			}
			e.GithubOrganization = ga
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Gsuite:
			e.Gsuite = &cloudflare.ZeroTrustAccessPolicyIncludeGsuiteArgs{Email: pulumi.String(v.Gsuite.Email), IdentityProviderId: pulumi.String(v.Gsuite.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Okta:
			e.Okta = &cloudflare.ZeroTrustAccessPolicyIncludeOktaArgs{Name: pulumi.String(v.Okta.Name), IdentityProviderId: pulumi.String(v.Okta.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Saml:
			e.Saml = &cloudflare.ZeroTrustAccessPolicyIncludeSamlArgs{AttributeName: pulumi.String(v.Saml.AttributeName), AttributeValue: pulumi.String(v.Saml.AttributeValue), IdentityProviderId: pulumi.String(v.Saml.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Oidc:
			e.Oidc = &cloudflare.ZeroTrustAccessPolicyIncludeOidcArgs{ClaimName: pulumi.String(v.Oidc.ClaimName), ClaimValue: pulumi.String(v.Oidc.ClaimValue), IdentityProviderId: pulumi.String(v.Oidc.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_AuthContext:
			e.AuthContext = &cloudflare.ZeroTrustAccessPolicyIncludeAuthContextArgs{Id: pulumi.String(v.AuthContext.Id), AcId: pulumi.String(v.AuthContext.AcId), IdentityProviderId: pulumi.String(v.AuthContext.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_AuthMethod:
			e.AuthMethod = &cloudflare.ZeroTrustAccessPolicyIncludeAuthMethodArgs{AuthMethod: pulumi.String(v.AuthMethod.AuthMethod)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_CommonName:
			e.CommonName = &cloudflare.ZeroTrustAccessPolicyIncludeCommonNameArgs{CommonName: pulumi.String(v.CommonName.CommonName)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Geo:
			e.Geo = &cloudflare.ZeroTrustAccessPolicyIncludeGeoArgs{CountryCode: pulumi.String(v.Geo.CountryCode)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_DevicePosture:
			e.DevicePosture = &cloudflare.ZeroTrustAccessPolicyIncludeDevicePostureArgs{IntegrationUid: pulumi.String(v.DevicePosture.IntegrationUid.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_ExternalEvaluation:
			e.ExternalEvaluation = &cloudflare.ZeroTrustAccessPolicyIncludeExternalEvaluationArgs{EvaluateUrl: pulumi.String(v.ExternalEvaluation.EvaluateUrl), KeysUrl: pulumi.String(v.ExternalEvaluation.KeysUrl)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_LoginMethod:
			e.LoginMethod = &cloudflare.ZeroTrustAccessPolicyIncludeLoginMethodArgs{Id: pulumi.String(v.LoginMethod.Id.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_ServiceToken:
			e.ServiceToken = &cloudflare.ZeroTrustAccessPolicyIncludeServiceTokenArgs{TokenId: pulumi.String(v.ServiceToken.TokenId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_AnyValidServiceToken:
			e.AnyValidServiceToken = &cloudflare.ZeroTrustAccessPolicyIncludeAnyValidServiceTokenArgs{}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_LinkedAppToken:
			e.LinkedAppToken = &cloudflare.ZeroTrustAccessPolicyIncludeLinkedAppTokenArgs{AppUid: pulumi.String(v.LinkedAppToken.AppUid.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_UserRiskScore:
			e.UserRiskScore = &cloudflare.ZeroTrustAccessPolicyIncludeUserRiskScoreArgs{UserRiskScores: riskLevels(v.UserRiskScore.UserRiskScore)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_CloudflareAccountMember:
			ctx.Log.Warn(cfAccountMemberWarning, nil)
			continue
		default:
			continue
		}
		out = append(out, e)
	}
	return out
}

// policyExcludes maps the access rules onto the provider's exclude array.
func policyExcludes(ctx *pulumi.Context, rules []*cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule) cloudflare.ZeroTrustAccessPolicyExcludeArray {
	out := cloudflare.ZeroTrustAccessPolicyExcludeArray{}
	for _, r := range rules {
		e := &cloudflare.ZeroTrustAccessPolicyExcludeArgs{}
		switch v := r.Rule.(type) {
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Email:
			e.Email = &cloudflare.ZeroTrustAccessPolicyExcludeEmailArgs{Email: pulumi.String(v.Email.Email)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_EmailDomain:
			e.EmailDomain = &cloudflare.ZeroTrustAccessPolicyExcludeEmailDomainArgs{Domain: pulumi.String(v.EmailDomain.Domain)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_EmailList:
			e.EmailList = &cloudflare.ZeroTrustAccessPolicyExcludeEmailListArgs{Id: pulumi.String(v.EmailList.Id.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Everyone:
			e.Everyone = &cloudflare.ZeroTrustAccessPolicyExcludeEveryoneArgs{}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Ip:
			e.Ip = &cloudflare.ZeroTrustAccessPolicyExcludeIpArgs{Ip: pulumi.String(v.Ip.Ip)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_IpList:
			e.IpList = &cloudflare.ZeroTrustAccessPolicyExcludeIpListArgs{Id: pulumi.String(v.IpList.Id.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Certificate:
			e.Certificate = &cloudflare.ZeroTrustAccessPolicyExcludeCertificateArgs{}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Group:
			e.Group = &cloudflare.ZeroTrustAccessPolicyExcludeGroupArgs{Id: pulumi.String(v.Group.Id.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_AzureAd:
			e.AzureAd = &cloudflare.ZeroTrustAccessPolicyExcludeAzureAdArgs{Id: pulumi.String(v.AzureAd.Id), IdentityProviderId: pulumi.String(v.AzureAd.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_GithubOrganization:
			ga := &cloudflare.ZeroTrustAccessPolicyExcludeGithubOrganizationArgs{IdentityProviderId: pulumi.String(v.GithubOrganization.IdentityProviderId.GetValue()), Name: pulumi.String(v.GithubOrganization.Name)}
			if v.GithubOrganization.Team != "" {
				ga.Team = pulumi.StringPtr(v.GithubOrganization.Team)
			}
			e.GithubOrganization = ga
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Gsuite:
			e.Gsuite = &cloudflare.ZeroTrustAccessPolicyExcludeGsuiteArgs{Email: pulumi.String(v.Gsuite.Email), IdentityProviderId: pulumi.String(v.Gsuite.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Okta:
			e.Okta = &cloudflare.ZeroTrustAccessPolicyExcludeOktaArgs{Name: pulumi.String(v.Okta.Name), IdentityProviderId: pulumi.String(v.Okta.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Saml:
			e.Saml = &cloudflare.ZeroTrustAccessPolicyExcludeSamlArgs{AttributeName: pulumi.String(v.Saml.AttributeName), AttributeValue: pulumi.String(v.Saml.AttributeValue), IdentityProviderId: pulumi.String(v.Saml.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Oidc:
			e.Oidc = &cloudflare.ZeroTrustAccessPolicyExcludeOidcArgs{ClaimName: pulumi.String(v.Oidc.ClaimName), ClaimValue: pulumi.String(v.Oidc.ClaimValue), IdentityProviderId: pulumi.String(v.Oidc.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_AuthContext:
			e.AuthContext = &cloudflare.ZeroTrustAccessPolicyExcludeAuthContextArgs{Id: pulumi.String(v.AuthContext.Id), AcId: pulumi.String(v.AuthContext.AcId), IdentityProviderId: pulumi.String(v.AuthContext.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_AuthMethod:
			e.AuthMethod = &cloudflare.ZeroTrustAccessPolicyExcludeAuthMethodArgs{AuthMethod: pulumi.String(v.AuthMethod.AuthMethod)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_CommonName:
			e.CommonName = &cloudflare.ZeroTrustAccessPolicyExcludeCommonNameArgs{CommonName: pulumi.String(v.CommonName.CommonName)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Geo:
			e.Geo = &cloudflare.ZeroTrustAccessPolicyExcludeGeoArgs{CountryCode: pulumi.String(v.Geo.CountryCode)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_DevicePosture:
			e.DevicePosture = &cloudflare.ZeroTrustAccessPolicyExcludeDevicePostureArgs{IntegrationUid: pulumi.String(v.DevicePosture.IntegrationUid.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_ExternalEvaluation:
			e.ExternalEvaluation = &cloudflare.ZeroTrustAccessPolicyExcludeExternalEvaluationArgs{EvaluateUrl: pulumi.String(v.ExternalEvaluation.EvaluateUrl), KeysUrl: pulumi.String(v.ExternalEvaluation.KeysUrl)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_LoginMethod:
			e.LoginMethod = &cloudflare.ZeroTrustAccessPolicyExcludeLoginMethodArgs{Id: pulumi.String(v.LoginMethod.Id.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_ServiceToken:
			e.ServiceToken = &cloudflare.ZeroTrustAccessPolicyExcludeServiceTokenArgs{TokenId: pulumi.String(v.ServiceToken.TokenId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_AnyValidServiceToken:
			e.AnyValidServiceToken = &cloudflare.ZeroTrustAccessPolicyExcludeAnyValidServiceTokenArgs{}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_LinkedAppToken:
			e.LinkedAppToken = &cloudflare.ZeroTrustAccessPolicyExcludeLinkedAppTokenArgs{AppUid: pulumi.String(v.LinkedAppToken.AppUid.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_UserRiskScore:
			e.UserRiskScore = &cloudflare.ZeroTrustAccessPolicyExcludeUserRiskScoreArgs{UserRiskScores: riskLevels(v.UserRiskScore.UserRiskScore)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_CloudflareAccountMember:
			ctx.Log.Warn(cfAccountMemberWarning, nil)
			continue
		default:
			continue
		}
		out = append(out, e)
	}
	return out
}

// policyRequires maps the access rules onto the provider's require array.
func policyRequires(ctx *pulumi.Context, rules []*cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule) cloudflare.ZeroTrustAccessPolicyRequireArray {
	out := cloudflare.ZeroTrustAccessPolicyRequireArray{}
	for _, r := range rules {
		e := &cloudflare.ZeroTrustAccessPolicyRequireArgs{}
		switch v := r.Rule.(type) {
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Email:
			e.Email = &cloudflare.ZeroTrustAccessPolicyRequireEmailArgs{Email: pulumi.String(v.Email.Email)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_EmailDomain:
			e.EmailDomain = &cloudflare.ZeroTrustAccessPolicyRequireEmailDomainArgs{Domain: pulumi.String(v.EmailDomain.Domain)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_EmailList:
			e.EmailList = &cloudflare.ZeroTrustAccessPolicyRequireEmailListArgs{Id: pulumi.String(v.EmailList.Id.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Everyone:
			e.Everyone = &cloudflare.ZeroTrustAccessPolicyRequireEveryoneArgs{}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Ip:
			e.Ip = &cloudflare.ZeroTrustAccessPolicyRequireIpArgs{Ip: pulumi.String(v.Ip.Ip)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_IpList:
			e.IpList = &cloudflare.ZeroTrustAccessPolicyRequireIpListArgs{Id: pulumi.String(v.IpList.Id.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Certificate:
			e.Certificate = &cloudflare.ZeroTrustAccessPolicyRequireCertificateArgs{}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Group:
			e.Group = &cloudflare.ZeroTrustAccessPolicyRequireGroupArgs{Id: pulumi.String(v.Group.Id.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_AzureAd:
			e.AzureAd = &cloudflare.ZeroTrustAccessPolicyRequireAzureAdArgs{Id: pulumi.String(v.AzureAd.Id), IdentityProviderId: pulumi.String(v.AzureAd.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_GithubOrganization:
			ga := &cloudflare.ZeroTrustAccessPolicyRequireGithubOrganizationArgs{IdentityProviderId: pulumi.String(v.GithubOrganization.IdentityProviderId.GetValue()), Name: pulumi.String(v.GithubOrganization.Name)}
			if v.GithubOrganization.Team != "" {
				ga.Team = pulumi.StringPtr(v.GithubOrganization.Team)
			}
			e.GithubOrganization = ga
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Gsuite:
			e.Gsuite = &cloudflare.ZeroTrustAccessPolicyRequireGsuiteArgs{Email: pulumi.String(v.Gsuite.Email), IdentityProviderId: pulumi.String(v.Gsuite.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Okta:
			e.Okta = &cloudflare.ZeroTrustAccessPolicyRequireOktaArgs{Name: pulumi.String(v.Okta.Name), IdentityProviderId: pulumi.String(v.Okta.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Saml:
			e.Saml = &cloudflare.ZeroTrustAccessPolicyRequireSamlArgs{AttributeName: pulumi.String(v.Saml.AttributeName), AttributeValue: pulumi.String(v.Saml.AttributeValue), IdentityProviderId: pulumi.String(v.Saml.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Oidc:
			e.Oidc = &cloudflare.ZeroTrustAccessPolicyRequireOidcArgs{ClaimName: pulumi.String(v.Oidc.ClaimName), ClaimValue: pulumi.String(v.Oidc.ClaimValue), IdentityProviderId: pulumi.String(v.Oidc.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_AuthContext:
			e.AuthContext = &cloudflare.ZeroTrustAccessPolicyRequireAuthContextArgs{Id: pulumi.String(v.AuthContext.Id), AcId: pulumi.String(v.AuthContext.AcId), IdentityProviderId: pulumi.String(v.AuthContext.IdentityProviderId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_AuthMethod:
			e.AuthMethod = &cloudflare.ZeroTrustAccessPolicyRequireAuthMethodArgs{AuthMethod: pulumi.String(v.AuthMethod.AuthMethod)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_CommonName:
			e.CommonName = &cloudflare.ZeroTrustAccessPolicyRequireCommonNameArgs{CommonName: pulumi.String(v.CommonName.CommonName)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_Geo:
			e.Geo = &cloudflare.ZeroTrustAccessPolicyRequireGeoArgs{CountryCode: pulumi.String(v.Geo.CountryCode)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_DevicePosture:
			e.DevicePosture = &cloudflare.ZeroTrustAccessPolicyRequireDevicePostureArgs{IntegrationUid: pulumi.String(v.DevicePosture.IntegrationUid.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_ExternalEvaluation:
			e.ExternalEvaluation = &cloudflare.ZeroTrustAccessPolicyRequireExternalEvaluationArgs{EvaluateUrl: pulumi.String(v.ExternalEvaluation.EvaluateUrl), KeysUrl: pulumi.String(v.ExternalEvaluation.KeysUrl)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_LoginMethod:
			e.LoginMethod = &cloudflare.ZeroTrustAccessPolicyRequireLoginMethodArgs{Id: pulumi.String(v.LoginMethod.Id.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_ServiceToken:
			e.ServiceToken = &cloudflare.ZeroTrustAccessPolicyRequireServiceTokenArgs{TokenId: pulumi.String(v.ServiceToken.TokenId.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_AnyValidServiceToken:
			e.AnyValidServiceToken = &cloudflare.ZeroTrustAccessPolicyRequireAnyValidServiceTokenArgs{}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_LinkedAppToken:
			e.LinkedAppToken = &cloudflare.ZeroTrustAccessPolicyRequireLinkedAppTokenArgs{AppUid: pulumi.String(v.LinkedAppToken.AppUid.GetValue())}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_UserRiskScore:
			e.UserRiskScore = &cloudflare.ZeroTrustAccessPolicyRequireUserRiskScoreArgs{UserRiskScores: riskLevels(v.UserRiskScore.UserRiskScore)}
		case *cloudflarezerotrustaccesspolicyv1.CloudflareAccessRule_CloudflareAccountMember:
			ctx.Log.Warn(cfAccountMemberWarning, nil)
			continue
		default:
			continue
		}
		out = append(out, e)
	}
	return out
}
