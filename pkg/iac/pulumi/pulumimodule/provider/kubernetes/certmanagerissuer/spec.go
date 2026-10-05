// Package certmanagerissuer renders the shared CertManagerIssuerConfig proto
// into the cert-manager.io/v1 IssuerSpec JSON shape plus the credential
// Secrets the modules must materialize.
//
// ClusterIssuer and Issuer share an IDENTICAL spec upstream, and the Planton
// kinds share the CertManagerIssuerConfig proto — this package is the third
// leg of that guarantee: ONE builder renders the CR spec for both kinds, so
// their behavior can never drift. The Terraform modules render the same
// shape through their locals; keep field mappings in lockstep.
//
// CREDENTIAL MODEL: wherever the CRD expects a secretRef, the proto carries
// the credential VALUE (sensitive). The builder returns the Secrets to
// create (deterministic names, derived from the resource name) and wires the
// CR's secretRefs to them. Secret data keys are fixed per provider and
// documented on each mapping.
package certmanagerissuer

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	"sigs.k8s.io/yaml"
)

// CredentialSecret is one Kubernetes Secret a module must create in the
// issuer's secrets namespace before applying the CR.
type CredentialSecret struct {
	// Secret metadata.name (deterministic: "<resource-name>-<purpose>").
	Name string
	// stringData entries.
	Data map[string]string
}

// Result carries everything a module needs to apply the issuer.
type Result struct {
	// The cert-manager.io/v1 Issuer/ClusterIssuer spec, in CRD JSON shape.
	Spec map[string]interface{}
	// Credential Secrets to materialize before the CR.
	Secrets []CredentialSecret
	// Name of the ACME account private key Secret cert-manager will create
	// (empty for non-ACME backends). Exported as an output.
	AcmeAccountKeySecretName string
}

// BuildSpec renders the issuer config. resourceName is the CR's
// metadata.name — it prefixes every derived Secret name.
func BuildSpec(resourceName string, config *kubernetesprovider.CertManagerIssuerConfig) (*Result, error) {
	result := &Result{Spec: map[string]interface{}{}}

	switch {
	case config.GetAcme() != nil:
		if err := buildAcme(resourceName, config.GetAcme(), result); err != nil {
			return nil, err
		}
	case config.GetCa() != nil:
		result.Spec["ca"] = buildCa(config.GetCa())
	case config.GetSelfSigned() != nil:
		result.Spec["selfSigned"] = buildSelfSigned(config.GetSelfSigned())
	case config.GetVault() != nil:
		if err := buildVault(resourceName, config.GetVault(), result); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("issuer config must select a backend (acme, ca, self_signed, or vault)")
	}

	return result, nil
}

// ---------------------------------------------------------------- ACME ----

func buildAcme(resourceName string, acme *kubernetesprovider.CertManagerAcmeConfig, result *Result) error {
	result.AcmeAccountKeySecretName = resourceName + "-acme-account-key"

	acmeSpec := map[string]interface{}{
		"email":  acme.GetEmail(),
		"server": acme.GetServer(),
		"privateKeySecretRef": map[string]interface{}{
			"name": result.AcmeAccountKeySecretName,
		},
	}
	if acme.GetProfile() != "" {
		acmeSpec["profile"] = acme.GetProfile()
	}
	if acme.GetPreferredChain() != "" {
		acmeSpec["preferredChain"] = acme.GetPreferredChain()
	}
	if acme.GetCaBundle() != "" {
		acmeSpec["caBundle"] = acme.GetCaBundle()
	}
	if acme.GetSkipTlsVerify() {
		acmeSpec["skipTLSVerify"] = true
	}
	if acme.GetDisableAccountKeyGeneration() {
		acmeSpec["disableAccountKeyGeneration"] = true
	}
	if acme.GetEnableDurationFeature() {
		acmeSpec["enableDurationFeature"] = true
	}

	if eab := acme.GetExternalAccountBinding(); eab != nil {
		secretName := resourceName + "-acme-eab"
		result.Secrets = append(result.Secrets, CredentialSecret{
			Name: secretName,
			Data: map[string]string{"key": eab.GetHmacKey()},
		})
		acmeSpec["externalAccountBinding"] = map[string]interface{}{
			"keyID": eab.GetKeyId(),
			"keySecretRef": map[string]interface{}{
				"name": secretName,
				"key":  "key",
			},
		}
	}

	solvers := make([]interface{}, 0, len(acme.GetSolvers()))
	for i, solver := range acme.GetSolvers() {
		rendered, err := buildSolver(resourceName, i, solver, result)
		if err != nil {
			return errors.Wrapf(err, "solver %d", i)
		}
		solvers = append(solvers, rendered)
	}
	acmeSpec["solvers"] = solvers

	result.Spec["acme"] = acmeSpec
	return nil
}

