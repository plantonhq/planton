package module

import (
	"reflect"
	"testing"

	auth0emailproviderv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0emailprovider/v1alpha1"
	"github.com/plantonhq/planton/shared"
)

func iacInput(spec *auth0emailproviderv1alpha1.Auth0EmailProviderSpec) *auth0emailproviderv1alpha1.Auth0EmailProviderIacInput {
	return &auth0emailproviderv1alpha1.Auth0EmailProviderIacInput{
		Target: &auth0emailproviderv1alpha1.Auth0EmailProvider{
			Metadata: &shared.CatalogObjectMetadata{Name: "email-provider"},
			Spec:     spec,
		},
	}
}

func str(v string) *string { return &v }
func num(v int) *int       { return &v }
func flag(v bool) *bool    { return &v }

func TestServiceArm(t *testing.T) {
	cases := []struct {
		name        string
		spec        *auth0emailproviderv1alpha1.Auth0EmailProviderSpec
		wantName    string
		wantCreds   Credentials
		wantHeaders *SmtpHeaders
		wantMessage *Message
	}{
		{
			name: "smtp sends the host, port, user and password, and no settings without headers",
			spec: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec{
				Service: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec_Smtp{Smtp: &auth0emailproviderv1alpha1.Auth0EmailProviderSmtp{
					Host: "smtp.resend.com", Port: 587, User: "resend", Password: "re_key",
				}},
			},
			wantName:  "smtp",
			wantCreds: Credentials{SmtpHost: str("smtp.resend.com"), SmtpPort: num(587), SmtpUser: str("resend"), SmtpPass: str("re_key")},
		},
		{
			name: "smtp headers are sent only as set",
			spec: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec{
				Service: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec_Smtp{Smtp: &auth0emailproviderv1alpha1.Auth0EmailProviderSmtp{
					Host: "email-smtp.eu-west-1.amazonaws.com", Port: 587, User: "AKIA", Password: "secret",
					Headers: &auth0emailproviderv1alpha1.Auth0EmailProviderSmtpHeaders{XSesConfigurationSet: "auth0"},
				}},
			},
			wantName:    "smtp",
			wantCreds:   Credentials{SmtpHost: str("email-smtp.eu-west-1.amazonaws.com"), SmtpPort: num(587), SmtpUser: str("AKIA"), SmtpPass: str("secret")},
			wantHeaders: &SmtpHeaders{XSesConfigurationSet: str("auth0")},
		},
		{
			name: "empty smtp headers send no settings",
			spec: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec{
				Service: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec_Smtp{Smtp: &auth0emailproviderv1alpha1.Auth0EmailProviderSmtp{
					Host: "smtp.example.com", Port: 465, User: "u", Password: "p",
					Headers: &auth0emailproviderv1alpha1.Auth0EmailProviderSmtpHeaders{},
				}},
			},
			wantName:  "smtp",
			wantCreds: Credentials{SmtpHost: str("smtp.example.com"), SmtpPort: num(465), SmtpUser: str("u"), SmtpPass: str("p")},
		},
		{
			name: "ses sends its key pair and region, and its configuration set as a message setting",
			spec: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec{
				Service: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec_Ses{Ses: &auth0emailproviderv1alpha1.Auth0EmailProviderSes{
					AccessKeyId: "AKIA", SecretAccessKey: "secret", Region: "eu-west-1", ConfigurationSetName: "auth0",
				}},
			},
			wantName:    "ses",
			wantCreds:   Credentials{AccessKeyId: str("AKIA"), SecretAccessKey: str("secret"), Region: str("eu-west-1")},
			wantMessage: &Message{ConfigurationSetName: str("auth0")},
		},
		{
			name: "sendgrid sends only its api key",
			spec: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec{
				Service: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec_Sendgrid{Sendgrid: &auth0emailproviderv1alpha1.Auth0EmailProviderSendgrid{ApiKey: "SG.key"}},
			},
			wantName:  "sendgrid",
			wantCreds: Credentials{ApiKey: str("SG.key")},
		},
		{
			name: "sparkpost in the US sends no region",
			spec: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec{
				Service: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec_Sparkpost{Sparkpost: &auth0emailproviderv1alpha1.Auth0EmailProviderSparkpost{ApiKey: "sp-key"}},
			},
			wantName:  "sparkpost",
			wantCreds: Credentials{ApiKey: str("sp-key")},
		},
		{
			name: "mailgun in the EU sends its key, domain and region",
			spec: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec{
				Service: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec_Mailgun{Mailgun: &auth0emailproviderv1alpha1.Auth0EmailProviderMailgun{
					ApiKey: "mg-key", Domain: "mg.example.com", Region: "eu",
				}},
			},
			wantName:  "mailgun",
			wantCreds: Credentials{ApiKey: str("mg-key"), Domain: str("mg.example.com"), Region: str("eu")},
		},
		{
			name: "mandrill sends its view-content switch, false included, as a message setting",
			spec: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec{
				Service: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec_Mandrill{Mandrill: &auth0emailproviderv1alpha1.Auth0EmailProviderMandrill{
					ApiKey: "md-key", ViewContentLink: flag(false),
				}},
			},
			wantName:    "mandrill",
			wantCreds:   Credentials{ApiKey: str("md-key")},
			wantMessage: &Message{ViewContentLink: flag(false)},
		},
		{
			name: "azure_cs sends its connection string",
			spec: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec{
				Service: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec_AzureCs{AzureCs: &auth0emailproviderv1alpha1.Auth0EmailProviderAzureCs{ConnectionString: "endpoint=https://acs;accesskey=k"}},
			},
			wantName:  "azure_cs",
			wantCreds: Credentials{AzureCsConnectionString: str("endpoint=https://acs;accesskey=k")},
		},
		{
			name: "ms365 sends its app registration",
			spec: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec{
				Service: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec_Ms365{Ms365: &auth0emailproviderv1alpha1.Auth0EmailProviderMs365{
					TenantId: "tenant", ClientId: "client", ClientSecret: "secret",
				}},
			},
			wantName:  "ms365",
			wantCreds: Credentials{Ms365TenantId: str("tenant"), Ms365ClientId: str("client"), Ms365ClientSecret: str("secret")},
		},
		{
			name: "custom sends the empty credentials block",
			spec: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec{
				Service: &auth0emailproviderv1alpha1.Auth0EmailProviderSpec_Custom{Custom: &auth0emailproviderv1alpha1.Auth0EmailProviderCustom{}},
			},
			wantName:  "custom",
			wantCreds: Credentials{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.spec.DefaultFromAddress = "Acme <no-reply@acme.com>"
			locals := initializeLocals(iacInput(tc.spec))
			if locals.Name != tc.wantName {
				t.Errorf("name: got %q, want %q", locals.Name, tc.wantName)
			}
			if !reflect.DeepEqual(locals.Credentials, tc.wantCreds) {
				t.Errorf("credentials: got %+v, want %+v", locals.Credentials, tc.wantCreds)
			}
			if !reflect.DeepEqual(locals.Headers, tc.wantHeaders) {
				t.Errorf("headers: got %+v, want %+v", locals.Headers, tc.wantHeaders)
			}
			if !reflect.DeepEqual(locals.Message, tc.wantMessage) {
				t.Errorf("message: got %+v, want %+v", locals.Message, tc.wantMessage)
			}
		})
	}
}

func TestEnabled(t *testing.T) {
	sendgrid := &auth0emailproviderv1alpha1.Auth0EmailProviderSpec_Sendgrid{Sendgrid: &auth0emailproviderv1alpha1.Auth0EmailProviderSendgrid{ApiKey: "SG.key"}}
	cases := []struct {
		name    string
		enabled *bool
		want    bool
	}{
		{name: "unset is on", enabled: nil, want: true},
		{name: "true is on", enabled: flag(true), want: true},
		{name: "false is off", enabled: flag(false), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			locals := initializeLocals(iacInput(&auth0emailproviderv1alpha1.Auth0EmailProviderSpec{
				DefaultFromAddress: "Acme <no-reply@acme.com>",
				Enabled:            tc.enabled,
				Service:            sendgrid,
			}))
			if locals.Enabled != tc.want {
				t.Errorf("got %v, want %v", locals.Enabled, tc.want)
			}
		})
	}
}
