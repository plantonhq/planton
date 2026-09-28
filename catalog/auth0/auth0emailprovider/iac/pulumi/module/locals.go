package module

import (
	auth0emailproviderv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0emailprovider/v1alpha1"
)

// Locals holds the values the module computes from the stack input. It mirrors
// the Terraform module's locals.tf -- keep them in lockstep.
type Locals struct {
	// ResourceName is the resource's identity (the Pulumi resource name).
	ResourceName string

	DefaultFromAddress string

	// Enabled is the spec's enabled, true when unset.
	Enabled bool

	// Name is the service as Auth0 names it, from the arm the spec sets.
	Name string

	// Credentials are the credentials the arm carries; every other one is nil
	// and never sent. The custom arm leaves them all nil: the empty block the
	// provider requires.
	Credentials Credentials

	// Headers are the smtp arm's headers, nil unless at least one is set.
	Headers *SmtpHeaders

	// Message is the ses arm's configuration set or the mandrill arm's
	// view-content switch, nil unless set.
	Message *Message
}

// Credentials mirrors the provider's credentials block. A nil pointer is never
// sent.
type Credentials struct {
	SmtpHost *string
	SmtpPort *int
	SmtpUser *string
	SmtpPass *string

	AccessKeyId     *string
	SecretAccessKey *string

	ApiKey *string
	Region *string
	Domain *string

	AzureCsConnectionString *string

	Ms365TenantId     *string
	Ms365ClientId     *string
	Ms365ClientSecret *string
}

// SmtpHeaders mirrors the provider's settings.headers block.
type SmtpHeaders struct {
	XMcViewContentLink   *string
	XSesConfigurationSet *string
}

// Message mirrors the provider's settings.message block.
type Message struct {
	ConfigurationSetName *string
	ViewContentLink      *bool
}

func initializeLocals(stackInput *auth0emailproviderv1alpha1.Auth0EmailProviderStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	locals := &Locals{
		ResourceName:       target.Metadata.Name,
		DefaultFromAddress: spec.DefaultFromAddress,
		Enabled:            spec.Enabled == nil || spec.GetEnabled(),
	}
	applyService(locals, spec)
	return locals
}

// applyService maps the spec's service arm onto the provider's name,
// credentials and settings -- the rule the provider's own expander applies
// (internal/auth0/email/expand.go).
func applyService(locals *Locals, spec *auth0emailproviderv1alpha1.Auth0EmailProviderSpec) {
	switch {
	case spec.GetSmtp() != nil:
		smtp := spec.GetSmtp()
		port := int(smtp.Port)
		locals.Name = "smtp"
		locals.Credentials = Credentials{
			SmtpHost: &smtp.Host,
			SmtpPort: &port,
			SmtpUser: &smtp.User,
			SmtpPass: &smtp.Password,
		}
		if headers := smtp.GetHeaders(); headers != nil && (headers.XMcViewContentLink != "" || headers.XSesConfigurationSet != "") {
			locals.Headers = &SmtpHeaders{
				XMcViewContentLink:   optional(headers.XMcViewContentLink),
				XSesConfigurationSet: optional(headers.XSesConfigurationSet),
			}
		}
	case spec.GetSes() != nil:
		ses := spec.GetSes()
		locals.Name = "ses"
		locals.Credentials = Credentials{
			AccessKeyId:     &ses.AccessKeyId,
			SecretAccessKey: &ses.SecretAccessKey,
			Region:          &ses.Region,
		}
		if ses.ConfigurationSetName != "" {
			locals.Message = &Message{ConfigurationSetName: optional(ses.ConfigurationSetName)}
		}
	case spec.GetSendgrid() != nil:
		locals.Name = "sendgrid"
		locals.Credentials = Credentials{ApiKey: &spec.GetSendgrid().ApiKey}
	case spec.GetSparkpost() != nil:
		sparkpost := spec.GetSparkpost()
		locals.Name = "sparkpost"
		locals.Credentials = Credentials{
			ApiKey: &sparkpost.ApiKey,
			Region: optional(sparkpost.Region),
		}
	case spec.GetMailgun() != nil:
		mailgun := spec.GetMailgun()
		locals.Name = "mailgun"
		locals.Credentials = Credentials{
			ApiKey: &mailgun.ApiKey,
			Domain: &mailgun.Domain,
			Region: optional(mailgun.Region),
		}
	case spec.GetMandrill() != nil:
		mandrill := spec.GetMandrill()
		locals.Name = "mandrill"
		locals.Credentials = Credentials{ApiKey: &mandrill.ApiKey}
		if mandrill.ViewContentLink != nil {
			viewContentLink := mandrill.GetViewContentLink()
			locals.Message = &Message{ViewContentLink: &viewContentLink}
		}
	case spec.GetAzureCs() != nil:
		locals.Name = "azure_cs"
		locals.Credentials = Credentials{AzureCsConnectionString: &spec.GetAzureCs().ConnectionString}
	case spec.GetMs365() != nil:
		ms365 := spec.GetMs365()
		locals.Name = "ms365"
		locals.Credentials = Credentials{
			Ms365TenantId:     &ms365.TenantId,
			Ms365ClientId:     &ms365.ClientId,
			Ms365ClientSecret: &ms365.ClientSecret,
		}
	case spec.GetCustom() != nil:
		locals.Name = "custom"
	}
}

// optional maps the proto's "unset" (the empty string) to nil, so the setting
// is never sent.
func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