func buildSolver(resourceName string, index int, solver *kubernetesprovider.CertManagerAcmeSolver, result *Result) (map[string]interface{}, error) {
	out := map[string]interface{}{}

	if sel := solver.GetSelector(); sel != nil {
		selector := map[string]interface{}{}
		if len(sel.GetDnsZones()) > 0 {
			selector["dnsZones"] = toInterfaceSlice(sel.GetDnsZones())
		}
		if len(sel.GetDnsNames()) > 0 {
			selector["dnsNames"] = toInterfaceSlice(sel.GetDnsNames())
		}
		if len(sel.GetMatchLabels()) > 0 {
			selector["matchLabels"] = toInterfaceMap(sel.GetMatchLabels())
		}
		if len(selector) > 0 {
			out["selector"] = selector
		}
	}

	if http01 := solver.GetHttp01(); http01 != nil {
		out["http01"] = buildHttp01(http01)
		return out, nil
	}

	dns01 := solver.GetDns01()
	if dns01 == nil {
		return nil, errors.New("solver must configure http01 or dns01")
	}
	renderedDns01, err := buildDns01(resourceName, index, dns01, result)
	if err != nil {
		return nil, err
	}
	out["dns01"] = renderedDns01
	return out, nil
}

func buildHttp01(http01 *kubernetesprovider.CertManagerAcmeHttp01Solver) map[string]interface{} {
	if ing := http01.GetIngress(); ing != nil {
		ingress := map[string]interface{}{}
		if ing.GetIngressClassName() != "" {
			ingress["ingressClassName"] = ing.GetIngressClassName()
		}
		if ing.GetName() != "" {
			ingress["name"] = ing.GetName()
		}
		if ing.GetServiceType() != "" {
			ingress["serviceType"] = ing.GetServiceType()
		}
		return map[string]interface{}{"ingress": ingress}
	}

	gw := http01.GetGatewayHttpRoute()
	parentRefs := make([]interface{}, 0, len(gw.GetParentRefs()))
	for _, ref := range gw.GetParentRefs() {
		parentRef := map[string]interface{}{"name": ref.GetName()}
		if ref.GetNamespace() != "" {
			parentRef["namespace"] = ref.GetNamespace()
		}
		if ref.GetSectionName() != "" {
			parentRef["sectionName"] = ref.GetSectionName()
		}
		parentRefs = append(parentRefs, parentRef)
	}
	route := map[string]interface{}{"parentRefs": parentRefs}
	if len(gw.GetLabels()) > 0 {
		route["labels"] = toInterfaceMap(gw.GetLabels())
	}
	if gw.GetServiceType() != "" {
		route["serviceType"] = gw.GetServiceType()
	}
	return map[string]interface{}{"gatewayHTTPRoute": route}
}

