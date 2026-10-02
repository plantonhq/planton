package verify

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// The behavioral-jwt scenario's tokens, test-only: both name the issuer and
// audience e2e-jwt.planton.dev and the subject e2e-jwt-sender under the key id
// e2e-jwt-sender. jwtTrustedToken is signed by the key whose public half the
// scenario holds inline; jwtStrangerToken carries identical claims signed by a
// key the scenario never saw, so only the signature tells them apart. Neither
// expires. Both keys were destroyed after signing.
const (
	jwtTrustedToken  = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImUyZS1qd3Qtc2VuZGVyIn0.eyJpc3MiOiJlMmUtand0LnBsYW50b24uZGV2Iiwic3ViIjoiZTJlLWp3dC1zZW5kZXIiLCJhdWQiOiJlMmUtand0LnBsYW50b24uZGV2IiwiaWF0IjoxNzkwMDAwMDAwfQ.YJgrPEgOW6AP2FAdiIOQKV3t5uDfpgoO76eqwJthnrhkLPQke-LjRZtPY7tzqfWdfyOjg2r7vyS1Aiwue8rV4RMhLHugXxXnP_OPttvJni975eCymE_NbASMWJml3hCoZy9ovZPEJxnL6WBGUTCixaPjg49inigy4flAOfrevB3GXjDsZ-lFfo69sGbKf8QNoYvciVOxCBCdfI6ekqpuVJ3NBJVpShaEeYZw_rJLrYxKx9JgC5jwgfNRaYgKuNnqfcAQnHVu1Dr5w6pQcOSjC8ilcUdMZCVP4sIF99xIJGzcKJIrDc2qxDPg5tTmNyIcaW1Rm7fMALsC57ZeKUrMfw"
	jwtStrangerToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImUyZS1qd3Qtc2VuZGVyIn0.eyJpc3MiOiJlMmUtand0LnBsYW50b24uZGV2Iiwic3ViIjoiZTJlLWp3dC1zZW5kZXIiLCJhdWQiOiJlMmUtand0LnBsYW50b24uZGV2IiwiaWF0IjoxNzkwMDAwMDAwfQ.MozEVR_Afqy4hmQUTzXyYbQEF-jzXYj1-Jv-wqugj-8tluHCv8oHvlnQO7alt-T1CrNmvI_Prn6ymBr5-6ONx0okY9Upg7QKPAZjx2iknh1kNJ2GPpPs1QJSnu7cGff1uBJtaDYlklnQv5AcimTt95-QsVgSAL6yD8fh4B3BQqSyiRZaPZO1vwRel2zoEfrxULdEcVxw47RO7eLNCXpLbIqHGxoAn7l-s4tCLVw16qOSOg0z0ATsor8QpkOk_sb0Bhte2ZleXDoEfy6OTgODput2lBmKUaanp4j1DmLqXINLF8oUl3kq_P-Ue4c6O5JlzFw9Edtuiqk94Q03Tv3rqg"
)

// JwtBehavioralVerifier proves RequestAuthentication ENFORCEMENT in a real
// mesh, paired with the ALLOW policy every token check needs: a stranger's
// token is refused as a bad token (401), a request with no token passes the
// check and is refused by the policy (403), and the trusted token reaches the
// backend (200). After the check is destroyed the stranger's token is no
// longer judged at all and falls to the policy's 403, which proves the 401
// came from the destroyed object.
type JwtBehavioralVerifier struct {
	Namespace string
	CheckName string
	// ClientDeployment is the meshed curl client the probes exec into.
	ClientDeployment string
	// BackendURL is the in-cluster URL of the protected Service.
	BackendURL string
}

func (v *JwtBehavioralVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] behavioral jwt: check %q must judge tokens for %q\n", v.CheckName, v.BackendURL)

	if err := KubectlResourceExists(ctx, kubeconfig,
		"requestauthentications.security.istio.io", v.CheckName, v.Namespace); err != nil {
		return err
	}
	for _, deploy := range []string{v.ClientDeployment, "e2e-jwt-backend"} {
		if err := kubectlWait(ctx, kubeconfig, "deployment", deploy, v.Namespace,
			"condition=Available", 4*time.Minute); err != nil {
			return errors.Wrapf(err, "meshed workload %q not available", deploy)
		}
	}
	// Configuration propagation to the sidecar takes seconds, so each outcome
	// is polled; the stranger's 401 first, because it is the one only this
	// object produces.
	if err := v.pollStatus(ctx, kubeconfig, jwtStrangerToken, "401", 3*time.Minute,
		"a stranger's token was never refused as a bad token"); err != nil {
		return err
	}
	if err := v.pollStatus(ctx, kubeconfig, "", "403", time.Minute,
		"a request with no token was not refused by the paired ALLOW policy"); err != nil {
		return err
	}
	return v.pollStatus(ctx, kubeconfig, jwtTrustedToken, "200", time.Minute,
		"the trusted token did not reach the backend")
}

func (v *JwtBehavioralVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	if err := KubectlResourceAbsent(ctx, kubeconfig,
		"requestauthentications.security.istio.io", v.CheckName, v.Namespace); err != nil {
		return err
	}
	// The release proof: with the check gone nothing validates tokens, so the
	// stranger's token is no longer a 401 and the ALLOW policy (a fixture that
	// outlives the component) refuses it as it refuses everyone.
	return v.pollStatus(ctx, kubeconfig, jwtStrangerToken, "403", 3*time.Minute,
		"a stranger's token was still refused as a bad token after the check was destroyed")
}

// pollStatus execs curl inside the meshed client, with the token as a bearer
// credential when one is given, until the backend returns the wanted status.
func (v *JwtBehavioralVerifier) pollStatus(ctx context.Context, kubeconfig, token, want string, timeout time.Duration, failMsg string) error {
	args := []string{"--kubeconfig", kubeconfig, "exec", "deploy/" + v.ClientDeployment, "-n", v.Namespace, "-c", "app", "--",
		"curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "--max-time", "5"}
	if token != "" {
		args = append(args, "-H", "Authorization: Bearer "+token)
	}
	args = append(args, v.BackendURL)
	deadline := time.Now().Add(timeout)
	var last string
	for time.Now().Before(deadline) {
		out, err := exec.CommandContext(ctx, "kubectl", args...).CombinedOutput()
		status := strings.TrimSpace(string(out))
		if err == nil && status == want {
			fmt.Printf("  [verify] in-mesh request returned HTTP %s — as required\n", status)
			return nil
		}
		last = fmt.Sprintf("status=%q err=%v", status, err)
		time.Sleep(5 * time.Second)
	}
	return errors.Errorf("%s (last probe: %s)", failMsg, last)
}
