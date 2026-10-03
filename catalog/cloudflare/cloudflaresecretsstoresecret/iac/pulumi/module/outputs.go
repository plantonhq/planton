package module

const (
	// OpSecretId is the exported output containing the secret's ID
	// within its store.
	OpSecretId = "secret_id"
	// OpStoreId is the exported output containing the ID of the store
	// holding the secret -- echoed for consumers that need the store/secret
	// pair.
	OpStoreId = "store_id"
)
