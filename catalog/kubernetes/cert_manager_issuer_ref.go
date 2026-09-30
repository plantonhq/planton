package kubernetes

// CRDKind returns the cert-manager kind the reference names, as the
// cert-manager CRDs spell it ("Issuer" or "ClusterIssuer"), and its resolved
// name. Both engines' modules render an issuerRef from it; empty strings mean
// no issuer was chosen.
func (r *CertManagerIssuerRef) CRDKind() (kind, name string) {
	switch ref := r.GetIssuerType().(type) {
	case *CertManagerIssuerRef_ClusterIssuer:
		return "ClusterIssuer", ref.ClusterIssuer.GetName().GetValue()
	case *CertManagerIssuerRef_Issuer:
		return "Issuer", ref.Issuer.GetName().GetValue()
	}
	return "", ""
}
