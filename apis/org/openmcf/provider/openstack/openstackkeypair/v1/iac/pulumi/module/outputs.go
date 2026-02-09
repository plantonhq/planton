package module

const (
	// OpName is the exported stack output containing the keypair name.
	OpName = "name"
	// OpFingerprint is the exported stack output containing the MD5 fingerprint.
	OpFingerprint = "fingerprint"
	// OpPublicKey is the exported stack output containing the SSH public key.
	OpPublicKey = "public_key"
	// OpRegion is the exported stack output containing the OpenStack region.
	OpRegion = "region"
	// OpPrivateKey is the exported stack output containing the generated private key (secret).
	// Only populated when no public_key is provided in spec.
	OpPrivateKey = "private_key"
)
