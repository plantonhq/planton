package module

const (
	// OpSshKeyId is the exported stack output name that contains the
	// Hetzner Cloud numeric ID of the created SSH key.
	OpSshKeyId = "ssh_key_id"

	// OpFingerprint is the exported stack output name that contains the
	// MD5 fingerprint of the SSH public key.
	OpFingerprint = "fingerprint"
)