// buildDns01 renders one DNS-01 provider. CRD JSON quirks (verified against
// the pinned CRD schema): "cloudDNS", "azureDNS", "acmeDNS" capitalization;
// "clientSecretSecretRef"; Route53 nested auth.kubernetes.serviceAccountRef.
func buildDns01(resourceName string, index int, dns01 *kubernetesprovider.CertManagerAcmeDns01Solver, result *Result) (map[string]interface{}, error) {
	out := map[string]interface{}{}

	if dns01.GetCnameStrategy() != "" {
		// Proto vocabulary is lowercase; the CRD wants "None"/"Follow".
		if dns01.GetCnameStrategy() == "follow" {
			out["cnameStrategy"] = "Follow"
		} else {
			out["cnameStrategy"] = "None"
		}
	}

	secretPrefix := fmt.Sprintf("%s-solver%d", resourceName, index)

	switch {
	case dns01.GetCloudflare() != nil:
		cf := dns01.GetCloudflare()
		cloudflare := map[string]interface{}{}
		if token := cf.GetApiToken(); token != nil {
			secretName := secretPrefix + "-cloudflare"
			result.Secrets = append(result.Secrets, CredentialSecret{
				Name: secretName,
				Data: map[string]string{"api-token": token.GetToken()},
			})
			cloudflare["apiTokenSecretRef"] = secretKeyRef(secretName, "api-token")
		} else if key := cf.GetApiKey(); key != nil {
			secretName := secretPrefix + "-cloudflare"
			result.Secrets = append(result.Secrets, CredentialSecret{
				Name: secretName,
				Data: map[string]string{"api-key": key.GetKey()},
			})
			cloudflare["email"] = key.GetEmail()
			cloudflare["apiKeySecretRef"] = secretKeyRef(secretName, "api-key")
		}
		out["cloudflare"] = cloudflare

	case dns01.GetRoute53() != nil:
		r53 := dns01.GetRoute53()
		route53 := map[string]interface{}{"region": r53.GetRegion()}
		if r53.GetHostedZoneId() != "" {
			route53["hostedZoneID"] = r53.GetHostedZoneId()
		}
		if r53.GetAssumeRoleArn() != "" {
			route53["role"] = r53.GetAssumeRoleArn()
		}
		if static := r53.GetStaticCredentials(); static != nil {
			secretName := secretPrefix + "-route53"
			result.Secrets = append(result.Secrets, CredentialSecret{
				Name: secretName,
				Data: map[string]string{"secret-access-key": static.GetSecretAccessKey()},
			})
			route53["accessKeyID"] = static.GetAccessKeyId()
			route53["secretAccessKeySecretRef"] = secretKeyRef(secretName, "secret-access-key")
		}
		if sa := r53.GetServiceAccount(); sa != nil {
			serviceAccountRef := map[string]interface{}{"name": sa.GetServiceAccountName().GetValue()}
			if len(sa.GetAudiences()) > 0 {
				serviceAccountRef["audiences"] = toInterfaceSlice(sa.GetAudiences())
			}
			route53["auth"] = map[string]interface{}{
				"kubernetes": map[string]interface{}{
					"serviceAccountRef": serviceAccountRef,
				},
			}
		}
		out["route53"] = route53

	case dns01.GetAzureDns() != nil:
		az := dns01.GetAzureDns()
		azureDNS := map[string]interface{}{
			"subscriptionID":    az.GetSubscriptionId(),
			"resourceGroupName": az.GetResourceGroupName(),
		}
		if az.GetHostedZoneName() != "" {
			azureDNS["hostedZoneName"] = az.GetHostedZoneName()
		}
		if az.GetZoneType() != "" {
			// Proto vocabulary is lowercase; the CRD wants Azure*Zone.
			if az.GetZoneType() == "private" {
				azureDNS["zoneType"] = "AzurePrivateZone"
			} else {
				azureDNS["zoneType"] = "AzurePublicZone"
			}
		}
		if az.GetEnvironment() != "" {
			azureDNS["environment"] = az.GetEnvironment()
		}
		if az.GetClientSecret() != "" {
			secretName := secretPrefix + "-azure-dns"
			result.Secrets = append(result.Secrets, CredentialSecret{
				Name: secretName,
				Data: map[string]string{"client-secret": az.GetClientSecret()},
			})
			azureDNS["clientID"] = az.GetClientId()
			azureDNS["tenantID"] = az.GetTenantId()
			azureDNS["clientSecretSecretRef"] = secretKeyRef(secretName, "client-secret")
		}
		if mi := az.GetManagedIdentity(); mi != nil {
			managedIdentity := map[string]interface{}{}
			if mi.GetClientId() != "" {
				managedIdentity["clientID"] = mi.GetClientId()
			}
			if mi.GetResourceId() != "" {
				managedIdentity["resourceID"] = mi.GetResourceId()
			}
			azureDNS["managedIdentity"] = managedIdentity
		}
		out["azureDNS"] = azureDNS

	case dns01.GetGcpCloudDns() != nil:
		gcp := dns01.GetGcpCloudDns()
		cloudDNS := map[string]interface{}{"project": gcp.GetProjectId()}
		if gcp.GetHostedZoneName() != "" {
			cloudDNS["hostedZoneName"] = gcp.GetHostedZoneName()
		}
		if gcp.GetServiceAccountKeyJson() != "" {
			secretName := secretPrefix + "-clouddns"
			result.Secrets = append(result.Secrets, CredentialSecret{
				Name: secretName,
				Data: map[string]string{"key.json": gcp.GetServiceAccountKeyJson()},
			})
			cloudDNS["serviceAccountSecretRef"] = secretKeyRef(secretName, "key.json")
		}
		out["cloudDNS"] = cloudDNS

	case dns01.GetDigitalocean() != nil:
		secretName := secretPrefix + "-digitalocean"
		result.Secrets = append(result.Secrets, CredentialSecret{
			Name: secretName,
			Data: map[string]string{"access-token": dns01.GetDigitalocean().GetToken()},
		})
		out["digitalocean"] = map[string]interface{}{
			"tokenSecretRef": secretKeyRef(secretName, "access-token"),
		}

	case dns01.GetRfc2136() != nil:
		r := dns01.GetRfc2136()
		rfc2136 := map[string]interface{}{"nameserver": r.GetNameserver()}
		if r.GetTsigKeyName() != "" {
			secretName := secretPrefix + "-rfc2136"
			result.Secrets = append(result.Secrets, CredentialSecret{
				Name: secretName,
				Data: map[string]string{"tsig-secret": r.GetTsigSecret()},
			})
			rfc2136["tsigKeyName"] = r.GetTsigKeyName()
			rfc2136["tsigSecretSecretRef"] = secretKeyRef(secretName, "tsig-secret")
		}
		if r.GetTsigAlgorithm() != "" {
			rfc2136["tsigAlgorithm"] = r.GetTsigAlgorithm()
		}
		out["rfc2136"] = rfc2136

	case dns01.GetAcmeDns() != nil:
		ad := dns01.GetAcmeDns()
		secretName := secretPrefix + "-acme-dns"
		result.Secrets = append(result.Secrets, CredentialSecret{
			Name: secretName,
			Data: map[string]string{"acmedns.json": ad.GetAccountJson()},
		})
		out["acmeDNS"] = map[string]interface{}{
			"host":             ad.GetHost(),
			"accountSecretRef": secretKeyRef(secretName, "acmedns.json"),
		}

	case dns01.GetAkamai() != nil:
		ak := dns01.GetAkamai()
		secretName := secretPrefix + "-akamai"
		result.Secrets = append(result.Secrets, CredentialSecret{
			Name: secretName,
			Data: map[string]string{
				"client-token":  ak.GetClientToken(),
				"client-secret": ak.GetClientSecret(),
				"access-token":  ak.GetAccessToken(),
			},
		})
		out["akamai"] = map[string]interface{}{
			"serviceConsumerDomain": ak.GetServiceConsumerDomain(),
			"clientTokenSecretRef":  secretKeyRef(secretName, "client-token"),
			"clientSecretSecretRef": secretKeyRef(secretName, "client-secret"),
			"accessTokenSecretRef":  secretKeyRef(secretName, "access-token"),
		}

	case dns01.GetWebhook() != nil:
		wh := dns01.GetWebhook()
		webhook := map[string]interface{}{
			"groupName":  wh.GetGroupName(),
			"solverName": wh.GetSolverName(),
		}
		if wh.GetConfigYaml() != "" {
			var config interface{}
			if err := yaml.Unmarshal([]byte(wh.GetConfigYaml()), &config); err != nil {
				return nil, errors.Wrap(err, "failed to parse webhook config_yaml as YAML")
			}
			webhook["config"] = config
		}
		out["webhook"] = webhook

	default:
		return nil, errors.New("dns01 must configure a provider")
	}

	return out, nil
}

// ------------------------------------------------------------------ CA ----

func buildCa(ca *kubernetesprovider.CertManagerCaConfig) map[string]interface{} {
	out := map[string]interface{}{
		"secretName": ca.GetCaSecretName().GetValue(),
	}
	if len(ca.GetCrlDistributionPoints()) > 0 {
		out["crlDistributionPoints"] = toInterfaceSlice(ca.GetCrlDistributionPoints())
	}
	if len(ca.GetOcspServers()) > 0 {
		out["ocspServers"] = toInterfaceSlice(ca.GetOcspServers())
	}
	if len(ca.GetIssuingCertificateUrls()) > 0 {
		out["issuingCertificateURLs"] = toInterfaceSlice(ca.GetIssuingCertificateUrls())
	}
	return out
}

// ---------------------------------------------------------- self-signed ----

func buildSelfSigned(selfSigned *kubernetesprovider.CertManagerSelfSignedConfig) map[string]interface{} {
	out := map[string]interface{}{}
	if len(selfSigned.GetCrlDistributionPoints()) > 0 {
		out["crlDistributionPoints"] = toInterfaceSlice(selfSigned.GetCrlDistributionPoints())
	}
	return out
}

// --------------------------------------------------------------- Vault ----

func buildVault(resourceName string, vault *kubernetesprovider.CertManagerVaultConfig, result *Result) error {
	out := map[string]interface{}{
		"server": vault.GetServer(),
		"path":   vault.GetPath(),
	}
	if vault.GetVaultNamespace() != "" {
		out["namespace"] = vault.GetVaultNamespace()
	}
	if vault.GetCaBundle() != "" {
		out["caBundle"] = vault.GetCaBundle()
	}
	if vault.GetServerName() != "" {
		out["serverName"] = vault.GetServerName()
	}

	auth := map[string]interface{}{}
	switch {
	case vault.GetTokenAuth() != nil:
		secretName := resourceName + "-vault-token"
		result.Secrets = append(result.Secrets, CredentialSecret{
			Name: secretName,
			Data: map[string]string{"token": vault.GetTokenAuth().GetToken()},
		})
		auth["tokenSecretRef"] = secretKeyRef(secretName, "token")

	case vault.GetAppRoleAuth() != nil:
		ar := vault.GetAppRoleAuth()
		secretName := resourceName + "-vault-approle"
		result.Secrets = append(result.Secrets, CredentialSecret{
			Name: secretName,
			Data: map[string]string{"secret-id": ar.GetSecretId()},
		})
		auth["appRole"] = map[string]interface{}{
			"path":   ar.GetPath(),
			"roleId": ar.GetRoleId(),
			"secretRef": map[string]interface{}{
				"name": secretName,
				"key":  "secret-id",
			},
		}

	case vault.GetKubernetesAuth() != nil:
		ka := vault.GetKubernetesAuth()
		serviceAccountRef := map[string]interface{}{"name": ka.GetServiceAccountName().GetValue()}
		if len(ka.GetAudiences()) > 0 {
			serviceAccountRef["audiences"] = toInterfaceSlice(ka.GetAudiences())
		}
		auth["kubernetes"] = map[string]interface{}{
			"role":              ka.GetRole(),
			"mountPath":         "/v1/auth/" + ka.GetMountPath(),
			"serviceAccountRef": serviceAccountRef,
		}

	default:
		return errors.New("vault config must select an auth method (token_auth, app_role_auth, or kubernetes_auth)")
	}
	out["auth"] = auth

	result.Spec["vault"] = out
	return nil
}

// ------------------------------------------------------------- helpers ----

func secretKeyRef(name, key string) map[string]interface{} {
	return map[string]interface{}{"name": name, "key": key}
}

func toInterfaceSlice(in []string) []interface{} {
	out := make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

func toInterfaceMap(in map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
